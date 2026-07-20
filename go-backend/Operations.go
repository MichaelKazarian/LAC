// Цей файл містить реєстрацію та реалізацію технологічних операцій.
//
// # Структура файлу
//
// Кожна нетривіальна операція організована у власний блок з чіткою структурою:
//
//	buildXxx()   — фабрика: повертає []Step, може містити локальні змінні-замикання.
//	stepXxx()    — конструктор одного Step: збирає Do/Wait/Cleanup в єдину структуру.
//	doXxx()      — виконує фізичну дію (миттєво, без очікувань).
//	waitXxx()    — очікує завершення фізичного процесу (може блокуватися).
//	cleanupXxx() — прибирає після кроку (аналог defer, не викликається при EmergencyStop).
//
// Щоб додати нову операцію:
//  1. Зареєструй її в RegisterOperations() через r.Add(id, displayName, buildXxx).
//  2. Реалізуй buildXxx() та відповідні step/do/wait/cleanup функції нижче.
//  3. Прості операції з одним кроком можна реєструвати inline через StepDoWait.

// Термінологія
// Поняття : "Вихідне положення (Home)" і "Положення на осі (Axis)".
// Дії: "Переміщення вперед (MoveAxis)", це перехід з вихідного в робоче положення (на вісь)
// і "переміщення назад (MoveHome)" — зворотня дія
// Вузли:
// Заштовхувач (Pusher) безпосередньо вводить деталь у цангу.
// Завантажувач (Loader) підносить деталь до осі або забирає її.
// Важливо не плутати Pusher та Loader

package main

import (
	"fmt"
	"time"
  "strings"
  "sync"
)

const (
	EncoderStartRangeMin = 81  // Початок вікна старт-позиції
	EncoderStartRangeMax = 106 // Кінець вікна старт-позиції
  VFDLeaveStartTimeout = 2000 * time.Millisecond
)

var (
  stepStartTime time.Time
  trayBgMutex sync.Mutex
	isTrayFillingBg bool // чи крутиться лоток у фоні
)

func buildOperationLightOn() []Step {
	return []Step{
    stepOperationLightOn(),
	}
}

func stepOperationLightOn() Step {
	return Step{
		Name: "Зелений ліхтар увімкнено",
		Do: func(c *Controller) {
			c.apply(func() {
				c.state.Device20Out[OutGreenLight] = 1
        c.state.Device20Out[OutRedLight] = 0
			})
		},
		Wait: waitAlwaysOK,
	}
}

func buildOperationLightOff() []Step {
	return []Step{
    stepOperationLightOff(),
	}
}

func stepOperationLightOff() Step {
	return Step{
		Name: "Ліхтарі вимкнено",
		Do: func(c *Controller) {
			c.apply(func() {
				c.state.Device20Out[OutGreenLight] = 0
        c.state.Device20Out[OutRedLight] = 0
			})
		},
		Wait: waitAlwaysOK,
	}
}

func buildMoveToSafePosition() []Step {
	return []Step{
    stepCheckStartPosition(),
    {
      Name: "Примусово повертаємо вивантажувач",
      Do: doUnloaderHome,
      Wait: waitTime(1000 * time.Millisecond),
    },
    {
      Name: "Примусово повертаємо інструмент",
      Before: nil, // Перевірити в якому діапазоні розпредвал, щоб не зламати сверло.
      Do:   doToolHome,
      Wait: waitTime(1000 * time.Millisecond),
    },
    {
      Name: "Примусово повертаємо заштовхувач",
      Do:   doPusherHome,
      Wait: waitTime(1000 * time.Millisecond),
    },
    {
      Name: "Примусово повертаємо лоток",
      Do:   doLoaderHome,
      Wait: waitTime(1000 * time.Millisecond),
    },
    {
      Name: "Примусово відкриваємо цангу",
      Do:   doColletOpen,
      Wait: waitTime(100 * time.Millisecond),
    },
	}
}

func buildMagShutter() []Step {
	return []Step{
    {
      Name: "Перемикання шторки магазину",
      Do:   doMagShutterToggle,
      Wait: waitAlwaysOK, 
    },
	}
}

// Внутрішня функція БЕЗ c.apply (для виклику всередині інших apply)
func updateMagShutterLogic(c *Controller) {
  magStep := Step{Name: "MAG", LogMode: LogOnce}

  // Пряме зчитування без RLock, бо м'ютекс уже захоплено зверху!
  if c.state.Device10In[Pin25] == 1 {
    c.state.Device20Out[OutMagShutterOpen] = 1 // Закриваємо
    return
  }

  isHome := c.state.Device10In[PinMagShutterHome] == 1
  if isHome {
    c.state.Device20Out[OutMagShutterOpen] = 0 // Відкриваємо
    AppLog{}.Info(magStep, "Action: Opening shutter (waiting for part)")
  } else {
    c.state.Device20Out[OutMagShutterOpen] = 1
  }
}

// атомарне зміна положення магазина
func doMagShutterToggle(c *Controller) {
  c.apply(func() {
    updateMagShutterLogic(c)
  })
}

func buildTrayMove() []Step {
  return []Step{
    {
      Name: "Рух лотка, такт 1",
      Do:   doTrayStepToggle,
      Wait: waitTime(500 * time.Millisecond),
    },
    {
      Name: "Рух лотка, такт 2",
      Do:   doTrayStepToggle,
      Wait: waitTime(500 * time.Millisecond),
    },
  }
}

// shouldStopFeeding аналізує стан контролера і вирішує, чи треба заглушити подачу заготовок.
// Викликається ВИКЛЮЧНО всередині c.apply (пряме зчитування стану).
func shouldStopFeeding(c *Controller) bool {
  if c.state.IsSafetyLocked || c.state.Device10In[PinPartInLoader] == 1 {
    return true
  }
  // Якщо автомат вимкнено, натиснуто STOP, PAUSE, чи одинарний цикл (ModeSingle)
  // відпрацював і вийшов з runPhase (ActiveOperation порожня)
  isAutoMode := c.state.Mode == ModeAutomatic || c.state.Mode == ModeSingle
  if !isAutoMode || c.state.IsPaused ||
    (c.state.Mode == ModeSingle && c.state.ActiveOperation == "") {
    return true
  }
  return false
}

func doTrayStepToggle(c *Controller) {
  c.apply(func() {
    trayStep := Step{Name: "TRAY_BG", LogMode: LogOnce}
    isHome := c.state.Device10In[PinTrayGateHome] == 1
    isOpen := c.state.Device10In[PinTrayGateOpen] == 1

    switch {
    case isHome && !isOpen: // відкриваємо лоток
      c.state.Device20Out[OutTrayGateOpen] = 0

    case !isHome && isOpen: // закриваємо лоток
      c.state.Device20Out[OutTrayGateOpen] = 1

    case !isHome && !isOpen:
      AppLog{}.Error(trayStep, "Попередження: зависання циліндру лотка між кінцевиками")
      c.state.Device20Out[OutTrayGateOpen] = 1 // повертаємо в Home

    case isHome && isOpen:
      AppLog{}.Error(trayStep, "ПОМИЛКА: Одночасна активація датчиків лотка Home та Open")
      c.state.Device20Out[OutTrayGateOpen] = 1
    }
  })
}

//неблокуючий крок для конвеєра
func stepStartBackgroundTrayFill() Step {
	return Step{
		Name: "Запуск фонової підготовки наступної заготовки",
		Do: func(c *Controller) {
			startBgTrayFilling(c)
		},
		Wait: waitAlwaysOK, // Не блокує конвеєр, одразу йдемо далі
    LogMode: LogOnce,
	}
}

// startBgTrayFilling запускає фонову операцію, якщо вона ще не активна
func startBgTrayFilling(c *Controller) {
	trayBgMutex.Lock()
	if isTrayFillingBg {
		trayBgMutex.Unlock()
		return
	}
	isTrayFillingBg = true
	trayBgMutex.Unlock()

	go runTrayFeedingLoop(c)
}

// runTrayFeedingLoop виконує фоновий цикл подачі лотка до появи заготовки або аварії.
func runTrayFeedingLoop(c *Controller) {
  this := Step{
    Name:    "Фонова подача лотка",
    LogMode: LogOnce,
  }
  for {
    var needStop bool
    // 1. Атомарно перевіряємо умови зупинки та оновлюємо шторку
    c.apply(func() {
      needStop = shouldStopFeeding(c)
      if !needStop {
        updateMagShutterLogic(c)
      }
    })
    // Якщо спрацював тригер зупинки (аварія, деталь на місці, пауза чи кінець одиночного циклу)
    if needStop {
      AppLog{}.Info(this, "[BG_TRAY] Зупинка фонової подачі")
      break
    }
    // 2. Якщо все ОК — робимо такт лотка
    doTrayStepToggle(c)
    time.Sleep(500 * time.Millisecond)
  }
  // 3. При виході з циклу гарантовано переводимо обидва циліндри в безпечний закритий стан
  c.apply(func() {
    c.state.Device20Out[OutMagShutterOpen] = 1
    c.state.Device20Out[OutTrayGateOpen] = 1
  })
  trayBgMutex.Lock()
  isTrayFillingBg = false
  trayBgMutex.Unlock()
}

// stepInitBackgroundTrayIfNeeded перевіряє наявність заготовки на старті циклу.
// Якщо завантажувач порожній і фонова рутина не активна, запускає подачу.
func stepInitBackgroundTrayIfNeeded() Step {
	return Step{
		Name: "Перевірка необхідності первинного завантаження",
		Do: func(c *Controller) {
			if !isPartInLoader(c) {
				startBgTrayFilling(c)
			}
		},
		Wait: waitAlwaysOK,
	}
}

// БЛОКУЮЧИЙ ЗАХИСТ: чекаємо, поки деталь буде в лотку
func stepWaitBackgroundTrayReady() Step {
	return Step{
		Name: "Очікування завершення фонової подачі заготовки",
		Do:   func(c *Controller) {}, // Нічого не робимо, просто чекаємо
		Wait: func(c *Controller) StepResult {
			trayBgMutex.Lock()
			running := isTrayFillingBg
			trayBgMutex.Unlock()

			partFound := isPartInLoader(c)

			if !running && partFound {
				return StepResult{Status: StepOK, Message: "Заготовка успішно підготовлена у фоні"}
			}

			if running && partFound {
				return StepResult{Status: StepOK, Message: "Заготовка на місці (рутина ще гаситься)"}
			}

			// Якщо рутина ще працює, а деталі немає — чекаємо на цьому кроці
			return StepResult{Status: StepRepeat, Message: "Очікування фонової подачі деталі..."}
		},
    LogMode: LogOnce,
	}
}

// isPartInLoader перевіряє наявність заготовки в завантажувачі.
// Використовує RLock для забезпечення потокобезпечного зчитування стану
// з регістра Device10In в конкурентному середовищі.
// Повертає true, якщо датчик фізично бачить заготовку (сигнал == 1).
func isPartInLoader(c *Controller) bool {
	c.state.mu.RLock()
	defer c.state.mu.RUnlock()
	return c.state.Device10In[PinPartInLoader] == 1
}

// isPartAtMagazineExit перевіряє, чи заготовка пройшла шторку магазина.
// Повертає true, якщо датчик на виході магазина (Pin 25) активний.
func isPartAtMagazineExit(c *Controller) bool {
	c.state.mu.RLock()
	defer c.state.mu.RUnlock()
	return c.state.Device10In[Pin25] == 1
}

func buildLoader() []Step {
	return []Step {
    stepLogAllInputs(),
    stepCheckStartPosition(),
    stepInitBackgroundTrayIfNeeded(),
    //stepCheckCylinddresHome,
    // БЛОКУЮЧИЙ ЗАХИСТ: Перевіряємо, чи фонова рутина з минулого циклу
    // вже дотягнула деталь. Якщо ні — тут трохи почекаємо.
    stepWaitBackgroundTrayReady(),
    stepToolToHome(),
    stepUnloaderToAxis(),
    stepColletOpen(),
    stepEjectorForward(),
    stepAirBlastPulse(),
    stepUnloaderHome(),
    stepLoaderToAxis(), // Завантажувач йде вперед.
    stepPusherToAxis(2),
    stepColletClose(),
    stepPusherHome(),
    stepLoaderHome(), // Завантажувач очікує заготовку
    //stepTrayAutoFill(), // подаємо наступну заготовку,
		stepStartBackgroundTrayFill(), // Запускаємо лоток у фоні
		// Ці кроки будуть виконуватися ПАРАЛЕЛЬНО з роботою лотка:
    stepToolToAxis(),
    stepVFDEnable(),
    stepVFDSpeed1(),
    stepCheckStopZeroDegree(),
  }
}

func stepLogAllInputs() Step {
  return Step{
    Name:    "Діагностика: зчитування стану входів",
    LogMode: LogNormal,
    Before: func(c *Controller) StepResult {
      fmt.Println("dadasd")
      AppLog{}.Info(Step{Name: "ДІАГНОСТИКА"}, "=== СТАН УСІХ ЦИФРОВИХ ВХОДІВ ВЕРСТАТА ===")
      for pin, description := range PinNamesIn {
        val := c.state.Device10In[pin]
        statusIcon := "OFF [0]"
        if val == 1 {
          statusIcon = "ON [1]"
        }
        if pin == Pin9 { // Особливий маркер для нормально замкнутої аварійної зупинки
          if val == 1 {
            statusIcon = "Аварійна кнопка [1] (ОК: Кнопка віджата)"
          } else {
            statusIcon = "Аварійна кнопка [0] (АВАРІЯ: Натиснуто/Обрив)"
          }
        }

        msg := fmt.Sprintf("Pin %2d | %-22s | %s", pin, statusIcon, description)
        AppLog{}.Info(Step{Name: ""}, msg)
      }
      AppLog{}.Info(Step{Name: "ДІАГНОСТИКА"}, "=========================================")
      return StepResult{Status: StepOK}
    },
    Do: nil,
    Wait: func(c *Controller) StepResult {
      return StepResult{Status: StepOK}
    },
  }
}

///
func stepCheckStartPosition() Step {
  stepStartTime = time.Now()
	return StepDoWait(
		"Перевірка та безпечне вирівнювання розпредвалу",
		func(c *Controller) {},
		func(c *Controller) StepResult {
			currentPos := int(c.state.EncoderValue / 2)

			// 1. УСПІХ: Вал на місці — глушимо мотори і йдемо далі
			if isEncoderAtStartPosition(currentPos) {
				c.apply(func() {
					c.state.Device20Out[OutVFDSpeed1] = 0
					c.state.Device20Out[OutVFDReverseBit] = 0
				})
				return StepResult{Status: StepOK}
			}

			// 2. АВАРІЯ: Недоліт
			if currentPos < 40 {
				c.emergencyStop("Розпредвал не дійшов до стартової позиції")
				return StepResult{Status: StepFail}
			}

			// 3. ПЕРЕЛІТ: Потребує реверсу
			if currentPos > 106 {
				// Якщо з моменту старту кроку пауза НЕ закінчилась —
				// гасимо швидкість вперед (якщо вона була) і просто чекаємо зупинки валу
				if !stepStartTime.IsZero() && time.Since(stepStartTime) < 1 * time.Second {
          fmt.Println("Поточний час старту кроку:", stepStartTime)
					c.apply(func() { c.state.Device20Out[OutVFDSpeed1] = 0 })
          time.Sleep(50 * time.Millisecond)
					return StepResult{Status: StepRepeat}
				}

				// 4. Пауза минула! Тепер безпечно вмикаємо реверс назад
				c.apply(func() {
          fmt.Println("BACK")
					c.state.Device20Out[OutVFDEnable] = 1
					c.state.Device20Out[OutVFDReverseBit] = 1
					c.state.Device20Out[OutVFDSpeed1] = 1
				})
			}
			return StepResult{Status: StepRepeat}
		},
	)
}

func stepCheckStopZeroDegree() Step {
	return StepDoWait(
		"Очікування завершення оберту валу",
		func(c *Controller) {},
		func(c *Controller) StepResult {
			currentPos := int(c.state.EncoderValue / 2)

			if isEncoderAtStartPosition(currentPos) {
				fmt.Printf("[LOADER] Цикл завершено успішно. Вал у точці: %d\n", currentPos)
				c.apply(func() {
          // c.state.Device20Out[OutVFsDEnable] = 0
					c.state.Device20Out[OutVFDSpeed1] = 0
				})
				return StepResult{Status: StepOK}
			}

			// Вал ще крутиться — чекаемнаступний такт опроса
			return StepResult{Status: StepRepeat}
		},
	)
}

func buildVFDToZero() []Step {
	return []Step{
    stepDrivePowerOn(),
    stepVFDEnable(),
		stepVFDToStartPosition(),
	}
}

func stepVFDToStartPosition() Step {
	return StepDoWait(
		"Очікування та зупинка розпредвалу з випередженням",
		func(c *Controller) {},
		func(c *Controller) StepResult {
			currentPos := int(c.state.EncoderValue / 2)

			if isEncoderAtStartPosition(currentPos) {
				return StepResult{Status: StepOK}
			}

			// ВИПЕРЕДЖЕННЯ: ловимо підліт до нуля, щоб вчасно зняти напругу
			if currentPos >= 75 && currentPos <= 100 {
				c.apply(func() {
					//c.state.Device20Out[OutVFDEnable] = 0
          c.state.Device20Out[OutVFDSpeed1] = 0
				})
				return StepResult{Status: StepOK}
			}

			c.apply(func() {
				c.state.Device20Out[OutVFDSpeed1] = 1
			})
			return StepResult{Status: StepRepeat}
		},
	)
}

func isEncoderAtStartPosition(val int) bool {
	return val >= EncoderStartRangeMin && val <= EncoderStartRangeMax
}

func stepToolToHome() Step {
  return Step {
    Name: "Відвід інструмента у вихідне (переміщення назад)",
      Before: func(c *Controller) StepResult {
        logPins(c, "[BEFORE]", PinToolHome, PinToolAxis)
        if c.state.Device10In[PinToolHome] == c.state.Device10In[PinToolAxis] {
          msg := "Помилка датчиків інструмента"
          c.emergencyStop(msg)
          return StepResult{
            Status:  StepFail,
            Message: msg,
          }
        }
        return StepResult{Status: StepOK}
      },
      Do: doToolHome,
      Wait: func(c *Controller) StepResult {
        res := waitCond(func(c *Controller) bool {
          // Очікуємо: вихідне (18) = 1, на осі (17) = 0
          return c.state.Device10In[PinToolHome] == 1 &&
            c.state.Device10In[PinToolAxis] == 0
        }, 2000*time.Millisecond)(c)
        logPins(c, "[AFTER] ", PinToolHome, PinToolAxis)
        return res
      },
    }
}

func doToolHome(c *Controller) {
  c.apply(func() {
    c.state.Device20Out[OutTool] = 1
  })
}

func stepUnloaderToAxis() Step {
	return Step{
		Name: "Вивантажувач на вісь (переміщення вперед)",
		Before: func(c *Controller) StepResult {
			logPins(c, "[BEFORE]", PinUnloaderHome, PinUnloaderAxis)
			// Очікуємо: вихідне (15) = 1, на осі (16) = 0
			if c.state.Device10In[PinUnloaderHome] != 1 || c.state.Device10In[PinUnloaderAxis] != 0 {
				return StepResult{
					Status:  StepFail,
					Message: "Вивантажувач не у вихідному положенні перед виходом на вісь",
				}
			}
			return StepResult{Status: StepOK}
		},
		Do: func(c *Controller) {
			c.apply(func() {
				c.state.Device20Out[OutUnloader] = 1
			})
		},
		Wait: func(c *Controller) StepResult {
			res := waitCond(func(c *Controller) bool {
				// Очікуємо: вихідне (15) = 0, на осі (16) = 1
				return c.state.Device10In[PinUnloaderHome] == 0 &&
					c.state.Device10In[PinUnloaderAxis] == 1
			}, 2000*time.Millisecond)(c)
			logPins(c, "[AFTER] ", PinUnloaderHome, PinUnloaderAxis)
			return res
		},
	}
}

func stepColletOpen() Step {
	return Step{
		Name: "Розтискання цанги", // Without sensor
		Do: doColletOpen,
		Wait: waitTime(1000 * time.Millisecond),
	}
}

func doColletOpen(c *Controller) {
  c.apply(func() {
    c.state.Device20Out[OutCollet] = 1
  })
}

func stepAirBlastPulse() Step {
	return Step{
		Name: "Продування шпінделя",
		Do: func(c *Controller) {
			c.apply(func() {
				c.state.Device20Out[OutAirBlast] = 1
			})
		},
		Wait: func(c *Controller) StepResult {
			waitTime(250 * time.Millisecond)(c)
			c.apply(func() {
				c.state.Device20Out[OutAirBlast] = 0
			})
			
			return StepResult{Status: StepOK}
		},
	}
}

func stepEjectorForward() Step {
	return Step{
		Name: "Виштовхувач вперед", // Without sensor
		Do: doEjectorForward,
		Wait: waitTime(500 * time.Millisecond), // Час фіксований, датчиків нема
	}
}

func doEjectorForward(c *Controller) {
  c.apply(func() {
    c.state.Device20Out[OutEjector] = 1
  })
}

func stepUnloaderHome() Step {
	return Step{
		Name: "Вивантажувач: повернення у вихідне (Home)",
		Before: func(c *Controller) StepResult {
			logPins(c, "[BEFORE]", PinUnloaderHome, PinUnloaderAxis)
			// Очікуємо: на осі (16) = 1, вихідне (15) = 0
			if c.state.Device10In[PinUnloaderHome] != 0 || c.state.Device10In[PinUnloaderAxis] != 1 {
				return StepResult{
					Status:  StepFail,
					Message: "Вивантажувач не в робочому положенні перед поверненням додому",
				}
			}
			return StepResult{Status: StepOK}
		},
		Do: doUnloaderHome,
		Wait: func(c *Controller) StepResult {
			res := waitCond(func(c *Controller) bool {
				// Очікуємо: вихідне (15) = 1, на осі (16) = 0
				return c.state.Device10In[PinUnloaderHome] == 1 &&
					c.state.Device10In[PinUnloaderAxis] == 0
			}, 2000*time.Millisecond)(c)

			logPins(c, "[AFTER] ", PinUnloaderHome, PinUnloaderAxis)
			return res
		},
	}
}

func doUnloaderHome(c *Controller) {
  c.apply(func() {
    c.state.Device20Out[OutUnloader] = 0
  })
}

func stepLoaderToAxis() Step {
	return Step{
		Name: "Завантажувач на вісь (вперед)",
		Before: func(c *Controller) StepResult {
			logPins(c, "[BEFORE]", PinLoaderHome, PinLoaderAxis)
			// Очікуємо: вихідне (20) = 1, на осі (21) = 0
			if c.state.Device10In[PinLoaderHome] != 1 || c.state.Device10In[PinLoaderAxis] != 0 {
				return StepResult{
					Status:  StepFail,
					Message: "Завантажувач не у вихідному положенні перед подачею",
				}
			}
			return StepResult{Status: StepOK}
		},
		Do: func(c *Controller) {
			c.apply(func() {
				c.state.Device20Out[OutLoader] = 1
			})
		},
		Wait: func(c *Controller) StepResult {
			res := waitCond(func(c *Controller) bool {
				// Очікуємо: 20:0, 21:1
				return c.state.Device10In[PinLoaderHome] == 0 &&
					c.state.Device10In[PinLoaderAxis] == 1
			}, 2000*time.Millisecond)(c)
			logPins(c, "[AFTER] ", PinLoaderHome, PinLoaderAxis)
			return res
		},
	}
}

func stepPusherToAxis(attempts int) Step {
	return Step{
		Name: "Заштовхувач в робоче (вперед)",
		Before: func(c *Controller) StepResult {
			logPins(c, "[BEFORE]", PinPusherHome, PinPusherAxis)
			// Очікуємо: вихідне (23) = 1, на осі (22) = 0
			if c.state.Device10In[PinPusherHome] != 1 || c.state.Device10In[PinPusherAxis] != 0 {
				return StepResult{
					Status:  StepFail,
					Message: "Заштовхувач не у вихідному положенні перед робочим ходом",
				}
			}
			return StepResult{Status: StepOK}
		},
		Do: func(c *Controller) {
			c.apply(func() {
        c.state.Device20Out[OutEjector] = 0
				c.state.Device20Out[OutPusher] = 1
			})
		},
		Wait: func(c *Controller) StepResult {
			for i := 1; i <= attempts; i++ {
				// Очікуємо спрацювання датчика "на осі"
				res := waitCond(func(c *Controller) bool {
					return c.state.Device10In[PinPusherHome] == 0 &&
						c.state.Device10In[PinPusherAxis] == 1
				}, 1500*time.Millisecond)(c)

				if res.Status == StepOK {
					if i > 1 {
						fmt.Printf("[CTRL] Успішно доштовхнуло з спроби %d\n", i)
					}
					return res
				}
        c.apply(func() { c.state.Device20Out[OutEjector] = 0 })
				if i < attempts { // Якщо не дотиснув і є ще спроби
					logPins(c, fmt.Sprintf("[RETRY] Спроба %d невдала, пересмикування...", i), PinPusherHome, PinPusherAxis)
					c.apply(func() {  c.state.Device20Out[OutPusher] = 0 })   // Відводимо назад
					waitTime(500 * time.Millisecond)(c)
					c.apply(func() { c.state.Device20Out[OutPusher] = 1 }) // Знову вперед
				} else {					// Спроби закінчилися — викликаємо Emergency Stop
					msg := fmt.Sprintf("Заштовхувач не зміг дослати деталь за %d спроб", attempts)
          c.apply(func() { c.state.Device20Out[OutPusher] = 0 })
          waitTime(500 * time.Millisecond)(c)
          c.apply(func() { c.state.Device20Out[OutEjector] = 1 })
          waitTime(500 * time.Millisecond)(c)
          c.apply(func() { c.state.Device20Out[OutEjector] = 0 })
          waitTime(500 * time.Millisecond)(c)
					c.emergencyStop(msg) 
					
					return StepResult{
						Status:  StepFail,
						Message: msg,
					}
				}
			}
			return StepResult{Status: StepFail}
		},
	}
}

func stepColletClose() Step {
	return Step{
		Name: "Затискання цанги",
		Do: func(c *Controller) {
			c.apply(func() {
				c.state.Device20Out[OutCollet] = 0
			})
		},
		Wait: waitTime(250 * time.Millisecond),
	}
}

func stepPusherHome() Step {
	return Step{
		Name: "Заштовхувач у вихідне (Home)",
		Before: func(c *Controller) StepResult {
			logPins(c, "[BEFORE]", PinPusherHome, PinPusherAxis)
			// Очікуємо: на осі (22) = 1, вихідне (23) = 0
			if c.state.Device10In[PinPusherHome] != 0 || c.state.Device10In[PinPusherAxis] != 1 {
				return StepResult{
					Status:  StepFail,
					Message: "Заштовхувач не в робочому положенні перед поверненням додому",
				}
			}
			return StepResult{Status: StepOK}
		},
		Do: doPusherHome,
		Wait: func(c *Controller) StepResult {
			res := waitCond(func(c *Controller) bool {
				// Очікуємо: 23:1, 22:0
				return c.state.Device10In[PinPusherHome] == 1 &&
					c.state.Device10In[PinPusherAxis] == 0
			}, 2000*time.Millisecond)(c)

			logPins(c, "[AFTER] ", PinPusherHome, PinPusherAxis)
			return res
		},
	}
}

func doPusherHome(c *Controller) {
			c.apply(func() {
				c.state.Device20Out[OutPusher] = 0
			})
		}

func stepLoaderHome() Step {
	return Step{
		Name: "Завантажувач у вихідне (Home)",
		Before: func(c *Controller) StepResult {
			logPins(c, "[BEFORE]", PinLoaderHome, PinLoaderAxis)
			// Очікуємо: вихідне (20) = 0, на осі (21) = 1
			if c.state.Device10In[PinLoaderHome] != 0 || c.state.Device10In[PinLoaderAxis] != 1 {
				return StepResult{
					Status:  StepFail,
					Message: "Завантажувач не в робочому положенні перед поверненням додому",
				}
			}
			return StepResult{Status: StepOK}
		},
		Do: doLoaderHome,
		Wait: func(c *Controller) StepResult {
			res := waitCond(func(c *Controller) bool {
				// Очікуємо: 20:1, 21:0
				return c.state.Device10In[PinLoaderHome] == 1 &&
					c.state.Device10In[PinLoaderAxis] == 0
			}, 2000*time.Millisecond)(c)

			logPins(c, "[AFTER] ", PinLoaderHome, PinLoaderAxis)
			return res
		},
	}
}

func doLoaderHome(c *Controller) {
  c.apply(func() {
    c.state.Device20Out[OutLoader] = 0
  })
}

func stepToolToAxis() Step {
	return Step {
		Name: "Відвід інструмента на вісь (вперед)",
		Before: func(c *Controller) StepResult {
			logPins(c, "[BEFORE]", PinToolHome, PinToolAxis)
			// Очікуємо: вихідне (18) = 1, на осі (17) = 0
			if c.state.Device10In[PinToolHome] != 1 || c.state.Device10In[PinToolAxis] != 0 {
				return StepResult{
					Status:  StepFail,
					Message: "Інструмент не у вихідному положенні перед подачею на вісь",
				}
			}
			return StepResult{Status: StepOK}
		},
		Do: func(c *Controller) {
			c.apply(func() {
				c.state.Device20Out[OutTool] = 0
			})
		},
		Wait: func(c *Controller) StepResult {
			res := waitCond(func(c *Controller) bool {
				// Очікуємо: вихідне (18) = 0, на осі (17) = 1
				return c.state.Device10In[PinToolHome] == 0 &&
					c.state.Device10In[PinToolAxis] == 1
			}, 2000*time.Millisecond)(c)
			logPins(c, "[AFTER] ", PinToolHome, PinToolAxis)
			return res
		},
	}
}
///
func buildDrivePowerOn() []Step {
  return []Step{
    stepDrivePowerOn(),
  }
}

func stepDrivePowerOn() Step {
  return StepDoWait(
    "Увімкнення живлення приводів",
    func(c *Controller) {
      c.apply(func() {
        c.state.Device20Out[OutDrivePower] = 1
      })
    },
    func(c *Controller) StepResult {
      time.Sleep(200 * time.Millisecond)

      // Якщо буде фідбек з контактора:
      // if c.state.Device20In[InDrivePowerOK] == 0 {
      //     return StepResult{Status: StepFail, Message: "Немає підтвердження живлення приводів"}
      // }

      return StepResult{Status: StepOK}
    },
  )
}

func buildDrivePowerOff() []Step {
	return []Step{
		stepDrivePowerOff(),
	}
}

// stepDrivePowerOff — крок, який можна перевикористовувати
func stepDrivePowerOff() Step {
	return StepDoWait(
		"Вимкнення живлення приводів",
		func(c *Controller) {
			c.apply(func() {
				c.state.Device20Out[OutDrivePower] = 0
			})
		},
		func(c *Controller) StepResult {
			time.Sleep(200 * time.Millisecond)
			return StepResult{Status: StepOK}
		},
	)
}

func buildSpindleOn() []Step {
  return []Step{
    stepDrivePowerOn(),
    stepSpindleMotorOn(),
  }
}

func stepSpindleMotorOn() Step {
	return Step{
		Name: "Увімкнення двигуна шпінделя",
		Do:   doSpindleMotorOn,
    Wait: waitTime(200 * time.Millisecond),
	}
}

func buildSpindleOff() []Step {
  return []Step{
    stepSpindleMotorOff(),
  }
}

func stepSpindleMotorOff() Step {
	return Step{
		Name: "Вимкнення двигуна шпінделя",
		Do:   doSpindleMotorOff,
    Wait: waitTime(200 * time.Millisecond),
	}
}

func doSpindleMotorOn(c *Controller) {
	c.apply(func() {
    c.state.Device20Out[OutSpindleMotor] = 1
  })
}

func doSpindleMotorOff(c *Controller) {
	c.apply(func() {
    c.state.Device20Out[OutDrivePower] = 0
    c.state.Device20Out[OutSpindleMotor] = 0
  })
}

// Крок активації ПЧВ
func stepVFDEnable() Step {
	return StepDoWait(
		"Активація ПЧВ (Enable)",
		func(c *Controller) {
			c.apply(func() { c.state.Device20Out[OutVFDEnable] = 1 })
		},
		func(c *Controller) StepResult {
			time.Sleep(500 * time.Millisecond)
			return StepResult{Status: StepOK}
		},
	)
}

// Крок вибору швидкості 1
func stepVFDSpeed1() Step {
	return StepDoWait(
		"ПЧВ: Швидкість 1",
		func(c *Controller) {
			c.apply(func() {
				c.state.Device20Out[OutVFDSpeed1] = 1
				c.state.Device20Out[OutVFDSpeed2] = 0
			})
		},
		// waitCond завершиться успішно (StepOK), як тільки вал вилетить за межі зони стопу
		waitCond(func(c *Controller) bool {
			currentPos := int(c.state.EncoderValue / 2)
      return currentPos > EncoderStartRangeMax
		}, VFDLeaveStartTimeout),
	)
}

// Крок вибору швидкості 2
func stepVFDSpeed2() Step {
	return StepDoWait(
		"ПЧВ: Швидкість 2",
		func(c *Controller) {
			c.apply(func() {
				c.state.Device20Out[OutVFDSpeed1] = 0
				c.state.Device20Out[OutVFDSpeed2] = 1
			})
		},
		waitAlwaysOK,
	)
}

func buildVFDSpeed1() []Step {
	return []Step{
		stepDrivePowerOn(), // 1. Силове живлення
		stepVFDEnable(),    // 2. Команда Enable
    stepVFDReverseOff(),  // 3. Перевірка, що ми їдемо вперед
		stepVFDSpeed1(),      // 4. Швидкість 1
	}
}

func buildVFDSpeed2() []Step {
	return []Step{
		stepDrivePowerOn(), // 1. Силове живлення
		stepVFDEnable(),    // 2. Команда Enable
		stepVFDReverseOff(),  // 3. Перевірка, що ми їдемо вперед
		stepVFDSpeed2(),      // 4. Швидкість 2
	}
}

func buildVFDReverse() []Step {
	return []Step{
		stepDrivePowerOn(),
    stepVFDSpeedsOff(),
		stepVFDEnable(),
		stepVFDSpeed1Reverse(),
	}
}

// Крок: Гарантоване вимкнення всіх пінів швидкості
func stepVFDSpeedsOff() Step {
	return StepDoWait(
		"Скидання швидкостей ПЧВ",
		func(c *Controller) {
			c.apply(func() { 
				c.state.Device20Out[OutVFDSpeed1] = 0 
				c.state.Device20Out[OutVFDSpeed2] = 0 
			})
		},
		func(c *Controller) StepResult {
			time.Sleep(100 * time.Millisecond)
			return StepResult{Status: StepOK}
		},
	)
}

// Крок: Швидкість 1 + Реверс
func stepVFDSpeed1Reverse() Step {
	return StepDoWait(
		"ПЧВ: Швидкість 1 (РЕВЕРС)",
		func(c *Controller) {
			c.apply(func() {
				c.state.Device20Out[OutVFDSpeed1] = 1
				c.state.Device20Out[OutVFDSpeed2] = 0
				c.state.Device20Out[OutVFDReverseBit] = 1 // Активація реверсу
			})
		},
		waitAlwaysOK,
	)
}

// Крок: Вимкнення реверсу (Forward mode)
func stepVFDReverseOff() Step {
	return StepDoWait(
		"Скидання реверсу",
		func(c *Controller) {
			c.apply(func() { 
				c.state.Device20Out[OutVFDReverseBit] = 0 
			})
		},
		func(c *Controller) StepResult {
			time.Sleep(100 * time.Millisecond)
			return StepResult{Status: StepOK}
		},
	)
}

// buildVFDStop — тепер повертає []Step, як і всі інші build-функції
func buildVFDStop() []Step {
    return []Step{
        stepVFDStop(),
    }
}

// Виносимо логіку в окремий step, щоб вона була атомарною
func stepVFDStop() Step {
  return StepDoWait(
    "Зупинка ПЧВ (Вимкнення Enable)",
    func(c *Controller) {
      c.apply(func() {
        c.state.Device20Out[OutVFDSpeed1] = 0
        c.state.Device20Out[OutVFDSpeed2] = 0
        c.state.Device20Out[OutVFDReverseBit] = 0
        //c.state.Device20Out[OutVFDEnable] = 0
      })
    },
    func(c *Controller) StepResult {
      time.Sleep(100 * time.Millisecond)
      return StepResult{Status: StepOK}
    },
  )
}

// =============================================================================
// op_safety_stop
// =============================================================================

func buildSafetyStop() []Step {
	return []Step{
		stepSafetyStop(),
	}
}

func stepSafetyStop() Step {
	return Step{
		Name:    "Стоп",
		Do:      doSafetyStop,
		Wait:    waitTime(1000 * time.Millisecond),
		Cleanup: cleanupSafetyStop,
	}
}

func doSafetyStop(c *Controller) {
	c.apply(func() { c.state.Device20Out[3] = 1 })
}

func cleanupSafetyStop(c *Controller) {
	c.apply(func() {
		for i := 0; i < 32; i++ { c.state.Device20Out[i] = 0 }
	})
}

// =============================================================================
// Wait-функції специфічні для операцій
// =============================================================================

// waitMotorOn — імітує очікування увімкнення мотору.
func waitMotorOn(c *Controller) StepResult {
  // 1. Тайм-аут: якщо мотор не розкрутився за 3 секунди — це помилка заліза
  timeout := time.After(3 * time.Second)
  ticker := time.NewTicker(20 * time.Millisecond)
  defer ticker.Stop()

  for {
    select {
    case <-timeout:
      return StepResult{Status: StepFail, Message: "Двигун не розкрутився (Timeout)"}
      
    case <-ticker.C:
      c.state.mu.RLock()
      // Читаємо реальний вхід з Device10In
      motorReady := c.state.Device10In[Pin12] == 1
      locked := c.state.IsSafetyLocked
      c.state.mu.RUnlock()

      if locked {
        return StepResult{Status: StepAbort, Message: "Зупинено через EmergencyStop"}
      }

      if motorReady {
        return StepResult{Status: StepOK} // Мотор готовий, йдемо до наступного кроку
      }
    }
  }
}

//  Mirror
func buildSyncMirror() []Step {
	return []Step{
		stepSyncMirror(),
	}
}

func stepSyncMirror() Step {
	return Step{
		Name:    "Синхронізація",
		Do:      doSyncMirror,
		Wait:    waitSyncMirror,
		Cleanup: cleanupSyncMirror,
	}
}

func doSyncMirror(c *Controller) {
	c.apply(func() {
		for i := 0; i < OutTestPin31; i++ {
			c.state.Device20Out[i] = c.state.Device10In[i]
		}
	})
}

func cleanupSyncMirror(c *Controller) {
	c.apply(func() {
		for i := 0; i < OutTestPin31; i++ { c.state.Device20Out[i] = 0 }
	})
}

// waitSyncMirror — протягом 5 секунд дзеркалює входи.
// Якщо сенсор потрапляє в аварійний діапазон — викликає EmergencyStop.
func waitSyncMirror(c *Controller) StepResult {
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		c.state.mu.RLock()
		val := c.state.EncoderValue / 2
		locked := c.state.IsSafetyLocked
		inputs := c.state.Device10In
		c.state.mu.RUnlock()

		if locked {
			return StepResult{Status: StepAbort, Message: "EmergencyStop already active"}
		}
		if val > 250 && val < 300 {
			c.Stop(fmt.Sprintf("Перевищено поріг сенсора: %d", val))
			return StepResult{Status: StepAbort, Message: "Sensor threshold exceeded"}
		}
		c.apply(func() {
			for i := 0; i < OutTestPin31; i++ {
				c.state.Device20Out[i] = inputs[i]
			}
		})
		time.Sleep(2 * time.Millisecond)
	}
	return StepResult{Status: StepOK}
}

// =============================================================================
// Do-функції для заглушок
// =============================================================================

func stepItWorks(c *Controller) {
	fmt.Println("✅ Це працює")
}

// logPins друкує стан вказаних вхідних пінів у зручному форматі.
// prefix — зазвичай "[BEFORE]" або "[ AFTER]"
func logPins(c *Controller, prefix string, pins ...int) {
    var reports []string
    
    // Блокуємо стан для читання, щоб отримати консистентний зріз
    c.state.mu.RLock()
    defer c.state.mu.RUnlock()

    for _, p := range pins {
        name := GetPinName(p, false) // false, бо нас цікавлять входи (sensors)
        val := c.state.Device10In[p]
        reports = append(reports, fmt.Sprintf("%d %s: %d", p, name, val))
    }

    // З'єднуємо всі звіти через роздільник
    fmt.Printf("%s %s\n", prefix, strings.Join(reports, " | "))
}

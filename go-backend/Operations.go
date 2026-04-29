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
)

// RegisterOperations реєструє всі технологічні операції.
func RegisterOperations(r *OperationRegistry) {
	r.Add("op_mag_shutter",  "Завантаження магазину",  buildMagShutter)
	r.Add("op_tray_move",  "Крок лотка",  buildTrayMove)
	r.Add("op_tray_move_auto", "Переміщення лотка",  buildTrayAutoFill)
	r.Add("op_loader",  "Підведення завантажувача",  buildLoader)
	r.Add("operation5",  "Операція 5",  func() []Step { return []Step{StepDoWait("DoSomething", stepItWorks, waitAlwaysOK)} })
	r.Add("operation6",  "Операція 6",  func() []Step { return []Step{StepDoWait("DoSomething", stepItWorks, waitAlwaysOK)} })
	r.Add("operation7",  "Операція 7",  func() []Step { return []Step{StepDoWait("DoSomething", stepItWorks, waitAlwaysOK)} })
	r.Add("operation8",  "Операція 8",  func() []Step { return []Step{StepDoWait("DoSomething", stepItWorks, waitAlwaysOK)} })
	r.Add("operation9",  "Операція 9",  func() []Step { return []Step{StepDoWait("DoSomething", stepItWorks, waitAlwaysOK)} })
	r.Add("sync_mirror",    "Дзеркалювання",   buildSyncMirror)
	r.Add("op_safety_stop", "Безпечна зупинка", buildSafetyStop)
	r.Add("operation10", "Операція 10", func() []Step { return []Step{StepDoWait("DoSomething", stepItWorks, waitAlwaysOK)} })
	r.Add("operation11", "Операція 11", func() []Step { return []Step{StepDoWait("DoSomething", stepItWorks, waitAlwaysOK)} })
	r.Add("operation12", "Операція 12", func() []Step { return []Step{StepDoWait("DoSomething", stepItWorks, waitAlwaysOK)} })
	r.Add("operation13", "Операція 13", func() []Step { return []Step{StepDoWait("DoSomething", stepItWorks, waitAlwaysOK)} })
	r.Add("operation14", "Операція 14", func() []Step { return []Step{StepDoWait("DoSomething", stepItWorks, waitAlwaysOK)} })
	r.Add("operation15", "Операція 15", func() []Step { return []Step{StepDoWait("DoSomething", stepItWorks, waitAlwaysOK)} })
	r.Add("operation16", "Операція 16", func() []Step { return []Step{StepDoWait("DoSomething", stepItWorks, waitAlwaysOK)} })
	r.Add("operation17", "Операція 17", func() []Step { return []Step{StepDoWait("DoSomething", stepItWorks, waitAlwaysOK)} })
	r.Add("operation18", "Операція 18", func() []Step { return []Step{StepDoWait("DoSomething", stepItWorks, waitAlwaysOK)} })
	r.Add("operation19", "Операція 19", func() []Step { return []Step{StepDoWait("DoSomething", stepItWorks, waitAlwaysOK)} })
	r.Add("operation20", "Операція 20", func() []Step { return []Step{StepDoWait("DoSomething", stepItWorks, waitAlwaysOK)} })
}

// =============================================================================
// sync_mirror
// =============================================================================

func buildMagShutter() []Step {
	return []Step{
    {
      Name: "Перемикання шторки магазину",
      Do:   doMagShutterToggle,
      Wait: waitAlwaysOK, 
    },
	}
}

func doMagShutterToggle(c *Controller) {
  c.apply(func() {
    // Вмикаємо двіжки
    // c.state.Device20Out[OutTestPin17] = 1
    // c.state.Device20Out[OutTestPin18] = 1
    // c.state.Device20Out[OutTestPin20] = 1

    // І вимикаємо
    // c.state.Device20Out[OutTestPin17] = 0
    // c.state.Device20Out[OutTestPin18] = 0
    // c.state.Device20Out[OutTestPin20] = 0
    fmt.Printf("30 - %b\n", c.state.Device10In[PinMagShutterHome])
    isHome := c.state.Device10In[PinMagShutterHome] == 1   
    fmt.Printf("[MAG] Shutter Home sensor: %v\n", isHome)
    if isHome {
      c.state.Device20Out[OutMagShutterOpen] = 0
      fmt.Println("[MAG] Action: Opening shutter")
    } else {
      c.state.Device20Out[OutMagShutterOpen] = 1
      fmt.Println("[MAG] Action: Closing shutter")
    }
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

func buildTrayAutoFill() []Step {
  return []Step{
    {
      Name: "Переміщення лотка доки заготовки не буде в завантажувачі",
      Do:   doTrayStepToggle,
      Wait: func(c *Controller) StepResult {
        time.Sleep(500 * time.Millisecond)

        c.state.mu.RLock()
        found := c.state.Device10In[PinPartInLoader] == 1
        c.state.mu.RUnlock()

        if found {
          return StepResult{Status: StepOK, Message: "Заготовка на місці"}
        }

        // Якщо не знайшли — повторюємо цей же крок (знову Do -> Wait)
        return StepResult{Status: StepRepeat, Message: "Деталі немає, наступний такт"}
      },
    },
  }
}

func doTrayStepToggle(c *Controller) {
  c.apply(func() {
    // якщо датчик бачить заготовку, припиняємо рух
    if c.state.Device10In[PinPartInLoader] == 1 {
      return
    }

    isHome := c.state.Device10In[PinTrayGateHome] == 1
    isOpen := c.state.Device10In[PinTrayGateOpen] == 1
    switch {
    case isHome && !isOpen: // відкриваємо
      c.state.Device20Out[OutTrayGateOpen] = 0

    case !isHome && isOpen: // закриваємо
      c.state.Device20Out[OutTrayGateOpen] = 1

    case !isHome && !isOpen:
      // ПРОБЛЕМА: Зависли посередині (немає повітря або циліндр застряг)
      // пробуємо повернути в Home (безпечний стан)
      c.state.Device20Out[OutTrayGateOpen] = 1
      // TODO: додати лог: "Попередження: втрата позиції лотка"

    case isHome && isOpen: // КРИТИЧНО: замикання або збій датчиків
      c.state.Device20Out[OutTrayGateOpen] = 1 // Вимикаємо
    }
  })
}

func buildLoader() []Step {
	return []Step{
    {
      Name: "Відвід інструмента у вихідне (переміщення назад)",
      Do:   func (c *Controller) {
        logPins(c, "[BEFORE]", PinToolAxis, PinToolHome)
        c.apply(func() {
          c.state.Device20Out[OutTool] = 1
        }) },
      Wait: func(c *Controller) StepResult {
        res := waitTime(2000 * time.Millisecond)(c)
        logPins(c, "[AFTER]", PinToolAxis, PinToolHome)
        return res
      },
    },
    {
      Name: "Вивантажувач на вісь (переміщення вперед)",
      Do: func(c *Controller) {
        logPins(c, "[BEFORE]", PinUnloaderHome, PinUnloaderAxis)
        c.apply(func() {
          c.state.Device20Out[OutUnloader] = 1
        })
      },
      Wait: func(c *Controller) StepResult {
        res := waitTime(2000 * time.Millisecond)(c)
        logPins(c, "[ AFTER]", PinUnloaderHome, PinUnloaderAxis)
        return res
      },
    },
    {
      Name: "Розтискання цанги", // Without sensor
      Do:   func (c *Controller) { c.apply(func() {
        c.state.Device20Out[OutCollet] = 1 
      }) },
      Wait: waitTime(1000 * time.Millisecond),
    },
    {
      Name: "Виштовхувач вперед", // Without sensor
      Do:   func (c *Controller) { c.apply(func() {
        c.state.Device20Out[OutEjector] = 1
      }) },
      Wait: waitTime(500 * time.Millisecond), // Час фіксований, датчиків нема
    },
    {
      Name: "Продування шпінделя вкл", // Without sensor
      Do:   func (c *Controller) { c.apply(func() {
        c.state.Device20Out[OutAirBlast] = 1
      }) },
      Wait: waitTime(250 * time.Millisecond),
    },
    {
      Name: "Продування шпінделя викл",
      Do:   func (c *Controller) { c.apply(func() {
        c.state.Device20Out[OutAirBlast] = 0
      }) },
      Wait: waitTime(250 * time.Millisecond),
    },
    {
      Name: "Вивантажувач: повернення у вихідне (назад)",
      Do: func(c *Controller) {
        logPins(c, "[BEFORE]", PinUnloaderHome, PinUnloaderAxis)
        c.apply(func() {
          c.state.Device20Out[OutUnloader] = 0
        })
      },
      Wait: func(c *Controller) StepResult {
        res := waitTime(2000 * time.Millisecond)(c)
        logPins(c, "[AFTER]", PinUnloaderHome, PinUnloaderAxis)
        return res
      },
    },
    {
      Name: "Завантажувач на вісь (вперед)",
      Do: func(c *Controller) {
        logPins(c, "[BEFORE]", PinLoaderHome, PinLoaderAxis)
        c.apply(func() {
          c.state.Device20Out[OutLoader] = 1 
        })
      },
      Wait: func(c *Controller) StepResult {
        res := waitTime(2000 * time.Millisecond)(c)
        logPins(c, "[AFTER]", PinLoaderHome, PinLoaderAxis)
        return res
      },
    },
    {
      Name: "Переміщення заштовхувача в робоче (вперед)",
      Do: func(c *Controller) {
        logPins(c, "[BEFORE]", PinPusherHome, PinPusherAxis)
        c.apply(func() {
          c.state.Device20Out[OutPusher] = 1
        })
      },
      Wait: func(c *Controller) StepResult {
        res := waitTime(2000 * time.Millisecond)(c)
        logPins(c, "[AFTER]", PinPusherHome, PinPusherAxis)
        return res
      },
    },
    {
      Name: "Затискання цанги",
      Do:   func (c *Controller) { c.apply(func() {
        c.state.Device20Out[OutCollet] = 0
      }) },
      Wait: waitTime(250 * time.Millisecond),
    },
    {
      Name: "Переміщення заштовхувача на вихідне (назад)",
      Do: func(c *Controller) {
        logPins(c, "[BEFORE]", PinPusherHome, PinPusherAxis)
        c.apply(func() {
          c.state.Device20Out[OutPusher] = 0
        })
      },
      Wait: func(c *Controller) StepResult {
        res := waitTime(500 * time.Millisecond)(c)
        logPins(c, "[AFTER]", PinPusherHome, PinPusherAxis)
        return res
      },
    },
    {
      Name: "Завантажувач у вихідне (назад)",
      Do: func(c *Controller) {
        logPins(c, "[BEFORE]", PinLoaderHome, PinLoaderAxis )
        c.apply(func() {
          c.state.Device20Out[OutLoader] = 0
        })
      },
      Wait: func(c *Controller) StepResult {
        res := waitTime(500 * time.Millisecond)(c)
        logPins(c, "[AFTER]", PinLoaderHome, PinLoaderAxis)
        return res
      },
    },
    {
      Name: "Відвід інструмента на вісь (вперед)",
      Do: func(c *Controller) {
        logPins(c, "[BEFORE]", PinToolHome, PinToolAxis)
        c.apply(func() {
          c.state.Device20Out[OutTool] = 0
        })
      },
      Wait: func(c *Controller) StepResult {
        res := waitTime(500 * time.Millisecond)(c)
        logPins(c, "[AFTER]", PinToolHome, PinToolAxis)
        return res
      },
    },
	}
}

func stepMotorOn() Step {
	return Step{
		Name: "Включення двигуна",
		Do:   doMotorOn,
		Wait: waitMotorOn,
	}
}

//  Mirror

func buildSyncMirror() []Step {
	return []Step{
		stepMotorOn(),
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

func doMotorOn(c *Controller) {
	c.apply(func() { c.state.Device20Out[OutTestPin31] = 1 })
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
		Wait:    waitStop2s,
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

// waitTime повертає функцію очікування для заданої тривалості.
// Використовується як: Wait: waitTime(2 * time.Second)
func waitTime(duration time.Duration) func(c *Controller) StepResult {
	return func(c *Controller) StepResult {
		deadline := time.Now().Add(duration)

		for time.Now().Before(deadline) {
			// 1. Перевірка Safety Lock (Emergency Stop)
			c.state.mu.RLock()
			locked := c.state.IsSafetyLocked
			c.state.mu.RUnlock()

			if locked {
				return StepResult{
					Status:  StepAbort,
					Message: "Очікування перервано: Safety Lock",
				}
			}

			// 2. Коротка пауза, щоб не перевантажувати CPU
			// 20-50мс достатньо для оперативності в промислових задачах
			time.Sleep(25 * time.Millisecond)
		}

		return StepResult{Status: StepOK}
	}
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
        reports = append(reports, fmt.Sprintf("%d (%s): %d", p, name, val))
    }

    // З'єднуємо всі звіти через роздільник
    fmt.Printf("%s %s\n", prefix, strings.Join(reports, " | "))
}

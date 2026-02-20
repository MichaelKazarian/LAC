package main

import (
	"fmt"
	"time"
)

const maxCommErrorStreak = 4

type OperationFunc func()

type OutputPowerControl interface {
    PowerOff() error
    PowerOn() error
}

type Controller struct {
	hw    HardwareService
	state *HardwareState
	opsMap      map[string]OperationInfo
  lastOutput [32]uint16
  firstRun   bool
  opQueue    chan string
  needsCounterReset bool
  commErrorStreak int
  commLost        bool
  lastCommError   string
  power PowerControl
  outputsLost bool // блокує запис на OUT-плату, поки її не відновлено у SafetyStart
}

func GetOperationsList(r *OperationRegistry) [][]string {
	list := make([][]string, 0, len(r.ops))
	for _, op := range r.ops {
		list = append(list, []string{op.ID, op.DisplayName})
	}
	return list
}

func NewController(hw HardwareService, state *HardwareState) *Controller {
	registry := NewOperationRegistry()
	RegisterOperations(registry)

	// Швидкий доступ по ID
	opMap := make(map[string]OperationInfo)
	for _, op := range registry.All() {
		opMap[op.ID] = op
	}

	state.OpsList = GetOperationsList(registry)
  power := &MockPowerControl{}

	return &Controller{
		hw:      hw,
		state:   state,
    power:   power,
		opsMap:  opMap,
		opQueue: make(chan string, 10),
	}
}

// Run — основний цикл контролера
func (c *Controller) Run() {
  c.firstRun = true // Ініціалізуємо перед циклом
  fmt.Println("[CTRL] Контролер логіки запущено")

  for {
    start := time.Now()
    sensor, inputs, err := c.hw.Read()
		if err == nil {
			c.updateEncoderState(sensor, true)
			c.updateInputsState(&inputs, true)
      c.resetCommError()        // успішний обмін = зв'язок відновився
		} else {
      c.state.mu.Lock()
      c.state.isEncoderOnline = false
      c.state.isInputsOnline = false
      c.state.mu.Unlock()
      c.trackCommError(err) //  рахуємо втрату зв'язку
		}

    c.syncHardware()
    c.updateCycleTime(time.Since(start).Milliseconds())
  }
}

func (c *Controller) logicWorker() {
  fmt.Println("[CTRL] Logic Worker запущено")
  
  for {
    // Перевіряємо режим та стан паузи
    c.state.mu.RLock()
    mode := c.state.Mode
    paused := c.state.IsPaused
    locked := c.state.IsSafetyLocked
    c.state.mu.RUnlock()

    // Якщо система заблокована (Emergency Stop) — чекаємо розблокування
    if locked {
      time.Sleep(10 * time.Millisecond)
      continue
    }

    // Пріоритетна черга ручних команд (Web/API)
    // Використовуємо select з default, щоб не блокуватися, якщо черга порожня
    select {
    case opID := <-c.opQueue:
      fmt.Printf("[CTRL] Виконання ручної команди: %s\n", opID)
      c.execSteps(opID)
    default:
      if (mode == ModeAutomatic || mode == ModeSingle) && !paused {
				c.runAutomaticCycleSteps()
				if mode == ModeSingle {
					c.SetMode(ModeManual)
					fmt.Println("[CTRL] Одиночний цикл завершено")
				}
			} else {
				time.Sleep(50 * time.Millisecond)
			}
    }
  }
}

// execSteps виконує послідовність Steps для операції з вказаним ID.
//
// ActiveOperation встановлюється на початку і гарантовано скидається
// через defer — незалежно від того, як завершилось виконання:
// штатно, через break або EmergencyStop.
//
// Порядок виконання кожного кроку:
//  1. Перевірка EmergencyStop перед Do      — міг прийти зовнішній сигнал
//  2. Do()                                  — фізична дія (миттєво)
//  3. Перевірка EmergencyStop перед Wait    — Do міг спровокувати зупинку
//  4. Wait()                                — очікування завершення процесу
//  5. Перевірка EmergencyStop перед Cleanup — Wait могла тривати довго
//  6. Cleanup()                             — аналог defer для Step,
//                                            не викликається при EmergencyStop
func (c *Controller) execSteps(opID string) {
  op, ok := c.opsMap[opID]
  if !ok {
    fmt.Printf("[CTRL]️ Операція %s не знайдена\n", opID)
    return
  }

  c.state.mu.Lock()
  c.state.ActiveOperation = opID
  c.state.mu.Unlock()

  defer func() {
    c.state.mu.Lock()
    c.state.ActiveOperation = ""
    c.state.mu.Unlock()
  }()

  steps := op.Build() 
  for _, step := range steps {
    if c.isEmergency() { break }
    if step.Do != nil { step.Do(c) }
    if c.isEmergency() { break }

    var result StepResult = StepResult{Status: StepOK}
    if step.Wait != nil { result = step.Wait(c) }

    // Cleanup — аналог defer, виконується завжди крім EmergencyStop
    if !c.isEmergency() && step.Cleanup != nil {
      step.Cleanup(c)
    }

    if result.Status == StepAbort {
      fmt.Printf("[CTRL] Step %s aborted: %s\n", step.Name, result.Message)
      break
    }
    if result.Status == StepFail {
      fmt.Printf("[CTRL]️ Step %s failed: %s\n", step.Name, result.Message)
      break
    }
  }
}

func (c *Controller) isEmergency() bool {
    c.state.mu.RLock()
    defer c.state.mu.RUnlock()
    return c.state.IsSafetyLocked
}

// --- Автоматичний цикл через Steps ---
func (c *Controller) runAutomaticCycleSteps() {
	c.state.mu.Lock()
	if c.needsCounterReset {
		c.state.Counter = 0
		c.needsCounterReset = false
	}
	c.state.mu.Unlock()

	for _, opEntry := range c.state.OpsList {
		c.waitIfPaused()

		c.state.mu.RLock()
		mode := c.state.Mode
		locked := c.state.IsSafetyLocked
		c.state.mu.RUnlock()

		if mode == ModeManual || locked {
			fmt.Printf("[CTRL] Цикл перервано перед кроком %s\n", opEntry[0])
			return
		}

		c.execSteps(opEntry[0])
	}

	c.state.mu.Lock()
	c.state.Counter++
	c.state.mu.Unlock()
}

// syncHardware перевіряє зміни в Device20Out та записує їх у залізо.
// Логіка роботи:
// Якщо система заблокована (EmergencyStop або операторська зупинка) — 
//   гарантуємо, що мотор  вимкнений перед записом (current[OutMainMotor]=0)
// Порівнюємо з останнім записаним станом, щоб уникнути зайвих записів.
// При помилці запису Modbus:
// - встановлюємо outputsLost = true, щоб не писати в плату виходів
// - скидання outputsLost тільки в SafetyStart()
// - фізично відключаємо живлення плати виходів через power.DisableOutputsPower()
// - викликаємо EmergencyStop, щоб очистити черги та гарантувати безпечний стан
// - Кешуємо новий стан лише при успішному записі.
//
// Сценарії:
// - Штатна зупинка (SAFE_LOCK):
//      Motor вимкнений, плати IN та OUT доступні, система чекає команди SafetyStart
// - Аварія IN-плати:
//      EmergencyStop через Modbus, мотор вимкнений, решта логіки заморожена
// - Аварія OUT-плати:
//      Латч outputsLost, живлення плати відключене, EmergencyStop викликається для гарантії безпечного стану
func (c *Controller) syncHardware() {
  c.state.mu.RLock()
  current := c.state.Device20Out
  locked := c.state.IsSafetyLocked
  outputsLost := c.outputsLost
  c.state.mu.RUnlock()

  if outputsLost {
    return
  }

  if locked { // не дає випадково включити мотор
    current[OutMainMotor] = 0
  }
  // Перевіряємо наявність змін
  if current == c.lastOutput && !c.firstRun {
    return
  }

  if err := c.hw.Write(current); err != nil {
    if !c.outputsLost { // Латчимо стан тільки ОДИН раз
      c.outputsLost = true
      fmt.Println("[CTRL] Outputs communication LOST → entering FAIL-SAFE")
      _ = c.power.DisableOutputsPower() // Фізично відрубуємо плату
      // Заморожуємо логіку стандартним шляхом
      c.emergencyStop("Втрачено зв'язок з платою виходів")
    }
    return
  }

  // Оновлюємо кеш тільки при успішному записі
  c.lastOutput = current
  c.firstRun = false
}

func (c *Controller) InvokeOperation(name string) error {
  if _, exists := c.opsMap[name]; !exists {
		return fmt.Errorf("[CTRL] Операція %s не знайдена", name)
	}
  select {
	case c.opQueue <- name:
		fmt.Printf("[CTRL] [%s] Команда додана в чергу: %s\n", time.Now().Format("15:04:05"), name)
	default:
		return fmt.Errorf("[CTRL] Черга переповнена, зачекайте")
	}
  return nil
}

// SetMode безпечно оновлює режим роботи контролера
func (c *Controller) SetMode(mode ControlMode) {
    c.state.mu.Lock()
    // Якщо поточний режим був ручний, а новий - автоматичний/одиночний
    if c.state.Mode == ModeManual && mode != ModeManual {
        c.needsCounterReset = true
    }
    c.state.IsPaused = false
    c.state.Mode = mode
    c.state.mu.Unlock()
    // Якщо ми переходимо в ручний режим, можна відразу скинути певні виходи, якщо треба
    if mode == ModeManual {
        fmt.Println("[CTRL] Контролер переведено в РУЧНИЙ режим")
    }
    fmt.Printf("[CTRL] Режим змінено на: %v\n", mode)
}

func (c *Controller) SetPause(paused bool) {
  c.state.mu.Lock()
  c.state.IsPaused = paused
  c.state.mu.Unlock()
  
  if paused {
    fmt.Println("[CTRL] ПАУЗА: Поточна операція допрацює, нові не почнуться.")
  } else {
    fmt.Println("[CTRL] ПРОДОВЖЕНО: Автоматичний цикл відновлено.")
  }
}

// Error start

func (c *Controller) trackCommError(err error) {
    c.commErrorStreak++
    c.lastCommError = err.Error()

    switch {
    case c.commErrorStreak == 1:
        // Логуємо лише першу помилку в серії
        fmt.Printf("[COMM] [%s] Помилка зв'язку: %v\n", time.Now().Format("15:04:05"), err)
    case c.commErrorStreak >= maxCommErrorStreak && !c.commLost:
        c.commLost = true
        c.emergencyStop(fmt.Sprintf("Втрата зв'язку: %d помилок підряд", c.commErrorStreak))
    }
}

func (c *Controller) resetCommError() {
  if c.commLost {
    c.state.mu.Lock()
    if c.state.IsSafetyLocked {
      c.state.StopReason = "Зв'язок відновлено. Натисніть РОЗБЛОКУВАТИ"
      fmt.Printf("[COMM] [%s] Зв'язок відновлено. Система заблокована\n",
        time.Now().Format("15:04:05"))
    }
    c.state.mu.Unlock()
  }

	c.commLost = false
	c.commErrorStreak = 0
	c.lastCommError = ""
}

// Error end

func (c *Controller) updateEncoderState(val uint16, online bool) {
	c.state.mu.Lock()
	defer c.state.mu.Unlock()
	c.state.EncoderValue = val
	c.state.isEncoderOnline = online
	c.state.LastUpdate = time.Now()
}

func (c *Controller) updateInputsState(data *[32]uint16, online bool) {
	c.state.mu.Lock()
	defer c.state.mu.Unlock()
	c.state.isInputsOnline = online
	if online && data != nil {
		c.state.Device10In = *data
	}
	c.state.LastUpdate = time.Now()
}

func (c *Controller) updateCycleTime(ms int64) {
	c.state.mu.Lock()
	defer c.state.mu.Unlock()
	c.state.ReadCycleMs = ms
}

func (c *Controller) waitIfPaused() {
    for {
        c.state.mu.RLock()
        paused := c.state.IsPaused
        c.state.mu.RUnlock()
        if !paused {
            break // Виходимо з циклу очікування, якщо паузу знято
        }
        time.Sleep(50 * time.Millisecond)
    }
}

func (c *Controller) apply(fn func()) {
  c.state.mu.Lock()
  if c.state.IsSafetyLocked { // Якщо заблоковано не виконуємо логіку функції
    c.state.mu.Unlock()
    return
  }
  fn()
  c.state.mu.Unlock()
}

// drainOpQueue очищає всі відкладені команди керування.
// Викликається при переході в SAFE_LOCK, щоб жодна запланована операція
// не виконалась після зупинки.
func (c *Controller) drainOpQueue() {
	for {
		select {
		case <-c.opQueue:
			// видаляємо всі накопичені команди
		default:
			return // черга порожня
		}
	}
}

// Stop виконує штатну керовану зупинку системи (операторська зупинка).
// Не є аварією і не змінює fail-safe стан обладнання.
//
// Логіка роботи:
// 1) Встановлює StopReason для відображення причини зупинки.
// 2) Вимикає головний мотор (Device20Out[OutMainMotor] = 0), щоб гарантувати стоп рухомого обладнання.
// 3) Блокує систему (IsSafetyLocked = true), щоб зупинити виконання Steps і автоматичний цикл.
// 4) Скидає IsPaused, бо Stop — це повна зупинка, а не пауза.
// 5) Скидає ActiveOperation та переводить режим у ModeManual.
// 6) Очищує чергу opQueue, щоб усі заплановані команди були видалені.
//
// Після Stop():
//   - рух зупинений
//   - нові операції не запускаються
//   - система чекає виклику SafetyStart()
func (c *Controller) Stop(reason string) {
	c.state.mu.Lock()
	c.state.StopReason = reason

	fmt.Printf("[CTRL] STOP: %s (Операція: %s)\n", reason, c.state.ActiveOperation)

	c.state.Device20Out[OutMainMotor] = 0 // force motor off
	c.state.IsSafetyLocked = true
	c.state.IsPaused = false
	c.state.ActiveOperation = ""
	c.state.Mode = ModeManual
	c.state.mu.Unlock()
  c.drainOpQueue()
}

// emergencyStop виконує аварійну зупинку (Fail-Safe).
// Це внутрішній метод контролера і не повинен викликатися оператором.
//
// Додатково до Stop():
//   - фіксує аварійний контекст у логах
//   - використовується при втраті зв'язку або апаратній помилці
//   - може викликатися після відключення живлення IO або інших fail-safe дій
func (c *Controller) emergencyStop(reason string) {
	fmt.Printf("[CTRL][FAULT] EMERGENCY STOP: %s\n", reason)

	// Усі механічні та логічні дії зупинки — через єдину точку
	c.Stop(reason)
}

// Reset повертає систему до робочого стану після зупинки або аварії.
// 1) Якщо outputsLost = true (плата виходів була відключена через помилку),
//      - включаємо живлення плати виходів через power.EnableOutputsPower()
//      - чекаємо 500 мс, щоб залізо прокинулось
//      - скидаємо outputsLost і встановлюємо firstRun = true для синхронізації виходів
// 2) Знімаємо блокування безпеки (IsSafetyLocked = false), очищаємо StopReason і ActiveOperation
// 3) Після виклику система готова для штатної роботи та прийняття команд
func (c *Controller) Reset() {
  if c.outputsLost {
    fmt.Println("[CTRL] Re-arming outputs power...")
    _ = c.power.EnableOutputsPower()   // Включаємо живлення DO-плати
    time.Sleep(500 * time.Millisecond) // Даємо залізу прокинутись
    c.outputsLost = false
    c.firstRun = true                  // Синхронізуємо виходи
  }
  c.state.mu.Lock()
  c.state.IsSafetyLocked = false
  c.state.StopReason = ""
  c.state.ActiveOperation = ""
  c.state.mu.Unlock()
  fmt.Println("[CTRL] Safety Start: блокування знято, система готова")
}

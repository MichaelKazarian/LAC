package main

import (
  "errors"
	"fmt"
	"time"
)

type OperationFunc func()

type Controller struct {
	hw    HardwareService
	state *HardwareState
	opsMap      map[string]OperationInfo
  lastOutput [32]uint16
  firstRun   bool
  opQueue    chan string
  needsCounterReset bool
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

	return &Controller{
		hw:      hw,
		state:   state,
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
		} else {
			c.handleError(err)
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

// Перевіряє зміни в стані та записує їх у залізо
func (c *Controller) syncHardware() {
  c.state.mu.RLock()
  current := c.state.Device20Out
  locked := c.state.IsSafetyLocked
  c.state.mu.RUnlock()

  // Якщо система заблокована, ми ігноруємо будь-які спроби
  // автоматичних або "доживаючих" операцій щось записати.
  if locked {
    // fmt.Println("[CTRL] Запис заблоковано: активний Safety Lock")
    current[OutMainMotor] = 0
  }
  // 2. Перевіряємо наявність змін
  if current == c.lastOutput && !c.firstRun {
    return
  }

  if err := c.hw.Write(current); err != nil {
    c.handleError(fmt.Errorf("[CTRL] Failed to update Outputs on addr %d: %w", AddrOutputs, err))
    return
  }

  // 4. Оновлюємо кеш тільки при успішному записі
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

func (c *Controller) handleError(err error) {
	c.updateInputsState(nil, false)
	c.updateEncoderState(0, false)

  var critErr *CriticalCommError
  if errors.As(err, &critErr) {
    EmergencyStop(c, fmt.Sprintf("Втрата зв'язку: %d помилок підряд", critErr.Streak))
    return
  }
	fmt.Printf("[CTRL] [%s] Помилка зв'язку: %v\n", time.Now().Format("15:04:05"), err)
}

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

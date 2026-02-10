package main

import (
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

func NewController(hw HardwareService, state *HardwareState) *Controller {
	registry := GetOperationsRegistry()

	// Швидкий доступ по ID
	opMap := make(map[string]OperationInfo)
	for _, op := range registry {
		opMap[op.ID] = op
	}
  state.OpsList=GetOperationsList()

	return &Controller{
		hw:       hw,
		state:    state,
		opsMap:   opMap,
		opQueue:  make(chan string, 10),
	}
}

// Run — основний цикл контролера
func (c *Controller) Run() {
  c.firstRun = true // Ініціалізуємо перед циклом
  fmt.Println("🚀 Контролер логіки запущено")

  for {
    start := time.Now()

    select {
		case opName := <-c.opQueue:
			c.exec(opName)
		default:
			// Якщо черга порожня — виконуємо звичайний цикл
		}

    // 1. Читання
    sensor, inputs, err := c.hw.Read()
		if err == nil {
			c.updateSlave3(sensor, true)
			c.updateSlave10(&inputs, true)
		} else {
			c.handleError(err)
		}

    c.switchMode()
    c.syncHardware()

    c.updateCycleTime(time.Since(start).Milliseconds())
  }
}

func (c *Controller) exec(opId string) {
	op, ok := c.opsMap[opId]
	if !ok {
		fmt.Printf("⚠️ Операція %s не знайдена\n", opId)
		return
	}

	c.state.mu.Lock()
	c.state.ActiveOperation = opId
	c.state.mu.Unlock()

	op.Action(c)

	c.state.mu.Lock()
	c.state.ActiveOperation = ""
	c.state.mu.Unlock()
}

// processLogic визначає поведінку контролера залежно від поточного режиму
func (c *Controller) switchMode() {
    c.state.mu.RLock()
    mode := c.state.Mode
    c.state.mu.RUnlock()

    switch mode {
    case ModeAutomatic: // просто крутимо логіку в кожному циклі
        c.executeLogic()
    case ModeSingle: // виконуємо логіку один раз і перемикаємо режим
        fmt.Println("🚀 Запуск одиночного циклу...")
        // Перемикаємо в Manual тільки якщо цикл реально ДОЙШОВ ДО КІНЦЯ
        c.executeLogic()
        c.SetMode(ModeManual)
        fmt.Println("✅ Одиночний цикл успішно завершено.")
    case ModeManual:
        // У ручному режимі контролер не втручається в Device20Out самостійно,
        // дозволяючи командам з InvokeOperation (Web) проходити без змін.
    }
}

// Перевіряє зміни в стані та записує їх у залізо
func (c *Controller) syncHardware() {
  // 1. Копіюємо стан під RLock
  c.state.mu.RLock()
  current := c.state.Device20Out
  locked := c.state.IsSafetyLocked
  c.state.mu.RUnlock()

  // Якщо система заблокована, ми ігноруємо будь-які спроби
  // автоматичних або "доживаючих" операцій щось записати.
  if locked {
    // fmt.Println("🚫 Запис заблоковано: активний Safety Lock")
    return
  }
  // 2. Перевіряємо наявність змін
  if current == c.lastOutput && !c.firstRun {
    return // Нічого не змінилося, виходимо
  }

  // 3. Записуємо в залізо
  if err := c.hw.Write(current); err != nil {
    c.handleError(fmt.Errorf("помилка запису Slave 20: %v", err))
    return
  }

  // 4. Оновлюємо кеш тільки при успішному записі
  c.lastOutput = current
  c.firstRun = false
  
  // fmt.Println("📡 [Modbus] Дані Slave 20 оновлено")
}

func (c *Controller) InvokeOperation(name string) error {
  if _, exists := c.opsMap[name]; !exists {
		return fmt.Errorf("операція %s не знайдена", name)
	}
  select {
	case c.opQueue <- name:
		fmt.Printf("[%s] 📨 Команда додана в чергу: %s\n", time.Now().Format("15:04:05"), name)
	default:
		return fmt.Errorf("черга переповнена, зачекайте")
	}
  return nil
}

// executeLogic послідовно виконує зареєстровані кроки сценарію
func (c *Controller) executeLogic() {
  // Якщо пауза натиснута після завершення попереднього повного циклу, 
  // ми не даємо почати новий цикл steps.
  //c.waitIfPaused()
  if c.needsCounterReset { //Перевірка на необхідність скидання лічильника
    c.state.mu.Lock()
    c.state.Counter = 0
    c.state.mu.Unlock()
    c.needsCounterReset = false
  }

  for _, op := range c.state.OpsList {
    c.waitIfPaused()
    // Після того, як пауза була знята (наприклад, через перехід у ручний режим),
		// перевіряємо, чи ми все ще маємо право виконувати цей цикл.
		c.state.mu.RLock()
		mode := c.state.Mode
		c.state.mu.RUnlock()

		if mode == ModeManual {
			fmt.Printf("🛑 Цикл перервано: режим змінено на ручний перед кроком %s\n", op[0])
			return // Виходимо з функції, решта операцій не виконається
		}
    c.exec(op[0])
  }
  c.state.mu.Lock()
  c.state.Counter++
  c.state.mu.Unlock()
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
        fmt.Println("🎛 Контролер переведено в РУЧНИЙ режим")
    }
    fmt.Printf("⚙️ Режим змінено на: %v\n", mode)
}

func (c *Controller) SetPause(paused bool) {
  c.state.mu.Lock()
  c.state.IsPaused = paused
  c.state.mu.Unlock()
  
  if paused {
    fmt.Println("⏸ ПАУЗА: Поточна операція допрацює, нові не почнуться.")
  } else {
    fmt.Println("▶️ ПРОДОВЖЕНО: Автоматичний цикл відновлено.")
  }
}

func (c *Controller) handleError(err error) {
	c.updateSlave10(nil, false)
	c.updateSlave3(0, false)
	fmt.Printf("[%s] Помилка зв'язку: %v\n", time.Now().Format("15:04:05"), err)
}

func (c *Controller) updateSlave3(val uint16, online bool) {
	c.state.mu.Lock()
	defer c.state.mu.Unlock()
	c.state.SensorValue = val
	c.state.IsOnline3 = online
	c.state.LastUpdate = time.Now()
}

func (c *Controller) updateSlave10(data *[32]uint16, online bool) {
	c.state.mu.Lock()
	defer c.state.mu.Unlock()
	c.state.IsOnline10 = online
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
  if c.state.IsSafetyLocked { // Якщо вже заблоковано, навіть не виконуємо логіку функції
        c.state.mu.Unlock()
        return
    }
  fn()
  c.state.mu.Unlock()
  c.syncHardware()
}

func (c *Controller) GetAllowedManualOps() []string {
  allowed := []string{}
  if c.state.SensorValue > 100 && c.state.SensorValue < 500 {
    allowed = append(allowed, c.opsMap["sync_mirror"].ID)
  }
  // Зупинка зазвичай дозволена завжди
  allowed = append(allowed, c.opsMap["op_safety_stop"].ID)
  return allowed
}

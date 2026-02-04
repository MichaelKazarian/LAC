package main

import (
	"fmt"
	"time"
)

type OperationFunc func()

type Controller struct {
	hw    HardwareService
	state *HardwareState
  operations map[string]OperationFunc
  lastOutput [32]uint16
  firstRun   bool
  opQueue    chan string
}

func NewController(hw HardwareService, state *HardwareState) *Controller {
    ctrl := &Controller{
      hw:    hw,
      state: state,
      operations: make(map[string]OperationFunc),
      opQueue:    make(chan string, 10),
    }

    // Реєструємо операції під зрозумілими іменами
    ctrl.operations["sync_mirror"] = ctrl.opSyncMirror
    ctrl.operations["op_safety_stop"]   = ctrl.opSafetyStop
    // Сюди додаватимете нові "ручні" дії
    
    return ctrl
}

// Run — основний цикл контролера
func (c *Controller) Run() {
  c.firstRun = true // Ініціалізуємо перед циклом
  fmt.Println("🚀 Контролер логіки запущено")

  for {
    start := time.Now()

    select {
		case opName := <-c.opQueue:
			if op, ok := c.operations[opName]; ok {
				op() // Виконуємо операцію (вона сама зробить syncHardware всередині)
			}
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

// processLogic визначає поведінку контролера залежно від поточного режиму
func (c *Controller) switchMode() {
    c.state.mu.RLock()
    mode := c.state.Mode
    c.state.IsPaused = false
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

// syncHardware перевіряє зміни в стані та записує їх у залізо
func (c *Controller) syncHardware() {
  // 1. Копіюємо стан під RLock
  c.state.mu.RLock()
  current := c.state.Device20Out
  c.state.mu.RUnlock()

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
  if _, exists := c.operations[name]; !exists {
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
  steps := []string{ // Визначаємо чергу операцій (порядок важливий)
    "sync_mirror",
    "op_safety_stop",
  }
  for _, opName := range steps {
    c.state.mu.RLock()
    p :=  c.state.IsPaused
    c.state.mu.RUnlock()
    if p {
      fmt.Printf("⏸ Логіку призупинено перед кроком: %s\n", opName)
      c.waitIfPaused()
    }
    if op, ok := c.operations[opName]; ok {
      op()
    } else {
      fmt.Printf("⚠️ Операція %s не знайдена в реєстрі\n", opName)
    }
  }
}

// SetMode безпечно оновлює режим роботи контролера
func (c *Controller) SetMode(mode ControlMode) {
    c.state.mu.Lock()
    defer c.state.mu.Unlock()
    c.state.IsPaused = false
    // Якщо ми переходимо в ручний режим, можна відразу скинути певні виходи, якщо треба
    if mode == ModeManual {
        fmt.Println("🎛 Контролер переведено в РУЧНИЙ режим")
    }
    c.state.Mode = mode
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
  fn()
  c.state.mu.Unlock()
  c.syncHardware()
}

// opSyncMirror - дзеркалювання з паузою
func (c *Controller) opSyncMirror() {
  fmt.Println("⏳ Початок синхронізації")
    // 1. Скидання
    c.apply(func() {
        for i := 0; i < 32; i++ { c.state.Device20Out[i] = 0 }
    })
    // time.Sleep(100 * time.Millisecond)

  // 2. Дзеркало
  c.apply(func() {
    for i := 0; i < 32; i++ {
      if c.state.Device20Out[i] != c.state.Device10In[i] {
        c.state.Device20Out[i] = c.state.Device10In[i]
      }
    }
  })
  time.Sleep(2 * time.Second)
  // 3. Фінальне скидання
  c.apply(func() {
    for i := 0; i < 32; i++ { c.state.Device20Out[i] = 0 }
  })
  fmt.Println("Кінець синхронізації")
}

// opSafetyStop - перемикання з паузою
func (c *Controller) opSafetyStop() {
	fmt.Println("⏳ Початок операції (3с)...")
  c.apply(func() { c.state.Device20Out[3] = 1})
  time.Sleep(2 * time.Second)
  c.apply(func() { c.state.Device20Out[3] = 0})
  fmt.Println("✅ Стан змінено")
}

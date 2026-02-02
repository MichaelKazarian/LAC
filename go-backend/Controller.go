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
}

func NewController(hw HardwareService, state *HardwareState) *Controller {
    ctrl := &Controller{
        hw:    hw,
        state: state,
        operations: make(map[string]OperationFunc),
    }

    // Реєструємо операції під зрозумілими іменами
    ctrl.operations["sync_mirror"] = ctrl.opSyncMirror
    ctrl.operations["op_safety_stop"]   = ctrl.opSafetyStop
    // Сюди додаватимете нові "ручні" дії
    
    return ctrl
}

// Run — основний цикл контролера
func (c *Controller) Run() {
	fmt.Println("🚀 Контролер логіки запущено")
	var lastOutput [32]uint16
	firstRun := true

	for {
		start := time.Now()

		// 1. Читання фізичних входів
		sensor, inputs, err := c.hw.Read()
		if err != nil {
			c.handleError(err)
			// Робимо невелику паузу при помилці, щоб не зациклити процесор
			time.Sleep(10 * time.Millisecond)
			continue
		}

		// Оновлюємо стан входів у HardwareState
		c.updateSlave3(sensor, true)
		c.updateSlave10(&inputs, true)

		// 2. Визначення режиму та виконання логіки
		c.state.mu.RLock()
		mode := c.state.Mode
		c.state.mu.RUnlock()

		switch mode {
		case ModeAutomatic:
			c.executeLogic()
		case ModeSingle:
			c.executeLogic()
			c.setMode(ModeManual) // Після одного проходу повертаємо в ручний
			fmt.Println("✅ Одиночний цикл завершено. Перехід у ModeManual")
		case ModeManual:
			// У ручному режимі нічого не робимо, чекаємо команд через Device20Out
		}

		// 3. Синхронізація з залізом (Slave 20)
		c.state.mu.RLock()
		currentOutput := c.state.Device20Out
		c.state.mu.RUnlock()

		if currentOutput != lastOutput || firstRun {
			if errWrite := c.hw.Write(currentOutput); errWrite == nil {
				lastOutput = currentOutput
				firstRun = false
			}
		}
		c.updateCycleTime(time.Since(start).Milliseconds())
	}
}

func (c *Controller) InvokeOperation(name string) error {
    op, exists := c.operations[name]
    if !exists {
        return fmt.Errorf("операція %s не знайдена", name)
    }

    // Захоплюємо mutex, бо ручний виклик — це теж втручання в стан
    c.state.mu.Lock()
    op()
    c.state.mu.Unlock()

    fmt.Printf("[%s] Manual Invoke: %s\n", time.Now().Format("15:04:05"), name)
    return nil
}

// executeLogic викликає послідовність операцій
func (c *Controller) executeLogic() {
    c.state.mu.Lock()
    defer c.state.mu.Unlock()

    // В автоматиці ми виконуємо базовий набір (наприклад, тільки sync_mirror)
    // Або всі зареєстровані операції послідовно
    // if op, ok := c.operations["sync_mirror"]; ok {
    //     op()
    // }
}

func (c *Controller) setMode(mode ControlMode) {
	c.state.mu.Lock()
	defer c.state.mu.Unlock()
	c.state.Mode = mode
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

// opSyncMirror - операція дзеркалювання входів на виходи
func (c *Controller) opSyncMirror() {
    changed := false
    for i := 0; i < 32; i++ {
        if c.state.Device20Out[i] != c.state.Device10In[i] {
            c.state.Device20Out[i] = c.state.Device10In[i]
            changed = true
        }
    }
    
    if changed {
        fmt.Printf("[%s] Logic: Outputs synchronized with inputs\n", 
            time.Now().Format("15:04:05"))
    }
}

func (c *Controller) opSafetyStop() {
    // Якщо сенсор вище норми — вимикаємо все
    // if c.state.SensorValue > 1000 {
    //     c.state.Device20Out[0] = 1 // Наприклад, вимкнути головний двигун
  // }

  fmt.Printf("[%s] cdas1 ", c.state.Device20Out[0]) 
  if c.state.Device20Out[0] == 0 {
    c.state.Device20Out[0] = 1
  } else {
    c.state.Device20Out[0] = 0
  }
  fmt.Printf("[%s] cdas2 ", c.state.Device20Out[0])
}

package main

import (
	"fmt"
	"time"
)

// OperationInfo описує метадані операції для UI та логіки
type OperationInfo struct {
  ID string
	UserName string
	Action   func(c *Controller)
}

func GetOperationsList() [][]string {
  opsList := GetOperationsRegistry()
	list := make([][]string, 0, len(opsList))
	for _, op := range opsList {
		list = append(list, []string{op.ID, op.UserName})
	}
	return list
}

// GetOperationsRegistry повертає карту всіх доступних операцій
func GetOperationsRegistry() []OperationInfo {
	return []OperationInfo {
    { ID:       "operation1",
      UserName: "Операція 1",
      Action:   opItWorks,
    },
      { ID:       "operation3",
        UserName: "Операція 3",
        Action:   opItWorks,
      },
      { ID:       "operation4",
        UserName: "Операція 4",
        Action:   opItWorks,
      },
      { ID:       "operation5",
        UserName: "Операція 5",
        Action:   opItWorks,
      },
      { ID:       "operation6",
        UserName: "Операція 6",
        Action:   opItWorks,
      },
      { ID:       "operation7",
        UserName: "Операція 7",
        Action:   opItWorks,
      },
      { ID:       "operation8",
        UserName: "Операція 8",
        Action:   opItWorks,
      },
      { ID:       "sync_mirror",
        UserName: "Дзеркалювання",
        Action:   opSyncMirror,
      },
      { ID:       "op_safety_stop",
        UserName: "Безпечна зупинка",
        Action:   opSafetyStop,
      },
      { ID:       "operation11",
        UserName: "Операція 11",
        Action:   opItWorks,
      },
      { ID:       "operation12",
        UserName: "Операція 12",
        Action:   opItWorks,
      },
      { ID:       "operation13",
        UserName: "Операція 13",
        Action:   opItWorks,
      },
      { ID:       "operation14",
        UserName: "Операція 14",
        Action:   opItWorks,
      },
      { ID:       "operation15",
        UserName: "Операція 15",
        Action:   opItWorks,
      },
      { ID:       "operation16",
        UserName: "Операція 16",
        Action:   opItWorks,
      },
      { ID:       "operation17",
        UserName: "Операція 17",
        Action:   opItWorks,
      },
      { ID:       "operation18",
        UserName: "Операція 18",
        Action:   opItWorks,
      },
      { ID:       "operation19",
        UserName: "Операція 19",
        Action:   opItWorks,
      },
      { ID:       "operation20",
        UserName: "Операція 20",
        Action:   opItWorks,
      },
    }
}

// GetAllowedManualOps визначає, які операції доступні для натискання в ручному режимі
func GetAllowedManualOps(state *HardwareState) []string {
    // 1. ПРІОРИТЕТ №1: Якщо активовано Emergency Stop (Safety Lock)
    // Система "завмерла", дозволяємо тільки кнопку розблокування (Start)
    if state.IsSafetyLocked {
        return []string{"op_safety_start"}
    }
    // 2. ПРІОРИТЕТ №2: Якщо виконується будь-яка операція (ручна або авто-крок)
    // Дозволяємо ТІЛЬКИ зупинку
    if state.ActiveOperation != "" {
        return []string{"op_safety_stop"}
    }
    // 3. ЗВИЧАЙНИЙ РЕЖИМ: Вираховуємо доступні зони
    allowed := make([]string, 0)
    val := state.SensorValue
    // Логіка зон (наприклад, для синхронізації дзеркал)
    if val > 100 && val < 500 {
        allowed = append(allowed, "sync_mirror")
    }
    // Операція "Стоп" доступна завжди, коли машина в спокої
    allowed = append(allowed, "op_safety_stop")
    // Додаткові умови по мережі/статусу пристроїв
    if state.IsOnline20 {
        // allowed = append(allowed, "some_network_op")
    }
    return allowed
}

// Заморожуємо стан, вимикаючи тільки головний двигун.
// Очищаємо черги операцй, щоб команди не виконались потім
func EmergencyStop(c *Controller, reason string) {
  c.state.mu.Lock()
  
  // Записуємо причину для фронтенду
  c.state.StopReason = reason
  fmt.Printf("🚨 EMERGENCY STOP: %s (Операція: %s)\n", reason, c.state.ActiveOperation)

  c.state.Device20Out[31] = 0
  c.state.StopReason = reason
  c.state.IsSafetyLocked = true
  c.state.IsPaused = false
  c.state.ActiveOperation = "" 
  c.state.Mode = ModeManual
  c.state.mu.Unlock()

  for len(c.opQueue) > 0 { 
    select {
    case <-c.opQueue:
    default:
    }
  }
}

// Оновлюємо opSafetyStart, щоб очищати причину
func SafetyStart(c *Controller) {
  c.state.mu.Lock()
  c.state.IsSafetyLocked = false
  c.state.StopReason = ""
  c.state.ActiveOperation = ""
  c.state.mu.Unlock()
}

func opSyncMirror(c *Controller) {
    fmt.Println("⏳ Початок синхронізації")

    // --- Внутрішня перевірка безпеки ---
    checkEmergency := func() bool {
        c.state.mu.RLock()
        val := c.state.SensorValue/2
        isLocked := c.state.IsSafetyLocked
        c.state.mu.RUnlock()

        if val > 250 && val < 300 {
            fmt.Printf("⚠️ АВТОМАТИЧНА ЗУПИНКА: Сенсор = %d\n", val)
            EmergencyStop(c, fmt.Sprintf("Перевищено поріг сенсора: %d", val))
            return true
        }
        if isLocked {
            return true
        }
        return false
    }

    // Крок 1: Вмикаємо двигун (тільки в пам'яті)
    c.apply(func() { c.state.Device20Out[31] = 1 })
    if checkEmergency() { return }
    // Даємо IO-циклу час
    if !interruptibleSleep(c, 2*time.Millisecond, checkEmergency) { return }
    // Крок 2: Синхронізація
    c.apply(func() {
        for i := 0; i < 31; i++ {
            c.state.Device20Out[i] = c.state.Device10In[i]
        }
    })
    if checkEmergency() { return }
    // Чекаємо завершення фізичного процесу
  if !interruptibleSleep(c, 2*5*time.Second, checkEmergency) { return }
    // Крок 3: Скидання (тільки в пам'яті)
    c.apply(func() {
        for i := 0; i < 31; i++ { c.state.Device20Out[i] = 0 }
    })
    if checkEmergency() { return }

    fmt.Println("✅ Кінець синхронізації")
}

func interruptibleSleep(c *Controller, d time.Duration, check func() bool) bool {
    start := time.Now()
    for time.Since(start) < d {
        c.state.mu.RLock()
        locked := c.state.IsSafetyLocked
        c.state.mu.RUnlock()

        if locked || (check != nil && check()) {
            return false // зупиняємо операцію негайно
        }
        time.Sleep(5 * time.Millisecond)
    }
    return true
}

func opItWorks(c *Controller) {
	fmt.Println("✅ Це працює")
  interruptibleSleep(c, 1*time.Second, func() bool {return false})
}

func opSafetyStop(c *Controller) {
  checkEmergency := func() bool {return false}
	fmt.Println("Початок Зупинки (2с)...")
	c.apply(func() { c.state.Device20Out[3] = 1 })
	if !interruptibleSleep(c, 2*time.Second, checkEmergency) { return }
  c.apply(func() {
    for i := 0; i < 32; i++ { c.state.Device20Out[i] = 0 }
  })
	fmt.Println("зупинено")
}

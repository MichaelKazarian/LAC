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
func EmergencyStop(c *Controller) {
  fmt.Printf("🚨 EMERGENCY STOP: Операція [%s] перервана. Вимкнення головного двигуна.\n", c.state.ActiveOperation)

  c.state.mu.Lock()
  c.state.Device20Out[31] = 0
  c.state.ActiveOperation = ""
  c.state.IsSafetyLocked = true
  c.state.IsPaused = false
  c.state.Mode = ModeManual
  c.state.mu.Unlock()
  for len(c.opQueue) > 0 {
    <-c.opQueue
  }
  // Негайний запис у залізо
  c.hw.Write(c.state.Device20Out)
  // Синхронізуємо кеш контролера
  c.lastOutput = c.state.Device20Out
}

func opSafetyStart(c *Controller) {
  c.state.mu.Lock()
  c.state.IsSafetyLocked = false
  c.state.mu.Unlock()
  fmt.Println("✅ Блокування знято. Система готова до роботи.")
}

func opSyncMirror(c *Controller) {
	fmt.Println("⏳ Початок синхронізації")
	c.apply(func() {
    c.state.Device20Out[31] = 1
		for i := 0; i < 31; i++ { c.state.Device20Out[i] = 0 }
	})

	c.apply(func() {
		for i := 0; i < 31; i++ {
			if c.state.Device20Out[i] != c.state.Device10In[i] {
				c.state.Device20Out[i] = c.state.Device10In[i]
			}
		}
	})
	time.Sleep(2 * time.Second)

	c.apply(func() {
		for i := 0; i < 31; i++ { c.state.Device20Out[i] = 0 }
	})
	fmt.Println("Кінець синхронізації")
}

func opSafetyStop(c *Controller) {
	fmt.Println("⏳ Початок операції (2с)...")
	c.apply(func() { c.state.Device20Out[3] = 1 })
	time.Sleep(2 * time.Second)
	c.apply(func() { c.state.Device20Out[31] = 0; c.state.Device20Out[3] = 0 })
	fmt.Println("✅ Стан змінено")
}

func opItWorks(c *Controller) {
	fmt.Println("✅ Це працює")
}

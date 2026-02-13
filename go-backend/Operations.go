package main

import (
	"fmt"
	"time"
)

// StepStatus описує результат виконання одного кроку (Step).
//
// Це стан логічного процесу.
// Контролер використовує цей статус, щоб вирішити — продовжувати операцію,
// зупинити її або аварійно перервати.
//
//   StepOK    — крок успішно завершений, можна переходити до наступного.
//   StepFail  — крок завершився штатною помилкою (технологічна невдача).
//               Операція зупиняється, але без EmergencyStop.
//   StepAbort — виконання перервано зовнішньою умовою (EmergencyStop,
//               блокування, зміна режиму тощо).
//
type StepStatus int

const (
	StepOK StepStatus = iota
	StepFail
	StepAbort
)

// StepResult — результат очікування завершення кроку.
//
// Використовується функцією Wait для передачі контролеру підсумку,
// коли фізичний процес завершився або був перерваний.
//
// Message — опціональне пояснення (для логів / UI).
//
type StepResult struct {
	Status  StepStatus
	Message string
}

// Step описує атомарний крок технологічної операції.
//
// Ідея Step-моделі:
//
//   * Технологічна операція розбивається на набор послідовних кроків.
//   * Контролер виконує цю послідовність.
//   * Сам Step НЕ керує циклом виконання.
//
// Це усуває ситуацію, коли кожна операція реалізує власний scheduler
//
// Життєвий цикл Step:
//
//   1. Controller викликає Do()
//      → змінюємо виходи, запускаємо фізичну дію.
//      → жодних очікувань усередині!
//
//   2. Controller викликає Wait()
//      → Wait сама визначає, коли процес завершився:
//          - по сенсору
//          - по таймауту
//          - по зовнішній події
//      → Wait МОЖЕ блокуватися, але це єдине дозволене місце очікування.
//
//   3. Якщо потрібно — викликається Cleanup()
//      → гарантує приведення системи в безпечний стан.
//
// Таким чином, Controller є єдиним execution engine,
// а Step декларативним описом технології
//
// Це робить поведінку:
//   передбачуваною
//   керованою (Pause / EmergencyStop)
//   спостережуваною (можна логувати Name)
//   розширюваною без вкладених state machine.
//
type Step struct {
	Name       string
	Do         func(c *Controller)
	Wait       func(c *Controller) StepResult
	Cleanup    func(c *Controller)
}

// OperationInfo описує технологічну операцію як послідовність Step.
// Тобто:
//   Операція = сценарій = []Step
// Контролер інтерпретує список Steps як маленький DSL процесу.
//
// Це дозволяє:
//   легко додавати нові операції без складної логіки
//   бачити поточний Step у UI
//   централізовано керувати Abort / Pause / Mode
//   уникнути дублювання interruptibleSleep у кожній операції
//
type OperationInfo struct {
	ID       string
	UserName string
	Steps    []Step
}

// GetOperationsRegistry повертає карту всіх доступних операцій
func GetOperationsRegistry() []OperationInfo {
	return []OperationInfo {
    {
			ID:       "operation1",
        UserName: "Операція 1",
        Steps: []Step{
          {Name: "DoSomething", Do: stepItWorks, Wait: waitAlwaysOK},
        },
      },
      {
			ID:       "operation2",
        UserName: "Операція 2",
        Steps: []Step{
          {Name: "DoSomething", Do: stepItWorks, Wait: waitAlwaysOK},
        },
      },
      {
			ID:       "operation3",
        UserName: "Операція 3",
        Steps: []Step{
          {Name: "DoSomething", Do: stepItWorks, Wait: waitAlwaysOK},
        },
      },
      {
			ID:       "operation4",
        UserName: "Операція 4",
        Steps: []Step{
          {Name: "DoSomething", Do: stepItWorks, Wait: waitAlwaysOK},
        },
      },
      {
			ID:       "operation5",
        UserName: "Операція 5",
        Steps: []Step{
          {Name: "DoSomething", Do: stepItWorks, Wait: waitAlwaysOK},
        },
      },
      {
			ID:       "operation6",
        UserName: "Операція 6",
        Steps: []Step{
          {Name: "DoSomething", Do: stepItWorks, Wait: waitAlwaysOK},
        },
      },
      {
			ID:       "operation7",
        UserName: "Операція 7",
        Steps: []Step{
          {Name: "DoSomething", Do: stepItWorks, Wait: waitAlwaysOK},
        },
      },
      {
			ID:       "operation8",
        UserName: "Операція 8",
        Steps: []Step{
          {Name: "DoSomething", Do: stepItWorks, Wait: waitAlwaysOK},
        },
      },
      {
			ID:       "operation9",
        UserName: "Операція 9",
        Steps: []Step{
          {Name: "DoSomething", Do: stepItWorks, Wait: waitAlwaysOK},
        },
      },
      {
			ID:       "sync_mirror",
        UserName: "Дзеркалювання",
        Steps: []Step{
          {
            Name: "Включення двигуна",
            Do: func(c *Controller) {
              c.apply(func() { c.state.Device20Out[31] = 1 })
            },
            Wait: waitMotorOn,
          },
          {
            Name: "Синхронізація",
            Do: func(c *Controller) {
              c.apply(func() {
                for i := 0; i < 31; i++ {
                  c.state.Device20Out[i] = c.state.Device10In[i]
                }
              })
            },
            Wait: waitSyncMirror,
          },
          {
            Name: "Вимкнення двигуна",
            Do: func(c *Controller) {
              c.apply(func() {
                for i := 0; i < 32; i++ {
                  c.state.Device20Out[i] = 0
                }
              })
            },
            Wait: waitAlwaysOK,
          },
        },
      },
      {
			ID:       "op_safety_stop",
        UserName: "Безпечна зупинка",
        Steps: []Step{
          {
            Name: "Стоп",
            Do: func(c *Controller) {
              c.apply(func() { c.state.Device20Out[3] = 1 })
            },
            Wait: waitStop2s,
            Cleanup: func(c *Controller) {
              c.apply(func() {
                for i := 0; i < 32; i++ {
                  c.state.Device20Out[i] = 0
                }
              })
            },
          },
        },
    },

      {
			ID:       "operation10",
        UserName: "Операція 10",
        Steps: []Step{
          {Name: "DoSomething", Do: stepItWorks, Wait: waitAlwaysOK},
        },
      },
      {
			ID:       "operation11",
        UserName: "Операція 11",
        Steps: []Step{
          {Name: "DoSomething", Do: stepItWorks, Wait: waitAlwaysOK},
        },
      },
      {
			ID:       "operation12",
        UserName: "Операція 12",
        Steps: []Step{
          {Name: "DoSomething", Do: stepItWorks, Wait: waitAlwaysOK},
        },
      },
      {
			ID:       "operation13",
        UserName: "Операція 13",
        Steps: []Step{
          {Name: "DoSomething", Do: stepItWorks, Wait: waitAlwaysOK},
        },
      },
      {
			ID:       "operation14",
        UserName: "Операція 14",
        Steps: []Step{
          {Name: "DoSomething", Do: stepItWorks, Wait: waitAlwaysOK},
        },
      },
      {
			ID:       "operation15",
        UserName: "Операція 15",
        Steps: []Step{
          {Name: "DoSomething", Do: stepItWorks, Wait: waitAlwaysOK},
        },
      },
      {
			ID:       "operation16",
        UserName: "Операція 16",
        Steps: []Step{
          {Name: "DoSomething", Do: stepItWorks, Wait: waitAlwaysOK},
        },
      },
      {
			ID:       "operation17",
        UserName: "Операція 17",
        Steps: []Step{
          {Name: "DoSomething", Do: stepItWorks, Wait: waitAlwaysOK},
        },
      },
      {
			ID:       "operation18",
        UserName: "Операція 8",
        Steps: []Step{
          {Name: "DoSomething", Do: stepItWorks, Wait: waitAlwaysOK},
        },
      },
      {
			ID:       "operation19",
        UserName: "Операція 19",
        Steps: []Step{
          {Name: "DoSomething", Do: stepItWorks, Wait: waitAlwaysOK},
        },
      },
      {
			ID:       "operation20",
        UserName: "Операція 20",
        Steps: []Step{
          {Name: "DoSomething", Do: stepItWorks, Wait: waitAlwaysOK},
        },
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

loop:
  for {
    select {
    case <-c.opQueue:
    default:
      break loop
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

// --- Wait-функції ---

func waitAlwaysOK(c *Controller) StepResult {
	return StepResult{Status: StepOK}
}

// Імітуємо очікування мотору включеним
func waitMotorOn(c *Controller) StepResult {
	timeout := time.After(100 * time.Millisecond)
	tick := time.Tick(5 * time.Millisecond)

	for {
		select {
		case <-timeout:
			return StepResult{Status: StepOK} // завершено
		case <-tick:
			c.state.mu.RLock()
			locked := c.state.IsSafetyLocked
			c.state.mu.RUnlock()
			if locked {
				return StepResult{Status: StepAbort, Message: "EmergencyStop"}
			}
		}
	}
}

// Очікуємо поки сенсор синхронізації не досягне порогу
func waitSyncMirror(c *Controller) StepResult {
	for {
		c.state.mu.RLock()
		val := c.state.SensorValue / 2
		locked := c.state.IsSafetyLocked
		c.state.mu.RUnlock()

		if locked {
			return StepResult{Status: StepAbort, Message: "EmergencyStop"}
		}
		if val > 250 && val < 300 {
			EmergencyStop(c, fmt.Sprintf("Перевищено поріг сенсора: %d", val))
			return StepResult{Status: StepAbort, Message: "Sensor threshold exceeded"}
		}
		if val > 200 {
			return StepResult{Status: StepOK}
		}
		time.Sleep(5 * time.Millisecond)
	}
}

// Стоп 2 секунди
func waitStop2s(c *Controller) StepResult {
	start := time.Now()
	for time.Since(start) < 2*time.Second {
		c.state.mu.RLock()
		locked := c.state.IsSafetyLocked
		c.state.mu.RUnlock()
		if locked {
			return StepResult{Status: StepAbort}
		}
		time.Sleep(5 * time.Millisecond)
	}
	return StepResult{Status: StepOK}
}

// --- Do-функції для простих Steps ---

func stepItWorks(c *Controller) {
	fmt.Println("✅ Це працює")
}


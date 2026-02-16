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
// # Ідея Step-моделі
//
//   - Технологічна операція розбивається на набір послідовних кроків.
//   - Контролер виконує цю послідовність як детермінований керуючий автомат.
//   - Сам Step НЕ керує циклом виконання — він декларативний опис технології.
//
// Це усуває ситуацію, коли кожна операція реалізує власний scheduler.
//
// # Модель виконання
//
// Система реалізує детермінований керуючий автомат через closure-factory:
//
//   OperationInfo  — шаблон (definition): описує що робити.
//   Build()        — інстанціювання (runtime instance): створює свіжий контекст.
//   []Step         — одноразовий execution-plan: виконали → викинули.
//
// Натиснули кнопку → Build() → отримали "свіжі" Steps → виконали → забули.
// Жодного shared state між запусками. Локальні змінні — нові на кожен запуск.
//
// Один execution flow, чіткі межі операції, ніяких fan-out / fan-in.
// Це НЕ ускладнення — це відкладене створення (lazy instantiation).
//
// # Життєвий цикл Step
//
//  1. Do() — опціональний.
//     Змінюємо виходи, запускаємо фізичну дію.
//     Має завершитись миттєво — жодних очікувань усередині.
//
//  2. Wait() — опціональний.
//     Wait сама визначає, коли процес завершився:
//       - по сенсору
//       - по таймауту
//       - по зовнішній події
//     Wait МОЖЕ блокуватися — це єдине дозволене місце очікування.
//     Wait повертає StepResult, який сигналізує контролеру про результат.
//     Якщо Wait == nil, крок вважається успішним (StepOK) миттєво.
//
//  3. Cleanup() — опціональний, аналог defer для Step.
//     Виконується після Wait якщо EmergencyStop не спрацював.
//     Гарантує приведення системи в штатний стан після:
//       - успішного завершення (StepOK)
//       - штатного переривання (StepAbort, StepFail)
//     **НЕ виконується** якщо спрацював EmergencyStop —
//     система вже переведена в безпечний стан самим EmergencyStop.
//
// # Порядок перевірок EmergencyStop в контролері
//
//   - перед Do      — зовнішній сигнал між кроками
//   - перед Wait    — Do міг спровокувати EmergencyStop
//   - перед Cleanup — Wait могла тривати довго
//
// Якщо EmergencyStop виявлено на будь-якому з цих етапів —
// виконання зупиняється негайно, Cleanup не викликається.
//
// # Властивості системи
//
// Таким чином, Controller є єдиним execution engine, а Step — декларативним описом технології.
// Це робить поведінку:
//   - передбачуваною
//   - керованою  (Pause / EmergencyStop)
//   - спостережуваною (Name для логів / UI)
//   - розширюваною без вкладених state machine
type Step struct {
	Name    string
	Do      func(c *Controller)
	Wait    func(c *Controller) StepResult
	Cleanup func(c *Controller)
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
//   вклинюватись між кроками (логування, валідація, interlock-и,
//   сервісні дії) без зміни коду самих операцій — достатньо
//   змінити поведінку контролера або вставити службовий Step.
//
type OperationInfo struct {
	ID       string
	DisplayName string
	Build func() []Step // Створює НОВИЙ сценарій виконання.
                      // Викликається КОЖЕН раз перед запуском операції.
}

func StepDoWait(name string, do func(c *Controller), wait func(c *Controller) StepResult) Step {
    return Step{Name: name, Do: do, Wait: wait}
}

// GetOperationsRegistry повертає карту всіх доступних операцій
func GetOperationsRegistry() []OperationInfo {
	return []OperationInfo{
		{ID: "operation1",  DisplayName: "Операція 1",  Build: func() []Step { return []Step{StepDoWait("DoSomething", stepItWorks, waitStop2s)} }},
		{ID: "operation2",  DisplayName: "Операція 2",  Build: func() []Step { return []Step{StepDoWait("DoSomething", stepItWorks, waitStop2s)} }},
		{ID: "operation3",  DisplayName: "Операція 3",  Build: func() []Step { return []Step{StepDoWait("DoSomething", stepItWorks, waitAlwaysOK)} }},
		{ID: "operation4",  DisplayName: "Операція 4",  Build: func() []Step { return []Step{StepDoWait("DoSomething", stepItWorks, waitAlwaysOK)} }},
		{ID: "operation5",  DisplayName: "Операція 5",  Build: func() []Step { return []Step{StepDoWait("DoSomething", stepItWorks, waitAlwaysOK)} }},
		{ID: "operation6",  DisplayName: "Операція 6",  Build: func() []Step { return []Step{StepDoWait("DoSomething", stepItWorks, waitAlwaysOK)} }},
		{ID: "operation7",  DisplayName: "Операція 7",  Build: func() []Step { return []Step{StepDoWait("DoSomething", stepItWorks, waitAlwaysOK)} }},
		{ID: "operation8",  DisplayName: "Операція 8",  Build: func() []Step { return []Step{StepDoWait("DoSomething", stepItWorks, waitAlwaysOK)} }},
		{ID: "operation9",  DisplayName: "Операція 9",  Build: func() []Step { return []Step{StepDoWait("DoSomething", stepItWorks, waitAlwaysOK)} }},
		{
			ID:       "sync_mirror",
			DisplayName: "Дзеркалювання",
			Build: func() []Step {
				return []Step{
					{
						Name: "Включення двигуна",
						Do:   func(c *Controller) { c.apply(func() { c.state.Device20Out[31] = 1 }) },
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
						Cleanup: func(c *Controller) {
							c.apply(func() {
								for i := 0; i < 31; i++ { c.state.Device20Out[i] = 0 }
							})
						},
					},
				}
			},
		},
		{
			ID:       "op_safety_stop",
			DisplayName: "Безпечна зупинка",
			Build: func() []Step {
				return []Step{
					{
						Name: "Стоп",
						Do:   func(c *Controller) { c.apply(func() { c.state.Device20Out[3] = 1 }) },
						Wait: waitStop2s,
						Cleanup: func(c *Controller) {
							c.apply(func() {
								for i := 0; i < 32; i++ { c.state.Device20Out[i] = 0 }
							})
						},
					},
				}
			},
		},
		{ID: "operation10", DisplayName: "Операція 10", Build: func() []Step { return []Step{StepDoWait("DoSomething", stepItWorks, waitAlwaysOK)} }},
		{ID: "operation11", DisplayName: "Операція 11", Build: func() []Step { return []Step{StepDoWait("DoSomething", stepItWorks, waitAlwaysOK)} }},
		{ID: "operation12", DisplayName: "Операція 12", Build: func() []Step { return []Step{StepDoWait("DoSomething", stepItWorks, waitAlwaysOK)} }},
		{ID: "operation13", DisplayName: "Операція 13", Build: func() []Step { return []Step{StepDoWait("DoSomething", stepItWorks, waitAlwaysOK)} }},
		{ID: "operation14", DisplayName: "Операція 14", Build: func() []Step { return []Step{StepDoWait("DoSomething", stepItWorks, waitAlwaysOK)} }},
		{ID: "operation15", DisplayName: "Операція 15", Build: func() []Step { return []Step{StepDoWait("DoSomething", stepItWorks, waitAlwaysOK)} }},
		{ID: "operation16", DisplayName: "Операція 16", Build: func() []Step { return []Step{StepDoWait("DoSomething", stepItWorks, waitAlwaysOK)} }},
		{ID: "operation17", DisplayName: "Операція 17", Build: func() []Step { return []Step{StepDoWait("DoSomething", stepItWorks, waitAlwaysOK)} }},
		{ID: "operation18", DisplayName: "Операція 18", Build: func() []Step { return []Step{StepDoWait("DoSomething", stepItWorks, waitAlwaysOK)} }},
		{ID: "operation19", DisplayName: "Операція 19", Build: func() []Step { return []Step{StepDoWait("DoSomething", stepItWorks, waitAlwaysOK)} }},
		{ID: "operation20", DisplayName: "Операція 20", Build: func() []Step { return []Step{StepDoWait("DoSomething", stepItWorks, waitAlwaysOK)} }},
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

// Протягом 5 секунд виконуємо "дзеркалювання" входів.
// Якщо в цей час сенсор потрапляє в аварійний діапазон — викликаємо EmergencyStop.
func waitSyncMirror(c *Controller) StepResult {
	deadline := time.Now().Add(5 * time.Second)

	for time.Now().Before(deadline) {
		c.state.mu.RLock()
		val := c.state.SensorValue / 2
		locked := c.state.IsSafetyLocked
		inputs := c.state.Device10In
		c.state.mu.RUnlock()

		// Якщо вже є блокування — просто перериваємо крок
		if locked {
			return StepResult{Status: StepAbort, Message: "EmergencyStop already active"}
		}

		if val > 250 && val < 300 {
			EmergencyStop(c, fmt.Sprintf("Перевищено поріг сенсора: %d", val))
			return StepResult{Status: StepAbort, Message: "Sensor threshold exceeded"}
		}

		c.apply(func() { // дзеркалювання
			for i := 0; i < 31; i++ {
				c.state.Device20Out[i] = inputs[i]
			}
		})
		time.Sleep(2 * time.Millisecond) // дрібний крок, щоб залишитись interruptible
	}
	return StepResult{Status: StepOK}
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


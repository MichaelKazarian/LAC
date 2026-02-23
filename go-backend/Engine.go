package main

import (
	"time"
)

// StepStatus описує результат виконання одного кроку (Step).
//
// Це стан логічного процесу.
// Контролер використовує цей статус, щоб вирішити — продовжувати операцію,
// зупинити її або аварійно перервати.
//
//	StepOK    — крок успішно завершений, можна переходити до наступного.
//	StepFail  — крок завершився штатною помилкою (технологічна невдача).
//	            Операція зупиняється, але без EmergencyStop.
//	StepAbort — виконання перервано зовнішньою умовою (EmergencyStop,
//	            блокування, зміна режиму тощо).
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
//	OperationInfo  — шаблон (definition): описує що робити.
//	Build()        — інстанціювання (runtime instance): створює свіжий контекст.
//	[]Step         — одноразовий execution-plan: виконали → викинули.
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
//     - по сенсору
//     - по таймауту
//     - по зовнішній події
//     Wait МОЖЕ блокуватися — це єдине дозволене місце очікування.
//     Wait повертає StepResult, який сигналізує контролеру про результат.
//     Якщо Wait == nil, крок вважається успішним (StepOK) миттєво.
//
//  3. Cleanup() — опціональний, аналог defer для Step.
//     Виконується після Wait якщо EmergencyStop не спрацював.
//     Гарантує приведення системи в штатний стан після:
//     - успішного завершення (StepOK)
//     - штатного переривання (StepAbort, StepFail)
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
//
//	Операція = сценарій = []Step
//
// Контролер інтерпретує список Steps як маленький DSL процесу.
//
// Це дозволяє:
//   - легко додавати нові операції без складної логіки
//   - бачити поточний Step у UI
//   - централізовано керувати Abort / Pause / Mode
//   - уникнути дублювання interruptibleSleep у кожній операції
//   - вклинюватись між кроками (логування, валідація, interlock-и)
//     без зміни коду самих операцій
type OperationInfo struct {
	ID          string
	DisplayName string
	Build       func() []Step // Створює НОВИЙ сценарій виконання.
	// Викликається КОЖЕН раз перед запуском операції.
}

// OperationRegistry — реєстр доступних операцій.
type OperationRegistry struct {
	ops []OperationInfo
}

func NewOperationRegistry() *OperationRegistry {
	return &OperationRegistry{}
}

// Add реєструє нову операцію.
func (r *OperationRegistry) Add(id, displayName string, build func() []Step) {
	r.ops = append(r.ops, OperationInfo{
		ID:          id,
		DisplayName: displayName,
		Build:       build,
	})
}

// All повертає всі зареєстровані операції.
func (r *OperationRegistry) All() []OperationInfo {
	return r.ops
}

// StepDoWait — helper для створення простого кроку з Do і Wait.
func StepDoWait(name string, do func(c *Controller), wait func(c *Controller) StepResult) Step {
	return Step{Name: name, Do: do, Wait: wait}
}

// --- Базові Wait-функції ---

// waitAlwaysOK — крок без очікування, завжди успішний.
func waitAlwaysOK(_ *Controller) StepResult {
	return StepResult{Status: StepOK}
}

// waitStop2s — очікування 2 секунди з перевіркою EmergencyStop.
func waitStop2s(c *Controller) StepResult {
	start := time.Now()
	for time.Since(start) < 2*time.Second {
		if c.isEmergency() {
			return StepResult{Status: StepAbort, Message: "Зупинено через SafetyLock"}
		}
		time.Sleep(10 * time.Millisecond)
	}
	return StepResult{Status: StepOK}
}

// --- Системні функції ---

// GetAllowedManualOpsFromView — UI-safe версія.
// Працює тільки з immutable snapshot, без доступу до runtime state.
func GetAllowedManualOpsFromView(v SystemView) []string {
	if v.IsSafetyLocked {
		return []string{"op_safety_start"}
	}
	if v.ActiveOperation != "" {
		return []string{"op_safety_stop"}
	}

	allowed := make([]string, 0)

	val := v.EncoderValue
	if val > 100 && val < 500 {
		allowed = append(allowed, "sync_mirror")
	}

	allowed = append(allowed, "op_safety_stop")

	// Тут вже НЕ можна перевіряти щось типу mutex / online flags напряму —
	// тільки те, що є у snapshot.
	return allowed
}

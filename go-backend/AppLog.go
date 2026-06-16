package main

import (
	"fmt"
	"log/slog"
	"sync"
)

type StepLogMode int

const (
	LogNormal StepLogMode = iota
	LogOnce
	LogSilent
)

type AppLog struct{}

// Структура для збереження стану логування конкретного кроку
type stepState struct {
	hasLoggedRepeat bool
}

var (
	// М'ютекс для захисту мапи від одночасного запису з різних горутин
	logMu sync.Mutex
	// Мапа: ключ — step.Name, значення — стан цього кроку
	registry = make(map[string]*stepState)
	// Трекер для визначення зміни кроку в межах системи
	lastStepName string
)

func (AppLog) Log(step Step, msg string) {
	if step.LogMode == LogSilent {
		return
	}

	logMu.Lock()
	// Якщо цей крок зайшов уперше взагалі, створюємо для нього запис
	state, exists := registry[step.Name]
	if !exists {
		state = &stepState{}
		registry[step.Name] = state
	}

	// Логіка зміни кроку: якщо ми перейшли до нового кроку,
	// очищаємо стан попереднього, щоб він міг логуватися знову в майбутньому циклі.
	if lastStepName != "" && lastStepName != step.Name {
		delete(registry, lastStepName)
	}
	lastStepName = step.Name

	// Перевірка режиму LogOnce
	if step.LogMode == LogOnce {
		// Якщо це повторне повідомлення або внутрішній лог, і ми ВЖЕ один раз щось вивели
		if state.hasLoggedRepeat {
			logMu.Unlock()
			return // Приглушуємо спам
		}
		
		// Фіксуємо, що перший вивід відбувся.
		// Наступні виклики для цього кроку (поки він не зміниться) будуть ігноруватися.
		state.hasLoggedRepeat = true
	}
	logMu.Unlock()

	// виводимо ПОЗА м'ютексом, щоб не блокувати інші потоки на час I/O операцій
	slog.Info(fmt.Sprintf("%s %s", step.Name, msg))
}

func (AppLog) LogForce(step Step, msg string) {
	slog.Warn(fmt.Sprintf("[FORCE] %s %s", step.Name, msg))
}

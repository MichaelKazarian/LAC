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

var (
	logMu    sync.Mutex
	registry = make(map[string]bool)
)

func (AppLog) log(step Step, level slog.Level, msg string) {
	if step.LogMode == LogSilent {
		return
	}

	logMu.Lock()
	if step.LogMode == LogOnce {
		// мовчимо, якщо хоч щось для цього кроку вже було залоговано
		if registry[step.Name] {
			logMu.Unlock()
			return
		}
		registry[step.Name] = true
	}
	logMu.Unlock()

	slog.Log(nil, level, fmt.Sprintf("%s %s", step.Name, msg))
}

// Reset викликається в defer для повного очищення пам'яті сесії
func (AppLog) Reset(step Step) {
	logMu.Lock()
	delete(registry, step.Name)
	logMu.Unlock()
}

func (a AppLog) Info(step Step, msg string)  { a.log(step, slog.LevelInfo, msg) }
func (a AppLog) Warn(step Step, msg string)  { a.log(step, slog.LevelWarn, msg) }
func (a AppLog) Error(step Step, msg string) { a.log(step, slog.LevelError, msg) }

func (a AppLog) LogForce(step Step, msg string) {
	slog.Warn(fmt.Sprintf("[FORCE] %s %s", step.Name, msg))
}

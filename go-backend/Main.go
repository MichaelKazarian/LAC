package main

import (
	"log"
	"sync"
	"time"
)

type ControlMode int

const (
    ModeManual    ControlMode = iota // 0: Ручний режим
    ModeSingle                       // 1: Одиночний цикл
    ModeAutomatic                    // 2: Автоматичний режим
)

// HardwareState - спільне сховище даних
type HardwareState struct {
	mu               sync.RWMutex
	SensorValue      uint16        `json:"sensor_value"`
	Device10In       [32]uint16    `json:"device10_in"`
  Device20Out      [32]uint16    `json:"-"` // "-" ігнорувати при маршалінгу
  Mode             ControlMode   `json:"mode"`
	LastUpdate       time.Time     `json:"last_update"`
	IsOnline3        bool          `json:"is_online_3"`
	IsOnline10       bool          `json:"is_online_10"`
	IsOnline20       bool          `json:"is_online_20"`
  IsSafetyLocked   bool          `json:"is_safety_locked"`
	ReadCycleMs      int64         `json:"read_cycle_ms"`
  IsPaused         bool          `json:"is_paused"`
  StopReason      string         `json:"stop_reason"`
  ActiveOperation  string        `json:"active_operation"`
  OpsList          [][]string    `json:"-"`
  Counter          int           `json:"counter"`
}

func runWebServer(state *HardwareState, controller *Controller) {
	webServer, err := NewWebServer(state, controller)
	if err != nil {
		log.Fatalf("❌ Помилка створення веб-сервера: %v", err)
	}

	if err := webServer.Start("localhost:8080"); err != nil {
		log.Fatalf("❌ Помилка запуску веб-сервера: %v", err)
	}
}

func main() {
    // 1. Спільний стан
    state := &HardwareState{
        Mode:ModeManual,
    }

    // 2. Створюємо низькорівневий сервіс (але не запускаємо опитування)
    hwService := NewModbusService("/dev/ttyUSB0", 38400)
    // defer тут не спрацює для hwService, якщо ми передаємо його в горутину, 
    // тому краще закривати його всередині runModbusPoll або через сигнал завершення системи.

    // 3. Створюємо контролер (тепер він доступний як для Modbus, так і для Web)
    controller := NewController(hwService, state)

    // 4. Запускаємо опитування Modbus у фоні
    go func() {
      defer hwService.Close()
      go controller.logicWorker()
      controller.Run()
    }()

    // 5. Створюємо Веб-сервер, передаючи йому і стан, і контролер
    runWebServer(state, controller)
}

package main

import (
	"log"
	"sync"
	"time"
)

// HardwareState - спільне сховище даних
type HardwareState struct {
	mu          sync.RWMutex
	SensorValue uint16    `json:"sensor_value"`
	Device10In  [32]uint16 `json:"device10_in"`
	LastUpdate  time.Time `json:"last_update"`
	IsOnline3   bool      `json:"is_online_3"`
	IsOnline10  bool      `json:"is_online_10"`
	IsOnline20  bool      `json:"is_online_20"`
	ReadCycleMs int64     `json:"read_cycle_ms"`
}

// runModbusPoll запускає контролер із Modbus сервісом
func runModbusPoll(state *HardwareState) {
	// Створюємо низькорівневий Modbus сервіс
	var hwService HardwareService = NewModbusService("/dev/ttyUSB0", 38400)

	// Гарантуємо закриття порту
	defer hwService.Close()

	// Створюємо контролер, який керує логікою
	controller := NewController(hwService, state)

	// Запускаємо цикл (тут ми не використовуємо go, щоб defer hwService.Close() спрацював коректно)
	controller.Run()
}

func runWebServer(state *HardwareState) {
	webServer, err := NewWebServer(state)
	if err != nil {
		log.Fatalf("❌ Помилка створення веб-сервера: %v", err)
	}

	if err := webServer.Start("localhost:8080"); err != nil {
		log.Fatalf("❌ Помилка запуску веб-сервера: %v", err)
	}
}

func main() {
	state := &HardwareState{}
	go runModbusPoll(state)
	runWebServer(state)
}

package main

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// HardwareState - спільне сховище даних
type HardwareState struct {
	mu          sync.RWMutex
	SensorValue uint16    `json:"sensor_value"` // Дані з Slave 3
	Device10In  [32]uint16 `json:"device10_in"`  // Дані з Slave 10 (32 регістри)
	LastUpdate  time.Time `json:"last_update"`
	IsOnline3   bool      `json:"is_online_3"`  // Стан Slave 3
	IsOnline10  bool      `json:"is_online_10"` // Стан Slave 10
  IsOnline20  bool      `json:"is_online_20"` // Стан Slave 20
  ReadCycleMs int64     `json:"read_cycle_ms"`
}

// ShowResults друкує масив та отримані регістри для перевірки
func ShowResults(data [32]uint16, r0, r1 uint16) {
	fmt.Printf("Reg0: %016b | Reg1: %016b\n", r0, r1)
	for i := 0; i < 32; i++ {
		fmt.Printf("%d:%d ", i, data[i])
		if (i+1)%16 == 0 {
			fmt.Println()
		}
	}
}

// UpdateSlave3 безпечно оновлює дані першого пристрою
func (s *HardwareState) UpdateSlave3(val uint16, online bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.SensorValue = val
	s.IsOnline3 = online
	s.LastUpdate = time.Now()
}

// UpdateSlave10 безпечно оновлює масив даних другого пристрою
func (s *HardwareState) UpdateSlave10(data *[32]uint16, online bool) {
  s.mu.Lock()
  defer s.mu.Unlock()
  s.IsOnline10 = online
  // Якщо пристрій онлайн і дані передані (data не nil)
  if online && data != nil {
    // Копіюємо вміст масиву за вказівником у стан системи
    s.Device10In = *data
  }
  s.LastUpdate = time.Now()
}

// Додамо окремий метод для оновлення часу циклу
func (s *HardwareState) UpdateCycleTime(ms int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ReadCycleMs = ms
}

// runModbusPoll запускає нескінченне фонове опитування Modbus
func runModbusPoll(state *HardwareState) {
  var hwService HardwareService = NewModbusService("/dev/ttyUSB0", 38400, state)
  
  // Якщо треба буде закрити порт при виході, можна зробити каст до ModbusService або додати Close в інтерфейс
  if ms, ok := hwService.(*ModbusService); ok {
    defer ms.Close()
  }
  go hwService.Run()
}

// runWebServer відповідає за HTTP інтерфейс та API
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

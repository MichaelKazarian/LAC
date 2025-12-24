package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/goburrow/modbus"
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

// UpdateSlave3 безпечно оновлює дані першого пристрою
func (s *HardwareState) UpdateSlave3(val uint16, online bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.SensorValue = val
	s.IsOnline3 = online
	s.LastUpdate = time.Now()
}

// UpdateSlave10 безпечно оновлює масив даних другого пристрою
func (s *HardwareState) UpdateSlave10(data []uint16, online bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.IsOnline10 = online
	if online && len(data) == 32 {
		copy(s.Device10In[:], data)
	}
	s.LastUpdate = time.Now()
}

// Додамо окремий метод для оновлення часу циклу
func (s *HardwareState) UpdateCycleTime(ms int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ReadCycleMs = ms
}

func InitModbus(device string, baud int, slaveID byte) (modbus.Client, *modbus.RTUClientHandler, error) {
	handler := modbus.NewRTUClientHandler(device)
	handler.BaudRate = baud
	handler.DataBits = 8
	handler.Parity = "N"
	handler.StopBits = 1
	handler.SlaveId = slaveID
	handler.Timeout = 1 * time.Second

	err := handler.Connect()
	if err != nil {
		return nil, nil, err
	}
	return modbus.NewClient(handler), handler, nil
}

func startServer(state *HardwareState) {
	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		state.mu.RLock()
		defer state.mu.RUnlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(state)
	})
	fmt.Println("Сервер запущено на http://localhost:8080/status")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func main() {
	client, handler, err := InitModbus("/dev/ttyUSB0", 9600, 3)
	if err != nil {
		log.Fatalf("Помилка порту: %v", err)
	}
	defer handler.Close()

	state := &HardwareState{}
	go startServer(state)

  // Буфер для порівняння змін
  var lastSyncedData [32]uint16 // Зберігає те, що ми ОСТАННІЙ РАЗ успішно записали в Slave 20
  firstRun := true

	for {
		// Фіксуємо час початку циклу
		start := time.Now()

		// --- Опитування Slave 3 ---
		handler.SlaveId = 3
		res3, err3 := client.ReadHoldingRegisters(0, 1)
		if err3 == nil {
			val := uint16(res3[0])<<8 | uint16(res3[1])
			state.UpdateSlave3(val, true)
		} else {
			state.UpdateSlave3(0, false)
		}

		time.Sleep(10 * time.Millisecond) // Міжкадровий інтервал

		// --- Опитування Slave 10 ---
		handler.SlaveId = 10
		res10, err10 := client.ReadHoldingRegisters(0, 32)

    if err10 == nil {
      // Створюємо слайс для поточних даних
      currentData := make([]uint16, 32)
      for i := 0; i < 32; i++ {
        currentData[i] = uint16(res10[i*2])<<8 | uint16(res10[i*2+1])
      }
      
      // Оновлюємо спільний стан (для JSON)
      state.UpdateSlave10(currentData, true)

      // 2. СИНХРОНІЗАЦІЯ ЗІ SLAVE 20 (запис у циклі)
      handler.SlaveId = 20
      for i := 0; i < 32; i++ {
        // Пишемо, якщо значення змінилося або це перший запуск
        if currentData[i] != lastSyncedData[i] || firstRun {
          time.Sleep(10 * time.Millisecond) // Пауза для Arduino
          _, err20 := client.WriteSingleRegister(uint16(i), currentData[i])
          
          if err20 == nil {
            lastSyncedData[i] = currentData[i]
            fmt.Printf("[%s] Регістр %d -> %d (OK)\n", time.Now().Format("15:04:05"), i, currentData[i])
          } else {
            fmt.Printf("Помилка запису регістра %d: %v\n", i, err20)
          }
        }
      }
      firstRun = false
    } else {
      state.UpdateSlave10(nil, false)
    }
    
		// Рахуємо скільки пройшло часу від start до завершення всіх читань
		duration := time.Since(start).Milliseconds()
		state.UpdateCycleTime(duration)

		// Пауза перед наступним колом
		time.Sleep(10 * time.Millisecond)
	}
}

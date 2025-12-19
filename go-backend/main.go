package main

import (
	"fmt"
	"log"
	// "os"
  "encoding/json"
  "net/http"
	"time"
  "sync"
	"github.com/goburrow/modbus"
)

type HardwareState struct {
	mu          sync.RWMutex
	SensorValue uint16    `json:"sensor_value"` // Теги для JSON
	LastUpdate  time.Time `json:"last_update"`
	IsOnline    bool      `json:"is_online"`
}

// Метод для безпечного оновлення даних (використовує Controller)
func (s *HardwareState) Update(val uint16, online bool) {
    s.mu.Lock()         // Блокуємо на запис
    defer s.mu.Unlock()
    s.SensorValue = val
    s.IsOnline = online
    s.LastUpdate = time.Now()
}

// Метод для безпечного читання (використовує UserUI)
func (s *HardwareState) GetSnapshot() (uint16, bool) {
    s.mu.RLock()        // Блокуємо тільки для читання (RLock)
    defer s.mu.RUnlock()
    return s.SensorValue, s.IsOnline
}

// InitModbus створює та підключає RTU обробник
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

	client := modbus.NewClient(handler)
	return client, handler, nil
}

// startServer запускає HTTP сервер
func startServer(state *HardwareState) {
	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		state.mu.RLock() // Читаємо безпечно
		defer state.mu.RUnlock()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(state)
	})

	fmt.Println("Сервер UserUI запущено на http://localhost:8080/status")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func main() {
    client, handler, err := InitModbus("/dev/ttyUSB0", 9600, 3)
    if err != nil {
        log.Fatalf("Помилка зв'язку: %v", err)
    }
    defer handler.Close()

    // Створюємо екземпляр спільного буфера
    state := &HardwareState{}

  fmt.Println("Контролер запущено. Запис у HardwareState...")
  // ЗАПУСКАЄМО СЕРВЕР В ОКРЕМІЙ ГОРУТИНІ
	go startServer(state)

    for {
        results, err := client.ReadHoldingRegisters(0, 1)
        
        if err != nil {
            // Оновлюємо стан: зв'язок втрачено
            state.Update(0, false)
            fmt.Printf("[%s] Error: %v\n", time.Now().Format("15:04:05"), err)
        } else {
            value := uint16(results[0])<<8 | uint16(results[1])
            
            // Записуємо в спільний буфер
            state.Update(value, true)
            
            // Для перевірки виведемо те, що реально лежить в буфері
            //v, ok := state.GetSnapshot()
            //fmt.Printf("[%s] State: Value=%d Online=%v\n", 
            //    time.Now().Format("15:04:05"), v, ok)
        }

        time.Sleep(20 * time.Millisecond)
    }
}

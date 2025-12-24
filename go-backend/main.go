package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
  "encoding/binary"
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

// Pack32 перетворює масив із 32 значень у два регістри uint16
func Pack32(data *[32]uint16) (uint16, uint16) {
	var reg0, reg1 uint16

	for i := 0; i < 16; i++ {
		if data[i] > 0 {
			reg0 |= (1 << i)
		}
		if data[i+16] > 0 {
			reg1 |= (1 << i)
		}
	}
	return reg0, reg1
}

// Unpack32 розпаковує два регістри uint16 назад у масив із 32 елементів
func Unpack32(reg0, reg1 uint16) [32]uint16 {
	var data [32]uint16
	for i := 0; i < 16; i++ {
		data[i] = (reg0 >> i) & 1
		data[i+16] = (reg1 >> i) & 1
	}
	return data
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
  firstRun := true
  var lastReg0 uint16
  var lastReg1 uint16

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
    var currentData [32]uint16

    if err10 == nil {
      for i := 0; i < 32; i++ {
        currentData[i] = uint16(res10[i * 2]) << 8 | uint16(res10[i * 2 + 1])
      }

      // Оновлюємо спільний стан (для JSON API)
      state.UpdateSlave10(&currentData, true)

      // 2. СИНХРОНІЗАЦІЯ ЗІ SLAVE 20 (стиснутий запис)
      handler.SlaveId = 20
      r0, r1 := Pack32(&currentData)

      // Перевіряємо, чи змінилося хоча б щось у запакованих даних
      if r0 != lastReg0 || r1 != lastReg1 || firstRun {
        // Формуємо слайс із двох регістрів
        packedBytes := make([]byte, 4)
        // Записуємо r0 (адреса 0)
        binary.BigEndian.PutUint16(packedBytes[0:2], r0)
        // Записуємо r1 (адреса 1)
        binary.BigEndian.PutUint16(packedBytes[2:4], r1)
    
        // Тепер передаємо []byte. Другий аргумент (2) - це кількість РЕГІСТРІВ
        time.Sleep(20 * time.Millisecond)
        _, err20 := client.WriteMultipleRegisters(0, 2, packedBytes)

        if err20 == nil {
          lastReg0 = r0
          lastReg1 = r1
          fmt.Printf("[%s] Slave 20 оновлено: Reg0=%04X, Reg1=%04X (OK)\n", 
            time.Now().Format("15:04:05"), r0, r1)
        } else {
          fmt.Printf("[%s] Помилка запису Slave 20: %v\n", time.Now().Format("15:04:05"), err20)
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

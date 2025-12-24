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
	client, handler, err := InitModbus("/dev/ttyUSB0", 38400, 3)
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
		// handler.SlaveId = 3
		// res3, err3 := client.ReadHoldingRegisters(0, 1)
		// if err3 == nil {
		// 	val := uint16(res3[0])<<8 | uint16(res3[1])
		// 	state.UpdateSlave3(val, true)
		// } else {
		// 	state.UpdateSlave3(0, false)
		// }

		// time.Sleep(10 * time.Millisecond) // Міжкадровий інтервал

    state.UpdateSlave3(0, true)
    
    // --- Опитування Slave 10 ---
    handler.SlaveId = 10
    // Читаємо лише 2 регістри (4 байти), які містять 32 біти стану
    res10, err10 := client.ReadHoldingRegisters(0, 2)

    var currentData [32]uint16

    if err10 == nil && len(res10) == 4 {
      // 1. Отримуємо два uint16 із отриманих байтів
      r0_in := binary.BigEndian.Uint16(res10[0:2])
      r1_in := binary.BigEndian.Uint16(res10[2:4])

      // 2. Розпаковуємо біти у масив [32]uint16
      // Проходимо по 16 біт для кожного регістру
      for i := 0; i < 16; i++ {
        // Перевіряємо i-й біт у першому регістрі
        if (r0_in & (1 << uint(i))) != 0 {
          currentData[i] = 1
        } else {
          currentData[i] = 0
        }
        // Перевіряємо i-й біт у другому регістрі (зміщення +16)
        if (r1_in & (1 << uint(i))) != 0 {
          currentData[i+16] = 1
        } else {
          currentData[i+16] = 0
        }
      }

      // Оновлюємо спільний стан (для JSON API)
      state.UpdateSlave10(&currentData, true)

      // 3. СИНХРОНІЗАЦІЯ ЗІ SLAVE 20 (стиснутий запис)
      handler.SlaveId = 20
      r0_out, r1_out := Pack32(&currentData)

      // Перевіряємо, чи змінилося хоча б щось у запакованих даних
      if r0_out != lastReg0 || r1_out != lastReg1 || firstRun {
        packedBytes := make([]byte, 4)
        binary.BigEndian.PutUint16(packedBytes[0:2], r0_out)
        binary.BigEndian.PutUint16(packedBytes[2:4], r1_out)
        time.Sleep(10 * time.Millisecond)
        _, err20 := client.WriteMultipleRegisters(0, 2, packedBytes)

        if err20 == nil {
          lastReg0 = r0_out
          lastReg1 = r1_out
          fmt.Printf("[%s] Slave 20 оновлено: Reg0=%04X, Reg1=%04X (OK)\n", 
            time.Now().Format("15:04:05"), r0_out, r1_out)
        } else {
          fmt.Printf("[%s] Помилка запису Slave 20: %v\n", time.Now().Format("15:04:05"), err20)
        }
      }
      firstRun = false
    } else {
      if err10 != nil {
        fmt.Printf("[%s] Помилка опитування Slave 10: %v\n", time.Now().Format("15:04:05"), err10)
      }
      state.UpdateSlave10(nil, false)
    }
		// Рахуємо скільки пройшло часу від start до завершення всіх читань
		duration := time.Since(start).Milliseconds()
		state.UpdateCycleTime(duration)

		// Пауза перед наступним колом
		time.Sleep(10 * time.Millisecond)
	}
}

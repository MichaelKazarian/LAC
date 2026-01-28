package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
  "html/template"
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

// runModbusPoll містить нескінченний цикл роботи з регістрами
func runModbusPoll(state *HardwareState, client modbus.Client, handler *modbus.RTUClientHandler) {
	firstRun := true
	var lastReg0, lastReg1 uint16
	fmt.Println("🚀 Цикл опитування Modbus запущено")
	
	for {
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
		time.Sleep(2 * time.Millisecond)
		
		// --- Опитування Slave 10 ---
		handler.SlaveId = 10
		res10, err10 := client.ReadHoldingRegisters(0, 2)
		
		if err10 == nil && len(res10) == 4 {
			r0_in := binary.BigEndian.Uint16(res10[0:2])
			r1_in := binary.BigEndian.Uint16(res10[2:4])
			currentData := Unpack32(r0_in, r1_in)
			
			// Оновлюємо стан для API
			state.UpdateSlave10(&currentData, true)
			
			// --- Синхронізація зі Slave 20 ---
			r0_out, r1_out := Pack32(&currentData)
			
			if r0_out != lastReg0 || r1_out != lastReg1 || firstRun {
				handler.SlaveId = 20
				packedBytes := make([]byte, 4)
				binary.BigEndian.PutUint16(packedBytes[0:2], r0_out)
				binary.BigEndian.PutUint16(packedBytes[2:4], r1_out)
				
				time.Sleep(2 * time.Millisecond)
				_, err20 := client.WriteMultipleRegisters(0, 2, packedBytes)
				if err20 == nil {
					lastReg0, lastReg1 = r0_out, r1_out
					firstRun = false
					fmt.Printf("[%s] Slave 20 OK: Reg0=%04X, Reg1=%04X\n", 
						time.Now().Format("15:04:05"), r0_out, r1_out)
				} else {
					fmt.Printf("[%s] Slave 20 Error: %v\n", 
						time.Now().Format("15:04:05"), err20)
				}
			}
		} else {
			if err10 != nil {
				fmt.Printf("[%s] Slave 10 Error: %v\n", 
					time.Now().Format("15:04:05"), err10)
			}
			state.UpdateSlave10(nil, false)
		}

    // --- ОНОВЛЕННЯ СЕРВІСНИХ ДАНИХ ---
		duration := time.Since(start).Milliseconds()
    state.UpdateCycleTime(duration)
		time.Sleep(2 * time.Millisecond)
	}
}

// runWebServer відповідає за HTTP інтерфейс та API
func runWebServer(state *HardwareState) {
    // 1. Карта функцій для шаблонів
    funcMap := template.FuncMap{
        "seq": func(start, end int) []int {
            var res []int
            for i := start; i <= end; i++ {
                res = append(res, i)
            }
            return res
        },
    }

    // 2. ЗАВАНТАЖЕННЯ ТЕМПЛЕЙТІВ ПРИ СТАРТІ
    // Створюємо базовий шаблон і парсимо всі файли відразу
    tmpl := template.New("index.html").Funcs(funcMap)
    
    // Парсимо сторінки
    var err error
    tmpl, err = tmpl.ParseGlob("../../webapp/templates/pages/*.html")
    if err != nil {
        log.Fatalf("❌ Критична помилка шаблонів сторінок: %v", err)
    }
    
    // Додаємо фрагменти (partials)
    _, err = tmpl.ParseGlob("../../webapp/templates/partials/*.html")
    if err != nil {
        log.Printf("⚠️ Попередження: partials не знайдено або помилка: %v", err)
    }

    // 3. Роздача статики
    fs := http.FileServer(http.Dir("../../webapp/static"))
    http.Handle("/static/", http.StripPrefix("/static/", fs))

    // 4. Головна сторінка (тепер працює миттєво)
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        opNames := map[int]string{
            1: "Одиничний цикл", 2: "Подача", 3: "Мотор шпінделя",
        }

        data := map[string]interface{}{
            "modes": []map[string]interface{}{
                {"id": "mode-auto", "name": "АВТОМАТ", "class": "btn-outline-success"},
                {"id": "mode-once-cycle", "name": "ОДИН ЦИКЛ", "class": "btn-outline-primary"},
                {"id": "mode-manual", "name": "РУЧНИЙ", "class": "btn-outline-secondary"},
            },
            "opNames": opNames,
        }

        // Використовуємо вже скомпільований tmpl
        // Якщо у вас кілька сторінок, використовуйте tmpl.ExecuteTemplate(w, "ім'я_файлу.html", data)
        err = tmpl.Execute(w, data)
        if err != nil {
            log.Printf("❌ Помилка виконання шаблону: %v", err)
            http.Error(w, "Internal Server Error", 500)
        }
    })

    // 5. API Стан (/state)
    http.HandleFunc("/state", func(w http.ResponseWriter, r *http.Request) {
        state.mu.RLock()
        defer state.mu.RUnlock()

        response := map[string]interface{}{
            "modeId":           "mode-manual",
            "modeState":        "ok",
            "modeDescription":  "Система в нормі",
            "operationState":   "idle",
            "quantity":         state.SensorValue,
            "degree":           int(state.SensorValue) % 720,
            "manualOperations": []string{"operation1", "operation2", "operation3", "operation9", "operation10"},
        }

        for i := 0; i < 18; i++ {
            key := fmt.Sprintf("operation%d", i+1)
            val := 1
            if i < len(state.Device10In) {
                val = int(state.Device10In[i])
            }
            response[key] = val
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(response)
    })

  // 5. Обробка команд (/radio, /modeset, /stop)
  http.HandleFunc("/radio", func(w http.ResponseWriter, r *http.Request) {
    id := r.URL.Query().Get("id")
    // Повертаємо логування, щоб бачити натискання кнопок
    log.Printf("🕹 [Web Command] Натиснуто операцію: %s", id)
    
    // Тут в майбутньому можна додати запис у Modbus Slave 20 
    // або зміну внутрішнього стану системи
    
    w.WriteHeader(http.StatusOK)
  })

  http.HandleFunc("/modeset", func(w http.ResponseWriter, r *http.Request) {
    mode := r.URL.Query().Get("id")
    log.Printf("🔄 [Web Command] Зміна режиму на: %s", mode)
    w.WriteHeader(http.StatusOK)
  })

  http.HandleFunc("/stop", func(w http.ResponseWriter, r *http.Request) {
    log.Println("🛑 [Web Command] EMERGENCY STOP TRIGGERED")
    w.WriteHeader(http.StatusOK)
  })

  http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		state.mu.RLock()
		defer state.mu.RUnlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(state)
	})
  
  fmt.Println("🌐 Веб-інтерфейс на http://localhost:8080")
  log.Fatal(http.ListenAndServe(":8080", nil))
}

func InitModbus(device string, baud int, slaveID byte) (modbus.Client, *modbus.RTUClientHandler, error) {
	handler := modbus.NewRTUClientHandler(device)
	handler.BaudRate = baud
	handler.DataBits = 8
	handler.Parity = "N"
	handler.StopBits = 1
	handler.SlaveId = slaveID
	handler.Timeout = 200 * time.Millisecond

	err := handler.Connect()
	if err != nil {
		return nil, nil, err
	}
	return modbus.NewClient(handler), handler, nil
}

// setupHardware відповідає за підключення до Modbus
func setupHardware() (modbus.Client, *modbus.RTUClientHandler) {
	client, handler, err := InitModbus("/dev/ttyUSB0", 38400, 3)
	if err != nil {
		log.Fatalf("❌ Помилка ініціалізації порту: %v", err)
	}
	fmt.Println("✅ Modbus порт успішно відкрито")
	return client, handler
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
	// 1. Ініціалізуємо залізо
	client, handler := setupHardware()
	defer handler.Close()

	// 2. Створюємо сховище стану
	state := &HardwareState{}

	// 3. Запускаємо фонове опитування Modbus
	go runModbusPoll(state, client, handler)

	// 4. Стартуємо веб-сервер (блокуючий виклик)
	runWebServer(state)
}

package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
)

// WebServer інкапсулює всю логіку веб-сервера
type WebServer struct {
	state *HardwareState
	tmpl  *template.Template
  mux   *http.ServeMux
  ctrl  *Controller
}

// NewWebServer створює новий екземпляр веб-сервера
func NewWebServer(state *HardwareState, ctrl *Controller) (*WebServer, error) {
	ws := &WebServer{
		state: state,
    ctrl:  ctrl,
	}
	
	if err := ws.initTemplates(); err != nil {
		return nil, fmt.Errorf("помилка ініціалізації шаблонів: %w", err)
	}
	
	return ws, nil
}

// initTemplates завантажує та компілює всі шаблони
func (ws *WebServer) initTemplates() error {
	funcMap := template.FuncMap{
		"seq": func(start, end int) []int {
			var res []int
			for i := start; i <= end; i++ {
				res = append(res, i)
			}
			return res
		},
	}
	
	tmpl := template.New("index.html").Funcs(funcMap)
	
	var err error
	tmpl, err = tmpl.ParseGlob("../../webapp/templates/pages/*.html")
	if err != nil {
		return fmt.Errorf("критична помилка шаблонів сторінок: %w", err)
	}
	
	_, err = tmpl.ParseGlob("../../webapp/templates/partials/*.html")
	if err != nil {
		log.Printf("⚠️ Попередження: partials не знайдено або помилка: %v", err)
	}
	
	ws.tmpl = tmpl
	return nil
}

// setupRoutes налаштовує всі маршрути HTTP
func (ws *WebServer) setupRoutes() {
  ws.mux = http.NewServeMux()

	// Статичні файли
	fs := http.FileServer(http.Dir("../../webapp/static"))
	ws.mux.Handle("/static/", http.StripPrefix("/static/", fs))
	
	// Сторінки
	ws.mux.HandleFunc("/", ws.handleIndex)
	
	// API endpoints
	ws.mux.HandleFunc("/state", ws.handleState)
	ws.mux.HandleFunc("/status", ws.handleStatus)
	
	// Команди
	ws.mux.HandleFunc("/radio", ws.handleManualOp)
	ws.mux.HandleFunc("/modeset", ws.handleModeSet)
	ws.mux.HandleFunc("/stop", ws.handleStop)
}

// handleIndex обробляє головну сторінку
func (ws *WebServer) handleIndex(w http.ResponseWriter, r *http.Request) {
	opNames := map[int]string{
		1: "Одиничний цикл", 
		2: "Подача", 
		3: "Мотор шпінделя",
	}
	
	data := map[string]interface{}{
		"modes": []map[string]interface{}{
			{"id": "mode-auto", "name": "АВТОМАТ", "class": "btn-outline-success"},
			{"id": "mode-once-cycle", "name": "ОДИН ЦИКЛ", "class": "btn-outline-primary"},
			{"id": "mode-manual", "name": "РУЧНИЙ", "class": "btn-outline-secondary"},
		},
		"opNames": opNames,
	}
	
	err := ws.tmpl.Execute(w, data)
	if err != nil {
		log.Printf("❌ Помилка виконання шаблону: %v", err)
		http.Error(w, "Internal Server Error", 500)
	}
}

// handleState повертає поточний стан системи для UI
func (ws *WebServer) handleState(w http.ResponseWriter, r *http.Request) {
	ws.state.mu.RLock()
	defer ws.state.mu.RUnlock()

  // Конвертуємо число режиму назад у рядок для JS
  modeStr := "mode-manual"
  switch ws.state.Mode {
  case ModeAutomatic: modeStr = "mode-auto"
  case ModeSingle:    modeStr = "mode-once-cycle"
  }

	response := map[string]interface{}{
		"modeId":          modeStr,
		"modeState":       "ok",
		"modeDescription": "Система в нормі",
		"operationState":  "idle",
		"quantity":        ws.state.SensorValue,
		"degree":          int(ws.state.SensorValue) % 720,
		"manualOperations": []string{"operation1", "operation2", "operation3", "operation9", "operation10"},
	}
	
	for i := 0; i < 18; i++ {
		key := fmt.Sprintf("operation%d", i+1)
		val := 1
		if i < len(ws.state.Device10In) {
			val = int(ws.state.Device10In[i])
		}
		response[key] = val
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleStatus повертає технічний статус системи
func (ws *WebServer) handleStatus(w http.ResponseWriter, r *http.Request) {
	ws.state.mu.RLock()
	defer ws.state.mu.RUnlock()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ws.state)
}

// handleRadio обробляє команди операцій
func (ws *WebServer) handleRadio(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	log.Printf("🕹 [Web Command] Натиснуто операцію: %s", id)
	
	// TODO: Додати логіку запису в Modbus Slave 20
	
	w.WriteHeader(http.StatusOK)
}

// handleModeSet обробляє зміну режиму через API
func (ws *WebServer) handleModeSet(w http.ResponseWriter, r *http.Request) {
    modeID := r.URL.Query().Get("id")
    log.Printf("🔄 [Web Command] Запит на зміну режиму: %s", modeID)

    var targetMode ControlMode

    switch modeID {
    case "mode-auto":
        targetMode = ModeAutomatic
    case "mode-once-cycle":
        targetMode = ModeSingle
    case "mode-manual":
        targetMode = ModeManual
    default:
        log.Printf("⚠️ Невідомий режим: %s", modeID)
        http.Error(w, "Invalid mode", http.StatusBadRequest)
        return
    }

    // Викликаємо метод контролера для безпечної зміни режиму
    ws.ctrl.SetMode(targetMode)
    w.WriteHeader(http.StatusOK)
}

// handleStop обробляє аварійну зупинку
func (ws *WebServer) handleStop(w http.ResponseWriter, r *http.Request) {
	log.Println("🛑 [Web Command] EMERGENCY STOP TRIGGERED")
	
	// TODO: Додати логіку аварійної зупинки
	
	w.WriteHeader(http.StatusOK)
}

// Start запускає веб-сервер (блокуючий виклик)
func (ws *WebServer) Start(addr string) error {
	ws.setupRoutes()
	fmt.Printf("🌐 Веб-інтерфейс на http://%s\n", addr)
	return http.ListenAndServe(addr, ws.mux)
}

// handleManualOp обробляє виклик будь-якої зареєстрованої операції
func (ws *WebServer) handleManualOp(w http.ResponseWriter, r *http.Request) {
    // Отримуємо id операції (наприклад, "operation10" або "sync_mirror")
    opID := r.URL.Query().Get("id")
    if opID == "" {
        http.Error(w, "Missing operation id", http.StatusBadRequest)
        return
    }

    log.Printf("🕹 [Web Command] Виконання операції: %s", opID)

    // Викликаємо метод контролера
    err := ws.ctrl.InvokeOperation(opID)
    if err != nil {
        log.Printf("⚠️ Операція %s не вдалася: %v", opID, err)
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }

    w.WriteHeader(http.StatusOK)
}

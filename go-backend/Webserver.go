package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os/exec"
	"time"
)

// WebServer інкапсулює всю логіку веб-сервера
type WebServer struct {
	tmpl  *template.Template
  mux   *http.ServeMux
  ctrl  *Controller
}

// NewWebServer створює новий екземпляр веб-сервера
func NewWebServer(ctrl *Controller) (*WebServer, error) {
	ws := &WebServer{
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

  // 1. Статичні файли (пріоритет)
  fs := http.FileServer(http.Dir("../../webapp/static"))
  ws.mux.Handle("/static/", http.StripPrefix("/static/", fs))
  
  // 2. API Endpoints (чіткі шляхи)
  ws.mux.HandleFunc("/state", ws.handleState)
  ws.mux.HandleFunc("/api/status", ws.handleStatusAPI) // JSON дані
  
  // 3. Команди
  ws.mux.HandleFunc("/radio", ws.handleManualOp)
  ws.mux.HandleFunc("/modeset", ws.handleModeSet)
  ws.mux.HandleFunc("/pause", ws.handlePause)
  ws.mux.HandleFunc("/safety", ws.handleEmergencyStop)
  ws.mux.HandleFunc("/api/io-map", ws.handleIOMap)
  ws.mux.HandleFunc("/logout", ws.handleLogout)
  ws.mux.HandleFunc("/shutdown", ws.handleShutdown)

  // 4. Сторінки
  ws.mux.HandleFunc("/status", ws.handleStatusPage)   // HTML сторінка
  
  // 5. Головна сторінка - реєструємо ОСТАННЬОЮ
  ws.mux.HandleFunc("/", ws.handleIndex)
  }

// handleIndex обробляє головну сторінку
func (ws *WebServer) handleIndex(w http.ResponseWriter, r *http.Request) {
  // Якщо шлях не рівно "/", то це 404, а не головна сторінка
  if r.URL.Path != "/" {
    http.NotFound(w, r)
    return
  }

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
  
  err := ws.tmpl.ExecuteTemplate(w, "index.html", data) // Явно вказуємо назву шаблону
  if err != nil {
    log.Printf("[WEB] Помилка виконання шаблону index: %v", err)
    http.Error(w, "Internal Server Error", 500)
  }
}

// handleState повертає поточний стан системи для UI
func (ws *WebServer) handleState(w http.ResponseWriter, r *http.Request) {
	view := ws.ctrl.GetView() // ← тільки snapshot

	// Конвертуємо режим у рядок для JS
	modeStr := "mode-manual"
	switch view.Mode {
	case ModeAutomatic:
		modeStr = "mode-auto"
	case ModeSingle:
		modeStr = "mode-once-cycle"
	}

	response := map[string]interface{}{
		"modeId":          modeStr,
		"modeState":       "ok",
		"isPaused":        view.IsPaused,
		"isLocked":        view.IsSafetyLocked,
		"stopReason":      view.StopReason,
		"modeDescription": "Система в нормі",
		"operationState":  "idle",
		"counter":         view.Counter,
		"OperationsList":  view.OpsList,
		"ActiveOperation": view.ActiveOperation,
		"degree":          int(view.EncoderValue) % 720,
		"manualOperations": GetAllowedManualOpsFromView(view),
	}

	for i := 0; i < 18; i++ {
		key := fmt.Sprintf("operation%d", i+1)
		val := int(view.Device10In[i])
		response[key] = val
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Обробник для сторінки
func (ws *WebServer) handleStatusPage(w http.ResponseWriter, r *http.Request) {
	// Передаємо порожні дані або загальну інфо, якщо треба
	err := ws.tmpl.ExecuteTemplate(w, "status.html", nil)
	if err != nil {
		log.Printf("❌ Помилка виконання шаблону status.html: %v", err)
		http.Error(w, "Internal Server Error", 500)
	}
}

// handleStatus повертає технічний статус системи
func (ws *WebServer) handleStatusAPI(w http.ResponseWriter, r *http.Request) {
	view := ws.ctrl.GetView() // snapshot від контролера

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(view); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// handleModeSet обробляє зміну режиму через API
func (ws *WebServer) handleModeSet(w http.ResponseWriter, r *http.Request) {
  modeID := r.URL.Query().Get("id")
  log.Printf("[Web] Запит на зміну режиму: %s", modeID)

  var targetMode ControlMode

  switch modeID {
  case "mode-auto":
    targetMode = ModeAutomatic
  case "mode-once-cycle":
    targetMode = ModeSingle
  case "mode-manual":
    targetMode = ModeManual
  default:
    log.Printf("[Web] Невідомий режим: %s", modeID)
    http.Error(w, "Invalid mode", http.StatusBadRequest)
    return
  }

  // Викликаємо метод контролера для безпечної зміни режиму
  ws.ctrl.SetMode(targetMode)
  w.WriteHeader(http.StatusOK)
}

// func (ws *WebServer) handleEmergencyStopTMP(w http.ResponseWriter, r *http.Request) {
//   ws.state.mu.RLock()
//   isLocked := ws.state.IsSafetyLocked
//   ws.state.mu.RUnlock()
//   if isLocked {
//     log.Println("[Web] Запит на розблокування (Safety Start)")
//     ws.ctrl.Reset()
//   } else {
//     log.Println("[Web] EMERGENCY STOP!")
//     ws.ctrl.Stop("Зупинка оператором")
//   }    
//   w.WriteHeader(http.StatusOK)
// }

func (ws *WebServer) handleEmergencyStop(w http.ResponseWriter, r *http.Request) {
  ws.execEmergencyStop(w)
}

// handlePause обробляє зупинку/продовження логіки
func (ws *WebServer) handlePause(w http.ResponseWriter, r *http.Request) {
	// Отримуємо значення з запиту (наприклад, /pause?set=true)
	val := r.URL.Query().Get("set")
	
	view := ws.ctrl.GetView()
	targetPause := !view.IsPaused // default toggle

	if val == "true" {
		targetPause = true
	} else {
		targetPause = false
	}
  fmt.Printf("[Web] %s\n", targetPause)
	ws.ctrl.SetPause(targetPause)
	
	// Повертаємо новий стан у JSON, щоб фронтенд міг оновити колір кнопки
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"paused": targetPause})
}

func (ws *WebServer) handleIOMap(w http.ResponseWriter, r *http.Request) {
    response := struct {
        In  map[int]string `json:"in"`
        Out map[int]string `json:"out"`
    }{
        In:  PinNamesIn,
        Out: PinNamesOut,
    }

    w.Header().Set("Content-Type", "application/json")
    if err := json.NewEncoder(w).Encode(response); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

// Start запускає веб-сервер (блокуючий виклик)
func (ws *WebServer) Start(addr string) error {
	ws.setupRoutes()
	fmt.Printf("[Web] Веб-інтерфейс на http://%s\n", addr)
	return http.ListenAndServe(addr, ws.mux)
}

// handleManualOp обробляє виклик будь-якої зареєстрованої операції
func (ws *WebServer) handleManualOp(w http.ResponseWriter, r *http.Request) {
    // Отримуємо id операції (наприклад, "operation10" або "sync_mirror")
    opID := r.URL.Query().Get("id")
    if opID == "" {
        http.Error(w, "[Web] Missing operation id", http.StatusBadRequest)
        return
    }

    log.Printf("[Web] Виконання операції: %s", opID)

    // Викликаємо метод контролера
    err := ws.ctrl.InvokeOperation(opID)
    if err != nil {
        log.Printf("[Web]️ Операція %s не вдалася: %v", opID, err)
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }

    w.WriteHeader(http.StatusOK)
}

// execEmergencyStop виконує перемикання блокування та відправляє статус OK
func (ws *WebServer) execEmergencyStop(w http.ResponseWriter) {
    ws.ctrl.ToggleSafetyLock()
    w.WriteHeader(http.StatusOK)
}

// handleShutdown спочатку зупиняє залізо, потім вимикає Raspberry
func (ws *WebServer) handleShutdown(w http.ResponseWriter, r *http.Request) {
    log.Println("[Web] Запит на ВИМКНЕННЯ системи")
    ws.execEmergencyStop(w)

    cmd := exec.Command("sudo", "shutdown", "-h", "now")
    if err := cmd.Start(); err != nil {
        log.Printf("[Web] Помилка виконання shutdown: %v", err)
        return
    }
}

func (ws *WebServer) handleLogout(w http.ResponseWriter, r *http.Request) {
  log.Println("[Web] Запит на вихід: закриття браузера")
  // Відправляємо відповідь клієнту, щоб він не "висів" у очікуванні
  w.WriteHeader(http.StatusOK)
  // Запускаємо завершення в окремій горутині, щоб встигнути закрити HTTP-з'єднання
  go func() {
    // Невелика затримка, щоб браузер отримав статус 200 OK
    time.Sleep(300 * time.Millisecond)
    exec.Command("pkill", "chromium").Run()
  }()
}

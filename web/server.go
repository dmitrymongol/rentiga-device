package web

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gotk3/gotk3/glib"

	"rentiga-device/app"
)

type WebServer struct {
	App *app.App
}

func (ws *WebServer) Start(port string) {
    // Проверка существования index.html
    if _, err := os.Stat("./web/static/index.html"); os.IsNotExist(err) {
        log.Fatal("index.html not found in static directory")
    }

    // Создаем кастомный обработчик файлов с CORS
    fileHandler := http.FileServer(http.Dir("./web/static"))

    // Настройка CORS для всех обработчиков
    corsMiddleware := func(h http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            w.Header().Set("Access-Control-Allow-Origin", "*")
            w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
            w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
            
            if r.Method == "OPTIONS" {
                return
            }
            
            h.ServeHTTP(w, r)
        })
    }

    // Главный обработчик маршрутов
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        // Обработка API-запросов
        if strings.HasPrefix(r.URL.Path, "/api/") {
            switch r.URL.Path {
            case "/api/status":
                ws.handleStatus(w, r)
            case "/api/start":
                ws.handleStart(w, r)
            case "/api/stop":
                ws.handleStop(w, r)
            default:
                http.NotFound(w, r)
            }
            return
        }

        // Для корневого пути отдаем index.html
        if r.URL.Path == "/" {
            http.ServeFile(w, r, "./web/static/index.html")
            return
        }

        // Пытаемся найти запрашиваемый файл
        path := filepath.Join("./web/static", r.URL.Path)
        if _, err := os.Stat(path); os.IsNotExist(err) {
            // Если файл не найден - отдаем index.html
            http.ServeFile(w, r, "./web/static/index.html")
            return
        }

        // Для существующих файлов используем FileServer с CORS
        corsMiddleware(fileHandler).ServeHTTP(w, r)
    })

    log.Printf("Web server started on port %s", port)
    if err := http.ListenAndServe(port, nil); err != nil {
        log.Fatal("Web server failed: ", err)
    }
}

func (ws *WebServer) handleStatus(w http.ResponseWriter, r *http.Request) {
	status := ws.App.GetStatus()
	respondJSON(w, http.StatusOK, status)
}

func (ws *WebServer) handleStart(w http.ResponseWriter, r *http.Request) {
	glib.IdleAdd(func() {
		ws.App.StartStream()
	})
	respondJSON(w, http.StatusOK, map[string]string{"status": "starting"})
}

func (ws *WebServer) handleStop(w http.ResponseWriter, r *http.Request) {
	glib.IdleAdd(func() {
		ws.App.StopStream()
	})
	respondJSON(w, http.StatusOK, map[string]string{"status": "stopping"})
}

func respondJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
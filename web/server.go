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
    App      *app.App
    Username string
    Password string
}

func (ws *WebServer) Start(port string) {
    // Проверка наличия учетных данных
    if ws.Username == "" || ws.Password == "" {
        log.Fatal("Basic auth credentials not configured")
    }

    // Проверка существования index.html
    if _, err := os.Stat("./web/static/index.html"); os.IsNotExist(err) {
        log.Fatal("index.html not found in static directory")
    }

    fileHandler := http.FileServer(http.Dir("./web/static"))

    // Главный обработчик с цепочкой middleware
    mainHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

        // Проверка существования файла
        path := filepath.Join("./web/static", r.URL.Path)
        if _, err := os.Stat(path); os.IsNotExist(err) {
            http.ServeFile(w, r, "./web/static/index.html")
            return
        }

        fileHandler.ServeHTTP(w, r)
    })

    // Цепочка middleware: CORS -> Auth -> Handler
    chain := middlewareChain(
        corsMiddleware,
        ws.basicAuthMiddleware,
    )

    http.Handle("/", chain(mainHandler))

    log.Printf("Web server started on port %s", port)
    if err := http.ListenAndServe(port, nil); err != nil {
        log.Fatal("Web server failed: ", err)
    }
}

// Middleware для Basic Auth
func (ws *WebServer) basicAuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        user, pass, ok := r.BasicAuth()
        
        if !ok || user != ws.Username || pass != ws.Password {
            w.Header().Set("WWW-Authenticate", `Basic realm="Restricted", charset="UTF-8"`)
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }
        
        next.ServeHTTP(w, r)
    })
}

// Middleware для CORS
func corsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
        
        if r.Method == "OPTIONS" {
            return
        }
        
        next.ServeHTTP(w, r)
    })
}

// Вспомогательная функция для объединения middleware
func middlewareChain(middlewares ...func(http.Handler) http.Handler) func(http.Handler) http.Handler {
    return func(final http.Handler) http.Handler {
        for i := len(middlewares) - 1; i >= 0; i-- {
            final = middlewares[i](final)
        }
        return final
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
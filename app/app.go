// app/app.go
package app

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"rentiga-device/certificate"
	"rentiga-device/interfaces"
	"rentiga-device/models"
	"rentiga-device/rabbitmq"
	"rentiga-device/streaming"
	"rentiga-device/ui"
	"sync"
	"time"

	"github.com/gotk3/gotk3/glib"
)

type App struct {
	certManager   *certificate.Manager
	streamer      *streaming.Streamer
	mainWindow    *ui.MainWindow
	config        *models.AppConfig
	stopHeartbeat chan struct{}
	rabbitClient  *rabbitmq.Client
	isStreaming bool
	isConnected bool
	mu          sync.Mutex
}

var _ interfaces.Application = (*App)(nil)

func New(cfg *models.AppConfig, rabbit *rabbitmq.Client) *App {
    return &App{
        certManager: certificate.NewManager(&cfg.Certificate),
        streamer:    streaming.NewStreamer(&cfg.Stream, cfg.Certificate.DeviceID),
        config:      cfg,
		rabbitClient: rabbit,
    }
}

func (a *App) Initialize() {
    // Инициализация UI
    a.mainWindow = ui.NewMainWindow(a)
    
    // Загрузка сертификатов и проверка соединения
    if a.hasSavedCertificate() {
        if err := a.loadSavedCertificate(); err == nil {
            log.Println("Loaded saved certificate successfully")
            a.checkConnection()
        }
    }
}

func (a *App) Start() {
    a.mainWindow.Window.ShowAll()
}

func (a *App) GetConfig() interface{} {
	return a.config
}

func (a *App) HasCertificate() bool {
    return a.certManager != nil && a.certManager.IsLoaded()
}

func (a *App) hasSavedCertificate() bool {
    requiredFiles := []string{"client.crt", "client.key", "ca.crt"}
    certDir := certificate.GetConfigPath()
    
    for _, file := range requiredFiles {
        if _, err := os.Stat(filepath.Join(certDir, file)); os.IsNotExist(err) {
            return false
        }
    }
    return true
}

func (a *App) loadSavedCertificate() error {
    certDir := certificate.GetConfigPath()
    return a.certManager.LoadFromDir(certDir)
}

func (a *App) LoadCertificate(zipPath string) error {
    if err := a.certManager.LoadFromZip(zipPath); err != nil {
        return err
    }
    
    if err := a.certManager.SaveCertificate(); err != nil {
        log.Printf("Failed to save certificate: %v", err)
    }
    
    a.checkConnection()
    a.mainWindow.UpdateCertificateButton()
    return nil
}

func (a *App) checkConnection() {
	deviceID := a.certManager.Config().DeviceID
	if deviceID == "" {
		a.UpdateConnectionStatus(false, "No device ID")
		return
	}

	url := fmt.Sprintf("https://localhost:8443/tls/devices/%s/login", deviceID)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		a.UpdateConnectionStatus(false, "Request creation failed")
		return
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := a.certManager.HTTPClient().Do(req)
	if err != nil {
		a.UpdateConnectionStatus(false, "Connection failed")
		return
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode == http.StatusOK:
		a.UpdateConnectionStatus(true, "Connected: "+deviceID)
		a.isConnected = true;
		go a.startHeartbeat()

	case resp.StatusCode == http.StatusUnauthorized:
		a.UpdateConnectionStatus(false, "Invalid credentials")
	default:
		a.UpdateConnectionStatus(false, "API error: "+resp.Status)
	}
}

func (a *App) startHeartbeat() {
	if a.stopHeartbeat != nil {
		close(a.stopHeartbeat)
	}
	a.stopHeartbeat = make(chan struct{})

	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			a.sendHeartbeat()
		case <-a.stopHeartbeat:
			return
		}
	}
}

func (a *App) sendHeartbeat() {
	deviceID := a.certManager.Config().DeviceID
	url := fmt.Sprintf("https://localhost:8443/tls/devices/%s/heartbeat", deviceID)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		log.Println("Heartbeat error:", err)
		return
	}

	resp, err := a.certManager.HTTPClient().Do(req)
	if err != nil {
		a.UpdateConnectionStatus(false, "Connection lost")
		a.isConnected = false;
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		log.Println("Unexpected heartbeat status:", resp.Status)
	}
}

func (a *App) StartStream() {
    a.mu.Lock()
    defer a.mu.Unlock()
    
    if a.isStreaming {
        return
    }
    
    if err := a.streamer.Start(); err != nil {
        log.Printf("Stream start error: %v", err)
        return
    }
    
    a.isStreaming = true
    a.UpdateConnectionStatus(true, "Streaming started")
}

func (a *App) StopStream() {
    a.mu.Lock()
    defer a.mu.Unlock()
    
    if !a.isStreaming {
        return
    }
    
    a.streamer.Stop()
    a.isStreaming = false
    a.UpdateConnectionStatus(false, "Streaming stopped")
}

func (a *App) UpdateConnectionStatus(connected bool, message string) {
	if a.mainWindow != nil {
		a.mainWindow.UpdateConnectionStatus(connected, message)
	}
}

func (a *App) StartCommandConsumer() {
	msgs, err := a.rabbitClient.Consume("stream_commands")
	if err != nil {
		log.Fatalf("Failed to start consumer: %v", err)
	}

	go func() {
		for msg := range msgs {
			var cmd models.CommandMessage
			if err := json.Unmarshal(msg.Body, &cmd); err != nil {
				log.Printf("Failed to parse command: %v", err)
				continue
			}

			// Проверяем, что команда для этого устройства
			if cmd.DeviceID != a.certManager.Config().DeviceID {
				continue
			}

            glib.IdleAdd(func() {
                switch cmd.Action {
                case "start":
                    a.StartStream()
                case "stop":
                    a.StopStream()
                default:
                    log.Printf("Unknown command action: %s", cmd.Action)
                }
            })
		}
	}()
}

func (a *App) GetStatus() map[string]interface{} {
    return map[string]interface{}{
        "streaming":  a.isStreaming,
        "connected":  a.isConnected,
        "device_id":  a.certManager.Config().DeviceID,
		"has_certificate": a.HasCertificate(),
    }
}
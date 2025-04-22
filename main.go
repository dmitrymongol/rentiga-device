// main.go
package main

import (
	"log"
	"rentiga-device/app"
	"rentiga-device/models"
	"rentiga-device/rabbitmq"
	"rentiga-device/web"

	"github.com/gotk3/gotk3/gtk"
)

func main() {
	cfg := loadConfig()
	
	rabbit := initRabbitMQ(cfg)
	defer rabbit.Close()

	application := app.New(cfg, rabbit)
	webServer := &web.WebServer{
		App:      application,
		Username: cfg.Web.Auth.Username,
		Password: cfg.Web.Auth.Password,
	}

	go webServer.Start(cfg.Web.Port)
	
	go application.StartCommandConsumer()

	// Инициализируем GTK только когда нужно показать окно
	gtk.Init(nil)
	application.Initialize()
	gtk.Main()
}

func loadConfig() *models.AppConfig {
	return &models.AppConfig{
		Certificate: models.CertificateConfig{
			TempDir:  "/tmp/certs",
			CertPath: "/tmp/certs/client.crt",
			KeyPath:  "/tmp/certs/client.key",
			CaPath:   "/tmp/certs/ca.crt",
		},
		Stream: models.StreamConfig{
			Device:     "/dev/video0",
			Resolution: "1920x1080",
			FontPath:   "/usr/share/fonts/...",
			QRPath:     "/tmp/qr.png",
		},
		Web: models.WebConfig{
			Port: ":8888",
			Auth: models.WebAuthConfig{
				Username: "admin",
				Password: "secret",
			},
		},
	}
}

func initRabbitMQ(cfg *models.AppConfig) *rabbitmq.Client {
	rabbit, err := rabbitmq.New("amqp://rabbitmq:vzukAkXJkkOypIpX@rentiga.ru:5672/")
	if err != nil {
		log.Fatal("RabbitMQ connection failed: ", err)
	}
	return rabbit
}
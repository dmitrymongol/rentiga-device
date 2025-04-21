package main

import (
	"log"
	"rentiga-device/app"
	"rentiga-device/models"
	"rentiga-device/rabbitmq"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

func main() {
    gtk.Init(nil)

    cfg := &models.AppConfig{
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
            QRPath:    "/tmp/qr.png",
        },
    }

	// Инициализация RabbitMQ клиента
	rabbit, err := rabbitmq.New("amqp://rabbitmq:vzukAkXJkkOypIpX@rentiga.ru:5672/")
	if err != nil {
		log.Fatal("RabbitMQ connection failed: ", err)
	}
	defer rabbit.Close()

	application := app.New(cfg, rabbit)

	go application.StartCommandConsumer()

    // Запускаем инициализацию в главном потоке GTK
    glib.IdleAdd(func() {
        application.Initialize()
        application.Start()
    })

    gtk.Main()
}
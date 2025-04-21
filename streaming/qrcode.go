package streaming

import (
	"fmt"
	"image/png"
	"os"

	"github.com/skip2/go-qrcode"
)

func GenerateQR(url string, qrPath string, size int) error {
    qr, err := qrcode.New(url, qrcode.Medium)
    if err != nil {
        return fmt.Errorf("ошибка создания QR-кода: %v", err)
    }

    file, err := os.Create(qrPath)
    if err != nil {
        return fmt.Errorf("ошибка создания файла: %v", err)
    }
    defer file.Close()

    if err := png.Encode(file, qr.Image(size)); err != nil {
        return fmt.Errorf("ошибка записи PNG: %v", err)
    }

    // Проверяем существование файла
    if _, err := os.Stat(qrPath); os.IsNotExist(err) {
        return fmt.Errorf("файл QR-кода не был создан")
    }

    return nil
}

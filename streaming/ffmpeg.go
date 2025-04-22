package streaming

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"rentiga-device/models"
)

type Streamer struct {
    config   *models.StreamConfig
    deviceID string
    cmd      *exec.Cmd
}

func NewStreamer(cfg *models.StreamConfig, deviceId string) *Streamer {
    return &Streamer{config: cfg, deviceID: deviceId}
}

func (s *Streamer) Start() error {
    if s.cmd != nil {
        return fmt.Errorf("stream already running")
    }

    deviceUrl := fmt.Sprintf("https://localhost:3000/devices/%s", s.deviceID)
    GenerateQR(deviceUrl, s.config.QRPath, 200)

    args := []string{
        "-f", "v4l2",
        "-input_format", "mjpeg",
        "-video_size", s.config.Resolution,
        "-i", s.config.Device,
        "-i", s.config.QRPath,
        "-filter_complex", fmt.Sprintf(
            "[1]scale=200:-1[qr];" +
            "[0]format=yuv420p[main];" +
            "[main][qr]overlay=50:50," +
            "drawtext=fontfile='%s':" +
            "text='%%{gmtime\\:%%H\\\\:%%M\\\\:%%S}':" +
            "fontcolor=white@0.9:fontsize=40:" +
            "box=1:boxcolor=black@0.5:boxborderw=5:" +
            "x=w-tw-50:y=50",
            s.config.FontPath,
        ),
        "-f", "sdl",        // Используем SDL для вывода
        "Rentiga Stream",   // Заголовок окна
    }

    s.cmd = exec.Command("ffmpeg", args...)
    
    // Устанавливаем переменные окружения
    s.cmd.Env = append(os.Environ(),
        "DISPLAY=:1",                // Указываем X11 display
        "SDL_VIDEODRIVER=x11",       // Форсируем X11 драйвер
    )

    // Перехват вывода для диагностики
    var stderr bytes.Buffer
    s.cmd.Stderr = &stderr

    if err := s.cmd.Start(); err != nil {
        log.Printf("FFmpeg command: %s", strings.Join(args, " "))
        log.Printf("FFmpeg error: %v\n%s", err, stderr.String())
        return fmt.Errorf("failed to start stream: %v", err)
    }

    return nil
}

func (s *Streamer) Stop() {
    if s.cmd != nil {
        s.cmd.Process.Signal(syscall.SIGINT)
        s.cmd.Wait()
        s.cmd = nil
    }
}
package streaming

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"rentiga-device/models"
)

type Streamer struct {
    config  *models.StreamConfig
	deviceID string
    cmd     *exec.Cmd
}

func NewStreamer(cfg *models.StreamConfig, deviceId string) *Streamer {
    return &Streamer{config: cfg, deviceID: deviceId}
}

func (s *Streamer) Start() error {
    if s.cmd != nil {
        return fmt.Errorf("stream already running")
    }

	deviceUrl:= fmt.Sprintf("https://localhost:3000/devices/%s", s.deviceID)
	GenerateQR(deviceUrl, s.config.QRPath, 200)

    s.cmd = exec.Command(
        "ffmpeg",
        "-f", "v4l2",
        "-input_format", "mjpeg",
        "-video_size", s.config.Resolution,
        "-i", s.config.Device,
        "-i", s.config.QRPath,
        "-filter_complex", fmt.Sprintf(
            "[1]scale=200:-1,format=rgba[qr];"+
                "[0][qr]overlay=50:50:format=auto,"+
                "drawtext=fontfile='%s':text='%%{gmtime\\:%%H\\\\\\:%%M\\\\\\:%%S}':"+
                "fontcolor=white@0.9:fontsize=40:box=1:"+
                "boxcolor=black@0.5:boxborderw=5:"+
                "x=w-tw-50:y=50,"+
                "format=yuv420p",
            s.config.FontPath,
        ),
        "-f", "sdl",
        "Video Preview",
    )

    s.cmd.Stderr = os.Stderr
    return s.cmd.Start()
}

func (s *Streamer) Stop() {
    if s.cmd != nil {
        s.cmd.Process.Signal(syscall.SIGINT)
        s.cmd.Wait()
        s.cmd = nil
    }
}
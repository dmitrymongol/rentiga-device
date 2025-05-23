package streaming

import (
	// "bytes"
	"fmt"
	"log"

	// "os"
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

    args := []string{
        "-v",
        "v4l2src", fmt.Sprintf("device=%s", s.config.Device),
        "!", "image/jpeg",
        "!", "jpegparse",
        "!", "vaapijpegdec",
        "!", "queue", "max-size-buffers=3", "leaky=downstream",
        "!", "kmssink",
        fmt.Sprintf("connector-id=%s", s.config.ConnectorID),
        "sync=false",
        "force-modesetting=true",
        "show-preroll-frame=false",
    }

    s.cmd = exec.Command("gst-launch-1.0", args...)

    var stderr strings.Builder
    s.cmd.Stderr = &stderr

    if err := s.cmd.Start(); err != nil {
        log.Printf("GStreamer command: %s", strings.Join(s.cmd.Args, " "))
        log.Printf("GStreamer error: %v\n%s", err, stderr.String())
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
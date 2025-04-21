package models

type AppConfig struct {
    Certificate CertificateConfig
    Stream      StreamConfig
	Web WebConfig
}

type CertificateConfig struct {
    CertPath    string
    KeyPath     string
    CaPath      string
    DeviceID    string
    TempDir     string
}

type StreamConfig struct {
    Device      string
    Resolution  string
    FontPath    string
    TempDir     string
    QRPath      string
}

type CommandMessage struct {
	DeviceID string `json:"device_id"`
	Action   string `json:"action"` // "start" или "stop"
}

type WebConfig struct {
    Port    string
}
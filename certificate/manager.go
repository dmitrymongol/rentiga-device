package certificate

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"rentiga-device/models"

	"github.com/mholt/archiver"
)

type Manager struct {
    config      *models.CertificateConfig
    httpClient  *http.Client
}

const (
    configDir = ".config/rentiga-device"
    certsDir  = "certs"
)

func NewManager(cfg *models.CertificateConfig) *Manager {
    return &Manager{config: cfg}
}

func GetConfigPath() string {
    home, _ := os.UserHomeDir()
    return filepath.Join(home, configDir, certsDir)
}

func (m *Manager) SaveCertificate() error {
    destDir := GetConfigPath()
    if err := os.MkdirAll(destDir, 0700); err != nil {
        return err
    }

    files := map[string]string{
        m.config.CertPath: filepath.Join(destDir, "client.crt"),
        m.config.KeyPath:  filepath.Join(destDir, "client.key"),
        m.config.CaPath:   filepath.Join(destDir, "ca.crt"),
    }

    for src, dest := range files {
        if err := copyFile(src, dest); err != nil {
            return err
        }
    }
    return nil
}

func copyFile(src, dest string) error {
    input, err := os.ReadFile(src)
    if err != nil {
        return err
    }
    return os.WriteFile(dest, input, 0600)
}

func (m *Manager) LoadFromZip(zipPath string) error {
    extractDir := filepath.Join(m.config.TempDir, "extract")
    if err := os.MkdirAll(extractDir, 0755); err != nil {
        return err
    }
    defer os.RemoveAll(extractDir)

    z := archiver.NewZip()
    if err := z.Unarchive(zipPath, extractDir); err != nil {
        return fmt.Errorf("extraction failed: %w", err)
    }

    files := map[string]string{
        "client.crt": m.config.CertPath,
        "client.key": m.config.KeyPath,
        "ca.crt":     m.config.CaPath,
    }

    for src, dest := range files {
        srcPath := filepath.Join(extractDir, src)
        if err := moveFile(srcPath, dest); err != nil {
            return err
        }
    }

    deviceID, err := m.extractDeviceID()
    if err != nil {
        return err
    }
    m.config.DeviceID = deviceID

    return m.initHTTPClient()
}

func (m *Manager) LoadFromDir(certDir string) error {
    // Добавляем логирование
    log.Printf("Loading certificates from: %s", certDir)
    
    files := map[string]string{
        "client.crt": m.config.CertPath,
        "client.key": m.config.KeyPath,
        "ca.crt":     m.config.CaPath,
    }

    for srcName, destPath := range files {
        srcPath := filepath.Join(certDir, srcName)
        log.Printf("Copying %s to %s", srcPath, destPath)
        if err := copyFile(srcPath, destPath); err != nil {
            return fmt.Errorf("failed to load %s: %w", srcName, err)
        }
    }

    deviceID, err := m.extractDeviceID()
    if err != nil {
        return fmt.Errorf("device ID extraction failed: %w", err)
    }
    
    log.Printf("Extracted device ID: %s", deviceID)
    m.config.DeviceID = deviceID

    if err := m.initHTTPClient(); err != nil {
        return fmt.Errorf("HTTP client initialization failed: %w", err)
    }
    
    return nil
}

func (m *Manager) extractDeviceID() (string, error) {
    certData, err := os.ReadFile(m.config.CertPath)
    if err != nil {
        return "", err
    }

    block, _ := pem.Decode(certData)
    cert, err := x509.ParseCertificate(block.Bytes)
    if err != nil {
        return "", err
    }

    return cert.Subject.CommonName, nil
}

func (m *Manager) initHTTPClient() error {
	if _, err := os.Stat(m.config.CertPath); os.IsNotExist(err) {
        return fmt.Errorf("certificate file missing: %s", m.config.CertPath)
    }

    cert, err := tls.LoadX509KeyPair(m.config.CertPath, m.config.KeyPath)
    if err != nil {
        return err
    }

    caCert, err := os.ReadFile(m.config.CaPath)
    if err != nil {
        return err
    }

    caCertPool := x509.NewCertPool()
    caCertPool.AppendCertsFromPEM(caCert)

    m.httpClient = &http.Client{
        Transport: &http.Transport{
            TLSClientConfig: &tls.Config{
                Certificates: []tls.Certificate{cert},
                RootCAs:      caCertPool,
            },
        },
    }
    log.Println("HTTP client initialized successfully")
    return nil
}

func moveFile(src, dest string) error {
    in, err := os.Open(src)
    if err != nil {
        return err
    }
    defer in.Close()

    out, err := os.Create(dest)
    if err != nil {
        return err
    }
    defer out.Close()

    _, err = io.Copy(out, in)
    if err != nil {
        return err
    }
    return os.Remove(src)
}

func (m *Manager) Config() *models.CertificateConfig {
    return m.config
}

func (m *Manager) HTTPClient() *http.Client {
    return m.httpClient
}

func (m *Manager) IsLoaded() bool {
    return m.httpClient != nil && m.config.DeviceID != ""
}
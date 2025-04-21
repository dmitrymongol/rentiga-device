// interfaces/app.go
package interfaces

type Application interface {
    LoadCertificate(string) error
    StartStream()
    StopStream()
    GetConfig() interface{}
    UpdateConnectionStatus(bool, string)
	HasCertificate() bool
}
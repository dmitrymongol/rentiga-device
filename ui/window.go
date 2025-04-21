package ui

import (
	"log"
	"rentiga-device/interfaces"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

type MainWindow struct {
	Window      *gtk.Window
	app         interfaces.Application
	connStatus  *gtk.Image
	connLabel   *gtk.Label
	btnLoadCert *gtk.Button
}

func NewMainWindow(app interfaces.Application) *MainWindow {
	win, _ := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	win.SetTitle("Video Controller")
	win.SetDefaultSize(300, 250)

	box, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 10)
	box.SetMarginTop(20)
	box.SetMarginBottom(20)
	box.SetMarginStart(20)
	box.SetMarginEnd(20)

	mw := &MainWindow{
		Window:    win,
		app:       app,
	}

	mw.createUI(box)
	win.Connect("destroy", mw.onDestroy)
	return mw
}

func (mw *MainWindow) createUI(box *gtk.Box) {
	connBox, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 5)
    mw.connStatus, _ = gtk.ImageNew()
    mw.connLabel, _ = gtk.LabelNew("")
    
    // Инициализируем начальное состояние
    mw.UpdateConnectionStatus(false, "Initializing...")

	connBox.Add(mw.connStatus)
	connBox.Add(mw.connLabel)

	mw.btnLoadCert, _ = gtk.ButtonNewWithLabel("Load Certificate")

	btnStart, _ := gtk.ButtonNewWithLabel("Start")
	btnStop, _ := gtk.ButtonNewWithLabel("Stop")
	btnExit, _ := gtk.ButtonNewWithLabel("Exit")

	mw.btnLoadCert.Connect("clicked", mw.onLoadCert)
	btnStart.Connect("clicked", mw.onStart)
	btnStop.Connect("clicked", mw.onStop)
	btnExit.Connect("clicked", mw.onExit)

	box.Add(connBox)
	box.Add(mw.btnLoadCert)
	box.Add(btnStart)
	box.Add(btnStop)
	box.Add(btnExit)

    mw.UpdateCertificateButton()
	mw.Window.Add(box)
}

func (mw *MainWindow) UpdateCertificateButton() {
    glib.IdleAdd(func() {
        if mw.app.HasCertificate() {
            mw.btnLoadCert.SetLabel("Update Certificate")
        } else {
            mw.btnLoadCert.SetLabel("Load Certificate")
        }
    })
}

func (mw *MainWindow) UpdateConnectionStatus(connected bool, message string) {
	glib.IdleAdd(func() {
		if mw.connStatus == nil || mw.connLabel == nil {
			log.Println("GUI elements not initialized")
			return
		}
		log.Printf("Updating status: connected=%v, message=%s", connected, message)
		
		iconName := "network-error"
		if connected {
			iconName = "network-transmit-receive" // Более подходящая иконка
		}

		mw.connStatus.SetFromIconName(iconName, gtk.ICON_SIZE_BUTTON)
		mw.connLabel.SetText(message)
		
		// Принудительное обновление элементов
		mw.connStatus.Show()
		mw.connLabel.Show()
		mw.Window.QueueDraw()
	})
}
// Обработчики событий
func (mw *MainWindow) onLoadCert() {
	dialog, _ := gtk.FileChooserNativeDialogNew(
        "Select Certificate Bundle",
        mw.Window,
        gtk.FILE_CHOOSER_ACTION_OPEN,
        "Select",
        "Cancel",
    )

    filter, _ := gtk.FileFilterNew()
    filter.AddPattern("*.zip")
    dialog.SetFilter(filter)

    if dialog.Run() == int(gtk.RESPONSE_ACCEPT) {
        file := dialog.GetFilename()
        if err := mw.app.LoadCertificate(file); err != nil {
            log.Println("Certificate processing failed:", err)
            mw.UpdateConnectionStatus(false, "Certificate error")
			mw.UpdateCertificateButton()
        }
    }
}
func (mw *MainWindow) onStart()    { /* ... */ }
func (mw *MainWindow) onStop()     { /* ... */ }
func (mw *MainWindow) onExit()     { /* ... */ }
func (mw *MainWindow) onDestroy()  { /* ... */ }
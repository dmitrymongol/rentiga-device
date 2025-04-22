// ui/windows.go
package ui

import (
	"github.com/gotk3/gotk3/gtk"
)

type MainWindow struct {
	Window *gtk.Window
}

func NewMainWindow() *MainWindow {
	win, _ := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	win.SetTitle("Video Stream")
	win.SetDefaultSize(800, 600)
	return &MainWindow{Window: win}
}
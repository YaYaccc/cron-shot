package gui

import (
	"fyne.io/fyne/v2"
)

var currentPopup fyne.Window

// NewSingletonWindow 创建一个单例窗口,会自动关闭之前创建的窗口
func NewSingletonWindow(title string) fyne.Window {
	if currentPopup != nil {
		currentPopup.Close()
		currentPopup = nil
	}
	w := fyne.CurrentApp().NewWindow(title)
	currentPopup = w
	return w
}

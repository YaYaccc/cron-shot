package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// 辅助函数：显示错误窗口
func showError(app fyne.App, title string, err error) {
	dialog := widget.NewLabel(title + ": " + err.Error())
	w := app.NewWindow("错误")
	w.SetContent(container.NewPadded(dialog))
	w.Resize(fyne.NewSize(300, 100))
	w.Show()
}

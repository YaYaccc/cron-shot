package win

import (
	"cron-shot/assets"
	"cron-shot/constants"
	"time"

	"fyne.io/systray"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	fynetooltip "github.com/dweymouth/fyne-tooltip"
)

// Stoppable 抽象出可停止的后台控制器，用于退出时优雅停止
type Stoppable interface{ Stop() }

// SetupSystemTray 初始化系统托盘菜单（显示/退出）并绑定操作
func SetupSystemTray(myApp fyne.App, myWindow fyne.Window, stopper Stoppable) {
	if d, ok := myApp.(desktop.App); ok {
		// 显示：短暂抑制隐藏，避免立即被最小化逻辑隐藏
		showItem := fyne.NewMenuItem(constants.TextTrayShow, func() {
			SuppressHideFor(2 * time.Second)
			myWindow.Show()
			myWindow.RequestFocus()
		})
		// 退出：优雅停止后台控制器，清理 Tooltip 图层并退出应用
		exitItem := fyne.NewMenuItem(constants.TextTrayExit, func() {
			if stopper != nil {
				stopper.Stop()
			}
			fynetooltip.DestroyWindowToolTipLayer(myWindow.Canvas())
			fyne.CurrentApp().Quit()
		})
		d.SetSystemTrayMenu(fyne.NewMenu(constants.TextAppTitle, showItem, exitItem))
		// 设置托盘悬停提示与标题（Windows/Mac 支持悬停提示文本显示）
		systray.SetTitle(constants.TextAppTitle)
		systray.SetTooltip(constants.TextAppTitle)
	}
}

// GetTrayIconResource 返回托盘/窗口图标资源（embed.FS 内置）
func GetTrayIconResource() fyne.Resource {
	b, _ := assets.Files.ReadFile("icons/cat.png")
	return fyne.NewStaticResource("app.png", b)
}

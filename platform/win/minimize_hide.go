package win

import (
	"cron-shot/constants"
	"cron-shot/logging"
	"sync/atomic"
	"syscall"
	"time"

	"fyne.io/fyne/v2"
	"github.com/lxn/win"
)

// StartHideOnMinimize 监听应用主窗口最小化并自动隐藏
// 使用轮询方式获取窗口句柄并检查最小化状态；当检测到最小化时调用 Fyne 的隐藏接口
func StartHideOnMinimize(myWindow fyne.Window) {
	go func() {
		defer logging.RecoverPanic("hideOnMinimizeLoop")
		title := constants.TextAppTitle
		var hwnd win.HWND
		// 初次尝试在限定次数内获取窗口句柄（应用启动阶段句柄可能尚未就绪）
		for i := 0; i < 20 && hwnd == 0; i++ {
			ptr, _ := syscall.UTF16PtrFromString(title)
			hwnd = win.FindWindow(nil, ptr)
			time.Sleep(100 * time.Millisecond)
		}
		for {
			// 轮询刷新句柄，防止窗口重建导致句柄变化
			ptr, _ := syscall.UTF16PtrFromString(title)
			newHwnd := win.FindWindow(nil, ptr)
			if newHwnd == 0 {
				time.Sleep(200 * time.Millisecond)
				continue
			}
			hwnd = newHwnd
			// 当被暂时抑制隐藏（例如托盘“显示”操作之后），跳过处理
			if IsHideSuppressed() {
				time.Sleep(200 * time.Millisecond)
				continue
			}
			// 检测最小化状态：最小化则隐藏窗口（避免出现在任务栏）
			if win.IsIconic(hwnd) {
				fyne.Do(func() { myWindow.Hide() })
				time.Sleep(500 * time.Millisecond)
			} else {
				time.Sleep(300 * time.Millisecond)
			}
		}
	}()
}

var hideSuppressUntil int64

// SuppressHideFor 在一段时间内抑制自动隐藏
// 用于避免“显示窗口”后立即被最小化逻辑再次隐藏
func SuppressHideFor(d time.Duration) {
	atomic.StoreInt64(&hideSuppressUntil, time.Now().Add(d).UnixNano())
}

// IsHideSuppressed 返回当前是否处于隐藏抑制期
func IsHideSuppressed() bool {
	return time.Now().UnixNano() < atomic.LoadInt64(&hideSuppressUntil)
}

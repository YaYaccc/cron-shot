package app

import (
	"cron-shot/sys_utils"
	"sync"
	"time"
)

// ProcessWindowController 负责维护当前选择的进程并周期刷新其窗口列表
// 通过回调 OnWindowsUpdated 将最新窗口标题传递给 UI 层
type ProcessWindowController struct {
	selectedProcess  string
	stopChan         chan struct{}
	mutex            sync.Mutex
	OnWindowsUpdated func([]string)
}

// NewProcessWindowController 创建进程窗口控制器
func NewProcessWindowController() *ProcessWindowController {
	return &ProcessWindowController{}
}

// SetProcess 设置监控的进程名，并启动/停止轮询
func (c *ProcessWindowController) SetProcess(processName string) {
	// 保证并发安全
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// 若进程名未变化则不处理
	if c.selectedProcess == processName {
		return
	}
	// 关闭已有轮询
	if c.stopChan != nil {
		close(c.stopChan)
		c.stopChan = nil
	}
	// 更新当前选择并立即刷新一次窗口列表
	c.selectedProcess = processName
	if processName == "" {
		if c.OnWindowsUpdated != nil {
			c.OnWindowsUpdated([]string{})
		}
		return
	}
	c.updateWindows()
	// 启动后台轮询
	c.stopChan = make(chan struct{})
	go c.pollWindows(c.stopChan)
}

// pollWindows 每隔固定时间刷新一次窗口列表
func (c *ProcessWindowController) pollWindows(stopChan chan struct{}) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-stopChan:
			return
		case <-ticker.C:
			// 定时触发更新
			c.updateWindows()
		}
	}
}

// updateWindows 枚举当前进程的可见窗口并通过回调通知 UI 层
func (c *ProcessWindowController) updateWindows() {
	if c.selectedProcess == "" {
		if c.OnWindowsUpdated != nil {
			c.OnWindowsUpdated([]string{})
		}
		return
	}
	// 获取窗口标题列表
	windows, _ := sys_utils.GetProcessWindows(c.selectedProcess)
	if c.OnWindowsUpdated != nil {
		c.OnWindowsUpdated(windows)
	}
}

// Stop 停止后台轮询（若存在）
func (c *ProcessWindowController) Stop() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.stopChan != nil {
		close(c.stopChan)
		c.stopChan = nil
	}
}

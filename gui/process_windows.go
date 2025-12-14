package gui

import (
	"cron-shot/constants"
	"cron-shot/logging"
	"cron-shot/sys_utils"
	"regexp"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// WindowStatusUI 组件
type WindowStatusUI struct {
	Container     *fyne.Container
	WindowList    *widget.List
	Windows       []string
	RulesProvider func() []WindowRule
}

// NewWindowStatusUI 创建进程窗口状态部分的UI
func NewWindowStatusUI() *WindowStatusUI {
	ui := &WindowStatusUI{
		Windows: []string{},
	}

	// 窗口列表组件
	ui.WindowList = widget.NewList(
		func() int {
			return len(ui.Windows)
		},
		func() fyne.CanvasObject {
			l := NewHoverLabel("Template")
			l.label.Wrapping = fyne.TextWrapOff
			l.label.Truncation = fyne.TextTruncateEllipsis
			return l
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			if i < len(ui.Windows) {
				name := ui.Windows[i]
				hl := false
				// 正则匹配
				if ui.RulesProvider != nil {
					for _, r := range ui.RulesProvider() {
						if !r.Enabled {
							continue
						}
						if r.Pattern == name {
							hl = true
							break
						}
						re, err := regexp.Compile(r.Pattern)
						if err == nil && re.MatchString(name) {
							hl = true
							break
						}
					}
				}
				lbl := o.(*HoverLabel)
				lbl.SetText(name)
				lbl.SetHighlighted(hl)
				lbl.Refresh()
			}
		},
	)

	// 使用公共组件创建列表容器
	listContainer := NewStyledListContainer(constants.TextWindowStatusHeader, ui.WindowList)

	content := container.NewVBox(
		listContainer,
	)

	ui.Container = content
	return ui
}

// UpdateWindows 更新窗口列表
func (ui *WindowStatusUI) UpdateWindows(windows []string) {
	// 确保在 UI 线程中更新数据和刷新列表
	// 避免 "Error in Fyne call thread" 错误
	fyne.Do(func() {
		ui.Windows = windows
		ui.WindowList.Refresh()
	})
}

// ProcessWindowManager 管理进程窗口的更新和轮询
type ProcessWindowManager struct {
	selectedProcess string
	statusUI        *WindowStatusUI
	stopChan        chan struct{}
	mutex           sync.Mutex
}

// NewProcessWindowManager 创建新的窗口管理器
func NewProcessWindowManager(statusUI *WindowStatusUI) *ProcessWindowManager {
	return &ProcessWindowManager{
		statusUI: statusUI,
	}
}

// SetProcess 设置当前选中的进程并开始轮询
func (m *ProcessWindowManager) SetProcess(processName string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 如果进程名没变，不做任何事
	if m.selectedProcess == processName {
		return
	}

	// 停止旧的轮询
	if m.stopChan != nil {
		close(m.stopChan)
		m.stopChan = nil
	}

	m.selectedProcess = processName

	// 如果选择了空进程，清空列表并返回
	if processName == "" {
		m.statusUI.UpdateWindows([]string{})
		return
	}

	// 立即进行一次更新
	m.updateWindows()

	// 启动新的轮询，每5秒更新一次
	m.stopChan = make(chan struct{})
	go m.pollWindows(m.stopChan)
}

// pollWindows 定时轮询窗口列表
func (m *ProcessWindowManager) pollWindows(stopChan chan struct{}) {
	defer logging.RecoverPanic("pollWindows")
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-stopChan:
			return
		case <-ticker.C:
			m.updateWindows()
		}
	}
}

// updateWindows 获取并更新窗口列表
func (m *ProcessWindowManager) updateWindows() {
	// 注意：这里不需要加锁，因为 SetProcess 里的锁主要保护状态切换
	// updateWindows 内部操作是安全的，或者是 UI 线程安全的调用

	windows, err := sys_utils.GetProcessWindows(m.selectedProcess)
	if err != nil {
		// 如果出错，可以在这里处理，例如显示错误信息
		// 目前简单处理：显示错误作为列表项
		m.statusUI.UpdateWindows([]string{"Error: " + err.Error()})
		return
	}

	if len(windows) == 0 {
		names, _ := sys_utils.GetProcessNames()
		found := false
		for _, n := range names {
			if n == m.selectedProcess {
				found = true
				break
			}
		}
		if !found {
			m.statusUI.UpdateWindows([]string{"(找不到进程)"})
		} else {
			m.statusUI.UpdateWindows([]string{"(无可见窗口)"})
		}
	} else {
		m.statusUI.UpdateWindows(windows)
	}
}

// Stop 停止轮询 (在程序退出或组件销毁时调用)
func (m *ProcessWindowManager) Stop() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.stopChan != nil {
		close(m.stopChan)
		m.stopChan = nil
	}
}

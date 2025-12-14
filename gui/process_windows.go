package gui

import (
	"cron-shot/constants"
	"regexp"

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

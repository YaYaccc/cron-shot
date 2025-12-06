package gui

import (
	"cron-shot/sys_utils"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// ProcessUI 组件
type ProcessUI struct {
	Container          *fyne.Container
	EntryProcessLocked *lockedEntry
	ButtonProcess      *widget.Button
}

// NewProcessUI 创建进程选择部分的UI
func NewProcessUI(app fyne.App, onProcessSelected func(string)) *ProcessUI {
	labelProcess := widget.NewLabel("进程:")
	labelProcess.TextStyle = fyne.TextStyle{Bold: true}
	entryProcessLocked := newLockedEntry()
	entryProcessLocked.PlaceHolder = "未选择进程"

	buttonProcess := widget.NewButton("选择", func() {
		// 获取进程列表
		processNames, err := sys_utils.GetProcessNames()
		if err != nil {
			showError(app, "获取进程失败", err)
			return
		}

		showSelectionWindow(app, "选择进程", processNames, func(selected string) {
			entryProcessLocked.SetText(selected)
			if onProcessSelected != nil {
				onProcessSelected(selected)
			}
		})
	})

	processRow := container.NewBorder(nil, nil, labelProcess, buttonProcess, entryProcessLocked)

	return &ProcessUI{
		Container:          processRow,
		EntryProcessLocked: entryProcessLocked,
		ButtonProcess:      buttonProcess,
	}
}

// lockedEntry 是一个只读的 Entry，通过拦截输入实现
type lockedEntry struct {
	widget.Entry
}

func newLockedEntry() *lockedEntry {
	entry := &lockedEntry{}
	entry.ExtendBaseWidget(entry)
	return entry
}

func (e *lockedEntry) TypedRune(r rune) {
	// 拦截输入，什么都不做
}

func (e *lockedEntry) TypedKey(key *fyne.KeyEvent) {
	// 拦截按键，什么都不做
}

// showSelectionWindow 在弹窗中显示进程列表并支持搜索
func showSelectionWindow(app fyne.App, title string, items []string, onSelected func(string)) {
	w := NewSingletonWindow(title)
	w.Resize(fyne.NewSize(400, 500))

	filtered := make([]string, len(items))
	copy(filtered, items)

	list := widget.NewList(
		func() int { return len(filtered) },
		func() fyne.CanvasObject { return widget.NewLabel("Template") },
		func(i widget.ListItemID, o fyne.CanvasObject) { o.(*widget.Label).SetText(filtered[i]) },
	)

	entry := widget.NewEntry()
	entry.PlaceHolder = "搜索进程..."
	entry.OnChanged = func(s string) {
		q := strings.ToLower(strings.TrimSpace(s))
		if q == "" {
			filtered = items
		} else {
			var out []string
			for _, it := range items {
				if strings.Contains(strings.ToLower(it), q) {
					out = append(out, it)
				}
			}
			filtered = out
		}
		list.Refresh()
	}

	list.OnSelected = func(id widget.ListItemID) {
		onSelected(filtered[id])
		w.Close()
	}

	content := container.NewBorder(entry, nil, nil, nil, list)
	w.SetContent(content)
	w.Show()
}

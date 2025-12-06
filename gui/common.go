package gui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// WindowRule 定义窗口匹配规则
type WindowRule struct {
	Pattern     string // 文本内容
	Enabled     bool   // 是否激活
	StorageRule string // 存储文件夹规则（固定文本或带捕获组的正则）
	FixedFolder string // 固定文件夹（若不为空，则优先在此文件夹下存储）
}

var AppCanvas fyne.Canvas

// NewStyledListContainer 创建一个统一风格的列表容器
// 包含标题、边框、固定高度限制和左侧对齐间距
func NewStyledListContainer(title string, list *widget.List) *fyne.Container {
	// 1. 设置列表样式
	// 禁止列表选中高亮
	list.OnSelected = func(id widget.ListItemID) {
		list.Unselect(id)
	}

	// 2. 创建标题
	labelTitle := widget.NewLabel(title)
	labelTitle.TextStyle = fyne.TextStyle{Bold: true}

	// 3. 创建列表容器背景框 (用于撑开高度和显示边框)
	// 设置列表最小高度为 5 个条目高度 (估算 150)
	rect := canvas.NewRectangle(color.Transparent)
	rect.SetMinSize(fyne.NewSize(0, 150))
	rect.StrokeWidth = 1
	rect.StrokeColor = theme.DisabledColor()

	// 使用 Stack 叠加边框和列表
	listWrapper := container.NewStack(list, rect)

	// 4. 创建左侧间距 (2px) 使边框与上方Label的文字对齐
	return container.NewBorder(labelTitle, nil, nil, nil, listWrapper)
}

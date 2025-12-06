package gui

import (
	"cron-shot/config"
	"cron-shot/constants"
	"regexp"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	fynetooltip "github.com/dweymouth/fyne-tooltip"
)

// RulesUI 组件
type RulesUI struct {
	Container      *fyne.Container
	Rules          []WindowRule
	RuleList       *widget.List
	OnRulesChanged func()
}

// NewRulesUI 创建规则列表部分的UI
func NewRulesUI(app fyne.App) *RulesUI {
	ui := &RulesUI{
		Rules: []WindowRule{},
	}

	// 正则表达式控件
	entryRegex := widget.NewEntry()
	entryRegex.PlaceHolder = "添加窗口名称或正则表达式，用于匹配"
	// 添加规则按钮
	buttonAddRule := widget.NewButton(constants.TextAdd, func() {
		regStr := entryRegex.Text
		if regStr == "" {
			return
		}
		// 验证正则表达式是否有效
		if _, err := regexp.Compile(regStr); err != nil {
			showError(app, "正则表达式错误", err)
			return
		}

		ui.Rules = append(ui.Rules, WindowRule{Pattern: regStr, Enabled: true})
		entryRegex.SetText("")
		ui.RuleList.Refresh()
		if ui.OnRulesChanged != nil {
			ui.OnRulesChanged()
		}
		var cfgRules []config.AppRule
		for _, r := range ui.Rules {
			cfgRules = append(cfgRules, config.AppRule{Pattern: r.Pattern, Enabled: r.Enabled, StorageRule: r.StorageRule, FixedFolder: r.FixedFolder})
		}
		config.SetRules(cfgRules)
	})

	// 规则列表组件
	ui.RuleList = widget.NewList(
		func() int {
			return len(ui.Rules)
		},
		func() fyne.CanvasObject {
			label := NewHoverLabel("Template")
			label.label.Wrapping = fyne.TextWrapOff
			label.label.Truncation = fyne.TextTruncateEllipsis

			btnToggle := widget.NewButton(constants.TextActivate, nil)
			btnConfig := widget.NewButton(constants.TextConfig, nil)
			btnDelete := widget.NewButton(constants.TextDelete, nil)

			smallSize := fyne.NewSize(50, 26)
			wrapToggle := container.NewGridWrap(smallSize, btnToggle)
			wrapConfig := container.NewGridWrap(smallSize, btnConfig)
			wrapDelete := container.NewGridWrap(smallSize, btnDelete)
			btnRow := container.NewHBox(wrapToggle, wrapConfig, wrapDelete)
			right := container.NewVBox(layout.NewSpacer(), btnRow, layout.NewSpacer())
			return container.NewBorder(nil, nil, nil, right, label)
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			c := o.(*fyne.Container)
			var label *HoverLabel
			var buttons []*widget.Button
			var walk func(obj fyne.CanvasObject)
			walk = func(obj fyne.CanvasObject) {
				switch t := obj.(type) {
				case *HoverLabel:
					if label == nil {
						label = t
					}
				case *widget.Button:
					buttons = append(buttons, t)
				case *fyne.Container:
					for _, child := range t.Objects {
						walk(child)
					}
				}
			}
			walk(c)

			if i >= len(ui.Rules) || label == nil || len(buttons) < 3 {
				return
			}

			rule := ui.Rules[i]
			label.SetText(rule.Pattern)

			toggle := buttons[0]
			configBtn := buttons[1]
			deleteBtn := buttons[2]
			if rule.Enabled {
				toggle.SetText(constants.TextDeactivate)
				toggle.Importance = widget.HighImportance
			} else {
				toggle.SetText(constants.TextActivate)
				toggle.Importance = widget.MediumImportance
			}
			toggle.Refresh()

			toggle.OnTapped = func() {
				if i < len(ui.Rules) {
					ui.Rules[i].Enabled = !ui.Rules[i].Enabled
					ui.RuleList.Refresh()
					if ui.OnRulesChanged != nil {
						ui.OnRulesChanged()
					}
					var cfgRules []config.AppRule
					for _, r := range ui.Rules {
						cfgRules = append(cfgRules, config.AppRule{Pattern: r.Pattern, Enabled: r.Enabled, StorageRule: r.StorageRule, FixedFolder: r.FixedFolder})
					}
					config.SetRules(cfgRules)
				}
			}
			configBtn.OnTapped = func() {
				w := NewSingletonWindow(constants.TextStorageRuleTitle)
				entryRule := widget.NewEntry()
				entryRule.PlaceHolder = constants.PlaceholderStorageRule
				entryRule.SetText(ui.Rules[i].StorageRule)
				entryFixed := widget.NewEntry()
				entryFixed.PlaceHolder = constants.PlaceholderFixedFolder
				entryFixed.SetText(ui.Rules[i].FixedFolder)
				btnSave := widget.NewButton(constants.TextSave, func() {
					ui.Rules[i].StorageRule = entryRule.Text
					ui.Rules[i].FixedFolder = entryFixed.Text
					ui.RuleList.Refresh()
					if ui.OnRulesChanged != nil {
						ui.OnRulesChanged()
					}
					var cfgRules []config.AppRule
					for _, r := range ui.Rules {
						cfgRules = append(cfgRules, config.AppRule{Pattern: r.Pattern, Enabled: r.Enabled, StorageRule: r.StorageRule, FixedFolder: r.FixedFolder})
					}
					config.SetRules(cfgRules)
					w.Close()
				})
				btnCancel := widget.NewButton(constants.TextCancel, func() { w.Close() })
				labelRule := widget.NewLabel(constants.TextStorageRuleTitle)
				labelFixed := widget.NewLabel(constants.TextFixedFolderTitle)
				inner := container.NewVBox(
					labelRule,
					entryRule,
					labelFixed,
					entryFixed,
					container.NewHBox(btnSave, btnCancel),
				)
				padded := container.NewPadded(inner)
				wrapped := fynetooltip.AddWindowToolTipLayer(padded, w.Canvas())
				w.SetContent(wrapped)
				w.Resize(fyne.NewSize(420, 220))
				w.SetOnClosed(func() {
					fynetooltip.DestroyWindowToolTipLayer(w.Canvas())
				})
				w.Show()
			}
			deleteBtn.OnTapped = func() {
				if i < len(ui.Rules) {
					ui.Rules = append(ui.Rules[:i], ui.Rules[i+1:]...)
					ui.RuleList.Refresh()
					if ui.OnRulesChanged != nil {
						ui.OnRulesChanged()
					}
					var cfgRules []config.AppRule
					for _, r := range ui.Rules {
						cfgRules = append(cfgRules, config.AppRule{Pattern: r.Pattern, Enabled: r.Enabled, StorageRule: r.StorageRule, FixedFolder: r.FixedFolder})
					}
					config.SetRules(cfgRules)
				}
			}
		},
	)

	// 1. 创建规则列表区域 (使用公共组件)
	listContainer := NewStyledListContainer(constants.TextRulesHeader, ui.RuleList)

	// 2. 创建添加规则区域
	labelWindow := widget.NewLabel(constants.TextRuleLabel)
	labelWindow.TextStyle = fyne.TextStyle{Bold: true}
	windowRow := container.NewBorder(nil, nil, labelWindow, buttonAddRule, entryRegex)

	// 3. 总体布局：上面是添加栏，下面是列表
	// 使用 VBox 布局，使列表部分保持由 rect 撑开的固定高度 (150)，不会自动填满整个窗口
	// 这样当条目超过5个时，列表内部会出现滚动条
	content := container.NewVBox(
		windowRow,
		listContainer,
	)

	ui.Container = content
	return ui
}

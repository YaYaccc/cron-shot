package gui

import (
	"fmt"
	"image"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"cron-shot/config"
	"cron-shot/constants"
	"cron-shot/logging"
	"cron-shot/sys_utils"
	"cron-shot/utils"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
	fynetooltip "github.com/dweymouth/fyne-tooltip"
	"github.com/lxn/win"
)

// Run 启动应用程序主界面与业务逻辑
func Run() {
	defer logging.RecoverPanic("gui.Run")
	myApp := app.New()
	myApp.Settings().SetTheme(&customTheme{})
	myApp.SetIcon(GetTrayIconResource())
	myWindow := myApp.NewWindow(constants.TextAppTitle)
	AppCanvas = myWindow.Canvas()
	config.Init()
	cfgDir, _ := os.UserConfigDir()
	baseCfg := filepath.Join(cfgDir, "CronShot")
	_ = logging.Init(baseCfg)
	logging.Info("config loaded")

	initAutostartRegistration()

	// 初始化各个模块
	rulesUI := NewRulesUI(myApp)
	windowStatusUI := NewWindowStatusUI()
	windowStatusUI.RulesProvider = func() []WindowRule { return rulesUI.Rules }
	processWindowManager := NewProcessWindowManager(windowStatusUI)
	rulesUI.OnRulesChanged = func() {
		windowStatusUI.UpdateWindows(windowStatusUI.Windows)
		// 持久化规则
		var cfgRules []config.AppRule
		for _, r := range rulesUI.Rules {
			cfgRules = append(cfgRules, config.AppRule{Pattern: r.Pattern, Enabled: r.Enabled, StorageRule: r.StorageRule, FixedFolder: r.FixedFolder})
		}
		config.SetRules(cfgRules)
	}

	var currentProcess string
	processUI := NewProcessUI(myApp, func(selectedProcess string) {
		// 进程选择回调
		// 切换监控的进程，这会立即更新一次窗口列表，并启动定时轮询
		processWindowManager.SetProcess(selectedProcess)
		currentProcess = selectedProcess
		config.SetCurrentProcess(selectedProcess)
	})

	// 从配置加载规则与进程
	if cfg := config.GetRules(); len(cfg) > 0 {
		rulesUI.Rules = nil
		for _, r := range cfg {
			rulesUI.Rules = append(rulesUI.Rules, WindowRule{Pattern: r.Pattern, Enabled: r.Enabled, StorageRule: r.StorageRule, FixedFolder: r.FixedFolder})
		}
		rulesUI.RuleList.Refresh()
	}
	if p := config.GetCurrentProcess(); strings.TrimSpace(p) != "" {
		currentProcess = p
		processWindowManager.SetProcess(p)
		processUI.EntryProcessLocked.SetText(p)
	}

	// 将规则列表和状态列表组合在中间区域
	// 使用 VBox 垂直排列
	centerContent := container.NewVBox(
		rulesUI.Container,
		windowStatusUI.Container,
	)

	autoEnabled := false
	dedupeEnabled := config.GetDedupeEnabled()
	var autoStopChan chan struct{}
	autoBtn := widget.NewButton(constants.TextOpenAutoShot, nil)
	autoBtn.Importance = widget.MediumImportance
	autoBtn.OnTapped = func() {
		autoEnabled = !autoEnabled
		if autoEnabled {
			autoBtn.SetText(constants.TextCloseAutoShot)
			autoBtn.Importance = widget.HighImportance
			autoBtn.Refresh()
			autoStopChan = make(chan struct{})
			go startAutoCaptureLoop(autoStopChan, &currentProcess, rulesUI, windowStatusUI, &dedupeEnabled)
		} else {
			autoBtn.SetText(constants.TextOpenAutoShot)
			autoBtn.Importance = widget.MediumImportance
			autoBtn.Refresh()
			if autoStopChan != nil {
				close(autoStopChan)
				autoStopChan = nil
			}
		}
	}
	settingsBtn := widget.NewButton(constants.TextSettings, func() {
		onSettingsButtonTapped(myApp, &autoEnabled, &autoStopChan, autoBtn, &dedupeEnabled, &currentProcess, rulesUI, windowStatusUI)
	})
	openPicturesBtn := widget.NewButton(constants.TextOpenPicturesFolder, func() {
		_ = sys_utils.OpenFolder(config.GetStorageRoot())
	})
	openConfigBtn := widget.NewButton(constants.TextOpenConfigFolder, func() {
		cfgDir, _ := os.UserConfigDir()
		base := filepath.Join(cfgDir, constants.TextAppTitle)
		_ = sys_utils.OpenFolder(base)
	})
	buttonsRow := container.NewGridWithColumns(1, autoBtn)
	bottomRow := container.NewBorder(nil, nil, nil, nil, buttonsRow)

	aboutBtn := widget.NewButton(constants.TextAbout, func() {
		w := NewSingletonWindow(constants.TextAbout)
		l1 := widget.NewLabel("版本:v0.0.2")
		l2 := widget.NewLabel("作者:YaYa")
		u, _ := url.Parse("https://github.com/YaYaccc/cron-shot")
		rowProject := widget.NewRichText(&widget.TextSegment{Text: "项目地址:"}, &widget.HyperlinkSegment{Text: "https://github.com/YaYaccc/cron-shot", URL: u})
		u2, _ := url.Parse("https://icons8.com/")
		rowIcons := widget.NewRichText(&widget.TextSegment{Text: "图标来源:"}, &widget.HyperlinkSegment{Text: "Icons8", URL: u2})
		form := container.NewVBox(l1, l2, rowProject, rowIcons)
		wrapped := fynetooltip.AddWindowToolTipLayer(container.NewPadded(form), w.Canvas())
		w.SetContent(wrapped)
		w.Resize(fyne.NewSize(420, 200))
		w.SetOnClosed(func() { fynetooltip.DestroyWindowToolTipLayer(w.Canvas()) })
		w.Show()
	})
	actionsTop := container.NewGridWithColumns(2, openPicturesBtn, openConfigBtn)
	actionsBottom := container.NewGridWithColumns(2, settingsBtn, aboutBtn)
	actionsRow := container.NewVBox(actionsTop, actionsBottom)
	centerContent = container.NewVBox(
		rulesUI.Container,
		windowStatusUI.Container,
		bottomRow,
		actionsRow,
	)

	content := container.NewBorder(
		processUI.Container,
		nil,
		nil, nil,
		centerContent,
	)

	wrapped := fynetooltip.AddWindowToolTipLayer(container.NewPadded(content), myWindow.Canvas())
	myWindow.SetContent(wrapped)
	myWindow.Resize(fyne.NewSize(600, 600))
	setupSystemTray(myApp, myWindow, processWindowManager)
	myWindow.SetCloseIntercept(myWindow.Hide)
	if config.GetSilentStartEnabled() {
		startHideOnMinimize(myWindow)
	}

	// 确保在窗口关闭时停止轮询
	myWindow.SetOnClosed(func() {
		processWindowManager.Stop()
		fynetooltip.DestroyWindowToolTipLayer(myWindow.Canvas())
	})

	if config.GetAutoCaptureEnabled() && strings.TrimSpace(currentProcess) != "" && !autoEnabled {
		autoBtn.OnTapped()
		autoBtn.Refresh()
	}
	if config.GetSilentStartEnabled() {
		myWindow.Hide()
		myApp.Run()
	} else {
		myWindow.ShowAndRun()
	}
}

// onSettingsButtonTapped 打开设置窗口并保存改动
func onSettingsButtonTapped(_ fyne.App, autoEnabled *bool, autoStopChanPtr *chan struct{}, _ *widget.Button, dedupeEnabled *bool, currentProcess *string, rulesUI *RulesUI, windowStatusUI *WindowStatusUI) {
	w := NewSingletonWindow(constants.TextSettingsTitle)
	entryRoot := widget.NewEntry()
	entryRoot.SetText(config.GetStorageRoot())
	entryRootWrap := container.NewGridWrap(fyne.NewSize(420, entryRoot.MinSize().Height), entryRoot)
	entryInterval := widget.NewEntry()
	entryInterval.SetText(fmt.Sprintf("%d", config.GetScreenshotIntervalSec()))
	toggleDedupe := widget.NewCheck(constants.TextDedupeTitle, func(v bool) {})
	toggleDedupe.SetChecked(config.GetDedupeEnabled())
	valueLabel := widget.NewLabel(fmt.Sprintf("%d", config.GetDedupeThreshold()))
	sliderThreshold := widget.NewSlider(1, 100)
	sliderThreshold.Step = 1
	sliderThreshold.Value = float64(config.GetDedupeThreshold())
	sliderThreshold.OnChanged = func(v float64) {
		valueLabel.SetText(fmt.Sprintf("%d", int(v)))
	}
	leftInfo := container.NewHBox(widget.NewLabel(constants.TextDedupeThreshold), valueLabel)
	thresholdRow := container.NewBorder(nil, nil, leftInfo, nil, sliderThreshold)
	thresholdRow.Hide()
	if toggleDedupe.Checked {
		thresholdRow.Show()
	}
	toggleDedupe.OnChanged = func(v bool) {
		if v {
			thresholdRow.Show()
		} else {
			thresholdRow.Hide()
		}
	}
	toggleAutoStart := widget.NewCheck(constants.TextAutoStartTitle, func(v bool) {})
	toggleAutoStart.SetChecked(config.GetAutostartEnabled())
	toggleAutoCapture := widget.NewCheck(constants.TextAutoCaptureTitle, func(v bool) {})
	toggleAutoCapture.SetChecked(config.GetAutoCaptureEnabled())
	toggleSilentStart := widget.NewCheck(constants.TextSilentStartTitle, func(v bool) {})
	toggleSilentStart.SetChecked(config.GetSilentStartEnabled())
	chooseBtn := widget.NewButton(constants.TextChoose, func() {
		if p, err := sys_utils.PickFolder(); err == nil && strings.TrimSpace(p) != "" {
			entryRoot.SetText(p)
		}
	})
	resetBtn := widget.NewButton(constants.TextResetDefault, func() {
		entryRoot.SetText(config.GetDefaultStorageRoot())
	})
	save := widget.NewButton(constants.TextSave, func() {
		root := entryRoot.Text
		config.SetStorageRoot(root)
		n := 5
		if v, err := strconv.Atoi(strings.TrimSpace(entryInterval.Text)); err == nil {
			if v > 0 {
				n = v
			}
		}
		config.SetScreenshotIntervalSec(n)
		config.SetDedupeEnabled(toggleDedupe.Checked)
		*dedupeEnabled = toggleDedupe.Checked
		// threshold
		th := int(sliderThreshold.Value)
		if th < 1 {
			th = 1
		} else if th > 100 {
			th = 100
		}
		config.SetDedupeThreshold(th)
		config.SetAutostartEnabled(toggleAutoStart.Checked)
		config.SetAutoCaptureEnabled(toggleAutoCapture.Checked)
		config.SetSilentStartEnabled(toggleSilentStart.Checked)
		exe, _ := os.Executable()
		if toggleAutoStart.Checked {
			_ = sys_utils.EnableAutoStart(constants.TextAppTitle, exe)
		} else {
			_ = sys_utils.DisableAutoStart(constants.TextAppTitle)
		}
		if *autoEnabled {
			if *autoStopChanPtr != nil {
				close(*autoStopChanPtr)
			}
			*autoStopChanPtr = make(chan struct{})
			go startAutoCaptureLoop(*autoStopChanPtr, currentProcess, rulesUI, windowStatusUI, dedupeEnabled)
		}
		w.Close()
	})
	cancel := widget.NewButton(constants.TextCancel, func() { w.Close() })
	form := container.NewVBox(
		widget.NewLabel(constants.TextStorageRootTitle),
		container.NewHBox(entryRootWrap, chooseBtn, resetBtn),
		widget.NewLabel(constants.TextIntervalTitle),
		entryInterval,
		toggleDedupe,
		thresholdRow,
		toggleAutoStart,
		toggleAutoCapture,
		toggleSilentStart,
		container.NewHBox(save, cancel),
	)
	wrapped := fynetooltip.AddWindowToolTipLayer(container.NewPadded(form), w.Canvas())
	w.SetContent(wrapped)
	w.Resize(fyne.NewSize(480, 280))
	w.SetOnClosed(func() { fynetooltip.DestroyWindowToolTipLayer(w.Canvas()) })
	w.Show()
}

// startAutoCaptureLoop 周期性根据规则与去重逻辑进行截图保存
func startAutoCaptureLoop(stop chan struct{}, currentProcess *string, rulesUI *RulesUI, windowStatusUI *WindowStatusUI, dedupeEnabled *bool) {
	defer logging.RecoverPanic("startAutoCaptureLoop")
	interval := time.Duration(config.GetScreenshotIntervalSec()) * time.Second
	if interval <= 0 {
		interval = time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			if *currentProcess == "" {
				continue
			}
			processScreenshotTick(*currentProcess, rulesUI.Rules, windowStatusUI, *dedupeEnabled)
		}
	}
}

// processScreenshotTick 执行一次截图循环：获取窗口、规则匹配、截图、去重与保存
func processScreenshotTick(currentProcess string, rules []WindowRule, windowStatusUI *WindowStatusUI, dedupeEnabled bool) {
	logging.Info("start screenshot tick for process: " + currentProcess)
	infos, err := sys_utils.GetProcessWindowsDetailed(currentProcess)
	if err != nil {
		logging.Error("GetProcessWindowsDetailed error: " + err.Error())
		return
	}
	if len(infos) == 0 {
		logging.Info("process not found: " + currentProcess)
		windowStatusUI.UpdateWindows([]string{"(找不到进程)"})
		return
	}
	base := time.Now()
	idx := 0
	for _, info := range infos {
		// 在匹配前刷新窗口标题：以当前时刻的标题进行规则判断，降低命名偏差
		titleForMatch := sys_utils.GetWindowTitleByHWND(info.HWND)
		if strings.TrimSpace(titleForMatch) == "" {
			// 无法获取到标题名称 跳过
			continue
		}
		rule, ok := matchRule(titleForMatch, rules)
		if !ok {
			continue
		}
		// 不对最小化窗口进行截图
		if win.IsIconic(info.HWND) {
			continue
		}
		// 执行截图
		img, err := sys_utils.CaptureWindowImage(info.HWND)
		if err != nil {
			logging.Error("capture failed: " + err.Error())
			continue
		}
		// 截图后再次获取标题，若两次不一致，则表示标题在快速变动，忽略本次截图事件
		// 否则可能造成存储目的地与预期不一致
		currTitle := sys_utils.GetWindowTitleByHWND(info.HWND)
		if strings.TrimSpace(currTitle) != titleForMatch {
			logging.Info("window title drift: " + titleForMatch + " -> " + currTitle)
			continue
		}
		// 使用匹配时刻的标题计算存储文件夹，保证规则与命名一致
		folder, fixed := resolveFolder(titleForMatch, rule)
		if dedupeEnabled && shouldSkipDueToDedupe(img, config.GetStorageRoot(), currentProcess, fixed, folder) {
			continue
		}
		p, err := sys_utils.SaveCronShot(img, config.GetStorageRoot(), currentProcess, fixed, folder, base.Add(time.Duration(idx)*time.Millisecond))
		if err != nil {
			logging.Error("save failed: " + err.Error())
		} else {
			logging.Info("screenshot saved: " + p)
		}
		idx++
	}
}

// matchRule 进行规则匹配：先精确匹配，再进行正则匹配
func matchRule(title string, rules []WindowRule) (*WindowRule, bool) {
	for _, r := range rules {
		if !r.Enabled {
			continue
		}
		if r.Pattern == title {
			return &r, true
		}
	}
	for _, r := range rules {
		if !r.Enabled {
			continue
		}
		re, err := regexp.Compile(r.Pattern)
		if err == nil && re.MatchString(title) {
			return &r, true
		}
	}
	return nil, false
}

// resolveFolder 解析存储文件夹名称与固定前缀
func resolveFolder(title string, rule *WindowRule) (string, string) {
	if rule == nil {
		return title, ""
	}
	folder := utils.ResolveStorageFolder(title, rule.StorageRule)
	return folder, rule.FixedFolder
}

// shouldSkipDueToDedupe 执行去重判定：
// - 构造目标保存目录并读取最新图片
// - 当阈值=100时用像素级比较；否则用 AHash16x16 + 汉明距离转换相似度
// - 返回 true 表示应跳过保存
func shouldSkipDueToDedupe(img *image.RGBA, storageRoot, processName, fixed, folder string) bool {
	proc := utils.SanitizeProcessName(processName)
	sub := utils.SanitizeFolderName(folder)
	var dir string
	if strings.TrimSpace(fixed) != "" {
		fix := utils.SanitizeFolderName(fixed)
		dir = filepath.Join(storageRoot, proc, fix, sub)
	} else {
		dir = filepath.Join(storageRoot, proc, sub)
	}
	prevImg, _ := utils.LatestPNGImage(dir)
	if prevImg == nil {
		return false
	}
	th := config.GetDedupeThreshold()
	if th >= 100 {
		if utils.ImagesEqualExact(img, prevImg) {
			logging.Info("skip save due to identical pixels")
			return true
		}
		return false
	}
	prevHash := utils.AHashFromImage(prevImg)
	currHash := utils.AHash16x16(img)
	dist := utils.Hamming256(prevHash, currHash)
	sim := 1.0 - float64(dist)/256.0
	if sim*100.0 >= float64(th) {
		logging.Info("skip save due to similarity >= threshold")
		return true
	}
	return false
}

// initAutostartRegistration 在启动时根据配置校验并注册开机自启
// 避免用户替换配置时，启动配置未生效
func initAutostartRegistration() {
	app := constants.TextAppTitle
	if config.GetAutostartEnabled() {
		exe, _ := os.Executable()
		ok, v, err := sys_utils.IsAutoStartRegistered(app)
		if err == nil {
			if !ok || strings.TrimSpace(v) != exe {
				_ = sys_utils.EnableAutoStart(app, exe)
			}
		}
	} else {
		ok, _, err := sys_utils.IsAutoStartRegistered(app)
		if err == nil && ok {
			_ = sys_utils.DisableAutoStart(app)
		}
	}
}

// setupSystemTray 初始化系统托盘菜单与图标
func setupSystemTray(myApp fyne.App, myWindow fyne.Window, processWindowManager *ProcessWindowManager) {
	if d, ok := myApp.(desktop.App); ok {
		showItem := fyne.NewMenuItem(constants.TextTrayShow, func() {
			myWindow.Show()
			myWindow.RequestFocus()
		})
		exitItem := fyne.NewMenuItem(constants.TextTrayExit, func() {
			processWindowManager.Stop()
			fynetooltip.DestroyWindowToolTipLayer(myWindow.Canvas())
			fyne.CurrentApp().Quit()
		})
		d.SetSystemTrayMenu(fyne.NewMenu(constants.TextAppTitle, showItem, exitItem))
	}
}

// startHideOnMinimize 监听窗口最小化并隐藏到托盘
func startHideOnMinimize(myWindow fyne.Window) {
	go func() {
		defer logging.RecoverPanic("hideOnMinimizeLoop")
		title := constants.TextAppTitle
		var hwnd win.HWND
		for i := 0; i < 20 && hwnd == 0; i++ {
			ptr, _ := syscall.UTF16PtrFromString(title)
			hwnd = win.FindWindow(nil, ptr)
			time.Sleep(100 * time.Millisecond)
		}
		if hwnd == 0 {
			return
		}
		for {
			if win.IsIconic(hwnd) {
				fyne.Do(func() { myWindow.Hide() })
				// 等待一段时间，避免紧凑重复触发
				time.Sleep(500 * time.Millisecond)
			} else {
				time.Sleep(300 * time.Millisecond)
			}
		}
	}()
}

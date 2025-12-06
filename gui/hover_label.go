package gui

import (
	"image/color"
	"time"

	"cron-shot/constants"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	fynetooltip "github.com/dweymouth/fyne-tooltip"
	ttwidget "github.com/dweymouth/fyne-tooltip/widget"
)

// HoverLabel 是一个支持鼠标悬停显示 Tooltip 的 Label
type HoverLabel struct {
	ttwidget.ToolTipWidget
	label       *widget.Label
	bg          *canvas.Rectangle
	highlighted bool
}

// NewHoverLabel 创建一个新的 HoverLabel
func NewHoverLabel(text string) *HoverLabel {
	l := &HoverLabel{}
	l.label = widget.NewLabel(text)
	l.bg = canvas.NewRectangle(theme.BackgroundColor())
	l.ExtendBaseWidget(l)
	l.SetToolTip(text)
	return l
}

func (l *HoverLabel) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(container.NewStack(l.bg, l.label))
}

func (l *HoverLabel) SetText(text string) {
	l.label.SetText(text)
	l.SetToolTip(text)
}

func (l *HoverLabel) SetHighlighted(h bool) {
	l.highlighted = h
	if h {
		l.bg.FillColor = color.NRGBA{R: 255, G: 236, B: 236, A: 255}
		l.bg.StrokeColor = color.NRGBA{R: 255, G: 200, B: 200, A: 255}
		l.bg.StrokeWidth = 1
	} else {
		l.bg.FillColor = theme.BackgroundColor()
		l.bg.StrokeWidth = 0
	}
	l.bg.Refresh()
	l.Refresh()
}

func (l *HoverLabel) TappedSecondary(ev *fyne.PointEvent) {
	fyne.CurrentApp().Clipboard().SetContent(l.label.Text)
	l.showCopiedBubble(ev.Position)
}

func (l *HoverLabel) showCopiedBubble(rel fyne.Position) {
	if AppCanvas == nil {
		return
	}
	fg := theme.ForegroundColor()
	base := color.NRGBAModel.Convert(fg).(color.NRGBA)
	text := canvas.NewText(constants.TextCopiedBubble, color.NRGBA{R: base.R, G: base.G, B: base.B, A: 0})
	text.Alignment = fyne.TextAlignCenter
	text.TextSize = theme.Size(theme.SizeNameCaptionText)
	bubble := container.NewPadded(text)
	pop := widget.NewPopUp(bubble, AppCanvas)
	fynetooltip.AddPopUpToolTipLayer(pop)
	pop.Resize(bubble.MinSize())
	pop.ShowAtRelativePosition(rel, l)

	steps := 8
	dur := 200 * time.Millisecond
	stepDur := dur / time.Duration(steps)

	go func() {
		// fade in
		for i := 1; i <= steps; i++ {
			a := uint8(i * 255 / steps)
			time.Sleep(stepDur)
			fyne.Do(func() {
				text.Color = color.NRGBA{R: base.R, G: base.G, B: base.B, A: a}
				text.Refresh()
			})
		}

		// hold
		time.Sleep(600 * time.Millisecond)

		// fade out
		for i := steps - 1; i >= 0; i-- {
			a := uint8(i * 255 / steps)
			time.Sleep(stepDur)
			fyne.Do(func() {
				text.Color = color.NRGBA{R: base.R, G: base.G, B: base.B, A: a}
				text.Refresh()
			})
		}

		fyne.Do(func() {
			pop.Hide()
			fynetooltip.DestroyPopUpToolTipLayer(pop)
		})
	}()
}

func (l *HoverLabel) MinSize() fyne.Size {
	s := l.label.MinSize()
	if s.Width > 10 {
		s.Width = 10
	}
	return s
}

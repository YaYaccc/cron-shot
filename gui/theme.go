package gui

import (
	"image/color"
	"os"
	"path/filepath"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type customTheme struct{}

var (
	customFont fyne.Resource
	loadOnce   sync.Once
)

func (t *customTheme) Color(n fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	return theme.DefaultTheme().Color(n, v)
}

func (t *customTheme) Font(s fyne.TextStyle) fyne.Resource {
	loadOnce.Do(func() {
		// 仅使用系统黑体 (SimHei) 以确保中文显示一致且稳定
		fontPath := filepath.Join(os.Getenv("SystemRoot"), "Fonts", "simhei.ttf")
		data, err := os.ReadFile(fontPath)
		if err != nil {
			// 备选：尝试硬编码路径
			fontPath = `C:\Windows\Fonts\simhei.ttf`
			data, err = os.ReadFile(fontPath)
		}

		if err == nil {
			customFont = fyne.NewStaticResource("simhei.ttf", data)
		}
	})

	if customFont != nil {
		return customFont
	}
	return theme.DefaultTheme().Font(s)
}

func (t *customTheme) Icon(n fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(n)
}

func (t *customTheme) Size(n fyne.ThemeSizeName) float32 {
	switch n {
	case theme.SizeNameText:
		return 18
	case theme.SizeNamePadding:
		return 8
	default:
		return theme.DefaultTheme().Size(n)
	}
}

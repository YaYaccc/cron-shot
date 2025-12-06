package gui

import (
	"cron-shot/assets"

	"fyne.io/fyne/v2"
)

func GetTrayIconResource() fyne.Resource {
	b, _ := assets.Files.ReadFile("icons/cat.png")
	return fyne.NewStaticResource("app.png", b)
}

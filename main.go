package main

import (
	_ "embed"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	"ytdlp-desktop/ui"
)

//go:embed logo/icon_256.png
var appIconData []byte

var compactWidget *ui.CompactWidget

func main() {
	a := app.NewWithID("ytdlp-desktop")
	a.SetIcon(&fyne.StaticResource{
		StaticName:    "icon.png",
		StaticContent: appIconData,
	})

	w := a.NewWindow("YTDLP Desktop")
	w.SetMaster()

	appUI := ui.NewApp(w)

	compactWidget = ui.NewCompactWidget(a, appUI)

	mainMenu := fyne.NewMainMenu(
		fyne.NewMenu("View",
			fyne.NewMenuItem("Toggle Widget", func() {
				if compactWidget != nil {
					compactWidget.Show()
				}
			}),
		),
	)
	w.SetMainMenu(mainMenu)

	w.Resize(fyne.NewSize(900, 700))
	w.SetContent(appUI.BuildUI())

	w.Show()

	a.Run()
}

package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	"ytdlp-desktop/ui"
)

var compactWidget *ui.CompactWidget

func main() {
	a := app.NewWithID("ytdlp-desktop")
	a.SetIcon(nil)

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

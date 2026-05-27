package ui

import (
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"ytdlp-desktop/internal/models"
)

type CompactWidget struct {
	window fyne.Window
	app    *App
}

func NewCompactWidget(parent fyne.App, a *App) *CompactWidget {
	w := parent.NewWindow("Drop URL")
	w.Resize(fyne.NewSize(320, 100))
	w.SetPadded(false)
	w.SetFixedSize(true)

	cw := &CompactWidget{window: w, app: a}

	bg := canvas.NewRectangle(&color.NRGBA{R: 20, G: 20, B: 40, A: 240})
	bg.CornerRadius = 8

	icon := canvas.NewText("⬇", &color.NRGBA{R: 255, G: 255, B: 255, A: 200})
	icon.TextSize = 24

	label := widget.NewLabel("Drop URL to download")
	label.Alignment = fyne.TextAlignCenter
	label.TextStyle = fyne.TextStyle{Italic: true}

	input := widget.NewEntry()
	input.SetPlaceHolder("Paste URL...")
	input.OnSubmitted = func(text string) {
		text = strings.TrimSpace(text)
		if text != "" {
			p := a.dropZone.presetSelector.SelectedPreset()
			if p == nil {
				p = &models.Preset{Format: "bestvideo+bestaudio/best"}
			}
			a.handleDropWithPreset(text, *p, FormatBuilderData{})
			input.SetText("")
		}
	}

	content := container.NewBorder(
		nil, nil,
		container.NewPadded(icon),
		nil,
		container.NewVBox(
			label,
			input,
		),
	)

	w.SetContent(container.NewStack(bg, container.NewPadded(content)))

	w.SetOnDropped(func(pos fyne.Position, uris []fyne.URI) {
		for _, uri := range uris {
			if uri != nil {
				url := strings.TrimSpace(uri.String())
				if url != "" && (strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")) {
					preset := a.dropZone.presetSelector.SelectedPreset()
					if preset != nil {
						a.handleDropWithPreset(url, *preset, FormatBuilderData{})
					}
				}
			}
		}
	})

	return cw
}

func (cw *CompactWidget) Show() {
	cw.window.Show()
}

func (cw *CompactWidget) Hide() {
	cw.window.Hide()
}

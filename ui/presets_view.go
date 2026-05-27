package ui

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"ytdlp-desktop/internal/models"
)

type PresetsView struct {
	widget.BaseWidget
	config     *models.Config
	onUpdate   func(cfg *models.Config)
	presetList *fyne.Container
}

func NewPresetsView(cfg *models.Config, onUpdate func(cfg *models.Config)) *PresetsView {
	pv := &PresetsView{
		config:   cfg,
		onUpdate: onUpdate,
	}
	pv.ExtendBaseWidget(pv)
	pv.refresh()
	return pv
}

func (pv *PresetsView) refresh() {
	pv.presetList = container.NewVBox()

	header := widget.NewLabelWithStyle("Download Presets", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	pv.presetList.Add(header)

	for i, p := range pv.config.Presets {
		card := pv.makePresetCard(i, p)
		pv.presetList.Add(card)
	}

	addBtn := widget.NewButton("+ New Preset", func() {
		newP := models.Preset{
			ID:        fmt.Sprintf("preset_%d", len(pv.config.Presets)+1),
			Name:      "New Preset",
			Format:    "bestvideo+bestaudio/best",
			AutoStart: true,
		}
		pv.config.Presets = append(pv.config.Presets, newP)
		pv.onUpdate(pv.config)
		pv.refresh()
	})
	pv.presetList.Add(container.NewPadded(addBtn))
}

func (pv *PresetsView) makePresetCard(index int, p models.Preset) *fyne.Container {
	nameEntry := widget.NewEntry()
	nameEntry.SetText(p.Name)
	nameEntry.OnChanged = func(s string) {
		pv.config.Presets[index].Name = s
	}

	outputEntry := widget.NewEntry()
	outputEntry.SetText(p.OutputDir)
	outputEntry.SetPlaceHolder("Default output directory")
	outputEntry.OnChanged = func(s string) {
		pv.config.Presets[index].OutputDir = s
	}

	autoCheck := widget.NewCheck("Auto-start on drop", func(b bool) {
		pv.config.Presets[index].AutoStart = b
	})
	autoCheck.SetChecked(p.AutoStart)

	limitEntry := widget.NewEntry()
	limitEntry.SetText(p.LimitRate)
	limitEntry.SetPlaceHolder("Rate limit (e.g. 5M)")
	limitEntry.OnChanged = func(s string) {
		pv.config.Presets[index].LimitRate = s
	}

	formatLabel := widget.NewLabel("Format: " + p.Format)
	formatLabel.Wrapping = fyne.TextWrapBreak
	formatLabel.TextStyle = fyne.TextStyle{Monospace: true}

	builderOpen := false
	builderBox := container.NewVBox()
	builderBox.Hide()

	data := ParsePresetToBuilderData(p)
	fb := NewFormatBuilder(data, func(format string) {
		pv.config.Presets[index].Format = format
		formatLabel.SetText("Format: " + format)
	})

	toggleBuilder := widget.NewButton("Edit format builder", nil)
	toggleBuilder.OnTapped = func() {
		builderOpen = !builderOpen
		if builderOpen {
			builderBox.Show()
			toggleBuilder.SetText("Hide format builder")
		} else {
			builderBox.Hide()
			toggleBuilder.SetText("Edit format builder")
		}
	}
	builderBox.Add(fb)

	rawEntry := widget.NewEntry()
	rawEntry.SetText(p.Format)
	rawEntry.OnChanged = func(s string) {
		pv.config.Presets[index].Format = s
		formatLabel.SetText("Format: " + s)
	}

	rawOpen := false
	rawBox := container.NewVBox(rawEntry)
	rawBox.Hide()

	toggleRaw := widget.NewButton("Edit raw format string", nil)
	toggleRaw.OnTapped = func() {
		rawOpen = !rawOpen
		if rawOpen {
			rawBox.Show()
			toggleRaw.SetText("Hide raw editor")
		} else {
			rawBox.Hide()
			toggleRaw.SetText("Edit raw format string")
		}
	}

	deleteBtn := widget.NewButton("Delete", func() {
		pv.config.Presets = append(pv.config.Presets[:index], pv.config.Presets[index+1:]...)
		pv.onUpdate(pv.config)
		pv.refresh()
	})
	deleteBtn.Importance = widget.DangerImportance

	saveBtn := widget.NewButton("Save", func() {
		pv.onUpdate(pv.config)
	})
	saveBtn.Importance = widget.HighImportance

	fields := container.NewVBox(
		container.NewBorder(nil, nil, widget.NewLabel("Name:"), nil, nameEntry),
		container.NewPadded(formatLabel),
		container.NewHBox(toggleBuilder, toggleRaw),
		builderBox,
		rawBox,
		container.NewBorder(nil, nil, widget.NewLabel("Output:"), nil, outputEntry),
		container.NewBorder(nil, nil, widget.NewLabel("Rate limit:"), nil, limitEntry),
		autoCheck,
		container.NewHBox(saveBtn, deleteBtn),
	)

	bg := canvas.NewRectangle(&color.NRGBA{R: 34, G: 34, B: 54, A: 220})
	bg.CornerRadius = 10
	card := container.NewStack(bg, container.NewPadded(fields))

	return card
}

func ParsePresetToBuilderData(p models.Preset) FormatBuilderData {
	data := FormatBuilderData{
		ExtractAudio: p.ExtractAudio,
		AudioFormat:  p.AudioFormat,
		Subs:         p.Subs,
		Thumbnail:    p.Thumbnail,
		LimitRate:    p.LimitRate,
		Resolution:   "Best",
		Codec:        "Auto",
		Container:    "Auto",
		AudioQuality: "Best",
	}

	data.RawMode = true
	data.RawFormat = p.Format

	return data
}

func (pv *PresetsView) CreateRenderer() fyne.WidgetRenderer {
	scroll := container.NewScroll(pv.presetList)
	return &presetsRenderer{scroll: scroll}
}

type presetsRenderer struct {
	scroll *container.Scroll
}

func (r *presetsRenderer) Layout(size fyne.Size) {
	r.scroll.Resize(size)
}

func (r *presetsRenderer) MinSize() fyne.Size {
	return fyne.NewSize(500, 300)
}

func (r *presetsRenderer) Refresh() {
	r.scroll.Refresh()
}

func (r *presetsRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.scroll}
}

func (r *presetsRenderer) Destroy() {}

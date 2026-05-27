package ui

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"ytdlp-desktop/internal/models"
)

type DropZone struct {
	entry          *widget.Entry
	downloadBtn    *widget.Button
	probeBtn       *widget.Button
	probeLoading   *widget.ProgressBarInfinite
	presetSelector *PresetSelector
	formatBuilder  *FormatBuilder
	customizeBtn   *widget.Button
	customizeOpen  bool
	customizeBox   *fyne.Container
	content        fyne.CanvasObject
	onDrop         func(url string, preset models.Preset, fbData FormatBuilderData)
	onProbe        func(url string)
}

func NewDropZone(
	presets []models.Preset,
	onDrop func(url string, preset models.Preset, fbData FormatBuilderData),
	onProbe func(url string),
) *DropZone {
	dz := &DropZone{
		onDrop:  onDrop,
		onProbe: onProbe,
	}

	dz.entry = widget.NewEntry()
	dz.entry.SetPlaceHolder("Paste video URL here...")
	dz.entry.OnSubmitted = func(text string) {
		dz.submit(text)
	}

	dz.presetSelector = NewPresetSelector(presets, "", func(p models.Preset) {
		if strings.HasPrefix(p.ID, "probed:") {
			data := dz.formatBuilder.GetData()
			data.RawMode = true
			data.RawFormat = p.Format
			dz.formatBuilder.SetData(data)
		}
	})

	dz.probeBtn = widget.NewButtonWithIcon("Probe", theme.SearchIcon(), func() {
		text := strings.TrimSpace(dz.entry.Text)
		if text != "" && (strings.HasPrefix(text, "http://") || strings.HasPrefix(text, "https://")) {
			if dz.onProbe != nil {
				dz.probeBtn.Hide()
				dz.probeLoading.Show()
				dz.probeLoading.Start()
				dz.onProbe(text)
			}
		}
	})
	dz.probeBtn.Importance = widget.MediumImportance

	dz.probeLoading = widget.NewProgressBarInfinite()
	dz.probeLoading.Hidden = true

	dz.downloadBtn = widget.NewButtonWithIcon("Download", theme.DownloadIcon(), func() {
		dz.submit(dz.entry.Text)
	})
	dz.downloadBtn.Importance = widget.HighImportance

	dz.formatBuilder = NewFormatBuilder(FormatBuilderData{
		Resolution:   "Best",
		Codec:        "Auto",
		Container:    "Auto",
		AudioQuality: "Best",
	}, nil)

	dz.customizeBtn = widget.NewButton("Customize format", func() {
		dz.customizeOpen = !dz.customizeOpen
		if dz.customizeOpen {
			dz.customizeBox.Show()
			dz.customizeBtn.SetText("Hide customization")
		} else {
			dz.customizeBox.Hide()
			dz.customizeBtn.SetText("Customize format")
		}
	})
	dz.customizeBox = container.NewVBox(dz.formatBuilder)
	dz.customizeBox.Hide()

	icon := widget.NewIcon(theme.DownloadIcon())
	label := widget.NewLabel("Drop YouTube / video URL here")
	label.Alignment = fyne.TextAlignCenter
	label.TextStyle = fyne.TextStyle{Bold: true}

	subtext := widget.NewLabel("Paste a URL or drag one from your browser")
	subtext.Alignment = fyne.TextAlignCenter
	subtext.TextStyle = fyne.TextStyle{Italic: true}

	top := container.NewVBox(
		container.NewCenter(icon),
		container.NewCenter(label),
		container.NewCenter(subtext),
	)

	inputRow := container.NewBorder(nil, nil, nil, container.NewHBox(dz.downloadBtn, dz.probeBtn, dz.probeLoading), dz.entry)

	bottom := container.NewVBox(
		inputRow,
		container.NewPadded(dz.presetSelector),
		container.NewCenter(dz.customizeBtn),
		container.NewPadded(dz.customizeBox),
	)

	dz.content = container.NewPadded(container.NewBorder(top, bottom, nil, nil))
	return dz
}

func (dz *DropZone) submit(text string) {
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}

	preset := dz.presetSelector.SelectedPreset()
	if preset == nil {
		preset = &models.Preset{
			ID:     "default",
			Name:   "Default",
			Format: "bestvideo+bestaudio/best",
		}
	}

	fbData := dz.formatBuilder.GetData()

	if dz.onDrop != nil {
		dz.onDrop(text, *preset, fbData)
	}
	dz.entry.SetText("")
}

func (dz *DropZone) SetProbeResults(url string, formats []models.FormatInfo, errMsg string) {
	dz.probeLoading.Stop()
	dz.probeLoading.Hide()
	dz.probeBtn.Show()

	if errMsg != "" || len(formats) == 0 {
		dz.presetSelector.ClearProbed()
		return
	}

	dz.presetSelector.SetProbedFormats(formats)
}

func (dz *DropZone) SetPresets(presets []models.Preset) {
	var selectedID string
	p := dz.presetSelector.SelectedPreset()
	if p != nil {
		selectedID = p.ID
	}
	dz.presetSelector.RefreshList(presets, selectedID)
}

func (dz *DropZone) Content() fyne.CanvasObject {
	return dz.content
}



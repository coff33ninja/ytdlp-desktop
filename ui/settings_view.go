package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type SettingsView struct {
	widget.BaseWidget
	config       *SettingsData
	onApply      func(cfg *SettingsData)
	maxConcEntry *widget.Entry
	ytdlpEntry   *widget.Entry
	themeSelect  *widget.Select
	formatEntry  *widget.Entry
	outputEntry  *widget.Entry
	audioCheck   *widget.Check
	subsCheck    *widget.Check
	thumbCheck   *widget.Check
	tabs         *container.AppTabs
}

type SettingsData struct {
	MaxConcurrent int
	YtdlpPath     string
	Theme         string
	Format        string
	OutputDir     string
	ExtractAudio  bool
	Subs          bool
	Thumbnail     bool
}

func NewSettingsView(cfg *SettingsData, onApply func(cfg *SettingsData)) *SettingsView {
	sv := &SettingsView{
		config:  cfg,
		onApply: onApply,
	}

	sv.maxConcEntry = widget.NewEntry()
	sv.maxConcEntry.SetText(intToStr(cfg.MaxConcurrent))

	sv.ytdlpEntry = widget.NewEntry()
	sv.ytdlpEntry.SetText(cfg.YtdlpPath)
	sv.ytdlpEntry.SetPlaceHolder("yt-dlp (auto-search PATH)")

	sv.themeSelect = widget.NewSelect([]string{"dark", "light"}, nil)
	sv.themeSelect.SetSelected(cfg.Theme)

	defaults := container.NewVBox(
		settingsRow("Max concurrent:", sv.maxConcEntry),
		settingsRow("yt-dlp path:", sv.ytdlpEntry),
		settingsRow("Theme:", sv.themeSelect),
	)

	applyBtn := widget.NewButton("Save Settings", func() {
		cfg.MaxConcurrent = strToInt(sv.maxConcEntry.Text, 2)
		cfg.YtdlpPath = sv.ytdlpEntry.Text
		cfg.Theme = sv.themeSelect.Selected
		cfg.Format = sv.formatEntry.Text
		cfg.OutputDir = sv.outputEntry.Text
		cfg.ExtractAudio = sv.audioCheck.Checked
		cfg.Subs = sv.subsCheck.Checked
		cfg.Thumbnail = sv.thumbCheck.Checked
		sv.onApply(cfg)
	})
	applyBtn.Importance = widget.HighImportance

	generalTab := container.NewVBox(
		widget.NewLabelWithStyle("General", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		defaults,
		container.NewPadded(applyBtn),
	)

	sv.formatEntry = widget.NewEntry()
	sv.formatEntry.SetText(cfg.Format)
	sv.formatEntry.SetPlaceHolder("bestvideo+bestaudio/best")

	sv.outputEntry = widget.NewEntry()
	sv.outputEntry.SetText(cfg.OutputDir)
	sv.outputEntry.SetPlaceHolder("%USERPROFILE%\\Downloads")

	sv.audioCheck = widget.NewCheck("Extract audio", nil)
	sv.audioCheck.SetChecked(cfg.ExtractAudio)

	sv.subsCheck = widget.NewCheck("Download subtitles", nil)
	sv.subsCheck.SetChecked(cfg.Subs)

	sv.thumbCheck = widget.NewCheck("Embed thumbnail", nil)
	sv.thumbCheck.SetChecked(cfg.Thumbnail)

	defaultsTab := container.NewVBox(
		widget.NewLabelWithStyle("Default Preset Options", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		settingsRow("Format string:", sv.formatEntry),
		settingsRow("Output directory:", sv.outputEntry),
		sv.audioCheck,
		sv.subsCheck,
		sv.thumbCheck,
	)

	sv.tabs = container.NewAppTabs(
		container.NewTabItem("General", container.NewPadded(generalTab)),
		container.NewTabItem("Defaults", container.NewPadded(defaultsTab)),
	)

	return sv
}

func settingsRow(label string, content fyne.CanvasObject) *fyne.Container {
	l := widget.NewLabel(label)
	return container.New(layout.NewGridLayout(2), l, content)
}

func (sv *SettingsView) CreateRenderer() fyne.WidgetRenderer {
	return &settingsRenderer{sv: sv}
}

type settingsRenderer struct {
	sv *SettingsView
}

func (r *settingsRenderer) Layout(size fyne.Size) {
	r.sv.tabs.Resize(size)
}

func (r *settingsRenderer) MinSize() fyne.Size {
	return fyne.NewSize(500, 300)
}

func (r *settingsRenderer) Refresh() {}

func (r *settingsRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.sv.tabs}
}

func (r *settingsRenderer) Destroy() {}

func intToStr(n int) string {
	if n == 0 {
		return "2"
	}
	return fmt.Sprintf("%d", n)
}

func strToInt(s string, def int) int {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return def
		}
		n = n*10 + int(c-'0')
	}
	if n < 1 {
		return def
	}
	return n
}

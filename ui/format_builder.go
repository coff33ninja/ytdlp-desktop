package ui

import (
	"fmt"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type FormatBuilderData struct {
	Resolution   string
	Codec        string
	Container    string
	AudioQuality string
	ExtractAudio bool
	AudioFormat  string
	Subs         bool
	Thumbnail    bool
	LimitRate    string
	RawMode      bool
	RawFormat    string
}

type FormatBuilder struct {
	widget.BaseWidget
	resolutionSelect *widget.Select
	codecSelect      *widget.Select
	containerSelect  *widget.Select
	audioSelect      *widget.Select
	extractCheck     *widget.Check
	audioFormatSel   *widget.Select
	subsCheck        *widget.Check
	thumbCheck       *widget.Check
	limitEntry       *widget.Entry
	rawCheck         *widget.Check
	rawEntry         *widget.Entry
	previewLabel     *widget.Label
	advancedBox      *fyne.Container
	content          *fyne.Container
	onChange         func(format string)
	onDataChange     func(data FormatBuilderData)
}

func NewFormatBuilder(data FormatBuilderData, onChange func(format string)) *FormatBuilder {
	fb := &FormatBuilder{
		onChange: onChange,
	}
	fb.ExtendBaseWidget(fb)

	if data.Resolution == "" {
		data.Resolution = "Best"
	}
	if data.Codec == "" {
		data.Codec = "Auto"
	}
	if data.Container == "" {
		data.Container = "Auto"
	}
	if data.AudioQuality == "" {
		data.AudioQuality = "Best"
	}
	if data.AudioFormat == "" {
		data.AudioFormat = "mp3"
	}

	fb.resolutionSelect = widget.NewSelect([]string{"Best", "2160p (4K)", "1440p", "1080p", "720p", "480p", "360p", "240p", "144p"}, nil)
	fb.resolutionSelect.SetSelected(data.Resolution)

	fb.codecSelect = widget.NewSelect([]string{"Auto", "H.264 (AVC)", "H.265 (HEVC)", "AV1", "VP9"}, nil)
	fb.codecSelect.SetSelected(data.Codec)

	fb.containerSelect = widget.NewSelect([]string{"Auto", "mp4", "mkv", "webm"}, nil)
	fb.containerSelect.SetSelected(data.Container)

	fb.audioSelect = widget.NewSelect([]string{"Best", "AAC", "MP3", "Opus", "FLAC", "WAV"}, nil)
	fb.audioSelect.SetSelected(data.AudioQuality)

	fb.extractCheck = widget.NewCheck("Extract audio", nil)
	fb.extractCheck.SetChecked(data.ExtractAudio)

	fb.audioFormatSel = widget.NewSelect([]string{"mp3", "aac", "flac", "opus", "wav", "m4a"}, nil)
	fb.audioFormatSel.SetSelected(data.AudioFormat)
	if !data.ExtractAudio {
		fb.audioFormatSel.Disable()
	}

	fb.subsCheck = widget.NewCheck("Download subtitles", nil)
	fb.subsCheck.SetChecked(data.Subs)

	fb.thumbCheck = widget.NewCheck("Embed thumbnail", nil)
	fb.thumbCheck.SetChecked(data.Thumbnail)

	fb.limitEntry = widget.NewEntry()
	fb.limitEntry.SetText(data.LimitRate)
	fb.limitEntry.SetPlaceHolder("Rate limit (e.g. 5M)")

	fb.rawCheck = widget.NewCheck("Raw format string (advanced)", nil)
	fb.rawCheck.SetChecked(data.RawMode)

	fb.rawEntry = widget.NewEntry()
	fb.rawEntry.SetText(data.RawFormat)
	fb.rawEntry.SetPlaceHolder("bestvideo+bestaudio/best")
	if !data.RawMode {
		fb.rawEntry.Disable()
	}

	fb.previewLabel = widget.NewLabel("")
	fb.previewLabel.Wrapping = fyne.TextWrapBreak
	fb.previewLabel.TextStyle = fyne.TextStyle{Monospace: true}

	fb.rawCheck.OnChanged = func(b bool) {
		if b {
			fb.rawEntry.SetText(fb.buildFormatString())
			fb.rawEntry.Enable()
		} else {
			fb.rawEntry.Disable()
		}
		fb.emitChange()
	}
	fb.rawEntry.OnChanged = func(s string) {
		if fb.rawCheck.Checked {
			fb.updatePreview()
		}
		fb.emitChange()
	}
	fb.limitEntry.OnChanged = func(s string) {
		fb.emitChange()
	}
	fb.resolutionSelect.OnChanged = func(s string) { fb.emitChange() }
	fb.codecSelect.OnChanged = func(s string) { fb.emitChange() }
	fb.containerSelect.OnChanged = func(s string) { fb.emitChange() }
	fb.audioSelect.OnChanged = func(s string) { fb.emitChange() }
	fb.extractCheck.OnChanged = func(b bool) {
		if b {
			fb.audioFormatSel.Enable()
		} else {
			fb.audioFormatSel.Disable()
		}
		fb.emitChange()
	}
	fb.audioFormatSel.OnChanged = func(s string) { fb.emitChange() }
	fb.subsCheck.OnChanged = func(b bool) { fb.emitChange() }
	fb.thumbCheck.OnChanged = func(b bool) { fb.emitChange() }

	fb.advancedBox = container.NewVBox(
		widget.NewLabel("Advanced"),
		fb.limitEntry,
		fb.rawCheck,
		fb.rawEntry,
	)

	basic := container.NewVBox(
		widget.NewLabelWithStyle("Format Builder", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		settingsRow("Resolution:", fb.resolutionSelect),
		settingsRow("Video codec:", fb.codecSelect),
		settingsRow("Container:", fb.containerSelect),
		settingsRow("Audio:", fb.audioSelect),
		fb.extractCheck,
		settingsRow("Audio format:", fb.audioFormatSel),
		fb.subsCheck,
		fb.thumbCheck,
	)

	fb.content = container.NewVBox(
		basic,
		fb.advancedBox,
		container.NewPadded(widget.NewLabelWithStyle("Preview:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})),
		container.NewPadded(fb.previewLabel),
	)

	fb.updatePreview()

	return fb
}

func (fb *FormatBuilder) buildFormatString() string {
	if fb.rawCheck.Checked {
		return fb.rawEntry.Text
	}

	res := fb.resolutionSelect.Selected
	codec := fb.codecSelect.Selected
	container := fb.containerSelect.Selected
	audio := fb.audioSelect.Selected

	heightFilter := ""
	switch {
	case strings.HasPrefix(res, "2160"):
		heightFilter = "[height<=2160]"
	case strings.HasPrefix(res, "1440"):
		heightFilter = "[height<=1440]"
	case strings.HasPrefix(res, "1080"):
		heightFilter = "[height<=1080]"
	case strings.HasPrefix(res, "720"):
		heightFilter = "[height<=720]"
	case strings.HasPrefix(res, "480"):
		heightFilter = "[height<=480]"
	case strings.HasPrefix(res, "360"):
		heightFilter = "[height<=360]"
	case strings.HasPrefix(res, "240"):
		heightFilter = "[height<=240]"
	case strings.HasPrefix(res, "144"):
		heightFilter = "[height<=144]"
	}

	codecFilter := ""
	switch codec {
	case "H.264 (AVC)":
		codecFilter = "[vcodec~='^((he|a)vc|h26[45])']"
	case "H.265 (HEVC)":
		codecFilter = "[vcodec~='^hevc']"
	case "AV1":
		codecFilter = "[vcodec~='^av01']"
	case "VP9":
		codecFilter = "[vcodec~='^vp0?9']"
	}

	containerFilter := ""
	if container != "Auto" {
		containerFilter = "[ext=" + container + "]"
	}

	audioFilter := ""
	switch audio {
	case "AAC":
		audioFilter = "[acodec~='^aac']"
	case "MP3":
		audioFilter = "[acodec~='^mp3']"
	case "Opus":
		audioFilter = "[acodec~='^opus']"
	case "FLAC":
		audioFilter = "[acodec~='^flac']"
	case "WAV":
		audioFilter = "[acodec~='^wav']"
	}

	videoSel := "bv*"
	audioSel := "ba"

	filters := heightFilter + codecFilter + containerFilter

	if filters != "" {
		videoSel = "bv*" + filters
	}

	if audioFilter != "" {
		audioSel = "ba" + audioFilter
	}

	formatStr := videoSel + "+" + audioSel

	if filters != "" || audioFilter != "" {
		formatStr += " / bv*+ba/b"
	} else {
		formatStr += "/b"
	}

	return formatStr
}

func (fb *FormatBuilder) updatePreview() {
	s := fb.buildFormatString()
	if s == "" {
		fb.previewLabel.SetText("(no format selected)")
	} else {
		fb.previewLabel.SetText("-f \"" + s + "\"")
	}
}

func (fb *FormatBuilder) emitChange() {
	fb.updatePreview()
	if fb.onChange != nil {
		fb.onChange(fb.buildFormatString())
	}
	if fb.onDataChange != nil {
		fb.onDataChange(fb.GetData())
	}
}

func (fb *FormatBuilder) GetData() FormatBuilderData {
	return FormatBuilderData{
		Resolution:   fb.resolutionSelect.Selected,
		Codec:        fb.codecSelect.Selected,
		Container:    fb.containerSelect.Selected,
		AudioQuality: fb.audioSelect.Selected,
		ExtractAudio: fb.extractCheck.Checked,
		AudioFormat:  fb.audioFormatSel.Selected,
		Subs:         fb.subsCheck.Checked,
		Thumbnail:    fb.thumbCheck.Checked,
		LimitRate:    fb.limitEntry.Text,
		RawMode:      fb.rawCheck.Checked,
		RawFormat:    fb.rawEntry.Text,
	}
}

func (fb *FormatBuilder) SetData(data FormatBuilderData) {
	fb.resolutionSelect.SetSelected(data.Resolution)
	fb.codecSelect.SetSelected(data.Codec)
	fb.containerSelect.SetSelected(data.Container)
	fb.audioSelect.SetSelected(data.AudioQuality)
	fb.extractCheck.SetChecked(data.ExtractAudio)
	fb.audioFormatSel.SetSelected(data.AudioFormat)
	fb.subsCheck.SetChecked(data.Subs)
	fb.thumbCheck.SetChecked(data.Thumbnail)
	fb.limitEntry.SetText(data.LimitRate)
	fb.rawCheck.SetChecked(data.RawMode)
	fb.rawEntry.SetText(data.RawFormat)
	if data.RawMode {
		fb.rawEntry.Enable()
	} else {
		fb.rawEntry.Disable()
	}
	fb.updatePreview()
}

func (fb *FormatBuilder) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(fb.content)
}

func formatFileSize(bytes float64) string {
	switch {
	case bytes >= 1_000_000_000:
		return fmt.Sprintf("%.1f GB", bytes/1_000_000_000)
	case bytes >= 1_000_000:
		return fmt.Sprintf("%.1f MB", bytes/1_000_000)
	case bytes >= 1_000:
		return fmt.Sprintf("%.1f KB", bytes/1_000)
	default:
		return fmt.Sprintf("%.0f B", bytes)
	}
}

func parseResolution(data FormatBuilderData) int {
	switch {
	case strings.HasPrefix(data.Resolution, "2160"):
		return 2160
	case strings.HasPrefix(data.Resolution, "1440"):
		return 1440
	case strings.HasPrefix(data.Resolution, "1080"):
		return 1080
	case strings.HasPrefix(data.Resolution, "720"):
		return 720
	case strings.HasPrefix(data.Resolution, "480"):
		return 480
	case strings.HasPrefix(data.Resolution, "360"):
		return 360
	case strings.HasPrefix(data.Resolution, "240"):
		return 240
	case strings.HasPrefix(data.Resolution, "144"):
		return 144
	default:
		return 99999
	}
}

func parseBitrate(s string) int {
	s = strings.TrimSpace(strings.ToUpper(s))
	if s == "" {
		return 0
	}
	mult := 1
	if strings.HasSuffix(s, "M") {
		mult = 1000000
		s = strings.TrimSuffix(s, "M")
	} else if strings.HasSuffix(s, "K") {
		mult = 1000
		s = strings.TrimSuffix(s, "K")
	}
	v, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return 0
	}
	return v * mult
}

func formatBitrate(bps float64) string {
	switch {
	case bps >= 1_000_000:
		return fmt.Sprintf("%.0f K", bps/1000)
	case bps >= 1_000:
		return fmt.Sprintf("%.0f K", bps/1000)
	default:
		return fmt.Sprintf("%.0f", bps)
	}
}

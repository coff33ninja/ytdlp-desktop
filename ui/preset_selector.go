package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"ytdlp-desktop/internal/models"
)

type PresetSelector struct {
	widget.BaseWidget
	selectWidget *widget.Select
	presets      []models.Preset
	probed       []models.Preset
	onChange     func(preset models.Preset)
	content      *fyne.Container
}

func NewPresetSelector(presets []models.Preset, selected string, onChange func(preset models.Preset)) *PresetSelector {
	ps := &PresetSelector{
		presets:  presets,
		onChange: onChange,
	}
	ps.ExtendBaseWidget(ps)

	names := ps.allNames()
	selectedName := ps.findName(selected)
	if selectedName == "" && len(names) > 0 {
		selectedName = names[0]
	}

	ps.selectWidget = widget.NewSelect(names, func(s string) {
		for _, list := range [][]models.Preset{ps.presets, ps.probed} {
			for _, p := range list {
				if p.Name == s {
					if ps.onChange != nil {
						ps.onChange(p)
					}
					return
				}
			}
		}
	})
	ps.selectWidget.SetSelected(selectedName)

	label := widget.NewLabel("Preset:")
	ps.content = container.NewBorder(nil, nil, label, nil, ps.selectWidget)

	return ps
}

func (ps *PresetSelector) RefreshList(presets []models.Preset, selectedID string) {
	ps.presets = presets
	ps.selectWidget.Options = ps.allNames()
	selectedName := ps.findName(selectedID)
	if selectedName == "" && len(presets) > 0 {
		selectedName = presets[0].Name
	}
	ps.selectWidget.SetSelected(selectedName)
	ps.BaseWidget.Refresh()
}

func (ps *PresetSelector) SetProbedFormats(formats []models.FormatInfo) {
	probed := make([]models.Preset, 0, len(formats))
	for _, f := range formats {
		resLine := f.FormatNote
		if resLine == "" {
			if f.Height > 0 {
				resLine = fmt.Sprintf("%dp", f.Height)
			} else {
				resLine = "audio"
			}
		}
		codec := f.VCodec
		if codec == "none" || codec == "" {
			codec = f.ACodec
		}
		extra := ""
		if f.Filesize > 0 {
			extra = formatFileSize(f.Filesize)
		} else if f.FilesizeApprox > 0 {
			extra = "~" + formatFileSize(f.FilesizeApprox)
		}
		if f.TBR > 0 {
			if extra != "" {
				extra += " "
			}
			extra += formatBitrate(f.TBR)
		}
		name := fmt.Sprintf("%s #%s %s %s", resLine, f.FormatID, codec, f.Ext)
		if extra != "" {
			name += " (" + extra + ")"
		}
		if len(name) > 60 {
			name = name[:60]
		}
		probed = append(probed, models.Preset{
			ID:     "probed:" + f.FormatID,
			Name:   name,
			Format: f.FormatID,
		})
	}
	ps.probed = probed
	ps.selectWidget.Options = ps.allNames()
	if len(probed) > 0 {
		ps.selectWidget.SetSelected(probed[0].Name)
	}
	ps.BaseWidget.Refresh()
}

func (ps *PresetSelector) ClearProbed() {
	ps.probed = nil
	ps.selectWidget.Options = ps.allNames()
	if len(ps.presets) > 0 {
		ps.selectWidget.SetSelected(ps.presets[0].Name)
	}
	ps.BaseWidget.Refresh()
}

func (ps *PresetSelector) SetSelected(id string) {
	name := ps.findName(id)
	if name != "" {
		ps.selectWidget.SetSelected(name)
	}
}

func (ps *PresetSelector) SelectedPreset() *models.Preset {
	name := ps.selectWidget.Selected
	for _, list := range [][]models.Preset{ps.presets, ps.probed} {
		for i, p := range list {
			if p.Name == name {
				return &list[i]
			}
		}
	}
	if len(ps.presets) > 0 {
		return &ps.presets[0]
	}
	return nil
}

func (ps *PresetSelector) allNames() []string {
	names := make([]string, 0, len(ps.presets)+len(ps.probed))
	for _, p := range ps.presets {
		names = append(names, p.Name)
	}
	if len(ps.probed) > 0 {
		names = append(names, "── Probed ──")
		for _, p := range ps.probed {
			names = append(names, p.Name)
		}
	}
	return names
}

func (ps *PresetSelector) findName(id string) string {
	for _, list := range [][]models.Preset{ps.presets, ps.probed} {
		for _, p := range list {
			if p.Name == id || p.ID == id {
				return p.Name
			}
		}
	}
	return ""
}

func (ps *PresetSelector) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(ps.content)
}

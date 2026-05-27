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

type HistoryView struct {
	widget.BaseWidget
	entries []models.HistoryEntry
	list    *fyne.Container
}

func NewHistoryView() *HistoryView {
	hv := &HistoryView{
		list: container.NewVBox(),
	}
	hv.ExtendBaseWidget(hv)
	return hv
}

func (hv *HistoryView) Load(entries []models.HistoryEntry) {
	hv.entries = entries
	hv.refresh()
}

func (hv *HistoryView) refresh() {
	hv.list.Objects = nil

	header := widget.NewLabelWithStyle("Download History", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	hv.list.Add(header)

	if len(hv.entries) == 0 {
		empty := widget.NewLabel("No downloads yet")
		empty.Alignment = fyne.TextAlignCenter
		empty.TextStyle = fyne.TextStyle{Italic: true}
		hv.list.Add(empty)
	} else {
		start := 0
		if len(hv.entries) > 50 {
			start = len(hv.entries) - 50
		}
		for i := len(hv.entries) - 1; i >= start; i-- {
			hv.list.Add(hv.makeEntryCard(hv.entries[i]))
		}
	}
}

func (hv *HistoryView) makeEntryCard(e models.HistoryEntry) *fyne.Container {
	title := widget.NewLabelWithStyle(e.Title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	title.Wrapping = fyne.TextWrapBreak

	meta := widget.NewLabel(fmt.Sprintf("%s  •  %s", e.Format, e.CompletedAt.Format("Jan 2 15:04")))
	meta.TextStyle = fyne.TextStyle{Italic: true}

	statusColor := &color.NRGBA{R: 102, G: 255, B: 102, A: 255}
	statusTxt := "✓ Done"
	if !e.Success {
		statusColor = &color.NRGBA{R: 255, G: 102, B: 102, A: 255}
		statusTxt = "✗ Failed"
	}
	status := canvas.NewText(statusTxt, statusColor)
	status.TextSize = 13

	bg := canvas.NewRectangle(&color.NRGBA{R: 34, G: 34, B: 54, A: 220})
	bg.CornerRadius = 10

	topRow := container.NewBorder(nil, nil, nil, status, title)
	body := container.NewVBox(topRow, meta)
	return container.NewStack(bg, container.NewPadded(body))
}

func (hv *HistoryView) CreateRenderer() fyne.WidgetRenderer {
	scroll := container.NewScroll(hv.list)
	return &historyRenderer{scroll: scroll}
}

type historyRenderer struct {
	scroll *container.Scroll
}

func (r *historyRenderer) Layout(size fyne.Size) {
	r.scroll.Resize(size)
}

func (r *historyRenderer) MinSize() fyne.Size {
	return fyne.NewSize(500, 300)
}

func (r *historyRenderer) Refresh() {
	r.scroll.Refresh()
}

func (r *historyRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.scroll}
}

func (r *historyRenderer) Destroy() {}

package ui

import (
	"image/color"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"ytdlp-desktop/internal/models"
	"ytdlp-desktop/internal/queue"
)

type QueueView struct {
	widget.BaseWidget
	mu           sync.RWMutex
	items        []*models.DownloadItem
	queueManager *queue.Manager
	listWidget   *widget.List
	emptyLabel   *widget.Label
	emptyBox     *fyne.Container
	content      *fyne.Container
}

const (
	cardCornerRadius = 10
)

var (
	cardBgColor    = &color.NRGBA{R: 34, G: 34, B: 54, A: 220}
	dimTextColor   = &color.NRGBA{R: 153, G: 153, B: 170, A: 255}
	runningColor   = &color.NRGBA{R: 102, G: 187, B: 255, A: 255}
	errorColor     = &color.NRGBA{R: 255, G: 102, B: 102, A: 255}
	completeColor  = &color.NRGBA{R: 102, G: 255, B: 102, A: 255}
	waitingColor   = &color.NRGBA{R: 170, G: 170, B: 204, A: 255}
	defaultColor   = &color.NRGBA{R: 136, G: 204, B: 136, A: 255}
)

func NewQueueView(qm *queue.Manager) *QueueView {
	qv := &QueueView{
		queueManager: qm,
	}

	qv.emptyLabel = widget.NewLabel("No downloads yet")
	qv.emptyLabel.Alignment = fyne.TextAlignCenter
	qv.emptyLabel.TextStyle = fyne.TextStyle{Italic: true}
	qv.emptyBox = container.NewCenter(qv.emptyLabel)

	qv.listWidget = widget.NewList(
		func() int {
			qv.mu.RLock()
			defer qv.mu.RUnlock()
			return len(qv.items)
		},
		func() fyne.CanvasObject {
			return createCardTemplate()
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			qv.mu.RLock()
			item := qv.items[id]
			qv.mu.RUnlock()
			updateCardRow(obj, item)
		},
	)

	qv.ExtendBaseWidget(qv)
	return qv
}

func createCardTemplate() fyne.CanvasObject {
	title := widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	title.Wrapping = fyne.TextWrapBreak

	status := canvas.NewText("", defaultColor)
	status.TextSize = 13

	meta := canvas.NewText("", dimTextColor)
	meta.TextSize = 12

	presetLabel := canvas.NewText("", dimTextColor)
	presetLabel.TextSize = 11

	progress := widget.NewProgressBar()

	header := container.NewBorder(nil, nil, nil, status, title)
	cardBox := container.NewVBox(header, presetLabel, meta, progress)

	bg := canvas.NewRectangle(cardBgColor)
	bg.CornerRadius = cardCornerRadius
	outer := container.NewStack(bg, container.NewPadded(cardBox))

	return outer
}

func updateCardRow(obj fyne.CanvasObject, item *models.DownloadItem) {
	stack := obj.(*fyne.Container)
	padded := stack.Objects[1].(*fyne.Container)
	cardBox := padded.Objects[0].(*fyne.Container)
	header := cardBox.Objects[0].(*fyne.Container)
	presetLabel := cardBox.Objects[1].(*canvas.Text)
	meta := cardBox.Objects[2].(*canvas.Text)
	progress := cardBox.Objects[3].(*widget.ProgressBar)
	title := header.Objects[0].(*widget.Label)
	status := header.Objects[1].(*canvas.Text)

	title.SetText(itemTitle(item))
	status.Text = string(item.State)
	status.Color = statusColorFor(item.State)
	meta.Text = metaText(item)
	presetLabel.Text = presetText(item)
	progress.Value = item.Progress
}

func (qv *QueueView) ProcessUpdate(up queue.Update) {
	switch up.Type {
	case "added":
		qv.mu.Lock()
		qv.items = qv.queueManager.Items()
		hasItems := len(qv.items) > 0
		qv.mu.Unlock()
		fyne.Do(func() {
			qv.listWidget.Refresh()
			qv.toggleEmpty(hasItems)
		})
	case "removed":
		qv.mu.Lock()
		qv.items = qv.queueManager.Items()
		hasItems := len(qv.items) > 0
		qv.mu.Unlock()
		fyne.Do(func() {
			qv.listWidget.Refresh()
			qv.toggleEmpty(hasItems)
		})
	case "progress", "info", "done":
		if up.Item == nil {
			return
		}
		qv.mu.Lock()
		idx := qv.indexOf(up.Item.ID)
		if idx >= 0 {
			clone := *up.Item
			qv.items[idx] = &clone
		}
		qv.mu.Unlock()
		if idx >= 0 {
			fyne.Do(func() {
				qv.listWidget.RefreshItem(idx)
			})
		}
	}
}

func (qv *QueueView) toggleEmpty(hasItems bool) {
	if hasItems {
		qv.emptyBox.Hide()
		qv.listWidget.Show()
	} else {
		qv.emptyBox.Show()
		qv.listWidget.Hide()
	}
}

func (qv *QueueView) indexOf(id string) int {
	for i, item := range qv.items {
		if item.ID == id {
			return i
		}
	}
	return -1
}

func itemTitle(item *models.DownloadItem) string {
	if item.Title != "" {
		return item.Title
	}
	t := item.URL
	if len(t) > 60 {
		t = t[:60] + "..."
	}
	return t
}

func statusColorFor(state models.DownloadState) color.Color {
	switch state {
	case models.StateRunning:
		return runningColor
	case models.StateError:
		return errorColor
	case models.StateComplete:
		return completeColor
	case models.StateWaiting:
		return waitingColor
	default:
		return defaultColor
	}
}

func metaText(item *models.DownloadItem) string {
	if item.Speed != "" {
		s := item.Speed
		if item.ETA != "" {
			s += "  ·  " + item.ETA
		}
		return s
	}
	return ""
}

func presetText(item *models.DownloadItem) string {
	if item.Format != "" {
		return item.Format
	}
	if item.PresetID != "" {
		return item.PresetID
	}
	return ""
}

func (qv *QueueView) CreateRenderer() fyne.WidgetRenderer {
	header := widget.NewLabelWithStyle("Downloads", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	stack := container.NewStack(qv.emptyBox, qv.listWidget)
	qv.content = container.NewBorder(container.NewPadded(header), nil, nil, nil, stack)
	return widget.NewSimpleRenderer(qv.content)
}

func (qv *QueueView) MinSize() fyne.Size {
	return fyne.NewSize(500, 300)
}




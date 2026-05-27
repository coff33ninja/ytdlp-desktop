package queue

import (
	"context"
	"sync"

	"ytdlp-desktop/internal/downloader"
	"ytdlp-desktop/internal/models"
)

type Update struct {
	Type   string
	ItemID string
	Item   *models.DownloadItem
}

type Manager struct {
	mu         sync.RWMutex
	items      []*models.DownloadItem
	active     int
	dl         *downloader.Downloader
	maxConcurrent int
	updates    chan Update
	ctx        context.Context
	cancel     context.CancelFunc
}

func New(dl *downloader.Downloader, maxConcurrent int) *Manager {
	ctx, cancel := context.WithCancel(context.Background())
	return &Manager{
		dl:            dl,
		maxConcurrent: maxConcurrent,
		updates:       make(chan Update, 100),
		ctx:           ctx,
		cancel:        cancel,
	}
}

func (qm *Manager) Updates() <-chan Update {
	return qm.updates
}

func (qm *Manager) Add(item *models.DownloadItem) {
	qm.mu.Lock()
	qm.items = append(qm.items, item)
	qm.mu.Unlock()

	qm.updates <- Update{Type: "added", ItemID: item.ID, Item: item}
	qm.tryDispatch()
}

func (qm *Manager) Remove(id string) {
	qm.mu.Lock()
	defer qm.mu.Unlock()
	for i, item := range qm.items {
		if item.ID == id {
			qm.items = append(qm.items[:i], qm.items[i+1:]...)
			qm.updates <- Update{Type: "removed", ItemID: id}
			return
		}
	}
}

func (qm *Manager) Items() []*models.DownloadItem {
	qm.mu.RLock()
	defer qm.mu.RUnlock()
	cp := make([]*models.DownloadItem, len(qm.items))
	for i, item := range qm.items {
		clone := *item
		cp[i] = &clone
	}
	return cp
}

func (qm *Manager) StopAll() {
	qm.cancel()
}

func (qm *Manager) SetMaxConcurrent(n int) {
	qm.mu.Lock()
	qm.maxConcurrent = n
	qm.mu.Unlock()
	qm.tryDispatch()
}

func (qm *Manager) tryDispatch() {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	for _, item := range qm.items {
		if item.State != models.StateWaiting {
			continue
		}
		if qm.active >= qm.maxConcurrent {
			return
		}
		qm.active++

		go func(it *models.DownloadItem) {
			qm.dl.Download(qm.ctx, it,
				func(id string, progress float64, speed, eta string) {
					qm.mu.Lock()
					it.Progress = progress
					it.Speed = speed
					it.ETA = eta
					qm.mu.Unlock()
					qm.updates <- Update{Type: "progress", ItemID: id, Item: it}
				},
				func(id string, title string) {
					qm.mu.Lock()
					it.Title = title
					qm.mu.Unlock()
					qm.updates <- Update{Type: "info", ItemID: id, Item: it}
				},
				func(id string, err error) {
					qm.mu.Lock()
					qm.active--
					if err != nil {
						it.State = models.StateError
						it.Error = err.Error()
					} else {
						it.State = models.StateComplete
						it.Progress = 1.0
					}
					qm.mu.Unlock()
					qm.updates <- Update{Type: "done", ItemID: id, Item: it}
					qm.tryDispatch()
				},
			)
		}(item)
	}
}

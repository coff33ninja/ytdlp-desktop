package ui

import (
	"context"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"

	"ytdlp-desktop/internal/config"
	"ytdlp-desktop/internal/downloader"
	"ytdlp-desktop/internal/logger"
	"ytdlp-desktop/internal/models"
	"ytdlp-desktop/internal/queue"
)

type App struct {
	window       fyne.Window
	config       *models.Config
	dl           *downloader.Downloader
	queueMgr     *queue.Manager
	dropZone     *DropZone
	queueView    *QueueView
	historyView  *HistoryView
	settingsView *SettingsView
	presetsView  *PresetsView
}

func NewApp(w fyne.Window) *App {
	a := &App{window: w}

	a.config = config.LoadConfig()
	a.dl = downloader.New()
	a.queueMgr = queue.New(a.dl, a.config.MaxConcurrent)

	if err := a.dl.EnsureInstalled(context.Background()); err != nil {
		logger.Error("failed to install yt-dlp: %v", err)
	}

	a.dropZone = NewDropZone(
		a.config.Presets,
		a.handleDropWithPreset,
		a.handleProbe,
	)

	a.queueView = NewQueueView(a.queueMgr)
	a.historyView = NewHistoryView()

	settingsData := &SettingsData{
		MaxConcurrent: a.config.MaxConcurrent,
		YtdlpPath:     a.config.YtdlpPath,
		Theme:         a.config.Theme,
	}
	if len(a.config.Presets) > 0 {
		p := a.config.Presets[0]
		settingsData.Format = p.Format
		settingsData.OutputDir = p.OutputDir
		settingsData.ExtractAudio = p.ExtractAudio
		settingsData.Subs = p.Subs
		settingsData.Thumbnail = p.Thumbnail
	} else {
		settingsData.Format = "bestvideo+bestaudio/best"
	}

	a.settingsView = NewSettingsView(settingsData, a.applySettings)
	a.presetsView = NewPresetsView(a.config, a.savePresets)

	loadHistory := config.LoadHistory()
	if loadHistory == nil {
		loadHistory = []models.HistoryEntry{}
	}
	a.historyView.Load(loadHistory)

	go a.listenUpdates()

	w.SetOnDropped(func(pos fyne.Position, uris []fyne.URI) {
		for _, uri := range uris {
			if uri != nil {
				url := strings.TrimSpace(uri.String())
				if url != "" && (strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")) {
					selectedPreset := a.dropZone.presetSelector.SelectedPreset()
					if selectedPreset != nil {
						a.handleDropWithPreset(url, *selectedPreset, FormatBuilderData{})
					} else {
						a.handleDropWithPreset(url, models.Preset{Format: "bestvideo+bestaudio/best"}, FormatBuilderData{})
					}
				}
			}
		}
	})

	return a
}

func (a *App) BuildUI() fyne.CanvasObject {
	downloadsTab := container.NewBorder(
		a.dropZone.Content(),
		nil, nil, nil,
		a.queueView,
	)

	tabs := container.NewAppTabs(
		container.NewTabItemWithIcon("Downloads", theme.DownloadIcon(), downloadsTab),
		container.NewTabItemWithIcon("Presets", theme.SettingsIcon(), a.presetsView),
		container.NewTabItemWithIcon("History", theme.HistoryIcon(), a.historyView),
		container.NewTabItemWithIcon("Settings", theme.SettingsIcon(), a.settingsView),
	)
	tabs.SetTabLocation(container.TabLocationTop)

	return tabs
}

func (a *App) handleDropWithPreset(text string, preset models.Preset, fbData FormatBuilderData) {
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}

	urls := extractURLs(text)
	for _, url := range urls {
		a.enqueueDownload(url, preset, fbData)
	}
}

func extractURLs(text string) []string {
	var urls []string
	lines := strings.Fields(text)
	for _, line := range lines {
		line = strings.Trim(line, `"'`)
		if strings.HasPrefix(line, "http://") || strings.HasPrefix(line, "https://") {
			urls = append(urls, line)
		}
	}
	if len(urls) == 0 && (strings.HasPrefix(text, "http://") || strings.HasPrefix(text, "https://")) {
		urls = append(urls, text)
	}
	return urls
}

func (a *App) enqueueDownload(url string, preset models.Preset, fbData FormatBuilderData) {
	format := preset.Format
	if fbData.RawMode && fbData.RawFormat != "" {
		format = fbData.RawFormat
	} else if !fbData.RawMode {
		builderFormat := a.dropZone.formatBuilder.buildFormatString()
		if builderFormat != "" && builderFormat != "bestvideo+bestaudio/best/b" {
			format = builderFormat
		}
	}

	item := &models.DownloadItem{
		ID:           time.Now().Format("150405.000000"),
		URL:          url,
		State:        models.StateWaiting,
		Format:       format,
		OutputDir:    preset.OutputDir,
		PresetID:     preset.ID,
		ExtractAudio: preset.ExtractAudio || fbData.ExtractAudio,
		AudioFormat:  preset.AudioFormat,
		Subs:         preset.Subs || fbData.Subs,
		Thumbnail:    preset.Thumbnail || fbData.Thumbnail,
		CreatedAt:    time.Now(),
	}

	if fbData.ExtractAudio && fbData.AudioFormat != "" {
		item.AudioFormat = fbData.AudioFormat
	}

	go func() {
		title, err := a.dl.FetchInfo(context.Background(), url)
		if err != nil {
			logger.Error("fetch info: url=%q err=%v", url, err)
		} else {
			item.Title = title
		}
	}()

	a.queueMgr.Add(item)
}

func (a *App) handleProbe(url string) {
	go func() {
		formats, err := a.dl.GetFormatsJSON(context.Background(), url)
		fyne.Do(func() {
			if err != nil {
				logger.Error("probe formats: url=%q err=%v", url, err)
				a.dropZone.SetProbeResults(url, nil, err.Error())
				return
			}
			a.dropZone.SetProbeResults(url, formats, "")
		})
	}()
}

func (a *App) applySettings(cfg *SettingsData) {
	a.config.MaxConcurrent = cfg.MaxConcurrent
	a.config.YtdlpPath = cfg.YtdlpPath
	a.config.Theme = cfg.Theme
	s := a.window.Canvas().Size()
	a.config.WindowWidth = int(s.Width)
	a.config.WindowHeight = int(s.Height)
	if err := config.SaveConfig(a.config); err != nil {
		logger.Error("save config: %v", err)
	}
	a.queueMgr.SetMaxConcurrent(cfg.MaxConcurrent)

	if len(a.config.Presets) > 0 {
		p := &a.config.Presets[0]
		p.Format = cfg.Format
		p.OutputDir = cfg.OutputDir
		p.ExtractAudio = cfg.ExtractAudio
		p.Subs = cfg.Subs
		p.Thumbnail = cfg.Thumbnail
		if err := config.SaveConfig(a.config); err != nil {
			logger.Error("save config (presets): %v", err)
		}
	}
}

func (a *App) savePresets(cfg *models.Config) {
	if err := config.SaveConfig(cfg); err != nil {
		logger.Error("save presets: %v", err)
	}
	a.dropZone.SetPresets(cfg.Presets)
}

func (a *App) listenUpdates() {
	for update := range a.queueMgr.Updates() {
		a.queueView.ProcessUpdate(update)

		if update.Type == "done" && update.Item != nil {
			if update.Item.State == models.StateError {
				logger.Error("download failed: url=%q title=%q err=%q", update.Item.URL, update.Item.Title, update.Item.Error)
				continue
			}

			if update.Item.State == models.StateComplete {
				now := time.Now()
				entry := models.HistoryEntry{
					URL:         update.Item.URL,
					Title:       update.Item.Title,
					Format:      update.Item.Format,
					OutputDir:   update.Item.OutputDir,
					Success:     true,
					CompletedAt: now,
				}
				entries := config.LoadHistory()
				entries = append(entries, entry)
				if err := config.SaveHistory(entries); err != nil {
					logger.Error("save history: %v", err)
				}
				a.historyView.Load(entries)
			}
		}
	}
}

func (a *App) Run() {
	a.window.Resize(fyne.NewSize(float32(a.config.WindowWidth), float32(a.config.WindowHeight)))
	a.window.SetContent(a.BuildUI())
	a.window.Show()
}

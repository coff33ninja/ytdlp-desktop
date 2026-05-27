package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	"ytdlp-desktop/internal/logger"
	"ytdlp-desktop/internal/models"
)

const (
	appDir     = "ytdlp-desktop"
	configFile = "config.json"
	historyFile = "history.json"
)

func getConfigDir() (string, error) {
	cfg, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(cfg, appDir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return dir, nil
}

func LoadConfig() *models.Config {
	cfg := defaultConfig()
	dir, err := getConfigDir()
	if err != nil {
		logger.Error("get config dir: %v", err)
		return cfg
	}
	data, err := os.ReadFile(filepath.Join(dir, configFile))
	if err != nil {
		logger.Warn("read config: %v (using defaults)", err)
		return cfg
	}
	if err := json.Unmarshal(data, cfg); err != nil {
		logger.Error("parse config: %v (using defaults)", err)
	}
	if cfg.MaxConcurrent < 1 {
		cfg.MaxConcurrent = 1
	}
	return cfg
}

func SaveConfig(cfg *models.Config) error {
	dir, err := getConfigDir()
	if err != nil {
		logger.Error("save config dir: %v", err)
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		logger.Error("save config marshal: %v", err)
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, configFile), data, 0644); err != nil {
		logger.Error("save config write: %v", err)
		return err
	}
	return nil
}

func LoadHistory() []models.HistoryEntry {
	dir, err := getConfigDir()
	if err != nil {
		logger.Error("get history dir: %v", err)
		return nil
	}
	data, err := os.ReadFile(filepath.Join(dir, historyFile))
	if err != nil {
		logger.Warn("read history: %v", err)
		return nil
	}
	var entries []models.HistoryEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		logger.Error("parse history: %v", err)
		return nil
	}
	if entries == nil {
		entries = []models.HistoryEntry{}
	}
	return entries
}

func SaveHistory(entries []models.HistoryEntry) error {
	dir, err := getConfigDir()
	if err != nil {
		logger.Error("save history dir: %v", err)
		return err
	}
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		logger.Error("save history marshal: %v", err)
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, historyFile), data, 0644); err != nil {
		logger.Error("save history write: %v", err)
		return err
	}
	return nil
}

func defaultConfig() *models.Config {
	return &models.Config{
		Presets: []models.Preset{
			{
				ID:   "best",
				Name: "Best Quality",
				Format: "bestvideo+bestaudio/best",
				AutoStart: true,
			},
			{
				ID:   "mp4",
				Name: "MP4 1080p",
				Format: "bestvideo[height<=1080]+bestaudio/best[height<=1080]",
				AutoStart: false,
			},
			{
				ID:   "audio",
				Name: "Audio Only (MP3)",
				Format: "bestaudio/best",
				ExtractAudio: true,
				AudioFormat: "mp3",
				AutoStart: false,
			},
		},
		DefaultPreset: "best",
		MaxConcurrent: 2,
		Theme:  "dark",
		WindowWidth:  900,
		WindowHeight: 700,
	}
}

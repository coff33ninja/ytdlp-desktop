package models

import "time"

type DownloadState string

const (
	StateWaiting  DownloadState = "waiting"
	StateRunning  DownloadState = "running"
	StateComplete DownloadState = "complete"
	StateError    DownloadState = "error"
	StatePaused   DownloadState = "paused"
)

type DownloadItem struct {
	ID            string        `json:"id"`
	URL           string        `json:"url"`
	Title         string        `json:"title"`
	State         DownloadState `json:"state"`
	Progress      float64       `json:"progress"`
	Speed         string        `json:"speed"`
	ETA           string        `json:"eta"`
	OutputDir     string        `json:"output_dir"`
	Format        string        `json:"format"`
	PresetID      string        `json:"preset_id"`
	ExtractAudio  bool          `json:"extract_audio"`
	AudioFormat   string        `json:"audio_format,omitempty"`
	Subs          bool          `json:"subs"`
	Thumbnail     bool          `json:"thumbnail"`
	CreatedAt     time.Time     `json:"created_at"`
	CompletedAt   *time.Time    `json:"completed_at,omitempty"`
	Error         string        `json:"error,omitempty"`
}

type Preset struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Format    string `json:"format"`
	OutputDir string `json:"output_dir"`
	AutoStart bool   `json:"auto_start"`
	ExtractAudio bool `json:"extract_audio"`
	AudioFormat string `json:"audio_format,omitempty"`
	Subs      bool   `json:"subs"`
	Thumbnail bool   `json:"thumbnail"`
	LimitRate string `json:"limit_rate,omitempty"`
}

type Config struct {
	Presets       []Preset `json:"presets"`
	DefaultPreset string   `json:"default_preset"`
	MaxConcurrent int      `json:"max_concurrent"`
	Theme         string   `json:"theme"`
	WindowWidth   int      `json:"window_width"`
	WindowHeight  int      `json:"window_height"`
	YtdlpPath     string   `json:"ytdlp_path"`
}

type HistoryEntry struct {
	URL       string    `json:"url"`
	Title     string    `json:"title"`
	Format    string    `json:"format"`
	OutputDir string    `json:"output_dir"`
	Success   bool      `json:"success"`
	CompletedAt time.Time `json:"completed_at"`
	FileSize  string    `json:"file_size,omitempty"`
}

type FormatInfo struct {
	FormatID     string  `json:"format_id"`
	Ext          string  `json:"ext"`
	Height       int     `json:"height"`
	Width        int     `json:"width"`
	VCodec       string  `json:"vcodec"`
	ACodec       string  `json:"acodec"`
	FPS          float64 `json:"fps"`
	Filesize     float64 `json:"filesize"`
	FilesizeApprox float64 `json:"filesize_approx"`
	TBR          float64 `json:"tbr"`
	ABR          float64 `json:"abr"`
	VBR          float64 `json:"vbr"`
	Container    string  `json:"container"`
	Protocol     string  `json:"protocol"`
	FormatNote   string  `json:"format_note"`
	Language     string  `json:"language"`
}

type VideoInfo struct {
	ID       string        `json:"id"`
	Title    string        `json:"title"`
	Formats  []FormatInfo  `json:"formats"`
	Duration float64       `json:"duration"`
}

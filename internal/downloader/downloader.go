package downloader

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/lrstanley/go-ytdlp"
	"ytdlp-desktop/internal/logger"
	"ytdlp-desktop/internal/models"
)

type ProgressCallback func(id string, progress float64, speed, eta string)
type InfoCallback func(id string, title string)
type DoneCallback func(id string, err error)

type Downloader struct {
	installed bool
}

func New() *Downloader {
	return &Downloader{}
}

func (d *Downloader) EnsureInstalled(ctx context.Context) error {
	if d.installed {
		return nil
	}
	ytdlp.MustInstall(ctx, nil)
	d.installed = true
	return nil
}

func (d *Downloader) FetchInfo(ctx context.Context, url string) (string, error) {
	res, err := ytdlp.New().
		SkipDownload().
		Print("%(title)s").
		Run(ctx, url)
	if err != nil {
		return "", fmt.Errorf("fetch info: %w", err)
	}
	return strings.TrimSpace(res.Stdout), nil
}

func (d *Downloader) Download(ctx context.Context, item *models.DownloadItem, onProgress ProgressCallback, onInfo InfoCallback, onDone DoneCallback) {
	item.State = models.StateRunning

	cmd := ytdlp.New().
		Newline().
		NoColors()
	cmd = applyItemFlags(cmd, item)

	cmd.ProgressFunc(500*time.Millisecond, func(update ytdlp.ProgressUpdate) {
		if update.Status == ytdlp.ProgressStatusDownloading {
			pct := update.Percent()
			item.Progress = pct / 100
			eta := update.ETA().Round(time.Second).String()
			speed := ""
			if update.DownloadedBytes > 0 && update.Finished.IsZero() {
				dur := update.Duration()
				if dur > 0 {
					bps := float64(update.DownloadedBytes) / dur.Seconds()
					speed = formatSpeed(bps)
				}
			}
			item.Speed = speed
			item.ETA = eta
			onProgress(item.ID, pct/100, speed, eta)
		}
	})

	cmd.StderrFunc(func(line string) {
		if item.Title == "" {
			if idx := strings.Index(line, "title:"); idx >= 0 {
				title := strings.TrimSpace(line[idx+6:])
				if title != "" {
					item.Title = title
					onInfo(item.ID, title)
				}
			}
		}
	})

	res, err := cmd.Run(ctx, item.URL)
	if err != nil {
		if ctx.Err() != nil {
			logger.Warn("download cancelled: id=%q url=%q", item.ID, item.URL)
			onDone(item.ID, ctx.Err())
			return
		}
		if res != nil && res.Stderr != "" {
			lines := strings.Split(res.Stderr, "\n")
			errMsg := strings.TrimSpace(lines[len(lines)-1])
			if errMsg == "" && len(lines) > 1 {
				errMsg = strings.TrimSpace(lines[len(lines)-2])
			}
			onDone(item.ID, fmt.Errorf("%s", errMsg))
			return
		}
		onDone(item.ID, err)
		return
	}

	onProgress(item.ID, 1.0, "", "")
	onDone(item.ID, nil)
}

func (d *Downloader) GetFormats(ctx context.Context, url string) (string, error) {
	res, err := ytdlp.New().ListFormats().Run(ctx, url)
	if err != nil {
		return "", fmt.Errorf("list formats: %w", err)
	}
	return res.Stdout, nil
}

func (d *Downloader) FetchFullInfo(ctx context.Context, url string) (*models.VideoInfo, error) {
	res, err := ytdlp.New().
		DumpJSON().
		NoPlaylist().
		Run(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("fetch full info: %w", err)
	}
	var info models.VideoInfo
	if err := json.Unmarshal([]byte(res.Stdout), &info); err != nil {
		return nil, fmt.Errorf("parse info: %w", err)
	}
	return &info, nil
}

func (d *Downloader) GetFormatsJSON(ctx context.Context, url string) ([]models.FormatInfo, error) {
	info, err := d.FetchFullInfo(ctx, url)
	if err != nil {
		return nil, err
	}
	return info.Formats, nil
}

func formatSpeed(bps float64) string {
	switch {
	case bps >= 1_000_000_000:
		return fmt.Sprintf("%.1f GB/s", bps/1_000_000_000)
	case bps >= 1_000_000:
		return fmt.Sprintf("%.1f MB/s", bps/1_000_000)
	case bps >= 1_000:
		return fmt.Sprintf("%.1f KB/s", bps/1_000)
	default:
		return fmt.Sprintf("%.0f B/s", bps)
	}
}

func applyItemFlags(cmd *ytdlp.Command, item *models.DownloadItem) *ytdlp.Command {
	if item.Format != "" {
		cmd = cmd.Format(item.Format)
	}

	outputTemplate := "%(title)s.%(ext)s"
	if item.OutputDir != "" {
		outputTemplate = item.OutputDir + string(os.PathSeparator) + outputTemplate
	}
	cmd = cmd.Output(outputTemplate)

	if item.ExtractAudio {
		cmd = cmd.ExtractAudio()
		if item.AudioFormat != "" {
			cmd = cmd.AudioFormat(item.AudioFormat)
		}
	}

	if item.Subs {
		cmd = cmd.WriteSubs().SubLangs("en,en.*")
	}

	if item.Thumbnail {
		cmd = cmd.EmbedThumbnail()
	}

	return cmd
}

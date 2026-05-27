# ytdlp-desktop

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="logo/logo.png">
  <source media="(prefers-color-scheme: light)" srcset="logo/logo.png">
  <img alt="ytdlp-desktop" src="logo/logo.png" width="96" align="right">
</picture>

A native Windows desktop frontend for [yt-dlp](https://github.com/yt-dlp/yt-dlp). Single static ~44 MB EXE, no Python runtime, no Web UI — just a clean Fyne-based GUI.

## Features

**Downloads tab** — drop or paste video URLs, probe available formats, select quality, monitor progress with speed/ETA, scroll through active downloads via a virtualized list view.

**Format Probe** — click Probe on any URL to fetch available formats from yt-dlp's JSON output. Probed formats appear in the preset selector dropdown, grouped by type (Video Only, Video+Audio, Audio Only), with resolution, codec, container, and file size.

**Format Builder** — visual format string construction with resolution, codec, container, and audio quality selectors. Toggle raw mode for direct yt-dlp format syntax. Live preview of the generated `-f` string.

**Presets tab** — create, edit, and delete named download presets. Each preset stores format, output directory, extract audio, subtitles, thumbnail, and rate limit settings. Includes a per-preset Format Builder.

**History tab** — browse the last 50 completed downloads with timestamps, format used, and success/failure status.

**Settings tab** — configure max concurrent downloads (default 2), yt-dlp binary path, theme (dark/light), and default preset options.

**Compact Widget** — floating overlay window (View → Toggle Widget) for quick URL drops from your browser.

**Drag-and-drop** — drop URLs directly onto the main window or compact widget.

**File-based logging** — all errors are logged to `%APPDATA%\ytdlp-desktop\ytdlp-desktop.log` with timestamps.

## Screenshots

*(Add screenshots here)*

## Download

[Download the latest release](https://github.com/coff33ninja/ytdlp-desktop/releases) — pre-built static EXE, no dependencies required.

## Build from source

Requires [Go 1.26+](https://go.dev/dl/) and [Zig](https://ziglang.org/download/) for CGo cross-compilation.

```powershell
$env:CC = "zig cc -target x86_64-windows-gnu"
go build -ldflags="-H windowsgui" -o ytdlp-desktop.exe .
```

- `-H windowsgui` suppresses the terminal window on launch
- Output: ~44 MB standalone `ytdlp-desktop.exe`

On first run, [go-ytdlp](https://github.com/lrstanley/go-ytdlp) auto-downloads and caches the yt-dlp binary internally.

## Architecture

```
ytdlp-desktop/
├── main.go                          # Entry point, window, compact widget
├── logo/
│   └── logo.png                     # App logo
├── ui/
│   ├── app.go                       # Controller: tabs, drop handling, probe, queue
│   ├── drop_zone.go                 # URL entry, download/probe buttons, preset selector
│   ├── format_builder.go            # Visual format string builder with filters
│   ├── preset_selector.go           # Preset dropdown with probed format support
│   ├── queue_view.go                # Downloads tab — virtualized widget.List
│   ├── presets_view.go              # Presets tab — editable preset cards
│   ├── history_view.go              # History tab — last 50 downloads
│   ├── settings_view.go             # Settings tab — concurrent, theme, defaults
│   └── widget.go                    # Compact floating overlay widget
├── internal/
│   ├── models/types.go              # DownloadItem, Preset, Config, FormatInfo, VideoInfo
│   ├── config/config.go             # JSON config/history persistence
│   ├── downloader/downloader.go     # yt-dlp wrapper via go-ytdlp, progress callbacks
│   ├── logger/logger.go             # File-based leveled logger
│   └── queue/queue.go               # Concurrent download queue with semaphore
├── go.mod
├── go.sum
└── .gitignore
```

### Key patterns

- **BaseWidget** — all tab views extend `widget.BaseWidget` with custom `CreateRenderer()`
- **Virtualized list** — `QueueView` uses `widget.List` so off-screen download cards are not rendered, keeping the UI smooth during progress updates
- **Thread-safe UI** — goroutines dispatch mutations via `fyne.Do()`, queue manager uses `sync.RWMutex`
- **Format string generation** — `FormatBuilder.buildFormatString()` constructs yt-dlp filter syntax like `bv*[height<=1080][vcodec~='^hevc']+ba/bv*+ba/b`
- **Probe integration** — results from yt-dlp's `--dump-json` populate virtual presets in the selector, auto-selecting raw mode with the chosen format ID
- **Error logging** — `internal/logger/` writes timestamped ERROR/WARN/INFO to `%APPDATA%\ytdlp-desktop\ytdlp-desktop.log`

### Design

- Dark theme with deep navy card backgrounds (`#222236`)
- Card corner radius: `10`, card alpha: `220`
- Color-coded status: blue (running), green (complete), red (error), gray (waiting)
- Monospace format string previews
- Consistent `500x300` MinSize across all tab views

## Dependencies

- [Fyne](https://fyne.io/) v2.7.4 — cross-platform GUI toolkit
- [go-ytdlp](https://github.com/lrstanley/go-ytdlp) v1.3.5 — Go-native yt-dlp bindings
- [Zig](https://ziglang.org/) — C compiler for CGo static linking

## License

MIT

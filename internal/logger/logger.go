package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	appDir    = "ytdlp-desktop"
	logFile   = "ytdlp-desktop.log"
)

var (
	mu     sync.Mutex
	buf    []byte
	logDir string
	initOnce sync.Once
)

func ensureDir() string {
	initOnce.Do(func() {
		cfg, err := os.UserConfigDir()
		if err != nil {
			logDir = os.TempDir()
			return
		}
		dir := filepath.Join(cfg, appDir)
		os.MkdirAll(dir, 0755)
		logDir = dir
	})
	return logDir
}

func write(level, msg string) {
	dir := ensureDir()
	ts := time.Now().Format("2006-01-02 15:04:05.000")
	line := fmt.Sprintf("%s [%s] %s\n", ts, level, msg)

	mu.Lock()
	defer mu.Unlock()

	f, err := os.OpenFile(filepath.Join(dir, logFile), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	f.WriteString(line)
}

func Info(format string, args ...any) {
	write("INFO", fmt.Sprintf(format, args...))
}

func Warn(format string, args ...any) {
	write("WARN", fmt.Sprintf(format, args...))
}

func Error(format string, args ...any) {
	write("ERROR", fmt.Sprintf(format, args...))
}

package dailylog

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Config struct {
	Timezone string `json:"timezone"`
	Path     string `json:"path"`
	Prefix   string `json:"prefix"`
	LogLevel string `json:"logLevel"`
}

type dailyWriter struct {
	path     string
	prefix   string
	timezone *time.Location

	mu      sync.Mutex
	curDate string
	file    *os.File
}

func Init(cfg *Config) {
	tz := time.FixedZone(cfg.Timezone, 0)
	dw := &dailyWriter{
		path:     cfg.Path,
		prefix:   cfg.Prefix,
		timezone: tz,
	}

	level := getLogLevelFromCfg(cfg)
	if err := dw.rotateIfNeeded(); err != nil {
		setSlog(os.Stderr, level, tz)
		slog.Error("Failed to initialize the daily writer. Will use 'os.Stderr'", "error", err)
		return
	}

	setSlog(io.MultiWriter(os.Stderr, dw), level, tz)
}

func (dw *dailyWriter) Write(p []byte) (n int, err error) {
	dw.mu.Lock()
	defer dw.mu.Unlock()
	if err := dw.rotateIfNeeded(); err != nil {
		return 0, err
	}
	return dw.file.Write(p)
}

func setSlog(w io.Writer, l slog.Level, loc *time.Location) {
	logger := slog.New(slog.NewTextHandler(w, &slog.HandlerOptions{
		Level: l,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				if t, ok := a.Value.Any().(time.Time); ok {
					a.Value = slog.TimeValue(t.In(loc))
				}
			}
			return a
		},
	}))
	slog.SetDefault(logger)
}

func getLogLevelFromCfg(cfg *Config) slog.Level {
	switch strings.ToLower(cfg.LogLevel) {
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelDebug
	}
}

func (dw *dailyWriter) filename(t time.Time) string {
	return filepath.Join(dw.path, fmt.Sprintf("%s_%s.log", dw.prefix, t.Format("2006-01-02")))
}

func (dw *dailyWriter) rotateIfNeeded() error {
	now := time.Now().In(dw.timezone)
	date := now.Format("2006-01-02")
	if dw.file != nil && date == dw.curDate {
		return nil
	}

	if dw.file != nil {
		_ = dw.file.Close()
		dw.file = nil
	}

	if err := os.MkdirAll(dw.path, 0o755); err != nil {
		return err
	}

	f, err := os.OpenFile(dw.filename(now), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}

	dw.file = f
	dw.curDate = date
	return nil
}

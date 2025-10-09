package dilog

import (
	"io"
	"log/slog"
	"os"
	"strings"
	"time"
)

var (
	w     io.Writer                                  = os.Stdout
	getDw func(*Config, *time.Location) *dailyWriter = func(cfg *Config, tz *time.Location) *dailyWriter {
		return &dailyWriter{
			path:     cfg.Path,
			prefix:   cfg.Prefix,
			timezone: tz,
		}
	}
)

type Config struct {
	Timezone string `json:"timezone"`
	Path     string `json:"path"`
	Prefix   string `json:"prefix"`
	LogLevel string `json:"logLevel"`
}

func Init(cfg *Config) {
	tz := getTimezone(cfg)
	Dw = getDw(cfg, tz)

	level := getLogLevelFromCfg(cfg)
	if err := Dw.rotateIfNeeded(); err != nil {
		lg := getLogger(w, level, tz)
		slog.SetDefault(lg)
		slog.Error("Failed to initialize the daily writer. Will use 'w'", "w", w, "error", err)
		return
	}

	lg := getLogger(io.MultiWriter(w, Dw), level, tz)
	slog.SetDefault(lg)
}

func WritelnToDw(msg string) {
	if Dw != nil {
		_, _ = Dw.Write([]byte(msg + "\n"))
	}
}

func getLogger(w io.Writer, l slog.Level, loc *time.Location) *slog.Logger {
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
	return logger
}

func getTimezone(cfg *Config) *time.Location {
	switch strings.ToLower(cfg.Timezone) {
	case "local", "":
		return time.Local
	case "utc":
		return time.UTC
	default:
		if loc, err := time.LoadLocation(cfg.Timezone); err == nil {
			return loc
		}
		return time.Local
	}
}

func getLogLevelFromCfg(cfg *Config) slog.Level {
	switch strings.ToLower(cfg.LogLevel) {
	case "error":
		return slog.LevelError
	case "warn":
		return slog.LevelWarn
	case "info":
		return slog.LevelInfo
	default:
		return slog.LevelDebug
	}
}

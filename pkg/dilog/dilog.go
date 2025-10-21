package dilog

import (
	"io"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/baudii/ada-ai/pkg/dilog/dailywriter"
	"github.com/baudii/ada-ai/pkg/dilog/simplehandler"
)

// LogConfig holds configuration settings for logging.
type LogConfig struct {
	Timezone string `json:"timezone"`
	Path     string `json:"path"`
	Prefix   string `json:"prefix"`
	Level    string `json:"level"`
}

// DefaultDailyLogger creates a simple slog.Logger based on the provided
// dilog.LogConfig and additional options.
func DefaultDailyLogger(cfg *LogConfig, opts ...dailywriter.Option) (*slog.Logger, error) {
	opts = append(opts,
		dailywriter.WithPrefix(cfg.Prefix),
		dailywriter.WithLocation(loc(cfg.Timezone)),
	)

	dw, err := dailywriter.New(cfg.Path, opts...)
	if err != nil {
		return nil, err
	}
	return slog.New(
		simplehandler.NewSimpleHandler(
			io.MultiWriter(os.Stdout, dw),
			simplehandler.WithLevel(level(cfg.Level)),
			simplehandler.WithLocation(loc(cfg.Timezone)),
		),
	), nil
}

func loc(loc string) *time.Location {
	switch strings.ToLower(loc) {
	case "local", "":
		return time.Local
	case "utc":
		return time.UTC
	default:
		if loc, err := time.LoadLocation(loc); err == nil {
			return loc
		}
		return time.Local
	}
}

func level(loglevel string) slog.Level {
	switch strings.ToLower(loglevel) {
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

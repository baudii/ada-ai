package common

import (
	"io"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/baudii/ada-ai/pkg/dilog"
	"github.com/baudii/ada-ai/pkg/dilog/simplehandler"
	"github.com/baudii/ada-ai/pkg/utils"
)

var defaultCfg = dilog.LogConfig{
	Timezone: "local",
	Path:     "logs",
	Prefix:   "pfx",
}

// LoadLogConfig loads the log configuration from the specified JSON file path.
// If loading fails, it returns a default configuration.
func LoadLogConfig(path string) *dilog.LogConfig {
	cfg, err := utils.ParseJSONConfigWithLocal[dilog.LogConfig](path)
	if err != nil {
		return &defaultCfg
	}

	return cfg
}

// DefaultSimpleLogger creates a default slog.Logger with a SimpleHandler writing
// to both stdout and a daily rotating log file. The log configuration is loaded
// from the specified JSON file path.
func DefaultSimpleLogger(cfg *dilog.LogConfig, opts ...dilog.Option) (*slog.Logger, error) {
	opts = append(opts,
		dilog.WithPrefix(cfg.Prefix),
		dilog.WithLocation(loc(cfg.Timezone)),
	)

	dw, err := dilog.NewDailyWriter(cfg.Path, opts...)
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

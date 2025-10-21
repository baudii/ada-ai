package logx

import (
	"log/slog"

	"github.com/baudii/ada-ai/pkg/dilog"
	"github.com/baudii/ada-ai/pkg/jsonx"
)

// DefaultCfg is the default logging configuration used when loading from file fails.
var DefaultCfg = dilog.LogConfig{
	Timezone: "local",
	Path:     "logs",
	Prefix:   "pfx",
}

// LoadLogConfigOrDefault loads the log configuration from the specified JSON file path.
// If loading fails, it returns a default configuration.
func LoadLogConfigOrDefault(path string) dilog.LogConfig {
	cfg, err := jsonx.LoadWithLocal[dilog.LogConfig](path)
	if err != nil {
		slog.Error("Failed to load configuration file. Using default.")
		return DefaultCfg
	}

	return cfg
}

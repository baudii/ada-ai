package dailylogger

import (
	"log/slog"

	"github.com/baudii/ada-ai/internal/infra/config"
)

// DefaultCfg is the default logging configuration used when loading from file fails.
var DefaultCfg = LogConfig{
	Timezone: "local",
	Path:     "logs",
	Prefix:   "pfx",
}

// LoadLogConfigOrDefault loads the log configuration from the specified JSON file path.
// If loading fails, it returns a default configuration.
func LoadLogConfigOrDefault(path string) LogConfig {
	cfg, err := config.LoadWithLocal[LogConfig](path)
	if err != nil {
		slog.Error("Failed to load configuration file. Using default.")
		return DefaultCfg
	}

	return cfg
}

package llm

import (
	"encoding/json"
	"os"
)

type Config struct {
	Provider string         `json:"provider" yaml:"provider"`
	Options  map[string]any `json:"options"  yaml:"options"`
}

func ParseCfg() Config {
	var cfg = Config{}
	file, err := os.ReadFile("llm-cfg.json")
	if err != nil {
		return cfg
	}

	json.Unmarshal(file, &cfg)
	return cfg
}

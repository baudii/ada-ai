package ai

import (
	"fmt"
	"path/filepath"

	"github.com/baudii/ada-ai/pkg/utils"
	"github.com/tmc/langchaingo/llms"
)

const configName string = "llm-provider.json"

// Represents the structure of the LLM provider configuration file.
type Config struct {
	Provider string            `json:"provider"`
	Options  map[string]string `json:"options"`
}

// Registers the LLM based on the specified configuration.
func Register(cfg *Config) (llms.Model, error) {
	switch cfg.Provider {
	case "ollama":
		return registerOllama(cfg.Options)
	default:
		return nil, fmt.Errorf("unsupported llm provider: %s", cfg.Provider)
	}
}

// Registers the LLM based on the configuration parsed
// from 'llm-provider.json' file located at path.
//
// Path will be joined with the current os.Executable() directory.
func RegisterFromFile(path string) (llms.Model, error) {
	relPath := filepath.Join(path, configName)
	cfg, err := utils.ParseJSONConfigWithLocal[Config](utils.GetAbsolutePath(relPath))
	if err != nil {
		return nil, fmt.Errorf("failed to parse llm config %q: %w", relPath, err)
	}

	return Register(cfg)
}

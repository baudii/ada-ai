package ai

import (
	"fmt"
	"path/filepath"

	"github.com/baudii/ada-ai/pkg/utils"
	"github.com/tmc/langchaingo/llms"
)

const configName string = "llm-provider.json"

// Config represents the structure of the LLM provider configuration file.
type Config struct {
	Options map[string]string `json:"options"`
}

// Register is used to register the LLM based on the specified configuration.
func Register(provider string, cfg *Config) (llms.Model, error) {
	switch provider {
	case "ollama":
		return registerOllama(cfg.Options)
	case "grok":
		return registerGrok(cfg.Options)
	default:
		return nil, fmt.Errorf("unsupported llm provider: %q", provider)
	}
}

// RegisterFromFile registers the LLM using the configuration from a file
// for the specified provider.
func RegisterFromFile(provider, folder string) (llms.Model, error) {
	path := filepath.Join(folder, configName)
	cfg, err := utils.ParseJSONConfigWithLocal[Config](path)
	if err != nil {
		return nil, fmt.Errorf("parse llm config %q: %w", path, err)
	}

	return Register(provider, cfg)
}

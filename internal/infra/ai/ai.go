package ai

import (
	"fmt"
	"path/filepath"

	"github.com/baudii/ada-ai/internal/infra/config"
	"github.com/tmc/langchaingo/llms"
)

// Config represents the structure of the LLM provider configuration file.
type Config struct {
	Options map[string]string `json:"options"`
}

// options defines a function type for modifying LLM registration options.
type options func(map[string]string)

// Register is used to register the LLM based on the specified configuration.
func Register(provider string, options map[string]string, opts ...options) (llms.Model, error) {
	switch provider {
	case "ollama":
		return registerOllama(options)
	case "grok":
		return registerGrok(options)
	default:
		return nil, fmt.Errorf("unsupported llm provider: %q", provider)
	}
}

// RegisterFromFile registers the LLM using the configuration from a file
// for the specified provider.
func RegisterFromFile(provider, folder string) (llms.Model, error) {
	path := filepath.Join(folder, fmt.Sprintf("%s.json", provider))
	cfg, err := config.LoadWithLocal[Config](path)
	if err != nil {
		return nil, fmt.Errorf("parse llm config: %w", err)
	}

	return Register(provider, cfg.Options)
}

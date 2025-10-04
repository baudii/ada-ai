package common

import (
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/baudii/ada-ai/pkg/utils"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
)

const configName string = "llm-provider.json"

type Config struct {
	Provider string            `json:"provider"`
	Options  map[string]string `json:"options"`
}

// Registers the LLM based on the configuration file data/configuration/llm-provider.json file.
func Register() (llms.Model, error) {
	relPath := filepath.Join(ConfigPath, configName)
	cfg, err := utils.ParseJSONConfigWithLocal[Config](utils.GetAbsolutePath(relPath))
	if err != nil {
		return nil, fmt.Errorf("failed to parse llm config '%s': %w", configName, err)
	}

	switch cfg.Provider {
	case "ollama":
		return registerOllama(cfg.Options)
	default:
		return nil, fmt.Errorf("unsupported llm provider: %s", cfg.Provider)
	}
}

func registerOllama(options map[string]string) (llms.Model, error) {
	model := options["model"]
	if model == "" {
		return nil, fmt.Errorf("ollama: options.model is required")
	}

	var opts []ollama.Option
	opts = append(opts, ollama.WithModel(model))

	if v := options["url"]; v != "" {
		opts = append(opts, ollama.WithServerURL(v))
	}
	if v := options["keep_alive"]; v != "" {
		opts = append(opts, ollama.WithKeepAlive(v))
	}
	if v := options["format"]; v != "" {
		opts = append(opts, ollama.WithFormat(v))
	}
	if v := options["custom_template"]; v != "" {
		opts = append(opts, ollama.WithCustomTemplate(v))
	}
	if v := options["http_timeout"]; v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			opts = append(opts, ollama.WithHTTPClient(&http.Client{Timeout: d}))
		}
	}

	l, err := ollama.New(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to register ollama: %w", err)
	}
	return l, nil
}

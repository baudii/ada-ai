package ai

import (
	"fmt"
	"net/http"
	"time"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
)

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

package ai

import (
	"fmt"
	"net/http"
	"time"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
)

func registerOllama(m map[string]string) (llms.Model, error) {
	model := m["model"]
	if model == "" {
		return nil, fmt.Errorf("ollama: options.model is required")
	}

	var opts []ollama.Option
	opts = append(opts, ollama.WithModel(model))

	if v := m["url"]; v != "" {
		opts = append(opts, ollama.WithServerURL(v))
	}
	if v := m["keep_alive"]; v != "" {
		opts = append(opts, ollama.WithKeepAlive(v))
	}
	if v := m["format"]; v != "" {
		opts = append(opts, ollama.WithFormat(v))
	}
	if v := m["custom_template"]; v != "" {
		opts = append(opts, ollama.WithCustomTemplate(v))
	}
	if v := m["http_timeout"]; v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			opts = append(opts, ollama.WithHTTPClient(&http.Client{Timeout: d}))
		}
	}

	return ollama.New(opts...)
}

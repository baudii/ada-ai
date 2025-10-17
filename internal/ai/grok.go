package ai

import (
	"fmt"
	"net/http"
	"time"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

func registerGrok(m map[string]string) (llms.Model, error) {
	model := m["model"]
	if model == "" {
		return nil, fmt.Errorf("grok: options.model is required")
	}

	var opts []openai.Option
	opts = append(opts, openai.WithModel(model))

	if v := m["url"]; v != "" {
		opts = append(opts, openai.WithBaseURL(v))
	}
	if v := m["token"]; v != "" {
		opts = append(opts, openai.WithToken(v))
	}
	if v := m["format"]; v == "json" {
		opts = append(opts, openai.WithResponseFormat(openai.ResponseFormatJSON))
	}
	if v := m["http_timeout"]; v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			opts = append(opts, openai.WithHTTPClient(&http.Client{Timeout: d}))
		}
	}

	return openai.New(opts...)
}

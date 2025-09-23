package llm

import (
	"fmt"
)

type Config struct {
	Provider string         `json:"provider" yaml:"provider"`
	Options  map[string]any `json:"options"  yaml:"options"`
}

type ProviderBuilder func(cfg map[string]any) (LLM, error)

var registry = map[string]ProviderBuilder{}

func Register(name string, builder ProviderBuilder) {
	registry[name] = builder
}

func Resolve(cfg Config) (LLM, error) {
	builder, ok := registry[cfg.Provider]
	if !ok {
		return nil, fmt.Errorf("unknown provider: %s", cfg.Provider)
	}
	return builder(cfg.Options)
}

package llm

import (
	"fmt"
	"path/filepath"

	"github.com/baudii/ada-ai/pkg/utils"
)

type config struct {
	Provider string         `json:"provider"`
	Options  map[string]any `json:"options"`
}

type ProviderBuilder func(cfg map[string]any) (LLM, error)

var registry = map[string]ProviderBuilder{}

func Register(name string, builder ProviderBuilder) {
	registry[name] = builder
}

func Resolve() (LLM, error) {
	relativePath := filepath.Join("cfg", "llm-provider.json")
	cfg, err := utils.ParseJsonConfigWithLocal[config](utils.GetAbsolutePath(relativePath))
	if err != nil {
		return nil, err
	}

	builder, ok := registry[cfg.Provider]
	if !ok {
		return nil, fmt.Errorf("unknown provider: %s", cfg.Provider)
	}
	return builder(cfg.Options)
}

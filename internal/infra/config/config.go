package config

import (
	"fmt"
	"path/filepath"

	"github.com/baudii/ada-ai/internal/app"
	"github.com/baudii/ada-ai/internal/infra/folders"
)

// ParseAppOptions parses the application options from a JSON file located
// in the specified path.
func ParseAppOptions(path string) (app.Options, error) {
	optsPath := filepath.Join(path, app.ConfigFile)
	opts, err := LoadWithLocal[app.Options](optsPath)
	if err != nil {
		return app.Options{}, fmt.Errorf("parse ada options: %w", err)
	}
	if opts.PromptsRoot == "" {
		opts.PromptsRoot = folders.Prompts
	}
	if opts.ProjectsRoot == "" {
		opts.ProjectsRoot = folders.Projects
	}
	return opts, nil
}

// MergeMaps merges dst map with the values from src. If a value exists
// in src, then it will be copied to dst. If dst already contains
// this kev/value pair, it will be overwritten
//
// Note: This function doesn't copy the values, it just assings them
// to dst. If the value is reference type, then changes will affect
// both src and dst.
func MergeMaps(dst, src map[string]any) {
	if dst == nil {
		return
	}

	for k, v := range src {
		if vMap, ok := v.(map[string]any); ok {
			if dMap, ok := dst[k].(map[string]any); ok {
				MergeMaps(dMap, vMap)
				continue
			}
		}
		dst[k] = v
	}
}

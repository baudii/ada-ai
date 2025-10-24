package config_test

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/baudii/ada-ai/internal/app"
	"github.com/baudii/ada-ai/internal/infra/config"
	"github.com/baudii/ada-ai/internal/infra/folders"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseAppOptions(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		filename string
		contents string
		err      string
		expected app.Options
	}{
		{"non-existent path", "non-existent-path", `
			{
				"requestTimeout": "5s",
				"projectsRoot": "testdata/projects",
				"promptsRoot": "testdata/prompts",
				"reflection": {
					"depth": 1,
					"threshhold": 0.4
				}
			}`, "parse ada options", app.Options{
			Timeout:      "5s",
			ProjectsRoot: "testdata/projects",
			PromptsRoot:  "testdata/prompts",
		}},
		{"valid config", app.ConfigFile, `{}`, "", app.Options{
			ProjectsRoot: folders.Projects,
			PromptsRoot:  folders.Prompts,
		}},
	}

	for _, v := range tests {
		t.Run(v.name, func(t *testing.T) {
			t.Parallel()
			d := t.TempDir()
			err := os.WriteFile(filepath.Join(d, v.filename), []byte(v.contents), 0o644)
			require.NoError(t, err)
			opts, err := config.ParseAppOptions(d)
			if v.err != "" {
				assert.ErrorContains(t, err, v.err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, v.expected, opts)
			}
		})
	}
}

func TestMerge(t *testing.T) {
	t.Parallel()
	tests := []struct {
		dst      map[string]any
		src      map[string]any
		expected map[string]any
	}{
		{nil, map[string]any{"a": nil}, nil},
		{map[string]any{"b": "s"}, nil, map[string]any{"b": "s"}},
		{map[string]any{"a": "s"}, map[string]any{"a": nil}, map[string]any{"a": nil}},
		{map[string]any{"a": "s"}, map[string]any{"b": nil}, map[string]any{"a": "s", "b": nil}},
		{
			map[string]any{
				"nil1": nil,
				"int":  12,
				"e": map[string]any{
					"nil2":     nil,
					"ovwrtSt":  "st",
					"ovwrtInt": 12,
				},
			},
			map[string]any{"e": map[string]any{
				"nil2":     "not nil",
				"ovwrtSt":  12,
				"ovwrtInt": "st",
				"new":      "new",
			}},
			map[string]any{
				"nil1": nil,
				"int":  12,
				"e": map[string]any{
					"nil2":     "not nil",
					"ovwrtSt":  12,
					"ovwrtInt": "st",
					"new":      "new",
				},
			},
		},
	}

	for i, v := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			config.MergeMaps(v.dst, v.src)
			assert.Equal(t, v.dst, v.expected)
		})
	}
}

package ai

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegister_invalid(t *testing.T) {
	t.Parallel()
	_, err := Register("random", map[string]string{})
	require.Error(t, err)
}

func TestRegister_ollama(t *testing.T) {
	tests := []struct {
		name   string
		cfg    *Config
		hasErr bool
	}{
		{
			name:   "fail: missing config",
			cfg:    &Config{nil},
			hasErr: true,
		},
		{
			name:   "success: invalid url",
			cfg:    &Config{map[string]string{"model": "gemma3:4b", "url": "invalid"}},
			hasErr: false,
		},
		{
			name:   "success: valid config",
			cfg:    &Config{map[string]string{"model": "gemma3:4b"}},
			hasErr: false,
		},
		{
			// Full options as per ollama.go defaults
			cfg: &Config{
				map[string]string{
					"model":           "gemma3:4b",
					"url":             "http://localhost:11434",
					"keep_alive":      "1m",
					"format":          "text",
					"custom_template": "{{.Content}}",
					"http_timeout":    "30s",
				},
			},
			hasErr: false,
		},
	}

	for i, v := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			_, err := Register("ollama", v.cfg.Options)
			assert.Equal(t, (err != nil), v.hasErr)
		})
	}
}

func TestRegister_grok(t *testing.T) {
	tests := []struct {
		cfg    *Config
		hasErr bool
	}{
		{&Config{nil}, true},
		{&Config{map[string]string{"model": "groq/compound"}}, true},
		{&Config{map[string]string{"model": "groq/compound", "token": "api-key"}}, false},
		{
			// Full options as per grok.go defaults
			&Config{
				map[string]string{
					"model":        "groq/compound",
					"url":          "https://api.groq.com/openai/v1",
					"token":        "your-api-key",
					"format":       "json",
					"http_timeout": "60s",
				},
			},
			false,
		},
	}

	for i, v := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			_, err := Register("grok", v.cfg.Options)
			assert.Equal(t, v.hasErr, (err != nil))
		})
	}
}

func TestRegisterFromFile(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		provider string
		json     string
		err      string
	}{
		{
			name:     "success: ollama",
			provider: "ollama",
			json:     `{"options": {"model":"some"}}`,
		},
		{
			name:     "success: grok",
			provider: "grok",
			json:     `{"options": {"model":"some", "token": "api-key"}}`,
		},
		{
			name:     "fail: parse llm config",
			provider: "ollama",
			json:     `{`,
			err:      "parse llm config",
		},
		{
			name:     "fail: default provider",
			provider: "unknown",
			json:     `{"options": {"model":"some"}}`,
			err:      "unsupported llm provider",
		},
	}
	for _, v := range tests {
		t.Run(v.name, func(t *testing.T) {
			dir := t.TempDir()
			err := os.WriteFile(filepath.Join(dir, fmt.Sprintf("%s.json", v.provider)), []byte(v.json), 0644)
			require.NoError(t, err)
			_, err = RegisterFromFile(v.provider, dir)
			if v.err != "" {
				assert.ErrorContains(t, err, v.err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPromptFromTemplate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		template string
		args     []any
		hasErr   bool
	}{
		{"some content", []any{}, false},
		{"some content %v", []any{"val1"}, false},
		{"some content %s, %q", []any{"val1", "val2"}, false},
		{"", []any{}, true},
	}

	file := "test_template.txt"

	for i, v := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			tempdir := t.TempDir()
			if !v.hasErr {
				err := os.WriteFile(filepath.Join(tempdir, file), []byte(v.template), 0644)
				require.NoError(t, err)
			}
			gen := NewGenerator(nil)
			res, err := gen.BuildPrompt(filepath.Join(tempdir, file), v.args...)
			assert.Equal(t, v.hasErr, err != nil)
			assert.Equal(t, fmt.Sprintf(v.template, v.args...), res)
		})
	}
}

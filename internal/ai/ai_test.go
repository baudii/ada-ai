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

func TestRegisterJSON(t *testing.T) {
	t.Parallel()
	m := map[string]string{}
	_, err := RegisterJSON("invalid", m)
	assert.Equal(t, "json", m["format"])
	require.Error(t, err)
}

func TestRegister_invalid(t *testing.T) {
	t.Parallel()
	_, err := Register("random", map[string]string{})
	require.Error(t, err)
}

func TestRegister_ollama(t *testing.T) {
	tests := []struct {
		cfg    *Config
		hasErr bool
	}{
		{&Config{nil}, true},
		{&Config{map[string]string{"model": "gemma3:4b", "url": "invalid"}}, false},
		{&Config{map[string]string{"model": "gemma3:4b"}}, false},
		{
			// Full options as per ollama.go defaults
			&Config{
				map[string]string{
					"model":           "gemma3:4b",
					"url":             "http://localhost:11434",
					"keep_alive":      "1m",
					"format":          "text",
					"custom_template": "{{.Content}}",
					"http_timeout":    "30s",
				},
			},
			false,
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
		provider string
		json     string
		hasErr   bool
	}{
		{"ollama", `{"options": {"model":"some"}}`, false},
		{"grok", `{"options": {"model":"some", "token": "api-key"}}`, false},
		{"unsupported", "", true},
	}
	for i, v := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			dir := t.TempDir()
			if !v.hasErr {
				err := os.WriteFile(filepath.Join(dir, fmt.Sprintf("%s.json", v.provider)), []byte(v.json), 0644)
				require.NoError(t, err)
			}
			_, err := RegisterFromFile(v.provider, dir)
			if v.hasErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

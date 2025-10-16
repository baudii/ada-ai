package ai

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegister(t *testing.T) {
	tests := []struct {
		cfg    *Config
		hasErr bool
	}{
		{&Config{"unknown", nil}, true},
		{&Config{"ollama", nil}, true},
		{&Config{"ollama", map[string]string{"model": "gemma3:4b", "url": "invalid"}}, false},
		{&Config{"ollama", map[string]string{"model": "gemma3:4b"}}, false},
		{
			// Full options as per ollama.go defaults
			&Config{
				"ollama",
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
			_, err := Register(v.cfg)
			assert.Equal(t, (err != nil), v.hasErr)
		})
	}
}

func TestRegisterFromFile(t *testing.T) {
	t.Parallel()
	tests := []bool{true, false}
	for i, v := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			dir := t.TempDir()
			if !v {
				err := os.WriteFile(filepath.Join(dir, configName), []byte(`{"provider": "ollama", "options": {"model":"some"}}`), 0644)
				require.NoError(t, err)
			}
			_, err := RegisterFromFile(dir)
			if v {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

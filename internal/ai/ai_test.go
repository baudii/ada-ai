package ai

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/baudii/ada-ai/pkg/utils"
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

	for _, v := range tests {
		_, err := Register(v.cfg)
		assert.Equal(t, (err != nil), v.hasErr)
	}
}

func TestRegisterFromFile(t *testing.T) {
	origExe := utils.Executable
	t.Cleanup(func() { utils.Executable = origExe })
	tests := []struct {
		hasErr bool
	}{
		{true},
		{false},
		{true},
	}

	for _, v := range tests {
		dir := t.TempDir()
		utils.Executable = func() (string, error) { return filepath.Join(dir, "mock.exe"), nil }
		if !v.hasErr {
			err := os.WriteFile(filepath.Join(dir, configName), []byte(`{"provider": "ollama", "options": {"model":"some"}}`), 0644)
			require.NoError(t, err)
		}
		_, err := RegisterFromFile(dir)
		if v.hasErr {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}
	}
}

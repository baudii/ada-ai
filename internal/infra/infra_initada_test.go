package infra_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/baudii/ada-ai/internal/infra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAI_FailsToParseAIConfig(t *testing.T) {
	t.Parallel()
	ai, err := infra.NewAI("prov", t.TempDir())
	assert.Nil(t, ai)
	assert.ErrorContains(t, err, "parse llm config")
}

func TestNewAI_FailsToRegisterAI(t *testing.T) {
	t.Parallel()
	cfgroot := filepath.Join(t.TempDir(), "prov.json")
	err := os.WriteFile(cfgroot, []byte(`{"options":{}}`), 0o644)
	require.NoError(t, err)
	ai, err := infra.NewAI("prov", cfgroot)
	assert.Nil(t, ai)
	assert.ErrorContains(t, err, "register ai")
}

func TestNewAI_Success(t *testing.T) {
	t.Parallel()
	aicfgroot := filepath.Join(t.TempDir(), "ollama.json")
	cfgroot := t.TempDir()
	err := os.WriteFile(aicfgroot, []byte(`{"options":{"model":"some"}}`), 0o644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(cfgroot, "ada.json"), []byte(`{"option1":"value1"}`), 0o644)
	require.NoError(t, err)
	ai, err := infra.NewAI("ollama", aicfgroot)
	require.NotNil(t, ai)
	require.NoError(t, err)
}

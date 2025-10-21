package app_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/baudii/ada-ai/internal/app"
	"github.com/stretchr/testify/require"
)

func TestInitAda_FailsToParseAIConfig(t *testing.T) {
	t.Parallel()
	a := app.New()
	err := a.InitAda("prov", t.TempDir())
	require.ErrorContains(t, err, "parse llm config")
}

func TestInitAda_FailsToRegisterAI(t *testing.T) {
	t.Parallel()
	cfgroot := filepath.Join(t.TempDir(), "prov.json")
	a := app.New()
	err := os.WriteFile(cfgroot, []byte(`{"options":{}}`), 0o644)
	require.NoError(t, err)
	err = a.InitAda("prov", cfgroot)
	require.ErrorContains(t, err, "register ai")
}

func TestInitAda_Success(t *testing.T) {
	t.Parallel()
	aicfgroot := filepath.Join(t.TempDir(), "ollama.json")
	cfgroot := t.TempDir()
	a := app.New(app.WithOptions(app.Options{}))
	err := os.WriteFile(aicfgroot, []byte(`{"options":{"model":"some"}}`), 0o644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(cfgroot, "ada.json"), []byte(`{"option1":"value1"}`), 0o644)
	require.NoError(t, err)
	err = a.InitAda("ollama", aicfgroot)
	require.NoError(t, err)
}

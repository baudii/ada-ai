package infra_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/baudii/ada-ai/internal/core/project"
	"github.com/baudii/ada-ai/internal/core/seqdir"
	"github.com/baudii/ada-ai/internal/infra"
	"github.com/baudii/ada-ai/internal/infra/fsproject"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var mdpErr = &mockDirProvider{
	err: os.ErrClosed,
}

type mockDirProvider struct {
	path string
	mode seqdir.Mode
	err  error
}

func (m *mockDirProvider) ProjectFolder(base string) (string, error) {
	return m.path, m.err
}

func TestInitLocalProject_FailsProjectFolder(t *testing.T) {
	t.Parallel()
	lp, err := infra.NewFSProject(t.TempDir(), project.Context{}, mdpErr)
	assert.Nil(t, lp)
	assert.ErrorContains(t, err, "project folder")
}

func TestInitLocalProject_FailsToCreateProjectManager(t *testing.T) {
	t.Parallel()
	d := filepath.Join(t.TempDir(), fsproject.NavFolder)
	f, err := os.Create(d)
	require.NoError(t, err)
	_ = f.Close()
	ai, err := infra.NewFSProject(t.TempDir(), project.Context{}, &mockDirProvider{path: d, mode: seqdir.CreateNew})
	assert.Nil(t, ai)
	assert.ErrorContains(t, err, "create local project")
}

func TestInitLocalProject_Success(t *testing.T) {
	t.Parallel()
	ai, err := infra.NewFSProject(t.TempDir(), project.Context{}, &mockDirProvider{path: t.TempDir(), mode: seqdir.CreateNew})
	require.NotNil(t, ai)
	require.NoError(t, err)
}

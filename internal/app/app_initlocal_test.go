package app_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/baudii/ada-ai/internal/ada"
	"github.com/baudii/ada-ai/internal/app"
	"github.com/baudii/ada-ai/internal/project/folder"
	"github.com/baudii/ada-ai/internal/project/local"
	"github.com/stretchr/testify/require"
)

var mdpErr = &mockDirProvider{err: os.ErrClosed}

type mockDirProvider struct {
	path string
	err  error
}

func (m *mockDirProvider) ProjectFolder(base string, mode folder.Mode) (string, error) {
	return m.path, m.err
}

func TestInitLocalProject_FailsReadDir(t *testing.T) {
	t.Parallel()
	d := filepath.Join(t.TempDir(), "file.txt")
	f, err := os.Create(d)
	require.NoError(t, err)
	_ = f.Close()
	ada := ada.New(nil)
	a := app.New(
		app.WithGen(ada),
		app.WithDirProvider(mdpErr),
		app.WithOptions(app.Options{
			ProjectsRoot: t.TempDir(),
		}))
	err = a.InitLocalProject()
	require.ErrorContains(t, err, "project folder")
}

func TestInitLocalProject_FailsToCreateProjectManager(t *testing.T) {
	t.Parallel()
	ada := ada.New(nil)
	d := filepath.Join(t.TempDir(), local.NavFolder)
	f, err := os.Create(d)
	require.NoError(t, err)
	_ = f.Close()
	a := app.New(
		app.WithGen(ada),
		app.WithMode(folder.CreateNew),
		app.WithDirProvider(&mockDirProvider{path: d}),
		app.WithOptions(app.Options{
			ProjectsRoot: t.TempDir(),
		}))
	err = a.InitLocalProject()
	require.ErrorContains(t, err, "create local project")
}

func TestInitLocalProject_Success(t *testing.T) {
	t.Parallel()
	ada := ada.New(nil)
	a := app.New(app.WithGen(ada),
		app.WithMode(folder.CreateNew),
		app.WithDirProvider(&mockDirProvider{path: t.TempDir()}),
		app.WithOptions(app.Options{
			ProjectsRoot: t.TempDir(),
		}))
	err := a.InitLocalProject()
	require.NoError(t, err)
}

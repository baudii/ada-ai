package local

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/baudii/ada-ai/internal/project"
	"github.com/baudii/ada-ai/internal/project/nav/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockStore struct {
	navStore
	err error
}

func (m *mockStore) Materialize() error {
	return m.err
}

func TestNew(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		path string
		err  string
	}{
		{"valid absolute path", filepath.Join(t.TempDir(), "project"), ""},
		{"invalid relative path", "relative/path", "is not absolute"},
	}
	for _, v := range tests {
		t.Run(v.name, func(t *testing.T) {
			st, err := store.New(t.TempDir())
			require.NoError(t, err)
			lp, err := New(v.path, st)
			if v.err != "" {
				assert.ErrorContains(t, err, v.err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, v.path, lp.root)
			}
		})
	}
}

func TestNew_DirCreationFailure(t *testing.T) {
	t.Parallel()
	d := filepath.Join(t.TempDir(), "project")
	f, err := os.Create(d)
	require.NoError(t, err)
	_ = f.Close()
	st, err := store.New(t.TempDir())
	require.NoError(t, err)
	_, err = New(d, st)
	assert.ErrorContains(t, err, "create nav path:")
}

func TestMaterializeFail(t *testing.T) {
	t.Parallel()
	store := &mockStore{
		err: assert.AnError,
	}
	lp, err := New(t.TempDir(), store)
	require.NoError(t, err)
	err = lp.Materialize(nil, nil)
	assert.ErrorContains(t, err, "setup nav")
}

func TestMaterialize_Integration(t *testing.T) {
	t.Parallel()
	projRoot := filepath.Join(t.TempDir(), "project")
	st, err := store.New(projRoot)
	require.NoError(t, err)
	lp, err := New(projRoot, st)
	require.NoError(t, err)
	err = st.AddNav(project.Structure, "json", []byte(`{"a":0,"b":{"d":0,"e":0},"c":{}}`))
	require.NoError(t, err)
	err = lp.Materialize(project.DefaultFileHandler, project.DefaultFolderHandler)
	require.NoError(t, err)
	assert.True(t, isMaterialized(st.Tree(), lp.root))
	assert.FileExists(t, filepath.Join(projRoot, "nav", project.Structure+".json"))
}

func isMaterialized(m map[string]any, path string) bool {
	for k, v := range m {
		p := filepath.Join(path, k)
		info, err := os.Stat(p)
		if err != nil {
			return false
		}
		m2, ok := v.(map[string]any)
		if info.IsDir() != ok {
			return false
		}
		if ok && !isMaterialized(m2, p) {
			return false
		}
	}

	return true
}

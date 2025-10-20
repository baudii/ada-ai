package store

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockNavItem struct {
	err error
}

func (m *mockNavItem) Materialize() error {
	return m.err
}

func TestNew_DirCreationFailure(t *testing.T) {
	t.Parallel()
	d := filepath.Join(t.TempDir(), "file.txt")
	f, err := os.Create(d)
	require.NoError(t, err)
	_ = f.Close()
	_, err = New(d)
	require.ErrorContains(t, err, "create nav root")
}

func TestTree(t *testing.T) {
	t.Parallel()
	p := &store{tree: map[string]any{"p": "a"}}
	s := p.Tree()
	assert.Equal(t, p.tree, s)
}

func TestMaterialize_Unit(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		store *store
		err   string
	}{
		{"create folder fail", &store{root: ""}, "create nav folder"},
		{"save project structure file", &store{root: t.TempDir(), tree: map[string]any{"f": make(chan int)}}, "save project structure file"},
		{"create nav file", &store{
			root:  t.TempDir(),
			tree:  map[string]any{},
			items: map[string]navItem{"file.txt": &mockNavItem{err: fmt.Errorf("mock error")}},
		}, "create nav file"},
		{"success", &store{
			root:  t.TempDir(),
			tree:  map[string]any{},
			items: map[string]navItem{"file.txt": &mockNavItem{err: nil}},
		}, ""},
	}

	for _, v := range tests {
		t.Run(v.name, func(t *testing.T) {
			err := v.store.Materialize()
			if v.err != "" {
				assert.ErrorContains(t, err, v.err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAddNav(t *testing.T) {
	t.Parallel()
	tests := []struct {
		key     string
		ext     string
		content []byte
		err     string
	}{
		{"file1", "txt", []byte("content1"), ""},
		{"structure", "json", []byte(`{"a":"b"}`), ""},
		{"structure", "json", []byte(`invalid json`), "parse structure content"},
	}

	for i, v := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			n, err := New(t.TempDir())
			require.NoError(t, err)
			err = n.AddNav(v.key, v.ext, v.content)
			if v.err != "" {
				assert.ErrorContains(t, err, v.err)
			} else if assert.NoError(t, err) {
				assert.Contains(t, n.items, v.key)
			}
		})
	}
}

func TestTryLoadNav(t *testing.T) {
	t.Parallel()
	n, err := New(t.TempDir())
	require.NoError(t, err)
	fn := "file1"
	cont := []byte("{}")
	err = n.AddNav(fn, "txt", cont)
	require.NoError(t, err)
	err = n.Materialize()
	require.NoError(t, err)
	var res []byte
	ok := n.TryLoadNav(fmt.Sprintf("%v.%v", fn, "txt"), &res)
	require.True(t, ok)
	assert.Equal(t, cont, res)
	ok = n.TryLoadNav("file2.txt", &res)
	require.False(t, ok)
}

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
	content string
	err     error
}

func (m *mockNavItem) Materialize() error {
	return m.err
}

func (m *mockNavItem) CompactContent() (string, error) {
	return m.content, m.err
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

func TestNavContent(t *testing.T) {
	t.Parallel()
	s := &store{items: map[string]navItem{"key1": &mockNavItem{content: "{}", err: nil}}}
	tests := []struct {
		name string
		key  string
		err  string
	}{
		{"valid nav content", "key1", ""},
		{"invalid nav content", "key2", "not found"},
	}

	for _, v := range tests {
		t.Run(v.name, func(t *testing.T) {
			res, err := s.NavContent(v.key)
			if v.err != "" {
				assert.ErrorContains(t, err, v.err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, "{}", res)
			}
		})
	}
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
	res, err := n.LoadNav(fmt.Sprintf("%v.%v", fn, "txt"))
	require.NoError(t, err)
	assert.Equal(t, cont, res)
	_, err = n.LoadNav("file2.txt")
	require.Error(t, err)
}

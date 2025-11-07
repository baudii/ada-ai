package fsproject

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/baudii/ada-ai/internal/core/nav"
	"github.com/baudii/ada-ai/internal/core/project"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var empty = func(p string) error { return nil }

type mockStore struct {
	items map[string]nav.FileInfo
	file  *nav.FileInfo
}

func (m *mockStore) Add(key, path string, content []byte) {

}
func (m *mockStore) Get(k string) (*nav.FileInfo, bool) {
	return m.file, m.file != nil
}
func (m *mockStore) GetItems() map[string]nav.FileInfo {
	return m.items
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
			st := nav.New()
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
	st := nav.New()
	_, err = New(d, st)
	assert.ErrorContains(t, err, "create nav path:")
}

func TestAddNav(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		key      string
		ext      string
		content  []byte
		expected map[string]any
		err      string
	}{
		{
			name:    "success: add regular nav file",
			key:     "file1",
			ext:     "txt",
			content: []byte("content1"),
		},
		{
			name:     "success: parse structure content",
			key:      project.Structure,
			ext:      "json",
			content:  []byte(`{"a":"b"}`),
			expected: map[string]any{"a": "b"},
		},
		{
			name:    "fail: parse structure content",
			key:     project.Structure,
			ext:     "json",
			content: []byte(`invalid json`),
			err:     "parse structure content",
		},
	}

	for _, v := range tests {
		t.Run(v.name, func(t *testing.T) {
			ms := &mockStore{}
			n, err := New(t.TempDir(), ms)
			require.NoError(t, err)
			err = n.AddNav(v.key, v.ext, v.content)
			if v.err != "" {
				assert.ErrorContains(t, err, v.err)
			} else if assert.NoError(t, err) {
				assert.Equal(t, v.expected, n.Tree)
			}
		})
	}
}

func TestLoadNav_Fail(t *testing.T) {
	t.Parallel()
	lp, err := New(t.TempDir(), nav.New())
	require.NoError(t, err)
	_, err = lp.LoadNav("nonexistent.txt")
	assert.ErrorContains(t, err, "read file")
}

func TestContent(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		key      string
		content  string
		mockFile *nav.FileInfo
		expected string
		err      string
	}{
		{
			name: "invalid nav content",
			key:  "key2",
			err:  "not found"},
		{
			name:     "valid nav content",
			key:      "key1",
			mockFile: &nav.FileInfo{},
			content: `
			{
				"file1.txt":"content1",
				"folder1":{
					"file2.txt":"content2"
				}
			}`,
			expected: `{"file1.txt":"content1","folder1":{"file2.txt":"content2"}}`,
		},
		{
			name:     "valid nav content",
			key:      "key1",
			mockFile: &nav.FileInfo{},
			content:  `{`,
			err:      "compact content",
		},
	}

	for _, v := range tests {
		t.Run(v.name, func(t *testing.T) {
			s := &Data{store: &mockStore{
				file: v.mockFile,
			}}
			if v.mockFile != nil {
				v.mockFile.Content = []byte(v.content)
			}
			res, err := s.NavContent(v.key)
			if v.err != "" {
				assert.ErrorContains(t, err, v.err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, v.expected, res)
			}
		})
	}
}

func TestMaterialize_FailCreateFile(t *testing.T) {
	t.Parallel()
	d := t.TempDir()
	m := &mockStore{
		items: map[string]nav.FileInfo{"file1": {Filename: "folder"}},
	}

	s := &Data{navRoot: d, store: m}
	err := os.MkdirAll(filepath.Join(d, "folder"), 0o755)
	require.NoError(t, err)
	err = s.Materialize(empty, empty)
	assert.ErrorContains(t, err, "create nav file")
}

func TestMaterialize_FailMkdir(t *testing.T) {
	t.Parallel()
	d := filepath.Join(t.TempDir(), "project")
	f, err := os.Create(d)
	require.NoError(t, err)
	_ = f.Close()
	s := &Data{
		root:    d,
		navRoot: filepath.Join(d, NavFolder),
		store: &mockStore{
			items: map[string]nav.FileInfo{
				"file1": {
					Filename: "file1.txt",
					Content:  []byte("content1"),
				},
			},
		},
	}
	err = s.Materialize(empty, empty)
	assert.ErrorContains(t, err, "create nav root")
}

func TestCreateFile(t *testing.T) {
	t.Parallel()
	d := t.TempDir()
	s := &Data{navRoot: d}
	err := s.HandleFile(filepath.Join(d, "file.txt"), []byte("content"))
	require.NoError(t, err)
	data, err := os.ReadFile(filepath.Join(d, "file.txt"))
	require.NoError(t, err)
	assert.Equal(t, []byte("content"), data)
}

package local

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/baudii/ada-ai/internal/project"
	"github.com/baudii/ada-ai/internal/project/nav"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockStore struct {
	file *nav.File
	err  error
}

func (m *mockStore) Materialize(root string) error {
	return m.err
}
func (m *mockStore) Add(key, path string, content []byte) {

}
func (m *mockStore) Get(k string) (*nav.File, bool) {
	return m.file, m.file != nil
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
				assert.Equal(t, v.expected, n.tree)
			}
		})
	}
}

// func TestLoadNav(t *testing.T) {
// 	t.Parallel()
// 	n, err := New(t.TempDir(), &mockStore{})
// 	require.NoError(t, err)
// 	fn := "file1"
// 	cont := []byte("{}")
// 	err = n.AddNav(fn, "txt", cont)
// 	require.NoError(t, err)
// 	err = n.Materialize(nil, nil)
// 	require.NoError(t, err)
// 	res, err := n.LoadNav(fmt.Sprintf("%v.%v", fn, "txt"))
// 	require.NoError(t, err)
// 	assert.Equal(t, cont, res)
// 	_, err = n.LoadNav("file2.txt")
// 	require.Error(t, err)
// }

func TestTree(t *testing.T) {
	t.Parallel()
	p := &proj{tree: map[string]any{"p": "a"}}
	s := p.Tree()
	assert.Equal(t, p.tree, s)
}

func TestContent(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		key      string
		content  string
		mockFile *nav.File
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
			mockFile: &nav.File{},
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
			mockFile: &nav.File{},
			content:  `{`,
			err:      "compact content",
		},
	}

	for _, v := range tests {
		t.Run(v.name, func(t *testing.T) {
			s := &proj{store: &mockStore{
				file: v.mockFile,
			}}
			if v.mockFile != nil {
				v.mockFile.Content = []byte(v.content)
			}
			res, err := s.Content(v.key)
			if v.err != "" {
				assert.ErrorContains(t, err, v.err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, v.expected, res)
			}
		})
	}
}

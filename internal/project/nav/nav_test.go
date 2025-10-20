package nav

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Parallel()
	s := New()
	assert.NotNil(t, s)
}

func TestMaterialize_Unit(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		root  string
		store *store
		err   string
	}{
		{
			name:  "fail: create folder",
			store: &store{},
			root:  "",
			err:   "create nav folder",
		},
		{
			name: "success",
			store: &store{
				items: map[string]*File{
					"file": {
						Content:  nil,
						Filename: "file.txt",
					},
				},
			},
			root: t.TempDir(),
		},
	}

	for _, v := range tests {
		t.Run(v.name, func(t *testing.T) {
			err := v.store.Materialize(v.root)
			if v.err != "" {
				assert.ErrorContains(t, err, v.err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGet(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		expectedOk bool
	}{
		{
			name:       "found",
			expectedOk: true,
		},
	}

	for _, v := range tests {
		t.Run(v.name, func(t *testing.T) {
			s := &store{}
			if v.expectedOk {
				s.items = map[string]*File{
					"file1": {
						Filename: "file1.txt",
					},
				}
			}
			f, ok := s.Get("file1")
			assert.Equal(t, v.expectedOk, ok)
			assert.Equal(t, s.items["file1"], f)
		})
	}
}

func TestAdd(t *testing.T) {
	t.Parallel()
	s := &store{items: make(map[string]*File)}
	s.Add("key1", "file1.txt", []byte("content1"))
	f, ok := s.Get("key1")
	assert.True(t, ok)
	assert.Equal(t, "file1.txt", f.Filename)
	assert.Equal(t, []byte("content1"), f.Content)
}

func TestMaterialize_FailWrite(t *testing.T) { /*  */
	t.Parallel()
	s := &store{
		items: map[string]*File{
			"file": {
				Filename: string([]byte{0x00}), // Invalid filename
				Content:  []byte("content"),
			},
		},
	}
	err := s.Materialize(t.TempDir())
	assert.ErrorContains(t, err, "create nav file")
}

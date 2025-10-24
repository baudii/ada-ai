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
				s.items = map[string]FileInfo{
					"file1": {
						Filename: "file1.txt",
					},
				}
			}
			f, ok := s.Get("file1")
			assert.Equal(t, v.expectedOk, ok)
			assert.Equal(t, s.items["file1"], *f)
		})
	}
}

func TestGetItems(t *testing.T) {
	t.Parallel()
	s := &store{
		items: map[string]FileInfo{
			"file1": {
				Filename: "file1.txt",
			},
			"file2": {
				Filename: "file2.txt",
			},
		},
	}
	items := s.GetItems()
	assert.Equal(t, 2, len(items))
	assert.Equal(t, "file1.txt", items["file1"].Filename)
	assert.Equal(t, "file2.txt", items["file2"].Filename)
}

func TestAdd(t *testing.T) {
	t.Parallel()
	s := &store{items: make(map[string]FileInfo)}
	s.Add("key1", "file1.txt", []byte("content1"))
	f, ok := s.Get("key1")
	assert.True(t, ok)
	assert.Equal(t, "file1.txt", f.Filename)
	assert.Equal(t, []byte("content1"), f.Content)
}

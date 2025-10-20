package nav

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNavContent(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		content  string
		expected string
		err      string
	}{
		{
			name: "valid content",
			content: `{
						"file1.txt":"content1",
						"folder1":{
							"file2.txt":"content2"
						}
					}`,
			expected: `{"file1.txt":"content1","folder1":{"file2.txt":"content2"}}`,
		},
		{
			name:    "invalid content",
			content: `{`,
			err:     "compact content",
		},
	}
	for _, v := range tests {
		t.Run(v.name, func(t *testing.T) {
			n := New(t.TempDir(), []byte(v.content))
			content, err := n.NavContent()
			if v.err != "" {
				assert.Error(t, err)
				assert.ErrorContains(t, err, v.err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, v.expected, content)
			}
		})
	}
}

func TestMaterialize(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		item   *item
		hasErr bool
	}{
		{"valid materialize", New(filepath.Join(t.TempDir(), "file.txt"), []byte("content")), false},
		{"invalid materialize", New(t.TempDir(), []byte("content")), true},
	}
	for _, v := range tests {
		t.Run(v.name, func(t *testing.T) {
			err := v.item.Materialize()
			if v.hasErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

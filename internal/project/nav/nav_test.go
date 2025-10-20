package nav

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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

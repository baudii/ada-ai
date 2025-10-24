package project_test

import (
	"testing"

	"github.com/baudii/ada-ai/internal/core/project"
	"github.com/stretchr/testify/assert"
)

var noErrHandler = func(p string) error { return nil }

func TestTraverse(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		tree  map[string]any
		hfile project.FileHandler
		hfold project.FolderHandler
		err   string
	}{
		{
			name: "valid traverse",
			tree: map[string]any{
				"file1.txt": float64(0),
				"folder1": map[string]any{
					"file2.txt": float64(0),
				},
			},
			hfile: noErrHandler, hfold: noErrHandler,
		},
		{
			name:  "empty tree",
			tree:  map[string]any{},
			hfile: noErrHandler, hfold: noErrHandler,
		},
		{
			name:  "fail: invalid tree",
			tree:  map[string]any{"key": true},
			hfile: noErrHandler, hfold: noErrHandler,
			err: "invalid tree",
		},
		{
			name: "fail: handle file nested",
			tree: map[string]any{
				"folder1": map[string]any{
					"folder2": map[string]any{
						"file3.txt": float64(0),
					},
				},
			},
			hfile: func(s string) error { return assert.AnError },
			hfold: noErrHandler,
			err:   "handle file",
		},
		{
			name:  "fail: handle folder",
			tree:  map[string]any{"folder": map[string]any{}},
			hfile: noErrHandler,
			hfold: func(s string) error { return assert.AnError },
			err:   "handle folder(s)",
		},
	}

	for _, v := range tests {
		t.Run(v.name, func(t *testing.T) {
			err := project.Traverse("", v.tree, v.hfile, v.hfold)
			if v.err != "" {
				assert.ErrorContains(t, err, v.err)
			} else if err != nil {
				assert.NoError(t, err)
			}
		})
	}
}

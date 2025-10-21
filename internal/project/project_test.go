package project_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/baudii/ada-ai/internal/project"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var noErrHandler = func(p string) error { return nil }

func TestTraverse(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		tree  map[string]any
		hfile project.Handler
		hfold project.Handler
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

func TestDefaultFolderHandler(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		path string
		err  string
	}{
		{"valid folder creation", t.TempDir() + "/newfolder", ""},
		{"invalid folder creation", filepath.Join(t.TempDir(), "file.txt"), "create folder"},
	}
	for _, v := range tests {
		t.Run(v.name, func(t *testing.T) {
			if v.err != "" {
				f, err := os.Create(v.path)
				require.NoError(t, err)
				_ = f.Close()
				err = project.DefaultFolderHandler(v.path)
				assert.ErrorContains(t, err, v.err)
			} else {
				err := project.DefaultFolderHandler(v.path)
				assert.NoError(t, err)
			}
		})
	}
}

func TestDefaultFileHandler(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		path string
		err  string
	}{
		{
			name: "valid file creation",
			path: filepath.Join(t.TempDir(), "newfile.txt"),
		},
		{
			name: "invalid file creation",
			path: t.TempDir(),
			err:  "create file",
		},
	}

	for _, v := range tests {
		t.Run(v.name, func(t *testing.T) {
			err := project.DefaultFileHandler(v.path)
			if v.err != "" {
				assert.ErrorContains(t, err, v.err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

package local_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/baudii/ada-ai/internal/project"
	"github.com/baudii/ada-ai/internal/project/local"
	"github.com/baudii/ada-ai/internal/project/nav"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMaterialize_Integration(t *testing.T) {
	t.Parallel()
	// Arrange
	structureContent := []byte(`{"a":0,"b":{"d":0,"e":0},"c":{}}`)
	otherContent := []byte("This is file 1")
	projRoot := filepath.Join(t.TempDir(), "project")
	st := nav.New()
	lp, err := local.New(projRoot, st)
	require.NoError(t, err)

	// Act
	err = lp.AddNav(project.Structure, "json", structureContent)
	require.NoError(t, err)
	err = lp.AddNav("file1", "txt", otherContent)
	require.NoError(t, err)
	err = lp.Materialize(project.DefaultFileHandler, project.DefaultFolderHandler)
	require.NoError(t, err)
	r1, err := lp.LoadNav(project.Structure + ".json")
	require.NoError(t, err)
	r2, err := lp.LoadNav("file1.txt")
	require.NoError(t, err)

	// Assert
	assert.True(t, isMaterialized(lp.Tree(), projRoot))
	path := filepath.Join(projRoot, "nav", project.Structure+".json")
	assert.FileExists(t, path)
	assert.FileExists(t, filepath.Join(projRoot, "nav", "file1.txt"))
	assert.Equal(t, structureContent, r1)
	assert.Equal(t, otherContent, r2)
}

func isMaterialized(m map[string]any, path string) bool {
	for k, v := range m {
		p := filepath.Join(path, k)
		info, err := os.Stat(p)
		if err != nil {
			return false
		}
		m2, ok := v.(map[string]any)
		if info.IsDir() != ok {
			return false
		}
		if ok && !isMaterialized(m2, p) {
			return false
		}
	}

	return true
}

package folders_test

import (
	"fmt"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/baudii/ada-ai/internal/infra/folders"
	"github.com/stretchr/testify/assert"
)

func TestFromExecutable(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		relPath  string
		basePath string
		err      error
	}{
		{
			name:     "success: normal case",
			relPath:  filepath.Join("a", "b", "c"),
			basePath: t.TempDir(),
		},
		{
			name: "fail: executable error",
			err:  fmt.Errorf("failed to get base"),
		},
	}

	for i, v := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			executable := func() (string, error) { return filepath.Join(v.basePath, "exe.exe"), v.err }
			if v.err != nil {
				assert.PanicsWithError(t, v.err.Error(), func() { folders.FromExecutable(v.relPath, executable) })
			} else {
				abspath := folders.FromExecutable(v.relPath, executable)
				assert.Equal(t, filepath.Join(v.basePath, v.relPath), abspath)
			}
		})
	}
}

func TestInsertSuffix(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		path     string
		suffix   string
		expected string
	}{
		{
			name:     "success: normal case",
			path:     filepath.Join("a", "b", "somefile.txt"),
			suffix:   "_somesuffix",
			expected: filepath.Join("a", "b", "somefile_somesuffix.txt"),
		},
		{
			name:     "success: no extension",
			path:     filepath.Join("a", "b", "somefile"),
			suffix:   "_somesuffix",
			expected: filepath.Join("a", "b", "somefile_somesuffix"),
		},
		{
			name:     "success: empty suffix",
			path:     filepath.Join("a", "b", "somefile.txt"),
			suffix:   "",
			expected: filepath.Join("a", "b", "somefile.txt"),
		},
	}

	for _, v := range tests {
		t.Run(v.name, func(t *testing.T) {
			res := folders.InsertFsuffix(v.path, v.suffix)
			assert.Equal(t, v.expected, res)
		})
	}
}

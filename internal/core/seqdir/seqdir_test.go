package seqdir_test

import (
	"io/fs"
	"path/filepath"
	"testing"

	"github.com/baudii/ada-ai/internal/core/seqdir"
	"github.com/stretchr/testify/assert"
)

var mockEntries = []fs.DirEntry{
	mockDirEntry{name: "0", isDir: true},
	mockDirEntry{name: "1", isDir: true},
	mockDirEntry{name: "3", isDir: true},
	mockDirEntry{name: "text", isDir: true},
	mockDirEntry{name: "4", isDir: false},
}

type mockDirEntry struct {
	name  string
	isDir bool
}

func (m mockDirEntry) IsDir() bool {
	return m.isDir
}

func (m mockDirEntry) Info() (fs.FileInfo, error) {
	return nil, nil
}

func (m mockDirEntry) Type() fs.FileMode {
	return 0
}

func (m mockDirEntry) Name() string {
	return m.name
}

func TestProjectFolder(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		mockFiles    []fs.DirEntry
		mockErr      error
		mode         seqdir.Mode
		expectedName string
		expectErr    string
	}{
		{
			name:      "fails to read dir",
			mockErr:   assert.AnError,
			mode:      seqdir.UseLatest,
			expectErr: "assert.AnError",
		},
		{
			name:         "fails to read dir",
			mockErr:      fs.ErrNotExist,
			mode:         seqdir.UseLatest,
			expectedName: filepath.Join("basepath", "0"),
		},
		{
			name:         "create new with empty folder",
			mockFiles:    []fs.DirEntry{},
			mode:         seqdir.CreateNew,
			expectedName: filepath.Join("basepath", "0"),
		},
		{
			name:         "use latest with empty folder",
			mockFiles:    []fs.DirEntry{},
			mode:         seqdir.UseLatest,
			expectedName: filepath.Join("basepath", "0"),
		},
		{
			name:         "create new with existing folders",
			mockFiles:    mockEntries,
			mockErr:      nil,
			mode:         seqdir.CreateNew,
			expectedName: filepath.Join("basepath", "5"),
		},
		{
			name:         "use latest with existing folders",
			mockFiles:    mockEntries,
			mockErr:      nil,
			mode:         seqdir.UseLatest,
			expectedName: filepath.Join("basepath", "3"),
		},
		{
			name:         "unknown mode",
			mockFiles:    mockEntries,
			mockErr:      nil,
			mode:         seqdir.Mode(99),
			expectedName: "",
			expectErr:    "unknown mode",
		},
	}

	for _, v := range tests {
		t.Run(v.name, func(t *testing.T) {
			t.Parallel()
			sd := seqdir.New(func(name string) ([]fs.DirEntry, error) {
				return v.mockFiles, v.mockErr
			})
			folder, err := sd.ProjectFolder("basepath", v.mode)
			if v.expectErr != "" {
				assert.ErrorContains(t, err, v.expectErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, v.expectedName, folder)
			}
		})
	}
}

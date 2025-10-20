package folder_test

import (
	"io/fs"
	"path/filepath"
	"testing"

	"github.com/baudii/ada-ai/internal/project/folder"
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
		mode         folder.Mode
		expectedName string
		expectErr    string
	}{
		{"fails to read dir", nil, fs.ErrNotExist, folder.UseLatest, "", "read dir"},
		{"create new with empty folder", []fs.DirEntry{}, nil, folder.CreateNew, filepath.Join("basepath", "0"), ""},
		{"use latest with empty folder", []fs.DirEntry{}, nil, folder.UseLatest, filepath.Join("basepath", "0"), ""},
		{"create new with existing folders", mockEntries, nil, folder.CreateNew, filepath.Join("basepath", "5"), ""},
		{"use latest with existing folders", mockEntries, nil, folder.UseLatest, filepath.Join("basepath", "3"), ""},
		{"unknown mode", mockEntries, nil, folder.Mode(99), "", "unknown mode"},
	}

	for _, v := range tests {
		t.Run(v.name, func(t *testing.T) {
			t.Parallel()
			sd := folder.NewSeqDir(func(name string) ([]fs.DirEntry, error) {
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

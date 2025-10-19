package project_test

import (
	"io/fs"
	"path/filepath"
	"testing"

	"github.com/baudii/ada-ai/internal/project"
	"github.com/stretchr/testify/assert"
)

var smdr = &mockDirReader{
	files: []fs.DirEntry{
		mockDirEntry{name: "0", isDir: true},
		mockDirEntry{name: "1", isDir: true},
		mockDirEntry{name: "3", isDir: true},
		mockDirEntry{name: "text", isDir: true},
		mockDirEntry{name: "4", isDir: false},
	},
	err: nil,
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

type mockDirReader struct {
	files []fs.DirEntry
	err   error
}

func (m *mockDirReader) ReadDir(path string) ([]fs.DirEntry, error) {
	return m.files, m.err
}

func (m *mockDirReader) Open(name string) (fs.File, error) {
	return nil, nil
}

func TestProjectFolder(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		mockFiles    []fs.DirEntry
		mockErr      error
		mode         project.Mode
		expectedName string
		expectErr    string
	}{
		{"fails to read dir", nil, fs.ErrNotExist, project.UseLatest, "", "read dir"},
		{"create new with empty folder", []fs.DirEntry{}, nil, project.CreateNew, filepath.Join("basepath", "0"), ""},
		{"use latest with empty folder", []fs.DirEntry{}, nil, project.UseLatest, filepath.Join("basepath", "0"), ""},
		{"create new with existing folders", smdr.files, nil, project.CreateNew, filepath.Join("basepath", "5"), ""},
		{"use latest with existing folders", smdr.files, nil, project.UseLatest, filepath.Join("basepath", "3"), ""},
		{"unknown mode", smdr.files, nil, project.Mode(99), "", "unknown mode"},
	}

	for _, v := range tests {
		t.Run(v.name, func(t *testing.T) {
			t.Parallel()
			mde := &mockDirReader{files: v.mockFiles, err: v.mockErr}
			sd := project.NewSeqDir(mde)
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

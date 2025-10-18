package projects

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Parallel()
	tests := []struct {
		structure      string
		path           string
		expected       *localProj
		structureError string
		expectError    string
	}{
		{`{"a":"b"}`, filepath.Join(t.TempDir(), "/a/b/c"), &localProj{structure: map[string]any{"a": "b"}, lastFolder: LastFolder, navs: make(map[string]nav)}, "", ""},
		{`}`, filepath.Join(t.TempDir(), "/a/b/c"), nil, "invalid character", ""},
		{`{}`, "a/b/c", nil, "", "is not absolute"},
		{`{"a":"b"}`, filepath.Join(t.TempDir(), "/b/c"), nil, "", "create base path"},
	}
	for i, v := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			switch v.expectError {
			case "create base path":
				// simulate base path creation failure by creating a file in the middle
				// of the base path
				f, err := os.Create(filepath.Dir(v.path))
				require.NoError(t, err)
				_ = f.Close()
			}
			actualLocalProj, err := New(v.path)
			if v.expectError != "" {
				assert.ErrorContains(t, err, v.expectError)
			} else if assert.NoError(t, err) {
				err = json.Unmarshal([]byte(v.structure), &actualLocalProj.structure)
				if v.structureError != "" {
					assert.ErrorContains(t, err, v.structureError)
					return
				}
				v.expected.projectRoot = v.path
				v.expected.navPath = navPath(v.path)
				v.expected.lastFolder = nil
				actualLocalProj.lastFolder = nil
				assert.Equal(t, v.expected, actualLocalProj)
			}
		})
	}
}

func TestTraverseFails(t *testing.T) {
	tests := []struct {
		m   map[string]any
		err string
	}{
		{map[string]any{"a": true}, fmt.Sprintf("map or float64, got %T", true)},
		{map[string]any{"a": map[string]any{"b": true}}, fmt.Sprintf("map or float64, got %T", true)},
		{map[string]any{"a": float64(0)}, "handle file"},
		{map[string]any{"a": map[string]any{"b": float64(0)}}, "handle folder(s)"},
	}

	for i, v := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			dir := t.TempDir()
			a := path.Join(dir, "a")
			switch v.err {
			case "handle file":
				err := os.MkdirAll(a, 0744)
				require.NoError(t, err)
			case "handle folder(s)":
				f, err := os.Create(a)
				require.NoError(t, err)
				err = f.Close()
				require.NoError(t, err)
			}
			err := TraverseStructure(dir, v.m, DefaultFileHandler, DefaultFolderHandler)
			assert.ErrorContains(t, err, v.err)
		})
	}
}

func TestMaterialize_Unit(t *testing.T) {
	t.Parallel()
	tests := []struct {
		hfile handler
		hfold handler
		lproj *localProj
		err   string
	}{
		{nil, nil, &localProj{projectRoot: ""}, "create project root folder"},
		{nil, nil, &localProj{projectRoot: t.TempDir(), navPath: ""}, "create nav folder"},
		{nil, nil, &localProj{projectRoot: t.TempDir(), navPath: t.TempDir(), structure: map[string]any{"f": make(chan int)}}, "save project structure file"},
		{nil, nil, &localProj{
			projectRoot: t.TempDir(),
			navPath:     t.TempDir(),
			structure:   map[string]any{},
			navs:        map[string]nav{"file.txt": {filepath: filepath.Join(t.TempDir(), "invalid", "file.txt")}},
		}, "create nav file"},
		{nil, nil, &localProj{
			projectRoot: t.TempDir(),
			navPath:     t.TempDir(),
			structure:   map[string]any{},
			navs:        map[string]nav{"file.txt": {filepath: filepath.Join(t.TempDir(), "file.txt"), content: []byte("content")}},
		}, ""},
	}

	for i, v := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			err := v.lproj.Materialize(v.hfile, v.hfold)
			if v.err != "" {
				assert.ErrorContains(t, err, v.err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLastFolder_Fails(t *testing.T) {
	t.Parallel()
	dir := filepath.Join(t.TempDir(), "file.txt")
	f, err := os.Create(dir)
	require.NoError(t, err)
	_ = f.Close()
	_, err = LastFolder(dir)
	assert.ErrorContains(t, err, "The system cannot find the path")
}

func TestLastFolder(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	folder1 := filepath.Join(dir, "0")
	folder2 := filepath.Join(dir, "1")
	folder3 := filepath.Join(dir, "2")
	err := os.MkdirAll(folder1, 0755)
	require.NoError(t, err)
	err = os.MkdirAll(folder2, 0755)
	require.NoError(t, err)
	err = os.MkdirAll(folder3, 0755)
	require.NoError(t, err)

	last, err := LastFolder(dir)
	require.NoError(t, err)

	assert.Equal(t, 2, last)
}

func TestMaterialize_Integration(t *testing.T) {
	t.Parallel()
	lproj, err := New(t.TempDir())
	require.NoError(t, err)
	err = json.Unmarshal([]byte(`{"a":0,"b":{"d":0,"e":0},"c":{}}`), &lproj.structure)
	require.NoError(t, err)
	err = lproj.AddNav("file", "txt", []byte("content"))
	require.NoError(t, err)
	err = lproj.Materialize(DefaultFileHandler, DefaultFolderHandler)
	require.NoError(t, err)
	assert.True(t, isMaterialized(lproj.structure, lproj.projectRoot))
	assert.FileExists(t, filepath.Join(lproj.navPath, "file.txt"))
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

func TestStructure(t *testing.T) {
	t.Parallel()
	p := &localProj{structure: map[string]any{"p": "a"}}
	s := p.Structure()
	assert.Equal(t, p.structure, s)
}

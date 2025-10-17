package projects

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type resolver struct {
	path string
}

func (r *resolver) ResolveProjectPath() string {
	return r.path
}

func TestNew(t *testing.T) {
	t.Parallel()
	tests := []struct {
		structure   string
		resolver    *resolver
		expected    *localProj
		expectError string
	}{
		{`{"a":"b"}`, &resolver{filepath.Join(t.TempDir(), "/a/b/c")}, &localProj{structure: map[string]any{"a": "b"}, base: "/a/b/c", uniqFoldName: uniqueIndexFolder}, ""},
		{`}`, &resolver{"/a/b/c"}, nil, "unmarshal project structure"},
		{`{}`, &resolver{"a/b/c"}, nil, "is not absolute"},
		{`{"a":"b"}`, &resolver{filepath.Join(t.TempDir(), "/b/c")}, nil, "create base path"},
	}

	for i, v := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			switch v.expectError {
			case "create base path":
				// simulate base path creation failure by creating a file in the middle
				// of the base path
				f, err := os.Create(filepath.Dir(v.resolver.path))
				require.NoError(t, err)
				_ = f.Close()
			}
			actualLocalProj, err := New([]byte(v.structure), v.resolver)
			if v.expectError != "" {
				assert.ErrorContains(t, err, v.expectError)
			} else if assert.NoError(t, err) {
				v.expected.base = v.resolver.path
				v.expected.uniqFoldName = nil
				actualLocalProj.uniqFoldName = nil
				assert.Equal(t, v.expected, actualLocalProj)
			}
		})
	}
}

func TestMaterializeFails(t *testing.T) {
	tests := []struct {
		m   map[string]any
		err string
	}{
		{map[string]any{"a": true}, fmt.Sprintf("map or float64, got %T", true)},
		{map[string]any{"a": map[string]any{"b": true}}, fmt.Sprintf("map or float64, got %T", true)},
		{map[string]any{"a": float64(0)}, "create file"},
		{map[string]any{"a": map[string]any{"b": float64(0)}}, "create folder(s)"},
	}

	for i, v := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			dir := t.TempDir()
			a := path.Join(dir, "a")
			switch v.err {
			case "create file":
				err := os.MkdirAll(a, 0744)
				require.NoError(t, err)
			case "create folder(s)":
				f, err := os.Create(a)
				require.NoError(t, err)
				err = f.Close()
				require.NoError(t, err)
			}
			err := materializeStructure(dir, v.m)
			assert.ErrorContains(t, err, v.err)
		})
	}
}

func TestMaterialize(t *testing.T) {
	t.Parallel()
	tests := []struct {
		structure map[string]any
		uniqFold  func(string) (string, error)
		errmsg    string
	}{
		{map[string]any{"a": make(chan int)}, uniqueIndexFolder, "save project structure file"},
		{map[string]any{"a": map[string]any{"b": float64(0), "c": float64(0), "d": float64(0)}, "b": float64(0)}, uniqueIndexFolder, ""},
		{map[string]any{"a": float64(0)}, uniqueIndexFolder, "get unique folder name: readdir"},
		{nil, func(string) (string, error) { return "", nil }, "create project root folder"},
	}

	for i, v := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			d := t.TempDir()
			switch v.errmsg {
			case "get unique folder name: readdir":
				v.errmsg = "any"
				d = filepath.Join(d, "p")
				f, err := os.Create(d)
				require.NoError(t, err)
				t.Cleanup(func() { _ = f.Close() })
			case "create project root folder":
				d = ""
			}
			lp := &localProj{
				structure:    v.structure,
				projectRoot:  "",
				base:         d,
				uniqFoldName: v.uniqFold,
			}
			if v.errmsg == "" {
				err := os.MkdirAll(path.Join(d, "0"), 0744)
				require.NoError(t, err)
			}
			err := lp.Materialize()
			if v.errmsg == "" && assert.NoError(t, err) {
				assert.True(t, isMaterialized(v.structure, lp.projectRoot))
			} else if v.errmsg == "any" {
				assert.Error(t, err)
			} else {
				assert.ErrorContains(t, err, v.errmsg)
			}
		})
	}
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

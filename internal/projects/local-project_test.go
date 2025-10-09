package projects

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
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
		base        string
		isValidJson bool
		isAbs       bool
		expected    *localProj
	}{
		{`}`, "a/b/c", false, false, nil},
		{`{}`, "a/b/c", true, false, nil},
		{`{"}`, "fold", false, true, nil},
		{`{"a":"b"}`, "fold", true, true, &localProj{map[string]any{"a": "b"}, "fold", "", uniqueIndexFolder}},
	}

	for i, v := range tests {
		d := ""
		if v.isAbs {
			d = t.TempDir()

			if v.expected != nil {
				v.expected.base = filepath.Join(d, v.expected.base)
			}
		}
		base := filepath.Join(d, v.base)
		actualLocalProj, err := New([]byte(v.structure), &resolver{base})
		if !v.isValidJson {
			assert.ErrorContains(t, err, "unmarshal project structure", "test: %v", i)
		} else if !v.isAbs {
			assert.ErrorContains(t, err, "is not absolute", "test: %v", i)
		} else {
			v.expected.uniqFoldName = nil
			actualLocalProj.uniqFoldName = nil
			assert.Equal(t, v.expected, actualLocalProj, "test: %v", i)
		}
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
		dir := t.TempDir()
		a := path.Join(dir, "a")
		switch v.err {
		case "create file":
			os.MkdirAll(a, 0744)
		case "create folder(s)":
			f, err := os.Create(a)
			require.NoError(t, err)
			f.Close()
		}
		err := materialize(dir, v.m)
		assert.ErrorContains(t, err, v.err, "test: %v", i)
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
		d := t.TempDir()
		switch v.errmsg {
		case "get unique folder name: readdir":
			d = filepath.Join(d, "p")
			f, err := os.Create(d)
			if !assert.NoError(t, err, "test: %v", i) {
				continue
			}
			t.Cleanup(func() { f.Close() })
		case "create project root folder":
			d = ""
		}
		lp := &localProj{v.structure, d, "", v.uniqFold}
		if v.errmsg == "" {
			err := os.MkdirAll(path.Join(d, "0"), 0744)
			if !assert.NoError(t, err) {
				continue
			}
		}
		err := lp.Materialize()
		if v.errmsg == "" && assert.NoError(t, err, "test: %v", i) {
			assert.True(t, isMaterialized(v.structure, lp.projectRoot), "test: %v", i)
			continue
		}
		assert.ErrorContains(t, err, v.errmsg, "test: %v", i)
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

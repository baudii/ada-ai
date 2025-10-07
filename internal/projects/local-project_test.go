package projects

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Parallel()
	tests := []struct {
		structure   string
		base        string
		isValidJson bool
		isAbs       bool
		expected    *LocalProj
	}{
		{`}`, "a/b/c", false, false, nil},
		{`{}`, "a/b/c", true, false, nil},
		{`{"}`, "fold", false, true, nil},
		{`{"a":"b"}`, "fold", true, true, &LocalProj{structure: map[string]any{"a": "b"}, projectRoot: "fold"}},
	}

	for _, v := range tests {
		d := ""
		if v.isAbs {
			d = t.TempDir()

			if v.expected != nil {
				v.expected.projectRoot = filepath.Join(d, v.expected.projectRoot)
			}
		}
		base := filepath.Join(d, v.base)
		p, err := New([]byte(v.structure), base)
		if !v.isValidJson {
			assert.ErrorContains(t, err, "unmarshal project structure: ")
		} else if !v.isAbs {
			assert.ErrorContains(t, err, "path must be absolute")
		} else {
			assert.Equal(t, v.expected, p)
		}
	}
}

func TestMaterialize(t *testing.T) {
	t.Parallel()
	tests := []struct {
		structure map[string]any
		projroot  string
		errmsg    string
	}{
		{map[string]any{"a": make(chan int)}, "p", "save project structure file"},
		{map[string]any{"a": map[string]any{"b": float64(0), "c": float64(0), "d": float64(0)}, "b": float64(0)}, "p", ""},
		{map[string]any{"a": map[string]any{"b": map[string]any{"c": true}}}, "p", "invalid structure: value must be either a map or float64, got"},
		{map[string]any{"a": map[string]any{}}, "p", "remove existing project root"},
		{map[string]any{"a": float64(0)}, "p", "is not a directory"},
		{map[string]any{"a": float64(0)}, "p", "create project root folder"},
	}

	for _, v := range tests {
		d := filepath.Join(t.TempDir(), v.projroot)
		switch v.errmsg {
		case "remove existing project root":
			err := os.MkdirAll(d, 0o755)
			require.NoError(t, err)
			f, err := os.Create(filepath.Join(d, "a"))
			require.NoError(t, err)
			t.Cleanup(func() { f.Close() })
		case "is not a directory":
			f, err := os.Create(d)
			require.NoError(t, err)
			t.Cleanup(func() { f.Close() })
		case "create project root folder":
			d = ""
		}
		lp := &LocalProj{v.structure, d}
		err := lp.Materialize()
		if v.errmsg == "" {
			assert.True(t, isMaterialized(v.structure, d))
			continue
		}
		assert.ErrorContains(t, err, v.errmsg)
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
	p := &LocalProj{structure: map[string]any{}}
	s := p.Structure()
	assert.Equal(t, p.structure, s)
}

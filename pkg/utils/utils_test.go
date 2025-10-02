package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetAbsolutePath(t *testing.T) {
	t.Cleanup(func() { executable = os.Executable })
	tests := []struct {
		relPath  string
		basePath string
		err      error
	}{
		{"anything", t.TempDir(), nil},
		{"a/b/c/", t.TempDir(), nil},
		{"/a/s/d", t.TempDir(), nil},
		{"", t.TempDir(), nil},
		{"", "", fmt.Errorf("failed to get base")},
	}

	for _, v := range tests {
		v.basePath = filepath.Join(v.basePath, "exe.exe")
		executable = func() (string, error) { return v.basePath, v.err }
		if v.err != nil {
			require.PanicsWithError(t, v.err.Error(), func() { GetAbsolutePath(v.relPath) })
			continue
		}

		ap := GetAbsolutePath(v.relPath)
		expected := filepath.Join(filepath.Dir(v.basePath), v.relPath)
		require.Equal(t, ap, expected)
	}
}

func TestDeepCopyMap(t *testing.T) {
	s1 := struct{ s string }{"hello"}
	s2 := struct{ s any }{"world"}
	var (
		nilSl  []int          = nil
		nilPtr *int           = nil
		nilMap map[string]any = nil
		nilIfc any
	)
	m1 := map[string]any{
		"a": "bbb",
		"b": map[string]any{"c": 123, "s": nil},
		"d": []int{1, 2, 3},
		"e": nil,
		"f": &s1,
		"g": nilSl,
		"h": nilMap,
		"i": nilPtr,
		"j": nilIfc,
	}

	m3 := DeepCopyMap(nil)
	m2 := DeepCopyMap(m1)
	m1["a"] = "zzz"
	m1["d"].([]int)[0] = 100
	delete(m1, "b")
	m1["e"] = "hello"
	m1["f"] = &s2

	require.Nil(t, m3)
	require.Equal(t, map[string]any{"a": "zzz", "d": []int{100, 2, 3}, "e": "hello", "f": &s2, "g": nilSl, "h": nilMap, "i": nilPtr, "j": nilIfc}, m1)
	require.Equal(t, map[string]any{
		"a": "bbb",
		"b": map[string]any{"c": 123, "s": nil},
		"d": []int{1, 2, 3},
		"e": nil,
		"f": &s1,
		"g": nilSl,
		"h": nilMap,
		"i": nilPtr,
		"j": nilIfc,
	}, m2)
}

func TestMergeMap(t *testing.T) {
	tests := []struct {
		dst      map[string]any
		src      map[string]any
		expected map[string]any
	}{
		{nil, map[string]any{"a": nil}, nil},
		{map[string]any{"b": "s"}, nil, map[string]any{"b": "s"}},
		{map[string]any{"a": "s"}, map[string]any{"a": nil}, map[string]any{"a": nil}},
		{map[string]any{"a": "s"}, map[string]any{"b": nil}, map[string]any{"a": "s", "b": nil}},
		{
			map[string]any{
				"nil1": nil,
				"int":  12,
				"e": map[string]any{
					"nil2":     nil,
					"ovwrtSt":  "st",
					"ovwrtInt": 12,
				},
			},
			map[string]any{"e": map[string]any{
				"nil2":     "not nil",
				"ovwrtSt":  12,
				"ovwrtInt": "st",
				"new":      "new",
			}},
			map[string]any{
				"nil1": nil,
				"int":  12,
				"e": map[string]any{
					"nil2":     "not nil",
					"ovwrtSt":  12,
					"ovwrtInt": "st",
					"new":      "new",
				},
			},
		},
	}

	for _, v := range tests {
		MergeMap(v.dst, v.src)
		require.Equal(t, v.dst, v.expected)
	}
}

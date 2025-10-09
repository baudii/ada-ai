package utils

import (
	"fmt"
	"io"
	"path/filepath"
	"reflect"
	"strings"
)

// Returns absolute that is calculated from the current
// executable path. Panics if os.Executable() returns an error.
func AbsolutePath(relativePath string, executable func() (string, error)) string {
	exe, err := executable()
	if err != nil {
		panic(err)
	}

	return filepath.Join(filepath.Dir(exe), relativePath)
}

func InsertFsuffix(path string, postfix string) string {
	ext := filepath.Ext(path)
	name := strings.TrimSuffix(path, ext)
	return name + postfix + ext
}

// Merges dst map with the values from src. If a value exists
// in src, then it will be copied to dst. If dst already contains
// this kev/value pair, it will be overwritten
//
// Note: This function doesn't copy the values, it just assings them
// to dst. If the value is reference type, then changes will affect
// both src and dst.
func MergeMap(dst, src map[string]any) {
	if dst == nil {
		return
	}

	for k, v := range src {
		if vMap, ok := v.(map[string]any); ok {
			if dMap, ok := dst[k].(map[string]any); ok {
				MergeMap(dMap, vMap)
				continue
			}
		}
		dst[k] = v
	}
}

// Creates a new map and recursively fills it with copy
// of every value from src
func DeepCopyMap(src map[string]any) map[string]any {
	if src == nil {
		return nil
	}

	dst := make(map[string]any, len(src))
	for k, v := range src {
		if v != nil {
			dst[k] = DeepCopy(v)
		} else {
			dst[k] = nil
		}
	}
	return dst
}

// Creates a copy of given value using reflect and taking
// into account its value type.
func DeepCopy(v any) any {
	return deepCopyR(reflect.ValueOf(v)).Interface()
}

func deepCopyR(v reflect.Value) reflect.Value {
	switch v.Kind() {
	case reflect.Map:
		if v.IsNil() {
			return reflect.Zero(v.Type())
		}
		out := reflect.MakeMapWithSize(v.Type(), v.Len())
		for _, key := range v.MapKeys() {
			out.SetMapIndex(key, deepCopyR(v.MapIndex(key)))
		}
		return out

	case reflect.Slice:
		if v.IsNil() {
			return reflect.Zero(v.Type())
		}
		out := reflect.MakeSlice(v.Type(), v.Len(), v.Cap())
		for i := 0; i < v.Len(); i++ {
			out.Index(i).Set(deepCopyR(v.Index(i)))
		}
		return out

	case reflect.Pointer:
		if v.IsNil() {
			return reflect.Zero(v.Type())
		}
		out := reflect.New(v.Type().Elem())
		out.Elem().Set(deepCopyR(v.Elem()))
		return out

	case reflect.Interface:
		if v.IsNil() {
			return reflect.Zero(v.Type())
		}
		return deepCopyR(v.Elem())

	default:
		return v
	}
}

// PrintTree writes a textual tree representation of the given structure to w.
// NOTE: since map is unordered by nature, there is not guarantee of an order in
// this function either.
//
// The structure is expected to be a nested map[string]any, where each key is
// treated as a directory or file name, and nested maps represent subdirectories.
//
// The output uses Unicode box-drawing characters (├──, └──, │) to visualize
// the hierarchy, similar to the Unix `tree` command.
//
// The prefix argument is used internally to align child elements and should
// normally be an empty string when called from outside.
func PrintTree(w io.Writer, structure map[string]any, prefix string) {
	i := 0
	for k, v := range structure {
		i++
		conn := "├── "
		nextPrefix := prefix + "│   "
		if i == len(structure) {
			conn = "└── "
			nextPrefix = prefix + "    "
		}
		fmt.Fprintf(w, "%s%s%s\n", prefix, conn, k)
		if m, ok := v.(map[string]any); ok {
			PrintTree(w, m, nextPrefix)
		}
	}
}

package utils

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

var executable = os.Executable

// Returns absolute that is calculated from the current
// executable path. Panics if os.Executable() returns an error.
func GetAbsolutePath(relativePath string) string {
	exe, err := executable()
	if err != nil {
		// TODO: Replace panic with error handling and fix comment
		panic(err)
	}

	base := filepath.Dir(exe)
	return filepath.Join(base, relativePath)
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

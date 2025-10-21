package clone

import "reflect"

// DeepCopyMap creates a new map and recursively fills it with copy
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

// DeepCopy creates a copy of given value using reflect and taking
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

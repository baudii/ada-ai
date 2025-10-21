package maps

// Merge merges dst map with the values from src. If a value exists
// in src, then it will be copied to dst. If dst already contains
// this kev/value pair, it will be overwritten
//
// Note: This function doesn't copy the values, it just assings them
// to dst. If the value is reference type, then changes will affect
// both src and dst.
func Merge(dst, src map[string]any) {
	if dst == nil {
		return
	}

	for k, v := range src {
		if vMap, ok := v.(map[string]any); ok {
			if dMap, ok := dst[k].(map[string]any); ok {
				Merge(dMap, vMap)
				continue
			}
		}
		dst[k] = v
	}
}

package utils

import (
	"os"
	"path/filepath"
)

func GetAbsolutePath(relativePath string) string {
	exe, err := os.Executable()
	if err != nil {
		panic(err)
	}

	base := filepath.Dir(exe)
	return filepath.Join(base, relativePath)
}

func PathExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func DeepMerge(dst, src map[string]any) {
	for k, v := range src {
		if vMap, ok := v.(map[string]any); ok {
			if dMap, ok := dst[k].(map[string]any); ok {
				DeepMerge(dMap, vMap)
				continue
			}
		}
		dst[k] = v
	}
}

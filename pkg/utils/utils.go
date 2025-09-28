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

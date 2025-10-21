package pathx

import (
	"path/filepath"
	"strings"
)

// FromExecutable returns an absolute path resolved relative to the directory
// reported by the provided executable locator. It panics if the callback
// returns an error.
func FromExecutable(relativePath string, executable func() (string, error)) string {
	exe, err := executable()
	if err != nil {
		panic(err)
	}

	return filepath.Join(filepath.Dir(exe), relativePath)
}

// InsertFsuffix inserts a postfix before the file extension
// in the given path. If there is no extension, postfix
// will be appended to the path.
func InsertFsuffix(path string, suffix string) string {
	ext := filepath.Ext(path)
	name := strings.TrimSuffix(path, ext)
	return name + suffix + ext
}

package ada

import (
	"os"
	"path/filepath"
)

var projDescrFile string
var projRoot string

const Artifacts string = "artifacts"

func createDirectory() error {
	path := filepath.Join(Artifacts, cfg.ProjRoot, cfg.UserName, cfg.ProjName)
	_, err := os.Stat(path)
	if err == nil {
		projRoot = path
		return &ProjExist{Path: path}
	}

	if !os.IsNotExist(err) {
		return err
	}

	err = os.MkdirAll(path, 0o755)
	if err != nil {
		return err
	}

	projRoot = path
	return nil
}

func saveProjectStructure(content []byte) (err error) {
	if err := createDirectory(); err != nil {
		return err
	}
	projDescrFile = filepath.Join(projRoot, "project-structure.json")
	return os.WriteFile(projDescrFile, content, 0644)
}

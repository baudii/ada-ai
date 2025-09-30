package adacore

import (
	"errors"
	"log/slog"
	"os"
	"path/filepath"
)

const Artifacts string = "artifacts"

func (ada *Ada) EnsureSaved(data []byte) error {
	var err error
	if err = ada.saveProjectStructure(data); err != nil {
		var nfErr *ProjExist
		if !errors.As(err, &nfErr) {
			return err
		}
	}

	return nil
}

func (ada *Ada) Materialize(data []byte) error {
	if err := ada.EnsureSaved(data); err != nil {
		slog.Error("failed to save data", "error", err)
	}

	node, err := parseNode(data)
	if err != nil {
		return err
	}

	node.materialize(ada.Ctx.projRoot)
	return nil
}

func (ada *Ada) saveProjectStructure(content []byte) (err error) {
	if err := ada.createDirectory(); err != nil {
		return err
	}
	ada.Ctx.projDescFile = filepath.Join(ada.Ctx.projRoot, "project-structure.json")
	return os.WriteFile(ada.Ctx.projDescFile, content, 0644)
}

func (ada *Ada) createDirectory() error {
	path := filepath.Join(Artifacts, ada.cfg.ProjRoot, ada.Ctx.UserName, ada.Ctx.ProjName)
	_, err := os.Stat(path)
	if err == nil {
		ada.Ctx.projRoot = path
		return &ProjExist{Path: path}
	}

	if !os.IsNotExist(err) {
		return err
	}

	err = os.MkdirAll(path, 0o755)
	if err != nil {
		return err
	}

	ada.Ctx.projRoot = path
	return nil
}

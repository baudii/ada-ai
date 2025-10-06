package projects

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/baudii/ada-ai/internal/adacore"
	"github.com/baudii/ada-ai/pkg/utils"
)

type LocalProj struct {
	structure   map[string]any
	projectRoot string
}

// NewLocalProj creates a new project descriptor from a JSON structure and base
// that determines the root base folder of where the project will be created
//
// It returns an error if the JSON cannot be unmarshaled or the base path
// is invalid.
func NewLocalProj(tr []byte, base string) (*LocalProj, error) {
	d := &LocalProj{}
	if err := json.Unmarshal([]byte(tr), &d.structure); err != nil {
		return nil, fmt.Errorf("unmarshal project structure: %w", err)
	}
	if !filepath.IsAbs(base) {
		return nil, fmt.Errorf("path must be absolute")
	}

	d.projectRoot = base
	return d, nil
}

// Materialize creates the directory and file structure defined in the descriptor,
// and also persists a JSON representation of this structure to the project root.
//
// The structure is created under the project root directory, which must already
// exist and be specified as an absolute path.
//
// If any part of the structure cannot be created or the JSON description cannot
// be saved, an error is returned.
func (d *LocalProj) Materialize() error {
	if info, err := os.Stat(d.projectRoot); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("stat project root: %w", err)
		}
	} else if !info.IsDir() {
		return fmt.Errorf("project root %q is not a directory", d.projectRoot)
	} else {
		err := os.RemoveAll(d.projectRoot)
		if err != nil {
			return err
		}
	}

	if err := os.MkdirAll(d.projectRoot, 0644); err != nil {
		return fmt.Errorf("create project root folder %q: %w", d.projectRoot, err)
	}

	if err := utils.SaveJSONToFile(d.structure, filepath.Join(d.projectRoot, adacore.ProjectStrucutreFile)); err != nil {
		return fmt.Errorf("save project structure file: %w", err)
	}

	return materialize(d.projectRoot, d.structure)
}

func (d *LocalProj) Structure() map[string]any {
	return d.structure
}

func materialize(base string, structure map[string]any) error {
	for name, v := range structure {
		path := filepath.Join(base, name)
		if m, ok := v.(map[string]any); ok {
			if err := os.MkdirAll(path, 0o755); err != nil {
				return fmt.Errorf("create folder(s): %w", err)
			}
			if err := materialize(path, m); err != nil {
				return err
			}
		} else if _, ok := v.(float64); ok {
			f, err := os.Create(path)
			if err != nil {
				return fmt.Errorf("create file: %w", err)
			}
			f.Close()
		} else {
			return fmt.Errorf("invalid structure: value must be either a map or float64, got %T", v)
		}
	}
	return nil
}

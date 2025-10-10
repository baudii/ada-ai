package projects

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/baudii/ada-ai/pkg/utils"
)

type projectPathResolver interface {
	ResolveProjectPath() string
}

type localProj struct {
	structure    map[string]any
	base         string
	projectRoot  string
	uniqFoldName func(string) (string, error)
}

// New creates a new local project descriptor from structure that represents
// the json folder structure of the project and a resolver that resolves the
// base that determines the root folder of where the project will be created.
// The base path must be absolute and will be created if it doesn't exist.
//
// It returns an error if the JSON cannot be unmarshaled or the base path
// is invalid.
func New(structure []byte, resolver projectPathResolver) (*localProj, error) {
	d := &localProj{}
	if err := json.Unmarshal([]byte(structure), &d.structure); err != nil {
		return nil, fmt.Errorf("unmarshal project structure: %w", err)
	}

	base := resolver.ResolveProjectPath()
	if !filepath.IsAbs(base) {
		return nil, fmt.Errorf("path %q is not absolute", base)
	}

	if err := os.MkdirAll(base, 0744); err != nil {
		return nil, fmt.Errorf("create base path: %w", err)
	}

	d.base = base
	d.uniqFoldName = uniqueIndexFolder
	return d, nil
}

// Structure returns a structure map of this LocalProj.
func (d *localProj) Structure() map[string]any {
	return d.structure
}

// Materialize creates the directory and file structure defined in the descriptor,
// and also persists a JSON representation of this structure to the project root.
//
// The structure is created under the project root directory, which must already
// exist and be specified as an absolute path.
//
// If any part of the structure cannot be created or the JSON description cannot
// be saved, an error is returned.
func (d *localProj) Materialize() error {
	uif, err := d.uniqFoldName(d.base)
	if err != nil {
		return fmt.Errorf("get unique folder name: %w", err)
	}

	d.projectRoot = filepath.Join(d.base, uif)
	if err := os.MkdirAll(d.projectRoot, 0744); err != nil {
		return fmt.Errorf("create project root folder %q: %w", d.projectRoot, err)
	}

	if err := utils.SaveJSONToFile(d.structure, filepath.Join(d.projectRoot, "project-structure.json")); err != nil {
		return fmt.Errorf("save project structure file: %w", err)
	}

	return materialize(d.projectRoot, d.structure)
}

func materialize(base string, structure map[string]any) error {
	for name, v := range structure {
		path := filepath.Join(base, name)
		switch r := v.(type) {
		case map[string]any:
			if err := os.MkdirAll(path, 0o755); err != nil {
				return fmt.Errorf("create folder(s): %w", err)
			}
			if err := materialize(path, r); err != nil {
				return err
			}
		case float64:
			f, err := os.Create(path)
			if err != nil {
				return fmt.Errorf("create file: %w", err)
			}
			_ = f.Close()
		default:
			return fmt.Errorf("invalid structure: value must be either a map or float64, got %T", v)
		}
	}
	return nil
}

func uniqueIndexFolder(base string) (string, error) {
	entries, err := os.ReadDir(base)
	if err != nil {
		return "", err
	}
	max := -1
	for _, v := range entries {
		id, err := strconv.Atoi(v.Name())
		if err == nil && id > max {
			max = id
		}
	}

	return strconv.Itoa(max + 1), nil
}

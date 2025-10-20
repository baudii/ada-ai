package local

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/baudii/ada-ai/internal/project"
)

type local struct {
	root  string
	store navStore
}

type navStore interface {
	Materialize() error
	Tree() map[string]any
	AddNav(key, ext string, content []byte) error
}

// New creates a new LocalProj instance with the given project root path.
// It initializes the navigation path and prepares the navs map.
// An error is returned if the provided project root path is not absolute
// or if the base navigation path cannot be created.
func New(root string, store navStore) (*local, error) {
	if !filepath.IsAbs(root) {
		return nil, fmt.Errorf("path %q is not absolute", root)
	}

	if err := os.MkdirAll(root, 0744); err != nil {
		return nil, fmt.Errorf("create nav path: %w", err)
	}

	return &local{
		root:  root,
		store: store,
	}, nil
}

// Materialize creates the directory and file structure defined in the descriptor,
// and also persists a JSON representation of this structure to the project root.
//
// The structure is created under the project root directory, which must already
// exist and be specified as an absolute path.
//
// If any part of the structure cannot be created or the JSON description cannot
// be saved, an error is returned.
func (d *local) Materialize(hfile, hfold project.Handler) error {
	if err := d.store.Materialize(); err != nil {
		return fmt.Errorf("setup nav: %w", err)
	}

	return project.Traverse(d.root, d.store.Tree(), hfile, hfold)
}

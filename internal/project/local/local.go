package local

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/baudii/ada-ai/internal/project"
	"github.com/baudii/ada-ai/internal/project/nav"
)

var NavFolder = "nav"

type proj struct {
	tree    map[string]any
	root    string
	navRoot string
	store   navStore
}

type navStore interface {
	Materialize(root string) error
	Add(key, path string, content []byte)
	Get(k string) (*nav.File, bool)
}

// New creates a new LocalProj instance with the given project root path.
// It initializes the navigation path and prepares the navs map.
// An error is returned if the provided project root path is not absolute
// or if the base navigation path cannot be created.
func New(root string, store navStore) (*proj, error) {
	if !filepath.IsAbs(root) {
		return nil, fmt.Errorf("path %q is not absolute", root)
	}

	if err := os.MkdirAll(root, 0744); err != nil {
		return nil, fmt.Errorf("create nav path: %w", err)
	}

	return &proj{
		root:    root,
		navRoot: filepath.Join(root, NavFolder),
		store:   store,
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
func (l *proj) Materialize(hfile, hfold project.Handler) error {
	if err := l.store.Materialize(l.navRoot); err != nil {
		return fmt.Errorf("setup nav: %w", err)
	}

	return project.Traverse(l.root, l.tree, hfile, hfold)
}

// AddItem adds a navigation file to the local project descriptor. The filename
// is the name of the file to be created under the "nav" folder, and content is
// the byte content to be written to that file.
func (l *proj) AddNav(key, ext string, content []byte) error {
	if key == project.Structure {
		err := json.Unmarshal(content, &l.tree)
		if err != nil {
			return fmt.Errorf("parse structure content: %w", err)
		}
	}
	l.store.Add(key, fmt.Sprintf("%v.%v", key, ext), content)
	return nil
}

func (l *proj) Tree() map[string]any {
	return l.tree
}

// LoadNav retrieves a navigation file by its filename. It returns the file content
// as a byte slice and a boolean indicating whether the file was found.
func (l *proj) LoadNav(filename string) ([]byte, error) {
	path := filepath.Join(l.navRoot, filename)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}
	return data, nil
}

// Content retrieves the content of a navigation file by its key.
// It returns the content as a compacted JSON string.
func (l *proj) Content(key string) (string, error) {
	item, ok := l.store.Get(key)
	if !ok {
		return "", fmt.Errorf("nav item %q not found", key)
	}
	buf := &bytes.Buffer{}
	if err := json.Compact(buf, item.Content); err != nil {
		return "", fmt.Errorf("compact content: %w", err)
	}
	return buf.String(), nil
}

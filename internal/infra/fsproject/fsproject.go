package fsproject

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/baudii/ada-ai/internal/core/nav"
	"github.com/baudii/ada-ai/internal/core/project"
)

var NavFolder = "nav"

type Data struct {
	Tree    map[string]any
	root    string
	navRoot string
	store   navStore
}

type navStore interface {
	Add(key, path string, content []byte)
	Get(k string) (*nav.FileInfo, bool)
	GetItems() map[string]nav.FileInfo
}

// Default handlers for file and folder creation during materialization.
var (
	DefaultFileHandler = func(path string) error {
		f, err := os.Create(path)
		if err != nil {
			return fmt.Errorf("create file: %w", err)
		}
		defer func() { _ = f.Close() }()
		return nil
	}

	DefaultFolderHandler = func(path string) error {
		err := os.MkdirAll(path, 0755)
		if err != nil {
			return fmt.Errorf("create folder: %w", err)
		}
		return nil
	}
)

// New creates a new LocalProj instance with the given project root path.
// It initializes the navigation path and prepares the navs map.
// An error is returned if the provided project root path is not absolute
// or if the base navigation path cannot be created.
func New(root string, store navStore) (*Data, error) {
	if !filepath.IsAbs(root) {
		return nil, fmt.Errorf("path %q is not absolute", root)
	}

	if err := os.MkdirAll(root, 0744); err != nil {
		return nil, fmt.Errorf("create nav path: %w", err)
	}

	return &Data{
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
func (l *Data) Materialize(hfile project.FileHandler, hfold project.FolderHandler) error {
	if err := os.MkdirAll(l.navRoot, 0755); err != nil {
		return fmt.Errorf("create nav root: %w", err)
	}

	for _, item := range l.store.GetItems() {
		if err := os.WriteFile(filepath.Join(l.navRoot, item.Filename), item.Content, 0644); err != nil {
			return fmt.Errorf("create nav file %q: %w", item.Filename, err)
		}
	}

	if hfile == nil {
		hfile = DefaultFileHandler
	}
	if hfold == nil {
		hfold = DefaultFolderHandler
	}

	return project.Traverse(l.root, l.Tree, hfile, hfold)
}

// AddItem adds a navigation file to the local project descriptor. The filename
// is the name of the file to be created under the "nav" folder, and content is
// the byte content to be written to that file.
func (l *Data) AddNav(key, ext string, content []byte) error {
	if key == project.Structure {
		err := json.Unmarshal(content, &l.Tree)
		if err != nil {
			return fmt.Errorf("parse structure content: %w", err)
		}
	}
	l.store.Add(key, fmt.Sprintf("%v.%v", key, ext), content)
	return nil
}

func (l *Data) CreateFile(path string, content []byte) error {
	return os.WriteFile(path, content, 0644)
}

// LoadNav retrieves a navigation file by its filename. It returns the file content
// as a byte slice and a boolean indicating whether the file was found.
func (l *Data) LoadNav(filename string) ([]byte, error) {
	path := filepath.Join(l.navRoot, filename)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}
	return data, nil
}

// Content retrieves the content of a navigation file by its key.
// It returns the content as a compacted JSON string.
func (l *Data) Content(key string) (string, error) {
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

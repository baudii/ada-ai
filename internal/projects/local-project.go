package projects

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/baudii/ada-ai/pkg/utils"
)

const Structure = "structure"

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
		return os.MkdirAll(path, 0755)
	}
)

type handler func(string) error

type nav struct {
	filepath string
	content  []byte
}

type localProj struct {
	projectRoot string
	navPath     string
	structure   map[string]any
	navs        map[string]nav
	lastFolder  func(string) (int, error)
}

// New creates a new LocalProj based on the provided structure and project root path.
//
// The structure is provided as a JSON byte slice and is unmarshaled into a map.
func New(projectRoot string) (*localProj, error) {
	d := &localProj{}

	if !filepath.IsAbs(projectRoot) {
		return nil, fmt.Errorf("path %q is not absolute", projectRoot)
	}

	if err := os.MkdirAll(projectRoot, 0744); err != nil {
		return nil, fmt.Errorf("create base path: %w", err)
	}

	d.projectRoot = projectRoot
	d.navPath = navPath(projectRoot)
	d.lastFolder = LastFolder
	d.navs = make(map[string]nav)
	return d, nil
}

// AddNav adds a navigation file to the local project descriptor. The filename
// is the name of the file to be created under the "nav" folder, and content is
// the byte content to be written to that file.
func (d *localProj) AddNav(key, ext string, content []byte) error {
	path := filepath.Join(d.navPath, fmt.Sprintf("%v.%v", key, ext))
	if key == Structure {
		err := json.Unmarshal(content, &d.structure)
		if err != nil {
			return fmt.Errorf("parse structure content: %w", err)
		}
	}

	d.navs[key] = nav{filepath: path, content: content}
	return nil
}

// NavContent retrieves the content of a navigation file by its key.
// It returns the content as a byte slice and a boolean indicating whether
// the file was found.
func (d *localProj) NavContent(key string) ([]byte, error) {
	nav, ok := d.navs[key]
	if !ok {
		return nil, fmt.Errorf("nav key %q not found", key)
	}
	return nav.content, nil
}

// Structure returns a structure map of this LocalProj.
func (d *localProj) Structure() map[string]any {
	// TODO: remove ?
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
func (d *localProj) Materialize(hfile, hfold handler) error {
	if err := os.MkdirAll(d.projectRoot, 0744); err != nil {
		return fmt.Errorf("create project root folder %q: %w", d.projectRoot, err)
	}

	if err := d.materializeNav(); err != nil {
		return fmt.Errorf("setup nav: %w", err)
	}

	return TraverseStructure(d.projectRoot, d.structure, hfile, hfold)
}

// TraverseStructure traverses a structure map and calls the provided handlers
// for folders and files respectively.
//
// The structure map should have string keys representing names, and values
// that are either nested maps (for folders) or float64 (for files).
// If a value is neither a map nor a float64, an error is returned.
//
// The handleFold handler is called for each folder path, and the handleFile handler
// is called for each file path. If any handler returns an error, the traversal
// stops and the error is returned.
func TraverseStructure(base string, structure map[string]any, hfile, hfold handler) error {
	for name, v := range structure {
		path := filepath.Join(base, name)
		switch r := v.(type) {
		case map[string]any:
			if err := hfold(path); err != nil { //os.MkdirAll(path, 0755)
				return fmt.Errorf("handle folder(s): %w", err)
			}
			if err := TraverseStructure(path, r, hfile, hfold); err != nil {
				return err
			}
		case float64:

			if err := hfile(path); err != nil {
				return fmt.Errorf("handle file: %w", err)
			}
		default:
			return fmt.Errorf("invalid structure: value must be either a map or float64, got %T", v)
		}
	}
	return nil
}

// LastFolder returns the next available integer index for folder naming
// within the specified base directory.
//
// It reads the contents of the base directory, identifies existing folders
// and returns the next available index. If the directory cannot be read,
// an error is returned.
func LastFolder(base string) (int, error) {
	entries, err := os.ReadDir(base)
	if err != nil {
		return -2, err
	}
	max := -1
	for _, v := range entries {
		id, err := strconv.Atoi(v.Name())
		if err == nil && id > max {
			max = id
		}
	}

	return max, nil
}

// TryLoadNav loads the content of a navigation file from the nav folder.
// It returns the content as a byte slice and a boolean indicating whether
// the file was found and read successfully.
func (d *localProj) TryLoadNav(filename string, res *[]byte) bool {
	path := filepath.Join(d.navPath, filename)
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	*res = data
	return true
}

func (d *localProj) materializeNav() error {
	if err := os.MkdirAll(d.navPath, 0755); err != nil {
		return fmt.Errorf("create nav folder: %w", err)
	}

	if err := utils.SaveJSONToFile(d.structure, filepath.Join(d.navPath, "structure.json")); err != nil {
		return fmt.Errorf("save project structure file: %w", err)
	}

	for _, nav := range d.navs {
		if err := os.WriteFile(nav.filepath, nav.content, 0644); err != nil {
			return fmt.Errorf("create nav file %q: %w", nav.filepath, err)
		}
	}

	return nil
}

func navPath(projRoot string) string {
	return filepath.Join(projRoot, "nav")
}

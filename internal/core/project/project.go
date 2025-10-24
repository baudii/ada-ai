package project

import (
	"fmt"
	"path/filepath"
)

// Structure is the navigation key for the project structure descriptor.
const Structure = "structure"

// FileHandler and FolderHandler are function types used as callbacks
// during the traversal of a project structure.
type (
	FileHandler   func(string) error
	FolderHandler func(string) error
)

// Data is the context of the current project.
type Data struct {
	UserName string `json:"userName"`
	ProjName string `json:"projName"`
	Language string `json:"language"`
	Summary  string `json:"summary"`
}

// Traverse traverses a structure map and calls the provided handlers
// for folders and files respectively.
//
// The structure map should have string keys representing names, and values
// that are either nested maps (for folders) or float64 (for files).
// If a value is neither a map nor a float64, an error is returned.
//
// The handleFold handler is called for each folder path, and the handleFile handler
// is called for each file path. If any handler returns an error, the traversal
// stops and the error is returned.
func Traverse(base string, structure map[string]any, hfile FileHandler, hfold FolderHandler) error {
	for name, v := range structure {
		path := filepath.Join(base, name)
		switch r := v.(type) {
		case map[string]any:
			if err := hfold(path); err != nil {
				return fmt.Errorf("handle folder(s): %w", err)
			}
			if err := Traverse(path, r, hfile, hfold); err != nil {
				return err
			}
		case float64:
			if err := hfile(path); err != nil {
				return fmt.Errorf("handle file: %w", err)
			}
		default:
			return fmt.Errorf("invalid tree: value must be either a map or float64, got %T", v)
		}
	}
	return nil
}

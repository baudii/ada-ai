package project

import (
	"fmt"
	"os"
	"path/filepath"
)

// Handler defines a function type for handling file or folder creation
type Handler func(string) error

const Structure = "structure"

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
func Traverse(base string, structure map[string]any, hfile, hfold Handler) error {
	for name, v := range structure {
		path := filepath.Join(base, name)
		switch r := v.(type) {
		case map[string]any:
			if err := hfold(path); err != nil { //os.MkdirAll(path, 0755)
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

package projects

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/baudii/ada-ai/pkg/utils"
)

type NavHandler interface {
	TryLoadNav(key string, dest *[]byte) bool
	AddNav(key, ext string, content []byte) error
	NavContent(key string) ([]byte, error)
}

type nav struct {
	filepath string
	content  []byte
}

// TryLoadNav loads the content of a navigation file from the nav folder.
// It returns the content as a byte slice and a boolean indicating whether
// the file was found and read successfully.
func (d *localProj) TryLoadNav(filename string, dest *[]byte) bool {
	path := filepath.Join(d.navPath, filename)
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	*dest = data
	return true
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

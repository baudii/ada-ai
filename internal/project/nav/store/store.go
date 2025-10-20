package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/baudii/ada-ai/internal/project"
	"github.com/baudii/ada-ai/internal/project/nav"
	"github.com/baudii/ada-ai/pkg/utils"
)

var navFolder = "nav"

type store struct {
	root  string
	tree  map[string]any
	items map[string]navItem
}

type navItem interface {
	Materialize() error
}

func New(root string) (*store, error) {
	n := &store{
		root:  filepath.Join(root, navFolder),
		items: make(map[string]navItem),
	}

	if err := os.MkdirAll(n.root, 0744); err != nil {
		return nil, fmt.Errorf("create nav root: %w", err)
	}

	return n, nil
}

func (n *store) Tree() map[string]any {
	return n.tree
}

// AddNav adds a navigation file to the local project descriptor. The filename
// is the name of the file to be created under the "nav" folder, and content is
// the byte content to be written to that file.
func (n *store) AddNav(key, ext string, content []byte) error {
	path := filepath.Join(n.root, fmt.Sprintf("%v.%v", key, ext))
	if key == project.Structure {
		err := json.Unmarshal(content, &n.tree)
		if err != nil {
			return fmt.Errorf("parse structure content: %w", err)
		}
	}

	n.items[key] = nav.New(path, content)
	return nil
}

// TryLoadNav loads the content of a navigation file from the nav folder.
// It returns the content as a byte slice and a boolean indicating whether
// the file was found and read successfully.
func (n *store) TryLoadNav(filename string, dest *[]byte) bool {
	path := filepath.Join(n.root, filename)
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	*dest = data
	return true
}

func (n *store) Materialize() error {
	if err := os.MkdirAll(n.root, 0755); err != nil {
		return fmt.Errorf("create nav folder: %w", err)
	}

	if err := utils.SaveJSONToFile(n.tree, filepath.Join(n.root, "structure.json")); err != nil {
		return fmt.Errorf("save project structure file: %w", err)
	}

	for _, v := range n.items {
		if err := v.Materialize(); err != nil {
			return fmt.Errorf("create nav file: %w", err)
		}
	}

	return nil
}

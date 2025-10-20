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
	CompactContent() (string, error)
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

func (s *store) Tree() map[string]any {
	return s.tree
}

// AddNav adds a navigation file to the local project descriptor. The filename
// is the name of the file to be created under the "nav" folder, and content is
// the byte content to be written to that file.
func (s *store) AddNav(key, ext string, content []byte) error {
	path := filepath.Join(s.root, fmt.Sprintf("%v.%v", key, ext))
	if key == project.Structure {
		err := json.Unmarshal(content, &s.tree)
		if err != nil {
			return fmt.Errorf("parse structure content: %w", err)
		}
	}

	s.items[key] = nav.New(path, content)
	return nil
}

// LoadNav retrieves a navigation file by its filename. It returns the file content
// as a byte slice and a boolean indicating whether the file was found.
func (s *store) LoadNav(filename string) ([]byte, error) {
	path := filepath.Join(s.root, filename)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// NavContent retrieves the content of a navigation file by its key.
// It returns the content as a compacted JSON string.
func (s *store) NavContent(key string) (string, error) {
	item, ok := s.items[key]
	if !ok {
		return "", fmt.Errorf("nav item %q not found", key)
	}
	return item.CompactContent()
}

func (s *store) Materialize() error {
	if err := os.MkdirAll(s.root, 0755); err != nil {
		return fmt.Errorf("create nav folder: %w", err)
	}

	if err := utils.SaveJSONToFile(s.tree, filepath.Join(s.root, "structure.json")); err != nil {
		return fmt.Errorf("save project structure file: %w", err)
	}

	for _, v := range s.items {
		if err := v.Materialize(); err != nil {
			return fmt.Errorf("create nav file: %w", err)
		}
	}

	return nil
}

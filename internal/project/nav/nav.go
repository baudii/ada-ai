package nav

import (
	"fmt"
	"os"
	"path/filepath"
)

type store struct {
	items map[string]*File
}

type File struct {
	Filename string
	Content  []byte
}

func New() *store {
	return &store{
		items: make(map[string]*File),
	}
}

func (s *store) Add(k, p string, c []byte) {
	s.items[k] = &File{
		Filename: p,
		Content:  c,
	}
}

func (s *store) Get(k string) (*File, bool) {
	f, ok := s.items[k]
	return f, ok
}

// Materialize writes the navigation files and structure to the filesystem.
func (s *store) Materialize(root string) error {
	if err := os.MkdirAll(root, 0755); err != nil {
		return fmt.Errorf("create nav folder: %w", err)
	}

	for _, v := range s.items {
		if err := os.WriteFile(filepath.Join(root, v.Filename), v.Content, 0644); err != nil {
			return fmt.Errorf("create nav file: %w", err)
		}
	}

	return nil
}

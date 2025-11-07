package templateproject

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/baudii/ada-ai/internal/core/project"
)

// Data represents a template project used for initializing new projects.
type Data struct {
	Tree         map[string]any
	root         string
	templatePath string
}

func New(root, templatePath string) (*Data, error) {
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("template path %q does not exist: %w", templatePath, err)
	}

	return &Data{
		root:         root,
		templatePath: templatePath,
	}, nil
}

// Materialize creates the project structure on the filesystem based of the template
// project. It uses the provided file and folder handlers to handle existing files and
// folders of the template.
func (l *Data) Materialize(hfile project.FileHandler, hfold project.FolderHandler) error {
	if err := l.CopyTemplate(); err != nil {
		return fmt.Errorf("copy template: %w", err)
	}

	if err := l.addTree(); err != nil {
		return fmt.Errorf("add tree: %w", err)
	}

	return project.Traverse(l.root, l.Tree, hfile, hfold)
}

// HandleFile writes the provided content to the specified file path.
// It returns an error if the file does not exist or if the write operation fails.
func (l *Data) HandleFile(name string, content []byte) error {
	_, err := os.Stat(name)
	if err != nil {
		return fmt.Errorf("stat file: %w", err)
	}
	if err := os.WriteFile(name, content, 0644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}
	return nil
}

// CopyTemplate copies the template project files to the project root directory.
func (l *Data) CopyTemplate() error {
	dst := l.root
	if err := os.MkdirAll(dst, 0755); err != nil {
		return fmt.Errorf("create template dir: %w", err)
	}

	cmd := exec.Command("cp", "-r", l.templatePath, dst)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("copy template: %w", err)
	}

	return nil
}

func (l *Data) addTree() error {
	tree := make(map[string]any)
	fold := tree
	err := filepath.Walk(l.root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			fold[info.Name()] = make(map[string]any)
			fold = fold[info.Name()].(map[string]any)
		} else {
			fold[info.Name()] = float64(0)
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("walk template project: %w", err)
	}

	l.Tree = tree
	return nil
}

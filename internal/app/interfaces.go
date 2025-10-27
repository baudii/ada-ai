package app

import (
	"context"

	"github.com/baudii/ada-ai/internal/core/gen"
	"github.com/baudii/ada-ai/internal/core/project"
)

// generator defines methods for generating AI content and handling prompts.
type generator interface {
	BuildPrompt(filename string, input ...any) (string, error)
	GenerateWithSys(ctx context.Context, sys, user string, opts ...gen.Option) (string, error)
}

// materializer defines the interface for managing project structures
// including navigation files and materialization of the project layout.
type materializer interface {
	Materialize(hfile project.FileHandler, hfold project.FolderHandler) error
	CreateFile(name string, content []byte) error
}

// navigator defines methods for managing navigation files within a project.
type navigator interface {
	LoadNav(key string) ([]byte, error)
	AddNav(key, ext string, content []byte) error
	Content(key string) (string, error)
}

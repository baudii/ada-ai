package app

import (
	"context"

	"github.com/baudii/ada-ai/internal/project"
	"github.com/baudii/ada-ai/internal/project/seqdir"
	"github.com/tmc/langchaingo/llms"
)

// generator defines methods for generating AI content and handling prompts.
type generator interface {
	BuildPrompt(filename string, input ...any) (string, error)
	GenerateWithSys(ctx context.Context, sys, user string, callOptions ...llms.CallOption) (*llms.ContentResponse, error)
}

// folderProvider defines an interface for providing folder names based on a base path
// and a flag indicating whether to create a new folder.
type folderProvider interface {
	ProjectFolder(base string, mode seqdir.Mode) (string, error)
}

// materializer defines the interface for managing project structures
// including navigation files and materialization of the project layout.
type materializer interface {
	Materialize(hfile, hfold project.Handler) error
}

// navigator defines methods for managing navigation files within a project.
type navigator interface {
	LoadNav(key string) ([]byte, error)
	AddNav(key, ext string, content []byte) error
	Content(key string) (string, error)
}

package app

import (
	"context"

	"github.com/baudii/ada-ai/internal/project"
	"github.com/baudii/ada-ai/internal/project/folder"
	"github.com/tmc/langchaingo/llms"
)

// Generator defines methods for generating AI content and handling prompts.
type Generator interface {
	PromptFromTemplate(filename string, input ...any) (string, error)
	GenerateWithSys(ctx context.Context, sys, user string, callOptions ...llms.CallOption) (*llms.ContentResponse, error)
}

// DirProvider defines an interface for providing folder names based on a base path
// and a flag indicating whether to create a new folder.
type DirProvider interface {
	ProjectFolder(base string, mode folder.Mode) (string, error)
}

// Runner defines an interface for reading project data.
type Runner interface {
	ReadProjdata() ProjectData
}

// Manager defines the interface for managing project structures
// including navigation files and materialization of the project layout.
type Manager interface {
	Materialize(hfile, hfold project.Handler) error
}

// NavHandler defines methods for managing navigation files within a project.
type NavHandler interface {
	TryLoadNav(key string, dest *[]byte) bool
	AddNav(key, ext string, content []byte) error
	NavContent(key string) ([]byte, error)
}

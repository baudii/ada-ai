package infra

import (
	"fmt"
	"path/filepath"

	"github.com/baudii/ada-ai/internal/core/nav"
	"github.com/baudii/ada-ai/internal/core/project"
	"github.com/baudii/ada-ai/internal/core/seqdir"
	"github.com/baudii/ada-ai/internal/infra/ai"
	"github.com/baudii/ada-ai/internal/infra/config"
	"github.com/baudii/ada-ai/internal/infra/fsproject"
	"github.com/tmc/langchaingo/llms"
)

// Templates lists prompt template filenames used by the system prompts.
type Templates struct {
    Reflect      string
    ShortReflect string
    Improve      string
}

// folderProvider defines an interface for providing folder names based on a base path
// and a flag indicating whether to create a new folder.
type folderProvider interface {
	ProjectFolder(base string, mode seqdir.Mode) (string, error)
}

// NewAI initializes the Ada AI model with the specified LLM provider
// and configuration. It sets up both a JSON-capable AI model and a standard
// text AI model, and configures the Ada workflow with the provided options.
func NewAI(provider, configPath string) (llms.Model, error) {
	cfg, err := config.LoadWithLocal[ai.Config](configPath)
	if err != nil {
		return nil, fmt.Errorf("parse llm config: %w", err)
	}

	ai, err := ai.Register(provider, cfg.Options)
	if err != nil {
		return nil, fmt.Errorf("register ai %q: %w", provider, err)
	}

	return ai, nil
}

// NewFSProject initializes a filesystem-backed project under the user/project root.
// It chooses the project folder according to the mode and returns the project handle.
func NewFSProject(projRoot string, projData project.Data, mode seqdir.Mode, folderer folderProvider) (*fsproject.Data, error) {
	base := filepath.Join(projRoot, projData.UserName, projData.ProjName)
	projRoot, err := folderer.ProjectFolder(base, mode)
	if err != nil {
		return nil, fmt.Errorf("project folder: %w", err)
	}
	lp, err := fsproject.New(projRoot, nav.New())
	if err != nil {
		return nil, fmt.Errorf("create local project: %w", err)
	}
	return lp, nil
}

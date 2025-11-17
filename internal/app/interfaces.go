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

type materializer interface {
	Validate(ctx context.Context) error
	Materialize(ctx context.Context, hfile project.FileHandler, hfold project.FolderHandler) error
}

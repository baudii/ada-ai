package app

import (
	"context"

	"github.com/baudii/ada-ai/internal/core/aigen"
)

// Generator defines methods for generating AI content and handling prompts.
type Generator interface {
	BuildPrompt(filename string, input ...any) (string, error)
	GenerateWithSys(ctx context.Context, sys, user string, opts ...aigen.Option) (string, error)
}

type Materializer interface {
	Materialize(ctx context.Context, openapi string) error
	PrepareOutputDir() error
}

package app

import (
	"context"

	"github.com/baudii/ada-ai/internal/core/gen"
)

// generator defines methods for generating AI content and handling prompts.
type generator interface {
	BuildPrompt(filename string, input ...any) (string, error)
	GenerateWithSys(ctx context.Context, sys, user string, opts ...gen.Option) (string, error)
}

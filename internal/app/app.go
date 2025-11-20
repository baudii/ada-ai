package app

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/baudii/ada-ai/internal/core/gen"
)

// Constants that define folder names containing prompts.

// GenerateProject generates a project based on the provided OpenAPI specification.
func (a *App) GenerateProject(ctx context.Context, m Materializer) error {
	// Generate OpenAPI specification
	spec, err := a.GenerateOpenAPISpec(ctx)
	if err != nil {
		return fmt.Errorf("generate openapi spec: %w", err)
	}

	// Use it with the openapi spec generator to create the project
	if err := m.PrepareOutputDir(); err != nil {
		return fmt.Errorf("prepare output dir: %w", err)
	}

	if err := m.Materialize(ctx, spec); err != nil {
		return fmt.Errorf("materialize project: %w", err)
	}
	// Add additional logic to the generated project
	return nil
}

// SendInstructions sends instructions to the AI model using the specified prompt
// and arguments, returning the generated response as a byte slice.
func (a *App) SendInstructions(ctx context.Context, promptName string, args []any, opts ...gen.Option) ([]byte, error) {
	sys, hum, err := a.buildSysAndHumanPrompts(promptName, args...)
	if err != nil {
		return nil, fmt.Errorf("system and human prompts: %w", err)
	}

	fmt.Println(hum)

	resp, err := a.generator.GenerateWithSys(ctx, sys, hum, opts...)
	if err != nil {
		return nil, fmt.Errorf("generate with sys: %w", err)
	}

	return []byte(resp), nil
}

func (a *App) buildSysAndHumanPrompts(p string, args ...any) (string, string, error) {
	sys, err := a.generator.BuildPrompt(filepath.Join(p, "system.txt"))
	if err != nil {
		return "", "", fmt.Errorf("load system template: %w", err)
	}

	hum, err := a.generator.BuildPrompt(filepath.Join(p, "human.txt"), args...)
	if err != nil {
		return "", "", fmt.Errorf("load human template: %w", err)
	}

	return sys, hum, nil
}

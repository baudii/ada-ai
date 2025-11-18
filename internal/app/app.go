package app

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/baudii/ada-ai/internal/core/gen"
)

// Constants that define folder names containing prompts.
const (
	filler = "handler-generator"
)

// Run initializes and runs the CLI application. It sets up the AI model,
// configures the Ada AI workflow, and handles user input to generate and
// materialize a project based on the provided description.
//
// It also manages configuration loading and error handling throughout the process.
//
// Additional arguments can be passed to modify the behavior of the application.
func (a *app) Run(ctx context.Context) error {
	a.logger.Info("starting app session", "user", a.projectContext.UserName, "project", a.projectContext.Name)

	// Generate OpenAPI specification

	// Use it with the openapi spec generator to create the project
	if err := a.materializer.Validate(ctx); err != nil {
		return err
	}

	err := a.materializer.Materialize(
		ctx,
		func(path string) error {
			// Here I should use LLM to generate the contents of the file.
			// File already exists at 'path' and contains the description
			// of the method to implement.
			return nil
		},
		nil, // Folder handler not needed for OpenAPI project
	)

	if err != nil {
		return err
	}

	// Add additional logic to the generated project
	return nil
}

func (a *app) generateFromContent(ctx context.Context, path string) error {
	res, err := a.sendInstructions(ctx, filler, nil)
	if err != nil {
		return fmt.Errorf("generate file %w", err)
	}

	if err := a.materializer.ApplyContent(ctx, path, res); err != nil {
		return fmt.Errorf("write file: %w", err)
	}
	a.logger.Info("success", "path", path)
	return nil
}

func (a *app) sendInstructions(ctx context.Context, promptName string, args []any, opts ...gen.Option) ([]byte, error) {
	sys, hum, err := a.sysHumanPrompts(promptName, args...)
	if err != nil {
		return nil, fmt.Errorf("system and human prompts: %w", err)
	}

	resp, err := a.generator.GenerateWithSys(ctx, sys, hum, opts...)
	if err != nil {
		return nil, fmt.Errorf("generate with sys: %w", err)
	}

	return []byte(resp), nil
}

func (a *app) sysHumanPrompts(p string, args ...any) (string, string, error) {
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

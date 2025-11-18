package app

import (
	"context"
)

// Constants that define folder names containing prompts.
const (
	business  = "business"
	technical = "technical"
	scope     = "scope"
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

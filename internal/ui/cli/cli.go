package cli

import (
	"context"
	"log/slog"
	"path/filepath"

	"github.com/baudii/ada-ai/internal/app"
	"github.com/baudii/ada-ai/internal/core/project"
	"github.com/baudii/ada-ai/internal/infra/config"
)

type cliApp struct {
	logger         *slog.Logger
	app            *app.App
	materializer   app.Materializer
	projectContext project.Context
	read           func(string) string
	save           func(any, string) error
}

type option func(*cliApp)

// WithLogger sets the logger used by the CLI application.
func WithLogger(logger *slog.Logger) option {
	return func(c *cliApp) {
		c.logger = logger
	}
}

func WithApp(a *app.App) option {
	return func(c *cliApp) {
		c.app = a
	}
}

func WithMaterializer(m app.Materializer) option {
	return func(a *cliApp) {
		a.materializer = m
	}
}

// New creates a new instance of the CLI application that implements the Runner interface.
func New(opts ...option) *cliApp {
	c := &cliApp{logger: slog.Default(), read: ReadInput, save: config.Save}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Run initializes and runs the CLI application. It sets up the AI model,
// configures the Ada AI workflow, and handles user input to generate and
// materialize a project based on the provided description.
//
// It also manages configuration loading and error handling throughout the process.
//
// Additional arguments can be passed to modify the behavior of the application.
func (a *cliApp) Run(ctx context.Context) error {
	a.logger.Info("starting app session", "user", a.projectContext.UserName, "project", a.projectContext.Name)

	if err := a.app.GenerateProject(ctx, a.materializer); err != nil {
		return err
	}

	return nil
}

// GetProjectData retrieves project data from a JSON file or prompts the user for input
// if the file does not exist or cannot be parsed. It sends the project data
// through the provided channel and closes the channel when done.
func (c *cliApp) GetProjectData(userConfigRoot string) {
	path := filepath.Join(userConfigRoot, "project_context.json")
	ctx, err := config.Load[project.Context](path)
	if err != nil {
		username := c.read("Provide nickname")
		projname := c.read("Provide project name")
		plang := c.read("Provide programming language (go, python, js, etc)")
		summary := c.read("Provide a short summary of the project")
		ctx = project.Context{
			UserName: username,
			Name:     projname,
			Language: plang,
			Summary:  summary,
		}
	}

	if err = c.save(ctx, path); err != nil {
		c.logger.Error("failed to save project data to file", "path", path, "error", err)
	}

	c.projectContext = ctx
}

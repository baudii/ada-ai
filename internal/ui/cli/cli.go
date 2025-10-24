package cli

import (
	"log/slog"
	"path/filepath"

	"github.com/baudii/ada-ai/internal/core/project"
	"github.com/baudii/ada-ai/internal/infra/config"
)

type cliApp struct {
	logger *slog.Logger
	read   func(string) string
	save   func(any, string) error
}

type option func(*cliApp)

// WithLogger sets the logger used by the CLI application.
func WithLogger(logger *slog.Logger) option {
    return func(c *cliApp) {
        c.logger = logger
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

// GetProjectData retrieves project data from a JSON file or prompts the user for input
// if the file does not exist or cannot be parsed. It sends the project data
// through the provided channel and closes the channel when done.
func (cli *cliApp) GetProjectData(userConfigRoot string) project.Data {
	path := filepath.Join(userConfigRoot, "user_data.json")
	projectData, err := config.Load[project.Data](path)
	if err != nil {
		username := cli.read("Provide nickname")
		projname := cli.read("Provide project name")
		plang := cli.read("Provide programming language (go, python, js, etc)")
		summary := cli.read("Provide a short summary of the project")
		projectData = project.Data{
			UserName: username,
			ProjName: projname,
			Language: plang,
			Summary:  summary,
		}
	}

	if err = cli.save(projectData, path); err != nil {
		cli.logger.Error("failed to save project data to file", "path", path, "error", err)
	}

	return projectData
}

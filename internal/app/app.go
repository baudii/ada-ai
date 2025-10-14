package app

import (
	"log/slog"
	"path/filepath"

	"github.com/baudii/ada-ai/internal/adacore"
	"github.com/baudii/ada-ai/internal/ai"
	"github.com/baudii/ada-ai/internal/common"
	"github.com/baudii/ada-ai/pkg/utils"
)

var defaultCfg adacore.Options = adacore.Options{
	Timeout: "3m",
}

type Runner interface {
	Projdata(chan adacore.ProjectData)
}

// Run initializes and runs the CLI application. It sets up the AI model,
// configures the Ada AI workflow, and handles user input to generate and
// materialize a project based on the provided description.
//
// It also manages configuration loading and error handling throughout the process.
func Run(runner Runner) {
	ai, err := ai.RegisterFromFile(common.ConfigPath)
	if err != nil {
		slog.Error("failed to register llm", "error", err)
		return
	}

	var opts *adacore.Options
	cfgPath := filepath.Join(common.ConfigPath, "ada.json")
	opts, err = utils.ParseJSONConfigWithLocal[adacore.Options](cfgPath)
	if err != nil {
		slog.Error("failed to parse json configuration", "path", cfgPath, "error", err)
		opts = &defaultCfg
	} else {
		slog.Info("succesffully parsed json configuration", "cfg", opts)
	}

	if opts.PromptsRoot == "" {
		opts.PromptsRoot = common.PromptsPath
	}
	if opts.ProjectsRoot == "" {
		opts.ProjectsRoot = common.ProjectsPath
	}

	ada := adacore.New(ai, adacore.WithOptions(*opts))
	c := make(chan adacore.ProjectData)
	go runner.Projdata(c)
	data := <-c
	ada.AddProjectData(data)

	slog.Info("starting Ada AI session", "user", ada.Project.UserName, "project", ada.Project.ProjName)
}

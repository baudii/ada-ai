package app

import (
	"log/slog"
	"os"
	"path/filepath"

	"github.com/baudii/ada-ai/internal/adacore"
	"github.com/baudii/ada-ai/internal/ai"
	"github.com/baudii/ada-ai/internal/common"
	"github.com/baudii/ada-ai/internal/projects"
	"github.com/baudii/ada-ai/pkg/utils"
)

var defaultCfg adacore.Options = adacore.Options{
	Timeout: "3m",
}

type Runner interface {
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

	input := utils.ReadInput("Provide project description")
	step1Template, err := ada.PromptFromTemplate(common.ProjectStructurePrompt, ada.Project.ProjName, input)
	if err != nil {
		slog.Error("failed to get step1 prompt from template", "error", err)
		os.Exit(1)
	}
	data, err := ada.SendReflect(step1Template)
	if err != nil {
		slog.Error("failed send reflect", "error", err)
		os.Exit(1)
	}
	ld, err := projects.New(data, ada)
	if err != nil {
		slog.Error("failed to create new local project", "error", err)
		os.Exit(1)
	}
	ada.Proj = ld
	err = utils.PrintTree(os.Stdout, ada.Proj.Structure(), "")
	if err != nil {
		slog.Error("failed to print the tree", "error", err)
	}

	err = ada.Proj.Materialize()
	if err != nil {
		slog.Error("failed to materialize project", "error", err)
		os.Exit(1)
	}
}

// c := make(chan adacore.SessionOption)
// go runner.Options(c)
// var options []adacore.SessionOption
// for v := range c {
// 	options = append(options, v)
// }
// options = append(options, adacore.WithConfig(opts))

package app

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/baudii/ada-ai/internal/adacore"
	"github.com/baudii/ada-ai/internal/ai"
	"github.com/baudii/ada-ai/internal/common"
	"github.com/baudii/ada-ai/internal/projects"
	"github.com/baudii/ada-ai/pkg/utils"
)

var (
	new bool = false
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
//
// Additional arguments can be passed to modify the behavior of the application.
func Run(ada *adacore.Ada, runner Runner, args ...any) error {
	parseArgs(args...)
	if err != nil {
		return fmt.Errorf("business description: %w", err)
	}

	slog.Debug("received business description", "description", busnessDesc)
	techDesc, err := sendInstructions(ada, tech, ada.Projdata.Language, busnessDesc)
	if err != nil {
		return fmt.Errorf("technical description: %w", err)
	}

	slog.Debug("received technical description", "description", techDesc)
	scopeDesc, err := sendInstructions(ada, scope, busnessDesc, techDesc)
	if err != nil {
		return fmt.Errorf("scope description: %w", err)
	}

	slog.Debug("received scope description", "description", scopeDesc)
	projStructure, err := sendInstructions(ada, projstruct, busnessDesc, techDesc)
	if err != nil {
		return fmt.Errorf("project structure: %w", err)
	}

	slog.Debug("received project structure", "structure", projStructure)

	lp, err := projects.New(
		[]byte(projStructure),
		ada,
		projects.NewDesc("business-description.json", []byte(busnessDesc)),
		projects.NewDesc("technical-description.json", []byte(techDesc)),
	)
	if err != nil {
		return fmt.Errorf("create local project: %w", err)
	}

	slog.Debug("created local project")
	err = utils.PrintTree(os.Stdout, lp.Structure(), "")
	if err != nil {
		return fmt.Errorf("print tree: %w", err)
	}
	err = lp.Materialize()
	if err != nil {
		return fmt.Errorf("materialize project: %w", err)
	}

	slog.Debug("project materialized")
	return nil
}

func InitAda(provider string, runner Runner) (*adacore.Ada, error) {
	ai, err := ai.RegisterFromFile(provider, common.AiConfigPath)
	if err != nil {
		return nil, fmt.Errorf("register llm: %w", err)
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
	slog.Info("starting app session", "user", ada.Projdata.UserName, "project", ada.Projdata.ProjName)
	return ada, nil
}

func sendInstructions(ada *adacore.Ada, p prompt, args ...any) (string, error) {
	sys, err := ada.PromptFromTemplate(p.system)
	if err != nil {
		return "", fmt.Errorf("load system template: %w", err)
	}

	hum, err := ada.PromptFromTemplate(p.human, args...)
	if err != nil {
		return "", fmt.Errorf("load human template: %w", err)
	}

	resp, err := ada.GenerateWithSys(sys, hum)
	if err != nil {
		return "", fmt.Errorf("generate with sys: %w", err)
	}
func parseArgs(args ...any) {
	for _, v := range args {
		switch v := v.(type) {
		case bool:
			new = v
		default:
			panic("unknown argument")
		}
	}
}

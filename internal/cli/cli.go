package cli

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

var defaultCfg adacore.Config = adacore.Config{
	Timeout: "3m",
}

// Run initializes and runs the CLI application. It sets up the AI model,
// configures the Ada AI workflow, and handles user input to generate and
// materialize a project based on the provided description.
//
// It also manages configuration loading and error handling throughout the process.
func Run() {
	ai, err := ai.RegisterFromFile(common.ConfigPath)
	if err != nil {
		slog.Error("failed to register llm", "error", err)
		return
	}

	if Debug {
		enableDebugging(ai)
		return
	}

	var cfg *adacore.Config
	cfgPath := filepath.Join(common.ConfigPath, "ada.json")
	cfg, err = utils.ParseJSONConfigWithLocal[adacore.Config](cfgPath)
	if err != nil {
		slog.Error("failed to parse json configuration", "path", cfgPath, "error", err)
		cfg = &defaultCfg
	} else {
		slog.Info("succesffully parsed json configuration", "cfg", cfg)
	}

	options := getOptions()
	options = append(options, adacore.WithConfig(cfg))
	ada := adacore.New(ai, options...)

	input := utils.ReadInput("Provide project description")
	step1Template, err := ada.PromptFromTemplate(adacore.ProjectStructurePrompt, ada.Session.Project.ProjName, input)
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

func getOptions() []adacore.SessionOption {
	path := filepath.Join(common.DataPath, "user_data.json")
	var options []adacore.SessionOption
	options = append(options, adacore.WithProjectsRoot(common.ProjectsPath), adacore.WithPromptsRoot(common.PromptsPath))
	projectData, err := utils.ParseJSONFile[adacore.ProjectData](path)
	if err != nil {
		username := utils.ReadInput("Provide nickname")
		projname := utils.ReadInput("Provide project name")
		projectData = &adacore.ProjectData{UserName: username, ProjName: projname}
		err = utils.SaveJSONToFile(projectData, path)
		if err != nil {
			slog.Error("failed to save project data", "error", err)
		}
	}

	return append(options, adacore.WithProjectData(*projectData))
}

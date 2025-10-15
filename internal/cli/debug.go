package cli

import (
	"log"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/baudii/ada-ai/internal/adacore"
	"github.com/baudii/ada-ai/internal/common"
	"github.com/baudii/ada-ai/internal/projects"
	"github.com/baudii/ada-ai/pkg/utils"
	"github.com/tmc/langchaingo/llms"
)

const desc string = "An AI-powered app that suggests recipes based on the ingredients you already have at home"

var (
	Debug      bool
	DebugStage int
)

var (
	debugProjData = adacore.ProjectData{
		UserName: "debug-user",
		ProjName: "pantrypal",
	}
)

func enableDebugging(ai llms.Model) {
	switch DebugStage {
	case 0:
		debugProjectStructure(ai)
	case 1:
		debugStep1(ai)
	}
}

func debugStep1(ai llms.Model) {
	opts := adacore.Options{
		PromptsRoot:  common.PromptsPath,
		ProjectsRoot: common.ProjectsPath,
	}
	ada := adacore.New(ai, nil, adacore.WithOptions(opts))
	ada.AddProjectData(debugProjData)

	utils.ReadInput("Press Enter to continue")

	template, err := ada.PromptFromTemplate(common.ProjectStructurePrompt, ada.Projdata.ProjName, desc)
	if err != nil {
		log.Fatal("failed to get step1 prompt from template", "error", err)
	}

	data, err := ada.SendReflect(template)
	if err != nil {
		panic(err)
	}

	ld, err := projects.New(data, ada)
	if err != nil {
		log.Fatal("couldn't parse a description into a valid json")
	}
	ada.Proj = ld

	err = utils.PrintTree(os.Stdout, ada.Proj.Structure(), "")
	if err != nil {
		slog.Error("failed to print the tree", "error", err)
	}
	_ = ada.Proj.Materialize()
}

func debugProjectStructure(ai llms.Model) {
	path := filepath.Join(common.DebuggingPath, "structure-unparsed.json")
	f, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	opts := adacore.Options{
		PromptsRoot:  common.PromptsPath,
		ProjectsRoot: common.ProjectsPath,
	}
	ada := adacore.New(ai, nil, adacore.WithOptions(opts))
	ada.AddProjectData(debugProjData)

	if f, err = utils.TrimJSON(string(f)); err != nil {
		panic(err)
	}

	ld, err := projects.New(f, ada)
	if err != nil {
		panic(err)
	}

	ada.Proj = ld
	err = utils.PrintTree(os.Stdout, ada.Proj.Structure(), "")
	if err != nil {
		slog.Error("failed to print the tree", "error", err)
	}
	err = ld.Materialize()
	if err != nil {
		slog.Error("error occured whe materializing", "error", err)
	}
}

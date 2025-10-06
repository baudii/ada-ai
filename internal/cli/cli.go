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
	ProjRoot: ".projects",
	Timeout:  "3m",
}

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

	ada := adacore.New(ai, cfg)
	setUserData(ada)

	input := utils.ReadInput("Provide project description")
	step1Template, err := adacore.PromptFromTemplate(adacore.Step1PromptFile, ada.Ctx.ProjName, input)
	if err != nil {
		slog.Error("failed to get step1 prompt from template", "error", err)
		os.Exit(1)
	}
	data, err := ada.SendReflect(step1Template)
	if err != nil {
		slog.Error("failed send reflect", "error", err)
		os.Exit(1)
	}
	ld, err := projects.NewLocalProj(data, ada.ResolveProjectPath())
	if err != nil {
		slog.Error("failed to create new local project", "error", err)
		os.Exit(1)
	}
	ada.SetWorkspace(ld)
	utils.PrintTree(os.Stdout, ada.Ctx.Proj.Structure(), "")
	ada.Ctx.Proj.Materialize()
}

func setUserData(ada *adacore.Ada) {
	if ada.Ctx != nil && ada.Ctx.UserName != "" && ada.Ctx.ProjName != "" {
		slog.Info("recognized project context", "context", ada.Ctx)
		return
	}

	username := utils.ReadInput("Provide nickname")
	projname := utils.ReadInput("Provide project name")
	ada.AddProjCtx(username, projname)
	if err := ada.SaveCtx(); err != nil {
		slog.Warn("failed to save project context but will use it in this session", "context", ada.Ctx, "error", err)
	} else {
		slog.Info("saved project context", "context", ada.Ctx)
	}
}

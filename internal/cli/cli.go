package cli

import (
	"log"
	"log/slog"
	"path/filepath"

	"github.com/baudii/ada-ai/internal/adacore"
	"github.com/baudii/ada-ai/internal/ai"
	"github.com/baudii/ada-ai/internal/common"
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
	relativePath := filepath.Join(common.ConfigPath, "ada.json")
	cfgPath := utils.GetAbsolutePath(relativePath)
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
	step1Template, err := adacore.GetPromptFromTemplate(adacore.Step1PromptFile, ada.Ctx.ProjName, input)
	if err != nil {
		log.Fatal("failed to get step1 prompt from template", "error", err)
	}

	data, err := ada.SendWithReflection(step1Template)
	if err != nil {
		slog.Error("something went wrong when processing the request", "error", err)
	}

	adacore.Print(data)
	ada.EnsureSaved(data)
	ada.Materialize(data)
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
		slog.Error("failed to save project context but will use it in this session", "context", ada.Ctx, "error", err)
	} else {
		slog.Info("saved project context", "context", ada.Ctx)
	}
}

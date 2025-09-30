package cli

import (
	"log/slog"
	"path/filepath"

	"github.com/baudii/ada-ai/internal/adacore"
	"github.com/baudii/ada-ai/internal/common"
	"github.com/baudii/ada-ai/pkg/utils"
)

var defaultCfg adacore.Config = adacore.Config{
	ProjRoot: ".projects",
	Timeout:  "3m",
}

func Run() {
	ai := common.RegisterOllama()
	if Debug {
		enableDebugging()
		return
	}

	var cfg *adacore.Config
	var err error
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
	template := ada.GetTemplate(1, input)
	data, err := ada.SendWithReflection(template)
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
	if err := utils.SaveJSONToFile(ada.Ctx, utils.GetAbsolutePath(adacore.UserDataPath)); err != nil {
		slog.Error("failed to save project context but will use it in this session", "context", ada.Ctx, "error", err)
	} else {
		slog.Info("saved project context", "context", ada.Ctx)
	}
}

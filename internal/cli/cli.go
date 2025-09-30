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
	relativePath := filepath.Join(common.ConfigPath, "adacore.json")
	cfgPath := utils.GetAbsolutePath(relativePath)
	cfg, err = utils.ParseJSONConfigWithLocal[adacore.Config](cfgPath)
	if err != nil {
		cfg = &defaultCfg
	}

	ada := adacore.New(ai, cfg)
	setUserData(ada)
	input := utils.ReadInput("Provide project description")
	data, err := ada.SendWithReflection(input)
	if err != nil {
		slog.Error("something went wrong when processing the request", "error", err)
	}

	adacore.Print(data)
}

func setUserData(ada *adacore.Ada) {
	if ada.Ctx.UserName != "" && ada.Ctx.ProjName != "" {
		return
	}

	ada.Ctx.UserName = utils.ReadInput("Provide nickname")
	slog.Info("recognized username", "username", ada.Ctx.UserName)
	ada.Ctx.ProjName = utils.ReadInput("Provide project name")
	slog.Info("recognized project", "projname", ada.Ctx.ProjName)
	utils.SaveJSONToFile(ada.Ctx, utils.GetAbsolutePath(adacore.UserDataPath))
}

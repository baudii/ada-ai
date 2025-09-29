package cli

import (
	"log/slog"
	"path/filepath"

	"github.com/baudii/ada-ai/internal/ada"
	"github.com/baudii/ada-ai/internal/common"
	"github.com/baudii/ada-ai/pkg/utils"
)

var cfgPath string

var defaultCfg ada.Config = ada.Config{
	ProjRoot: ".projects",
	Timeout:  "3m",
}

func Run() {
	ai := common.RegisterOllama()
	if Debug {
		enableDebugging()
		return
	}

	var cfg *ada.Config
	var err error
	relativePath := filepath.Join(common.ConfigPath, "ada.json")
	cfgPath = utils.GetAbsolutePath(relativePath)
	cfg, err = utils.ParseJSONConfigWithLocal[ada.Config](cfgPath)
	if err != nil {
		cfg = &defaultCfg
	}

	ada.Init(ai, cfg)
	setUserData(cfg)
	input := utils.ReadInput("Provide project description")
	data, err := ada.SendWithReflection(input)
	if err != nil {
		slog.Error("something went wrong when processing the request", "error", err)
	}

	ada.Print(data)
}

func setUserData(cfg *ada.Config) {
	// TODO: switch to other way of storing user data
	if cfg.UserName == "" {
		cfg.UserName = utils.ReadInput("Provide nickname")
		utils.SaveJSONToFile(cfg, cfgPath)
	}
	slog.Info("recognized username", "username", cfg.UserName)

	if cfg.ProjName == "" {
		cfg.ProjName = utils.ReadInput("Provide project name")
		utils.SaveJSONToFile(cfg, cfgPath)
	}
	slog.Info("recognized project", "projname", cfg.ProjName)
}

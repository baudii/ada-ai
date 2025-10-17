package main

import (
	"log"
	"log/slog"
	"path/filepath"

	"github.com/baudii/ada-ai/internal/app"
	"github.com/baudii/ada-ai/internal/cli"
	"github.com/baudii/ada-ai/internal/common"
)

func main() {
	cfgPath := filepath.Join(common.ConfigPath, "dilog.json")
	cfg := common.LoadLogConfig(cfgPath)
	logger, err := common.DefaultSimpleLogger(cfg)
	if err != nil {
		log.Fatal("failed to create logger", "error", err)
	}
	slog.SetDefault(logger)
	logger.Info("starting application")
	err = app.Run(cli.New())
	if err != nil {
		slog.Error("application error", "error", err)
	}
}

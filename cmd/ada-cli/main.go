package main

import (
	"flag"
	"log/slog"
	"path/filepath"

	"github.com/baudii/ada-ai/internal/cli"
	"github.com/baudii/ada-ai/internal/common"
	"github.com/baudii/ada-ai/pkg/dilog"
	"github.com/baudii/ada-ai/pkg/utils"
)

func main() {
	initLogger()
	parseFlags()
	cli.Run()
}

func parseFlags() {
	dbg := flag.Bool("debug", false, "Enable an application in a Debug mode")
	stage := flag.Int("stage", 0, "Choose the stage you want to debug")
	flag.Parse()

	slog.Debug("debug flags parsed", "debug", *dbg, "stage", *stage)
	cli.Debug = *dbg
	cli.DebugStage = *stage
}

func initLogger() {
	cfgPath := filepath.Join(common.ConfigPath, "dilog.json")
	cfg, err := utils.ParseJSONConfigWithLocal[dilog.Config](cfgPath)
	if err != nil {
		cfg = &dilog.Config{
			Timezone: "local",
			Path:     "logs",
			Prefix:   "ada",
		}
	}

	dilog.Init(cfg)
}

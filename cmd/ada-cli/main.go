package main

import (
	"flag"
	"log"
	"log/slog"
	"path/filepath"

	"github.com/baudii/ada-ai/internal/ada"
	"github.com/baudii/ada-ai/pkg/dilog"
	"github.com/baudii/ada-ai/pkg/utils"
	"github.com/tmc/langchaingo/llms/ollama"
)

func main() {
	cfgPath := filepath.Join("cfg", "dilog.json")
	cfg, err := utils.ParseJsonConfigWithLocal[dilog.Config](utils.GetAbsolutePath(cfgPath))
	if err != nil {
		cfg = &dilog.Config{
			Timezone: "local",
			Path:     utils.GetAbsolutePath("logs"),
			Prefix:   "ada",
		}
	}

	dilog.Init(cfg)
	parseFlags()
	slog.Info("registering ollama")
	l, err := ollama.New(ollama.WithModel("gemma3:4b"))
	if err != nil {
		log.Fatal(err)
	}

	ada.Run(l)
}

func parseFlags() {
	dbg := flag.Bool("debug", false, "Enable an application in a Debug mode")
	stage := flag.Int("stage", 0, "Choose the stage you want to debug")
	flag.Parse()

	slog.Debug("debug flags parsed", "debug", *dbg, "stage", *stage)
	ada.Debug = *dbg
	ada.DebugStage = *stage
}

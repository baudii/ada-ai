package main

import (
	"flag"
	"log"
	"log/slog"
	"path/filepath"

	"github.com/baudii/ada-ai/internal/app"
	"github.com/baudii/ada-ai/internal/cli"
	"github.com/baudii/ada-ai/internal/common"
)

func main() {
	provider := flag.String("provider", "grok", "LLM provider to use (openai, ollama, etc.)")
	flag.Parse()
	cfgPath := filepath.Join(common.ConfigPath, "dilog.json")
	cfg := common.LoadLogConfig(cfgPath)
	logger, err := common.DefaultSimpleLogger(cfg)
	if err != nil {
		log.Fatalf("failed to create logger: %v", err)
	}

	logger.Info("starting application")
	runner := cli.New()
	ada, err := app.InitAda(*provider, runner)
	if err != nil {
		log.Fatalf("failed to initialize ada: %v", err)
	}

	err = app.Run(ada, runner)
	if err != nil {
		slog.Error("application error", "error", err)
	}
}

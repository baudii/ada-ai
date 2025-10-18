package main

import (
	"context"
	"flag"
	"log"
	"log/slog"
	"path/filepath"

	"github.com/baudii/ada-ai/internal/app"
	"github.com/baudii/ada-ai/internal/cli"
	"github.com/baudii/ada-ai/internal/common"
	"github.com/baudii/ada-ai/pkg/utils"
)

func main() {
	// Parse command-line flags
	provider := flag.String("provider", "grok", "LLM provider to use (openai, ollama, etc.)")
	new := flag.Bool("new", false, "create a new project folder even if one exists")
	deg := flag.Int("deg", 1, "degree of concurrency for project materialization")
	flag.Parse()

	// Set up logging
	cfgPath := filepath.Join(common.ConfigPath, "dilog.json")
	cfg := common.LoadLogConfig(cfgPath)
	logger := utils.Must(common.DefaultSimpleLogger(cfg))
	slog.SetDefault(logger)
	logger.Info("initialized logger", "config", cfgPath)

	// Create and run the application
	app := app.New(app.WithDegree(*deg), app.WithNew(*new))
	utils.MustErr(app.InitAda(*provider))
	utils.MustErr(app.InitProject())
	app.ReceiveProjdata(cli.New())
	if err := app.Run(context.Background()); err != nil {
		log.Fatalf("runtime error: %v", err)
	}
}

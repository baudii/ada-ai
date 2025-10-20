package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/baudii/ada-ai/internal/app"
	"github.com/baudii/ada-ai/internal/cli"
	"github.com/baudii/ada-ai/internal/common"
	"github.com/baudii/ada-ai/internal/project/folder"
	"github.com/baudii/ada-ai/pkg/utils"
)

func main() {
	// Parse command-line flags
	provider := flag.String("provider", "grok", "LLM provider to use (openai, ollama, etc.)")
	mode := flag.Int("mode", 0, "project mode: 0=reuse existing, 1=create new, default is 0")
	deg := flag.Int("deg", 4, "degree of concurrency for project materialization")
	flag.Parse()

	// Set up logging
	cfgPath := filepath.Join(common.ConfigPath, "dilog.json")
	cfg := common.LoadLogConfig(cfgPath)
	logger := utils.Must(common.DefaultSimpleLogger(cfg))
	slog.SetDefault(logger)
	logger.Info("initialized logger", "config", cfgPath)

	// Create and run the application
	opts := utils.Must(app.ParseAppOptions(common.ConfigPath))
	path := filepath.Join(common.AiConfigPath, fmt.Sprintf("%s.json", *provider))
	c := cli.New().GetProjectData()
	app := app.New(
		app.WithDegree(*deg),
		app.WithMode(folder.Mode(*mode)),
		app.WithOptions(opts),
		app.WithProjectData(c),
		app.WithDirProvider(folder.NewSeqDir(os.ReadDir)),
	)
	utils.MustErr(app.InitAda(*provider, path))
	utils.MustErr(app.InitLocalProject())
	if err := app.Run(context.Background()); err != nil {
		log.Fatalf("runtime error: %v", err)
	}
}

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
	"github.com/baudii/ada-ai/internal/folders"
	"github.com/baudii/ada-ai/internal/logx"
	"github.com/baudii/ada-ai/internal/project/seqdir"
	"github.com/baudii/ada-ai/pkg/dilog"
	"github.com/baudii/ada-ai/pkg/must"
)

func main() {
	// Parse command-line flags
	provider := flag.String("provider", "grok", "LLM provider to use (openai, ollama, etc.)")
	mode := flag.Int("mode", 0, "project mode: 0=reuse existing, 1=create new, default is 0")
	deg := flag.Int("deg", 4, "degree of concurrency for project materialization")
	flag.Parse()

	// Set up logging
	cfgPath := filepath.Join(folders.Config, "dilog.json")
	cfg := logx.LoadLogConfigOrDefault(cfgPath)
	logger := must.Value(dilog.DefaultDailyLogger(&cfg))
	slog.SetDefault(logger)
	logger.Info("initialized logger", "config", cfgPath)

	// Create and run the application
	opts := must.Value(app.ParseAppOptions(folders.Config))
	path := filepath.Join(folders.AiConfig, fmt.Sprintf("%s.json", *provider))
	c := cli.New().GetProjectData()
	app := app.New(
		app.WithDegree(*deg),
		app.WithMode(seqdir.Mode(*mode)),
		app.WithOptions(opts),
		app.WithProjectData(c),
		app.WithDirProvider(seqdir.New(os.ReadDir)),
	)
	must.Do(app.InitAda(*provider, path))
	must.Do(app.InitLocalProject())
	if err := app.Run(context.Background()); err != nil {
		log.Fatalf("runtime error: %v", err)
	}
}

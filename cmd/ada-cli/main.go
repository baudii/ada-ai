package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"path/filepath"

	"github.com/baudii/ada-ai/internal/app"
	"github.com/baudii/ada-ai/internal/infra"
	"github.com/baudii/ada-ai/internal/infra/ai"
	"github.com/baudii/ada-ai/internal/infra/config"
	"github.com/baudii/ada-ai/internal/infra/dailylogger"
	"github.com/baudii/ada-ai/internal/infra/folders"
	"github.com/baudii/ada-ai/internal/infra/openapiproject"
	"github.com/baudii/ada-ai/internal/ui/cli"
)

func main() {
	// Parse command-line flags
	provider := flag.String("provider", "grok", "LLM provider to use (openai, ollama, etc.)")
	// mode := flag.Int("mode", 0, "project mode: 0=reuse existing, 1=create new, default is 0")
	deg := flag.Int("deg", 4, "degree of concurrency for project materialization")
	flag.Parse()

	// Set up logging
	cfgPath := filepath.Join(folders.Config, "dilog.json")
	cfg := dailylogger.LoadLogConfigOrDefault(cfgPath)
	logger := cli.Must(dailylogger.DefaultDailyLogger(&cfg))
	logger.Info("initialized logger", "config", cfgPath)

	// Create and run the application
	opts := cli.Must(config.ParseAppOptions(folders.Config))
	path := filepath.Join(folders.AiConfig, fmt.Sprintf("%s.json", *provider))
	c := cli.New(cli.WithLogger(logger)).GetProjectData(folders.Artifacts)
	llm := cli.Must(infra.NewAI(*provider, path))
	gen := ai.NewGenerator(llm,
		ai.WithPromptsRoot(opts.PromptsRoot),
		ai.WithTimeout(opts.Timeout),
	)

	//lp := cli.Must(infra.NewFSProject(folders.Projects, c, seqdir.New(os.ReadDir, seqdir.WithMode(seqdir.Mode(*mode)))))
	materializer := openapiproject.New(filepath.Join(folders.Projects, "oapitest"))
	app := app.New(
		app.WithMaterializer(materializer),
		app.WithDegree(*deg),
		app.WithOptions(opts),
		app.WithProjectData(c),
		app.WithGenerator(gen),
		app.WithLogger(logger),
	)

	if err := app.Run(context.Background()); err != nil {
		log.Fatalf("runtime error: %v", err)
	}
}

package main

import (
	"flag"
	"log/slog"

	"github.com/baudii/ada-ai/internal/app"
	"github.com/baudii/ada-ai/internal/cli"
)

func main() {
	parseFlags()
	app.Run(cli.New())
}

func parseFlags() {
	dbg := flag.Bool("debug", false, "Enable an application in a Debug mode")
	stage := flag.Int("stage", 0, "Choose the stage you want to debug")
	flag.Parse()

	slog.Debug("debug flags parsed", "debug", *dbg, "stage", *stage)
	cli.Debug = *dbg
	cli.DebugStage = *stage
}

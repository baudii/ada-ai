package main

import (
	"log/slog"

	"github.com/baudii/ada-ai/internal/app"
	"github.com/baudii/ada-ai/internal/cli"
)

func main() {
	err := app.Run(cli.New())
	if err != nil {
		slog.Error("application error", "error", err)
	}
}

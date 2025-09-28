package main

import (
	"flag"
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/baudii/ada-ai/internal/ada"
	"github.com/baudii/ada-ai/pkg/dailylog"
	"github.com/baudii/ada-ai/pkg/llm/providers/ollama"
	"github.com/baudii/ada-ai/pkg/utils"
)

func main() {
	cfgPath := filepath.Join("cfg", "dailylog.json")
	cfg, err := utils.ParseJsonFile[dailylog.Config](utils.GetAbsolutePath(cfgPath))
	fmt.Println(cfg)
	if err != nil {
		cfg = &dailylog.Config{
			Timezone: "local",
			Path:     utils.GetAbsolutePath("logs"),
			Prefix:   "ada",
		}
	}

	dailylog.Init(cfg)
	dbg := flag.Bool("debug", false, "Enable an application in a Debug mode")
	flag.Parse()
	slog.Info("registering ollama")
	ollama.Init()
	ada.Run(*dbg)
}

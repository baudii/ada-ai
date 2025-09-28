package main

import (
	"flag"
	"log/slog"
	"path/filepath"

	"github.com/baudii/ada-ai/internal/ada"
	"github.com/baudii/ada-ai/pkg/dilog"
	"github.com/baudii/ada-ai/pkg/llm/providers/ollama"
	"github.com/baudii/ada-ai/pkg/utils"
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
	dbg := flag.Bool("debug", false, "Enable an application in a Debug mode")
	flag.Parse()
	slog.Info("registering ollama")
	ollama.Init()
	ada.Run(*dbg)
}

package main

import (
	"flag"

	"github.com/baudii/ada-ai/internal/ada"
	"github.com/baudii/ada-ai/pkg/llm/providers/ollama"
)

func main() {
	dbg := flag.Bool("debug", false, "Enable an application in a Debug mode")
	flag.Parse()
	ollama.Init()
	ada.Run(*dbg)
}

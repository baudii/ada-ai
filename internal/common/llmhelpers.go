package common

import (
	"log"
	"log/slog"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
)

func RegisterOllama() llms.Model {
	slog.Info("registering ollama")
	l, err := ollama.New(ollama.WithModel("gemma3:4b"))
	if err != nil {
		log.Fatal(err)
	}

	return l
}

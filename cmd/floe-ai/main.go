package main

import (
	"fmt"
	"os"
	"time"

	"github.com/baudii/vision-maker-ai/internal/ollama"
)

const defaultOllamaURL = "http://localhost:11434/api/chat"

func main() {
	start := time.Now()
	msg := ollama.Message{
		Role:    "user",
		Content: "Why is the sky blue?",
	}
	req := ollama.Request{
		Model:    "gemma3:4b",
		Stream:   false,
		Messages: []ollama.Message{msg},
	}
	resp, err := ollama.TalkToOllama(defaultOllamaURL, req)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(resp.Message.Content)
	fmt.Printf("Completed in %v", time.Since(start))
}

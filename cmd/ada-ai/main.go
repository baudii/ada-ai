package main

import (
	"fmt"
	"time"

	"github.com/baudii/ada-ai/internal/ada"
	"github.com/baudii/ada-ai/internal/llm/providers/ollama"
)

func main() {
	start := time.Now()
	ollama.Init()
	ada.Run()
	fmt.Printf("Completed in %v", time.Since(start))
}

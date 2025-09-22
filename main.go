package main

import (
	"fmt"
	"os"
	"time"

	"github.com/baudii/floe-ai/internal/llm"
)

func main() {
	start := time.Now()
	cfg := llm.ParseCfg()
	llm, err := llm.Resolve(cfg)
	if err != nil {
		panic(err)
	}

	var prompt string
	fmt.Scanln(&prompt)
	resp, err := llm.SendMessage(prompt)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(resp.Message.Content)
	fmt.Printf("Completed in %v", time.Since(start))
}

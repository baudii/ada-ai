package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/baudii/ada-ai/internal/llm"
	_ "github.com/baudii/ada-ai/internal/llm/providers/ollama"
)

func main() {
	start := time.Now()
	cfg := llm.ParseCfg()
	llm, err := llm.Resolve(cfg)
	if err != nil {
		panic(err)
	}

	var prompt = readInput()
	resp, err := llm.SendMessage(prompt)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("Ai Response: %v\n", resp.Message.Content)
	fmt.Printf("Completed in %v", time.Since(start))
}

func readInput() string {
	fmt.Print("Enter message:\n > ")
	scanner := bufio.NewScanner(os.Stdin)
	var line string
	if scanner.Scan() {
		line = scanner.Text()
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading from stdin: %v", err)
	}

	return line
}

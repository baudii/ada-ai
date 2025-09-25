package utils

import (
	"bufio"
	"fmt"
	"log"
	"os"
)

func ReadInput(msg string) string {
	fmt.Printf("%v > ", msg)
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

package internal

import (
	"bufio"
	"fmt"
	"log"
	"os"
)

func ReadInput() string {
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

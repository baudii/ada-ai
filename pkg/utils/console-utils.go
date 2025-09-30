package utils

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
)

func ReadInput(msg string) string {
	for {
		fmt.Printf("%v > ", msg)
		scanner := bufio.NewScanner(os.Stdin)
		var line string
		if scanner.Scan() {
			line = scanner.Text()
		}

		if err := scanner.Err(); err != nil {
			slog.Error("error reading from stdin", "error", err)
			continue
		}

		return line
	}
}

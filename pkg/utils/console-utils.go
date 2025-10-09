package utils

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"os"
)

var (
	reader io.Reader    = os.Stdin
	writer io.Writer    = os.Stdout
	logger *slog.Logger = slog.Default()
)

func ReadInput(msg string) string {
	for {
		_, err := fmt.Fprintf(writer, "%v > ", msg)
		if err != nil {
			panic(fmt.Errorf("failed to write to the writer %v", writer))
		}

		scanner := bufio.NewScanner(reader)
		var line string
		if scanner.Scan() {
			line = scanner.Text()
		}

		if err := scanner.Err(); err != nil {
			logger.Error("error reading from reader", "error", err, "reader", reader)
			continue
		}

		return line
	}
}

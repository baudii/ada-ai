package console

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

// ReadInput reads a line from the reader and returns it as a string.
// If there is an error reading from the reader, it logs the error and
// prompts the user again.
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

// PrintTree writes a textual tree representation of the given structure to w.
// NOTE: since map is unordered by nature, there is not guarantee of an order in
// this function either.
//
// The structure is expected to be a nested map[string]any, where each key is
// treated as a directory or file name, and nested maps represent subdirectories.
//
// The output uses connector strings to visualize the hierarchy, similar to the
// Unix `tree` command.
//
// The prefix argument is used internally to align child elements and should
// normally be an empty string when called from outside.
func PrintTree(w io.Writer, structure map[string]any, prefix string) error {
	i := 0
	for k, v := range structure {
		i++
		conn := "├── "
		nextPrefix := prefix + "│   "
		if i == len(structure) {
			conn = "└── "
			nextPrefix = prefix + "    "
		}
		_, err := fmt.Fprintf(w, "%s%s%s\n", prefix, conn, k)
		if err != nil {
			return fmt.Errorf("writing to writer %v: %w", w, err)
		}
		if m, ok := v.(map[string]any); ok {
			err := PrintTree(w, m, nextPrefix)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

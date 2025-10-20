package nav

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
)

type item struct {
	root    string
	content []byte
}

func New(root string, content []byte) *item {
	return &item{
		root:    root,
		content: content,
	}
}

// NavContent retrieves the content of a navigation file by its key.
// It returns the content as a byte slice and a boolean indicating whether
// the file was found.
func (n *item) NavContent() (string, error) {
	buf := &bytes.Buffer{}
	if err := json.Compact(buf, n.content); err != nil {
		return "", fmt.Errorf("compact content: %w", err)
	}
	return buf.String(), nil
}

// Materialize creates the navigation file on disk with its content.
func (n *item) Materialize() error {
	return os.WriteFile(n.root, n.content, 0644)
}

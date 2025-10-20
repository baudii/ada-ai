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

// New creates a new navigation item with the specified root path and content.
func New(root string, content []byte) *item {
	return &item{
		root:    root,
		content: content,
	}
}

// CompactContent returns the compacted JSON string content of the
// navigation item.
func (n *item) CompactContent() (string, error) {
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

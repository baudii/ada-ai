package adacore

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

type Node struct {
	Name     string  `json:"name"`
	Type     string  `json:"type"`
	Children []*Node `json:"-"`
}

type nodeAlias struct {
	Name     string             `json:"name"`
	Type     string             `json:"type"`
	Children []*json.RawMessage `json:"children"`
	Contents []*json.RawMessage `json:"contents"`
}

func Print(data []byte) error {
	node, err := parseNode(data)
	if err != nil {
		return err
	}

	printTree(node, "", true)
	return nil
}

func parseNode(b []byte) (*Node, error) {
	var a nodeAlias
	if err := json.Unmarshal(b, &a); err != nil {
		return nil, err
	}

	n := &Node{
		Name: a.Name,
		Type: a.Type,
	}

	raws := make([]*json.RawMessage, 0, len(a.Children)+len(a.Contents))
	raws = append(raws, a.Children...)
	raws = append(raws, a.Contents...)

	for _, r := range raws {
		if r == nil {
			continue
		}
		child, err := parseNode(*r)
		if err != nil {
			return nil, err
		}
		n.Children = append(n.Children, child)
	}
	return n, nil
}

func (n *Node) materialize(base string) error {
	path := filepath.Join(base, n.Name)
	switch n.Type {
	case "folder":
		if err := os.MkdirAll(path, 0o755); err != nil {
			return err
		}
		for _, c := range n.Children {
			if err := c.materialize(path); err != nil {
				return err
			}
		}
	case "file":
		if _, err := os.Stat(path); os.IsNotExist(err) {
			f, err := os.Create(path)
			if err != nil {
				return err
			}
			defer f.Close()
		}
	default:
		slog.Warn("invalid json was provided", "node_type", n.Type)
	}
	return nil
}

func printTree(n *Node, prefix string, isLast bool) {
	conn := "├── "
	nextPrefix := prefix + "│   "
	if isLast {
		conn = "└── "
		nextPrefix = prefix + "    "
	}
	fmt.Printf("%s%s%s\n", prefix, conn, n.Name)
	for i, c := range n.Children {
		printTree(c, nextPrefix, i == len(n.Children)-1)
	}
}

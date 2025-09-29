package ada

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

var projDescrFile string
var projRoot string

const Artifacts string = "artifacts"

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

func parseNode(data string) (n *Node, err error) {
	return parseNodeHelper([]byte(data))
}

func parseNodeHelper(b []byte) (*Node, error) {
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
		child, err := parseNodeHelper(*r) // <- recurse with alias again
		if err != nil {
			return nil, err
		}
		n.Children = append(n.Children, child)
	}
	return n, nil
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

func createDirectory() error {
	path := filepath.Join(Artifacts, cfg.ProjRoot, cfg.UserName, cfg.ProjName)
	_, err := os.Stat(path)
	if err == nil {
		projRoot = path
		return &ProjExist{Path: path}
	}

	if !os.IsNotExist(err) {
		return err
	}

	err = os.MkdirAll(path, 0o755)
	if err != nil {
		return err
	}

	projRoot = path
	return nil
}

func saveProjectStructure(content string) (err error) {
	if err := createDirectory(); err != nil {
		return err
	}
	projDescrFile = filepath.Join(projRoot, "project-structure.json")
	return os.WriteFile(projDescrFile, []byte(content), 0644)
}

func (n *Node) printTree() {
	printTree(n, "", true)
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

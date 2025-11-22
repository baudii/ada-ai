package codeanalyzer

import (
	"bytes"
	"go/ast"
	"go/printer"
	"go/token"
	"strings"
)

func CommentText(cg *ast.CommentGroup) string {
	if cg == nil {
		return ""
	}
	return cg.Text()
}

func NodeToString(fset *token.FileSet, n ast.Node) string {
	var b bytes.Buffer
	_ = printer.Fprint(&b, fset, n)
	return b.String()
}

func FieldListToStrings(fset *token.FileSet, fl *ast.FieldList) []string {
	if fl == nil || len(fl.List) == 0 {
		return nil
	}

	var out []string
	for _, f := range fl.List {
		typStr := NodeToString(fset, f.Type)

		if len(f.Names) > 0 {
			for _, n := range f.Names {
				out = append(out, strings.TrimSpace(n.Name+" "+typStr))
			}
			continue
		}

		out = append(out, strings.TrimSpace(typStr))
	}

	return out
}

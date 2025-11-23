package codeanalyzer

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"strings"
)

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

// ExtractTypesOf extracts all type definitions whose Type is of kind T
// (e.g. *ast.InterfaceType, *ast.StructType, etc.).
func ExtractTypesOf[T ast.Expr](src string, name string) (string, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	config := &printer.Config{
		Mode:     printer.UseSpaces | printer.TabIndent,
		Tabwidth: 8,
	}

	first := true

	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.TYPE {
			continue
		}

		for _, spec := range gen.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			if _, ok := ts.Type.(T); !ok {
				continue
			}
			if name != "" && ts.Name.Name != name {
				continue
			}

			doc := ts.Doc
			if doc == nil {
				doc = gen.Doc
			}

			newTS := &ast.TypeSpec{
				Name:       ast.NewIdent(ts.Name.Name),
				Type:       ts.Type,
				Assign:     ts.Assign,
				TypeParams: ts.TypeParams,
				Comment:    ts.Comment,
			}

			out := &ast.GenDecl{
				Tok:   token.TYPE,
				Specs: []ast.Spec{newTS},
				Doc:   doc,
			}

			if doc != nil {
				out.TokPos = doc.End() + 1
			}

			if !first {
				buf.WriteString("\n\n")
			}
			first = false

			if err := config.Fprint(&buf, fset, out); err != nil {
				return "", err
			}
		}
	}

	return buf.String(), nil
}

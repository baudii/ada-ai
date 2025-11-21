package openapiproject

import (
	"bufio"
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

type ProjectInterfaceMethod struct {
	Description string
	Name        string
	Params      []string
	Returns     []string
}

type ProjectInterface struct {
	Description string
	Name        string
	Methods     []ProjectInterfaceMethod
}

type ProjectInterfaces struct {
	m map[string]*ProjectInterface
	s string
}

func NewProjectInterface(name, description string) *ProjectInterface {
	return &ProjectInterface{
		Name:        name,
		Description: description,
		Methods:     []ProjectInterfaceMethod{},
	}
}

func (o *OpenAPIProject) ParseProjectInterfaces() (*ProjectInterfaces, error) {
	outInterfaces := filepath.Join(o.outputDir, INTERNAL_FOLDER, HANDLERS_FOLDER, "interfaces.go")

	srcBytes, err := os.ReadFile(outInterfaces)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	interfacesContent := ClearInterfacesContent(srcBytes)

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, outInterfaces, srcBytes, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parse file: %w", err)
	}

	projectInterfaces := &ProjectInterfaces{
		m: make(map[string]*ProjectInterface),
		s: interfacesContent,
	}

	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}

		for _, spec := range genDecl.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			iface, ok := ts.Type.(*ast.InterfaceType)
			if !ok {
				continue
			}

			docCg := ts.Doc
			if docCg == nil {
				docCg = genDecl.Doc
			}
			ifaceComment := strings.TrimSpace(commentText(docCg))

			currentInterface := NewProjectInterface(ts.Name.Name, ifaceComment)
			projectInterfaces.m[currentInterface.Name] = currentInterface

			if iface.Methods == nil {
				continue
			}

			for _, field := range iface.Methods.List {
				if len(field.Names) == 0 {
					continue
				}

				ft, ok := field.Type.(*ast.FuncType)
				if !ok {
					continue
				}

				params := fieldListToStrings(fset, ft.Params)
				returns := fieldListToStrings(fset, ft.Results)

				methodComment := strings.TrimSpace(commentText(field.Doc))
				if methodComment == "" {
					methodComment = strings.TrimSpace(commentText(field.Comment))
				}

				for _, name := range field.Names {
					method := ProjectInterfaceMethod{
						Description: methodComment,
						Name:        name.Name,
						Params:      params,
						Returns:     returns,
					}
					currentInterface.Methods = append(currentInterface.Methods, method)
				}
			}
		}
	}

	return projectInterfaces, nil
}

func ClearInterfacesContent(interfaces []byte) string {
	scanner := bufio.NewScanner(strings.NewReader(string(interfaces)))
	var (
		out           strings.Builder
		inHeader      = true
		inImportBlock = false
	)

	for scanner.Scan() {
		line := scanner.Text()
		if inHeader {
			if strings.HasPrefix(line, "package") {
				continue
			}

			if strings.HasPrefix(line, "import") {
				if strings.Contains(line, "(") {
					inImportBlock = true
				}
				continue
			}

			if inImportBlock {
				if strings.Contains(line, ")") {
					inImportBlock = false
				}
				continue
			}

			inHeader = false
		}

		out.WriteString(line + "\n")
	}

	return strings.TrimSpace(out.String())
}

func commentText(cg *ast.CommentGroup) string {
	if cg == nil {
		return ""
	}
	return cg.Text()
}

func fieldListToStrings(fset *token.FileSet, fl *ast.FieldList) []string {
	if fl == nil || len(fl.List) == 0 {
		return nil
	}

	var out []string
	for _, f := range fl.List {
		typStr := nodeToString(fset, f.Type)

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

func nodeToString(fset *token.FileSet, n ast.Node) string {
	var b bytes.Buffer
	_ = printer.Fprint(&b, fset, n)
	return b.String()
}

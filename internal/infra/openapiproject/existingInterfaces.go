package openapiproject

import (
	"bufio"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"sort"
	"strings"

	"github.com/baudii/ada-ai/internal/app/codeanalyzer"
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

func NewProjectInterface(name, description string) *ProjectInterface {
	return &ProjectInterface{
		Name:        name,
		Description: description,
		Methods:     []ProjectInterfaceMethod{},
	}
}

func (o *OpenAPIProject) WriteProjectInterfaces(m map[string]*ProjectInterface) error {
	interfacesPath := o.InterfacesFilePath()
	out := strings.Builder{}

	out.WriteString(HEADER_COMMENT)
	out.WriteString("package handlers\n\n")
	out.WriteString("import \"" + o.moduleName + "/internal/models\"\n\n")
	out.WriteString(o.GenerateInterfacesContent(m))

	return formatAndWrite(out.String(), interfacesPath)
}

func (o *OpenAPIProject) GenerateInterfacesContent(m map[string]*ProjectInterface) string {
	out := &strings.Builder{}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	first := true
	for _, key := range keys {
		if !first {
			out.WriteString("\n")
		}
		first = false
		iface := m[key]
		if iface.Description != "" {
			out.WriteString("// " + iface.Description + "\n")
		}
		out.WriteString("type " + iface.Name + " interface {\n")
		for _, method := range iface.Methods {
			if method.Description != "" {
				out.WriteString("\t// " + method.Description + "\n")
			}
			out.WriteString("\t" + method.Name + "(" + strings.Join(method.Params, ", ") + ")")
			if len(method.Returns) > 0 {
				out.WriteString(" (" + strings.Join(method.Returns, ", ") + ")")
			}
			out.WriteString("\n")
		}
		out.WriteString("}\n")
	}

	return out.String()
}

func (o *OpenAPIProject) ParseProjectInterfaces() (map[string]*ProjectInterface, error) {
	outInterfaces := o.InterfacesFilePath()
	if _, err := os.Stat(outInterfaces); os.IsNotExist(err) {
		return make(map[string]*ProjectInterface), nil
	}

	srcBytes, err := os.ReadFile(outInterfaces)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, outInterfaces, srcBytes, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parse file: %w", err)
	}

	projectInterfaces := make(map[string]*ProjectInterface)
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
			ifaceComment := strings.TrimSpace(docCg.Text())

			currentInterface := NewProjectInterface(ts.Name.Name, ifaceComment)
			projectInterfaces[currentInterface.Name] = currentInterface

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

				params := codeanalyzer.FieldListToStrings(fset, ft.Params)
				returns := codeanalyzer.FieldListToStrings(fset, ft.Results)

				methodComment := strings.TrimSpace(field.Doc.Text())
				if methodComment == "" {
					methodComment = strings.TrimSpace(field.Comment.Text())
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

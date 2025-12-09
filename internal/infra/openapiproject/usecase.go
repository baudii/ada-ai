package openapiproject

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

func (o *OpenAPIProject) GenerateService() error {
	err := os.MkdirAll(o.UseCaseFolder(), 0755)
	if err != nil {
		return fmt.Errorf("create usecase folder: %w", err)
	}

	interfaces, err := o.ParseProjectInterfaces()
	if err != nil {
		return fmt.Errorf("parse project interfaces: %w", err)
	}

	for _, iface := range interfaces {
		err := o.MaterializeFromInterface(iface)
		if err != nil {
			return fmt.Errorf("materialize from interface %s: %w", iface.Name, err)
		}
	}

	return nil
}

func (o *OpenAPIProject) MaterializeFromInterface(iface *ProjectInterface) error {
	filename := fmt.Sprintf("%s.go", PascalToSnake(iface.Name))
	h, err := os.OpenFile(filepath.Join(o.UseCaseFolder(), filename), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer h.Close()

	_, err = h.WriteString(InitialUseCaseContent(iface).String())
	if err != nil {
		return fmt.Errorf("write initial usecase content: %w", err)
	}

	return nil
}

func InitialUseCaseContent(iface *ProjectInterface) *strings.Builder {
	content := &strings.Builder{}
	runes := []rune(iface.Name)
	runes[0] = unicode.ToLower(runes[0])
	structName := string(runes)
	content.WriteString(HEADER_COMMENT)
	content.WriteString("package usecase\n\n")
	content.WriteString("// " + iface.Description + "\n")
	content.WriteString("type " + structName + " struct {\n")
	content.WriteString("}\n\n")
	for _, method := range iface.Methods {
		content.WriteString("// " + method.Description + "\n")
		content.WriteString("func (" + string(runes[0]) + " " + structName + ") " + method.Name + "(" + strings.Join(method.Params, ", ") + ")")
		if len(method.Returns) > 0 {
			content.WriteString(" (" + strings.Join(method.Returns, ", ") + ")")
		}
		content.WriteString(" {\n")
		content.WriteString("}\n\n")
	}
	return content
}

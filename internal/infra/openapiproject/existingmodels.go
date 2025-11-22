package openapiproject

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/baudii/ada-ai/internal/app/codeanalyzer"
)

type ProjectModel struct {
	Name        string
	Description string
	Fields      []FieldInfo
}

// ParseExistingModels parses existing model files to extract model definitions using ast package.
func (o *OpenAPIProject) ParseExistingModels(resourceName string) (map[string]*ProjectModel, error) {
	modelFilePath := filepath.Join(o.ModelsFolder(), resourceName+".go")
	if _, err := os.Stat(modelFilePath); os.IsNotExist(err) {
		return make(map[string]*ProjectModel), nil
	}

	data, err := os.ReadFile(modelFilePath)
	if err != nil {
		return nil, err
	}

	return o.ExtractModels(modelFilePath, string(data))
}

// ExtractModels extracts model definitions from the provided Go source code.
func (o *OpenAPIProject) ExtractModels(path, data string) (map[string]*ProjectModel, error) {
	if data == "none" {
		return make(map[string]*ProjectModel), nil
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, data, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parse file: %w", err)
	}

	existingModels := make(map[string]*ProjectModel)

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

			structType, ok := ts.Type.(*ast.StructType)
			if !ok {
				continue
			}

			docCg := ts.Doc
			if docCg == nil {
				docCg = genDecl.Doc
			}

			currentModel := &ProjectModel{
				Name:        ts.Name.Name,
				Description: strings.TrimSpace(docCg.Text()),
				Fields:      []FieldInfo{},
			}

			if structType.Fields == nil {
				continue
			}

			for _, field := range structType.Fields.List {
				fieldInfo := extractFieldInfo(fset, field)
				currentModel.Fields = append(currentModel.Fields, fieldInfo...)
			}
			existingModels[currentModel.Name] = currentModel
		}
	}

	return existingModels, nil
}

func (o *OpenAPIProject) GenerateModelsContent(m map[string]*ProjectModel) string {
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
		model := m[key]
		if model.Description != "" {
			out.WriteString("// " + model.Description + "\n")
		}
		out.WriteString("type " + model.Name + " struct {\n")
		for _, field := range model.Fields {
			if field.Doc != "" {
				out.WriteString("\t// " + field.Doc + "\n")
			}
			out.WriteString("\t" + field.Name + " " + field.Type)
			if field.Tags != "" {
				out.WriteString(" " + string(field.Tags))
			}
			out.WriteString("\n")
		}
		out.WriteString("}\n")
	}

	return out.String()
}

func (o *OpenAPIProject) WriteHandlerModels(packageName string, m map[string]*ProjectModel) error {
	modelsPath := filepath.Join(o.ModelsFolder(), packageName+".go")
	out := strings.Builder{}
	out.WriteString(HEADER_COMMENT)
	out.WriteString("package models\n\n")
	out.WriteString(o.GenerateModelsContent(m))
	return formatAndWrite(out.String(), modelsPath)
}

func extractFieldInfo(fset *token.FileSet, field *ast.Field) []FieldInfo {
	typeStr := codeanalyzer.NodeToString(fset, field.Type)
	fieldComment := strings.TrimSpace(field.Doc.Text())
	if fieldComment == "" {
		fieldComment = strings.TrimSpace(field.Comment.Text())
	}
	var fields []FieldInfo
	for _, f := range field.Names {
		fields = append(fields, FieldInfo{
			Name: f.Name,
			Type: typeStr,
			Doc:  fieldComment,
			Tags: reflect.StructTag(field.Tag.Value),
		})
	}
	return fields
}

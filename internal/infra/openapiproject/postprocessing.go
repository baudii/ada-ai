package openapiproject

import (
	"bufio"
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"

	"github.com/baudii/ada-ai/internal/app/codeanalyzer"
)

var (
	interfaceRx = regexp.MustCompile(`(?s)Interface:\s*([^\n]+)\n\s*Description:\s*([^\n]+)\s*Methods:\n(.*)`)
	methodRx    = regexp.MustCompile(`\s*-\s*([^\n]+)\s*Signature:\s*([^\n]+)\s*Description:\s*([^\n]+)`)
)

type llmResponseRaw struct {
	functionBody          string
	addedFields           string
	interfaces            string
	addedModels           string
	interfacesDescription string
}

type llmResponse struct {
	functionBody           string
	addedFields            string
	interfaces             string
	newModels              map[string]*ProjectModel
	interfacesDescriptions map[string]*ProjectInterface
}

func ReadLLMResponse(body string) (*llmResponseRaw, error) {
	functionBody := strings.Builder{}
	addedFields := strings.Builder{}
	interfaces := strings.Builder{}
	addedModels := strings.Builder{}
	interfacesDescription := strings.Builder{}

	scanner := bufio.NewScanner(strings.NewReader(string(body)))
	state := 0
	for scanner.Scan() {
		if scanner.Err() != nil {
			return nil, scanner.Err()
		}
		line := scanner.Text()
		switch state {
		case 0:
			if strings.HasPrefix(line, "## function") {
				state = 1
			}
		case 1:
			if strings.HasPrefix(line, "## added_fields") {
				state = 2
			} else {
				functionBody.WriteString(line + "\n")
			}
		case 2:
			if strings.HasPrefix(line, "## interfaces") {
				state = 3
			} else {
				addedFields.WriteString(line + "\n")
			}
		case 3:
			if strings.HasPrefix(line, "## added_models") {
				state = 4
			} else {
				interfaces.WriteString(line + "\n")
			}
		case 4:
			if strings.HasPrefix(line, "## interfaces_description") {
				state = 5
			} else {
				addedModels.WriteString(line + "\n")
			}
		case 5:
			interfacesDescription.WriteString(line + "\n")
		}
	}

	if state != 5 {
		return nil, fmt.Errorf("failed to parse body: invalid structure")
	}

	return &llmResponseRaw{
		functionBody:          functionBody.String(),
		addedFields:           addedFields.String(),
		interfaces:            interfaces.String(),
		addedModels:           addedModels.String(),
		interfacesDescription: interfacesDescription.String(),
	}, nil
}

func ParseInterfaceDescriptions(ifaceDesc string) map[string]*ProjectInterface {
	descriptions := make(map[string]*ProjectInterface)
	spl := strings.SplitSeq(ifaceDesc, "```")
	for text := range spl {
		regexpMatches := interfaceRx.FindAllStringSubmatch(text, -1)
		projInterface := ProjectInterface{}
		for _, match := range regexpMatches {
			interfaceName := match[1]
			interfaceDescription := match[2]
			methodsText := match[3]
			projInterface.Name = interfaceName
			projInterface.Description = interfaceDescription
			pims := []ProjectInterfaceMethod{}
			methodMatches := methodRx.FindAllStringSubmatch(methodsText, -1)
			for _, methodMatch := range methodMatches {
				methodName := methodMatch[1]
				methodSignature := methodMatch[2]
				methodDescription := methodMatch[3]
				_, params, returns := ParseMethodSignature(methodSignature)
				pims = append(pims, ProjectInterfaceMethod{
					Name:        methodName,
					Description: methodDescription,
					Params:      params,
					Returns:     returns,
				})
			}
			projInterface.Methods = pims
			descriptions[projInterface.Name] = &projInterface
		}
	}

	return descriptions
}

func ParseMethodSignature(signature string) (name string, params []string, returns []string) {
	signature = strings.TrimSpace(signature)
	openParenIndex := strings.Index(signature, "(")
	closeParenIndex := strings.Index(signature, ")")
	if openParenIndex == -1 || closeParenIndex == -1 || closeParenIndex < openParenIndex {
		return "", nil, nil
	}
	name = strings.TrimSpace(signature[:openParenIndex])
	params = []string{}
	returns = []string{}
	paramsList := signature[openParenIndex+1 : closeParenIndex]
	returnsList := strings.TrimSpace(signature[closeParenIndex+1:])
	if len(paramsList) > 0 {
		params = strings.Split(paramsList, ",")
		for i := range params {
			params[i] = strings.TrimSpace(params[i])
		}
	}
	if len(returnsList) > 0 {
		returnsList = strings.Trim(strings.TrimSpace(returnsList), "()")
		returns = strings.Split(returnsList, ",")
		for i := range returns {
			returns[i] = strings.TrimSpace(returns[i])
		}
	}
	return
}

func GetMdBlock(body, blockName string) string {
	return strings.TrimSuffix(strings.TrimPrefix(strings.TrimSpace(body), "```"+blockName+"\n"), "```")
}

func (o *OpenAPIProject) ProcessResponse(
	ctx context.Context,
	content []byte,
	oldInterfaces map[string]*ProjectInterface,
	existingModels map[string]*ProjectModel,
) (*llmResponse, error) {
	raw, err := ReadLLMResponse(string(content))
	if err != nil {
		return nil, err
	}

	updatedInterfaces, err := UpdateInterfaces(raw, oldInterfaces)
	if err != nil {
		return nil, err
	}

	generatedModels, err := o.ExtractGeneratedModels(raw)
	if err != nil {
		return nil, err
	}

	updatedModels, err := UpdateModels(generatedModels, existingModels)
	if err != nil {
		return nil, err
	}

	return &llmResponse{
		functionBody:           GetMdBlock(raw.functionBody, "go"),
		addedFields:            GetMdBlock(raw.addedFields, "go"),
		interfaces:             GetMdBlock(raw.interfaces, "go"),
		newModels:              updatedModels,
		interfacesDescriptions: updatedInterfaces,
	}, nil
}

func ParseFieldLine(line string) (*FieldInfo, error) {
	state := 0
	section := []rune{}
	fieldInfo := &FieldInfo{}
	for _, c := range line {
		switch state {
		case 0:
			if c == ' ' {
				state = 1
				fieldInfo.Name = string(section)
				section = []rune{}
			} else {
				section = append(section, c)
			}
		case 1:
			if c == ' ' {
				state = 2
				fieldInfo.Type = string(section)
				section = []rune{}
			} else {
				section = append(section, c)
			}
		case 2:
			section = append(section, c)
			switch c {
			case '`':
				state = 3
			case '/':
				state = 4
			default:
				break
			}
		case 3:
			section = append(section, c)
			if c == '`' {
				state = 4
				fieldInfo.Tags = reflect.StructTag(string(section))
				section = []rune{}
			}
		case 4:
			section = append(section, c)
		}
	}
	if len(section) > 0 {
		if state == 1 {
			fieldInfo.Type = string(section)
		} else if state == 4 && len(section) > 0 {
			fieldInfo.Doc = string(section)
		} else {
			return nil, fmt.Errorf("failed to parse field line: %s", line)
		}
	}
	return fieldInfo, nil
}

func (o *OpenAPIProject) InsertFieldsToServer(raw *llmResponse) error {
	if strings.TrimSpace(raw.addedFields) == "none" {
		return nil
	}
	fieldsSplit := strings.SplitSeq(raw.addedFields, "\n")
	newFields := []FieldInfo{}
	for line := range fieldsSplit {
		line = strings.TrimSpace(line)
		line = strings.TrimPrefix(line, "-")
		line = strings.TrimSpace(line)
		line = strings.Join(strings.Split(line, " "), " ")
		fieldInfo, err := ParseFieldLine(line)
		if err != nil {
			return err
		}
		newFields = append(newFields, *fieldInfo)
	}

	serverPath := filepath.Join(o.HandlersFolder(), SERVER_FILE)
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, serverPath, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("parse server file: %w", err)
	}

	for _, decl := range f.Decls {
		if structDecl, ok := decl.(*ast.GenDecl); ok && structDecl.Tok == token.TYPE {
			for _, spec := range structDecl.Specs {
				if typeSpec, ok := spec.(*ast.TypeSpec); ok && typeSpec.Name.Name == "Server" {
					if structType, ok := typeSpec.Type.(*ast.StructType); ok {
						for _, newField := range newFields {
							found := false
							for _, field := range structType.Fields.List {
								if field.Names[0].Name == newField.Name {
									if newField.Type != field.Type.(*ast.Ident).Name {
										return fmt.Errorf("field %s already exists with different type", newField.Name)
									}
									field.Tag = &ast.BasicLit{
										Kind:  token.STRING,
										Value: string(newField.Tags),
									}
									field.Doc = &ast.CommentGroup{
										List: []*ast.Comment{{Text: newField.Doc}},
									}
									found = true
								}
							}
							if !found {
								structType.Fields.List = append(structType.Fields.List, &ast.Field{
									Names: []*ast.Ident{
										{Name: newField.Name},
									},
									Type: &ast.Ident{Name: newField.Type},
									Tag: &ast.BasicLit{
										Kind:  token.STRING,
										Value: string(newField.Tags),
									},
									Comment: &ast.CommentGroup{
										List: []*ast.Comment{{Text: newField.Doc}},
									},
								})
							}
						}
					}
				}
			}
		}
	}

	if err := formatAndWrite(codeanalyzer.NodeToString(fset, f), serverPath); err != nil {
		return fmt.Errorf("write updated server file: %w", err)
	}

	return nil
}

func (o *OpenAPIProject) ExtractGeneratedModels(raw *llmResponseRaw) (map[string]*ProjectModel, error) {
	modelsBlock := GetMdBlock(raw.addedModels, "go")
	if modelsBlock == "none" {
		return make(map[string]*ProjectModel), nil
	}
	goFile := "package models\n\n" + modelsBlock

	generatedModels, err := o.ExtractModels("", goFile)
	if err != nil {
		return nil, fmt.Errorf("extract models: %w", err)
	}
	return generatedModels, nil
}

func UpdateModels(
	generatedModels map[string]*ProjectModel,
	existingModels map[string]*ProjectModel,
) (map[string]*ProjectModel, error) {
	for name, genModel := range generatedModels {
		if existModel, ok := existingModels[name]; ok {
			fieldMap := make(map[string]FieldInfo)
			for _, field := range existModel.Fields {
				fieldMap[field.Name] = field
			}
			for _, genField := range genModel.Fields {
				if _, ok := fieldMap[genField.Name]; !ok {
					existModel.Fields = append(existModel.Fields, genField)
				}
			}
		} else {
			existingModels[name] = genModel
		}
	}

	return existingModels, nil
}

func UpdateInterfaces(
	raw *llmResponseRaw,
	oldInterfaces map[string]*ProjectInterface,
) (map[string]*ProjectInterface, error) {
	newInterfaces := ParseInterfaceDescriptions(raw.interfacesDescription)

	for _, new := range newInterfaces {
		if old, ok := oldInterfaces[new.Name]; ok {
			for _, newMethod := range new.Methods {
				for i, oldMethod := range old.Methods {
					if oldMethod.Name == newMethod.Name {
						if !reflect.DeepEqual(oldMethod.Params, newMethod.Params) {
							// TODO: Handle new params in the future for now return error
							return nil, fmt.Errorf("existing method %s in interface %s has different parameters", newMethod.Name, new.Name)
						}
						if !reflect.DeepEqual(oldMethod.Returns, newMethod.Returns) {
							// TODO: Handle new returns in the future for now return error
							return nil, fmt.Errorf("existing method %s in interface %s has different return values", newMethod.Name, new.Name)
						}
						old.Methods[i].Description = newMethod.Description
					} else {
						old.Methods = append(old.Methods, newMethod)
					}
				}
			}
		} else {
			oldInterfaces[new.Name] = new
		}
	}
	return oldInterfaces, nil
}

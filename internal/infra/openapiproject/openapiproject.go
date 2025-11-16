package openapiproject

import (
	"context"
	"fmt"
	"go/ast"
	"go/types"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"golang.org/x/tools/go/packages"
)

type OpenAPIProject struct {
	logger      *slog.Logger
	specPath    string
	oapiCfgPath string
	outputDir   string
}

type FieldInfo struct {
	Name string
	Type string
	Tags reflect.StructTag
	Doc  string
}

type ParamInfo struct {
	Name       string
	Type       string
	IsStruct   bool
	StructName string
	Fields     []FieldInfo
}

const (
	OAPI_FOLDER         = "oapi"
	SERVER_INTERFACE    = "ServerInterface"
	OAPI_GENERATED_FILE = "server.gen.go"
)

type option func(*OpenAPIProject)

func WithLogger(logger *slog.Logger) option {
	return func(p *OpenAPIProject) {
		p.logger = logger
	}
}

func New(outputDir string, opts ...option) *OpenAPIProject {
	specPath := filepath.Join(outputDir, OAPI_FOLDER, "openapi.json")
	oapiCfgPath := filepath.Join(outputDir, OAPI_FOLDER, "cfg.yaml")
	project := &OpenAPIProject{
		specPath:    specPath,
		outputDir:   outputDir,
		oapiCfgPath: oapiCfgPath,
	}
	for _, opt := range opts {
		opt(project)
	}
	return project
}

func (p *OpenAPIProject) Materialize(ctx context.Context) error {
	// Copy OpenAPI spec and config files to output directory

	// Run oapi-codegen to generate server code
	if err := p.runOAPICodegen(ctx); err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Join(p.outputDir, "handlers"), 0755); err != nil {
		return err
	}

	// Implement interface and create every handler from the generated file
	if err := p.ExtractServerInterface(p.MaterializeHandler); err != nil {
		return err
	}

	// Use SQLite to handle storage operations (maybe add other later?)

	return nil
}

func (p *OpenAPIProject) MaterializeHandler(method *types.Func, cg *ast.CommentGroup, pkg *packages.Package) {
	var out strings.Builder

	out.WriteString(fmt.Sprintf("// Handler for %s\n", method.Name()))

	if cg != nil {
		for _, c := range cg.List {
			out.WriteString(c.Text + "\n")
		}
	}

	params, _ := p.AnalyzeMethodParams(method, pkg)

	out.WriteString("\n// PARAMETERS:\n")
	for _, prm := range params {
		out.WriteString(fmt.Sprintf("//   %s %s\n", prm.Name, prm.Type))
	}

	for _, prm := range params {
		if prm.IsStruct {
			out.WriteString(fmt.Sprintf("\n// STRUCT %s:\n", prm.StructName))
			for _, f := range prm.Fields {
				out.WriteString(fmt.Sprintf("//   FIELD: %s (%s)\n", f.Name, f.Type))

				if tag := f.Tags.Get("json"); tag != "" {
					out.WriteString(fmt.Sprintf("//     TAG json: %s\n", tag))
				}
				if tag := f.Tags.Get("form"); tag != "" {
					out.WriteString(fmt.Sprintf("//     TAG form: %s\n", tag))
				}

				if f.Doc != "" {
					out.WriteString("//     DOC: " + f.Doc)
				}
			}
		}
	}

	file := filepath.Join(p.outputDir, "handlers", method.Name()+".go")
	_ = os.WriteFile(file, []byte(out.String()), 0644)
}

func (p *OpenAPIProject) AnalyzeMethodParams(method *types.Func, pkg *packages.Package) ([]ParamInfo, error) {
	sig := method.Type().(*types.Signature)

	params := []ParamInfo{}

	// Для удобного поиска AST TypeSpec
	typeSpecs := map[string]*ast.TypeSpec{}
	for _, f := range pkg.Syntax {
		for _, decl := range f.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}
			for _, spec := range gd.Specs {
				if ts, ok := spec.(*ast.TypeSpec); ok {
					typeSpecs[ts.Name.Name] = ts
				}
			}
		}
	}

	for i := 0; i < sig.Params().Len(); i++ {
		pv := sig.Params().At(i)
		param := ParamInfo{
			Name: pv.Name(),
			Type: pv.Type().String(),
		}

		// Проверяем, структура ли это
		if named, ok := pv.Type().(*types.Named); ok {
			if st, ok := named.Underlying().(*types.Struct); ok {
				param.IsStruct = true
				param.StructName = named.Obj().Name()

				// Находим AST TypeSpec для структуры
				if ts, ok := typeSpecs[named.Obj().Name()]; ok {
					if structNode, ok := ts.Type.(*ast.StructType); ok {
						for idx, field := range structNode.Fields.List {
							fi := FieldInfo{
								Name: field.Names[0].Name,
								Type: st.Field(idx).Type().String(),
							}

							if field.Tag != nil {
								tag := reflect.StructTag(field.Tag.Value[1 : len(field.Tag.Value)-1])
								fi.Tags = tag
							}

							if field.Doc != nil {
								for _, c := range field.Doc.List {
									fi.Doc += c.Text + "\n"
								}
							}

							param.Fields = append(param.Fields, fi)
						}
					}
				}
			}
		}

		params = append(params, param)
	}

	return params, nil
}

// ExtractServerInterface reads the generated server code and extracts method information
// from the ServerInterface.
func (p *OpenAPIProject) ExtractServerInterface(callback func(*types.Func, *ast.CommentGroup, *packages.Package)) error {
	cfg := &packages.Config{
		Mode: packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedImports,
	}

	pkgs, err := packages.Load(cfg, filepath.Join(p.outputDir, OAPI_GENERATED_FILE))
	if err != nil {
		return err
	}

	for _, pkg := range pkgs {
		scope := pkg.Types.Scope()

		obj := scope.Lookup(SERVER_INTERFACE)
		if obj == nil {
			continue
		}

		if iface, ok := obj.Type().Underlying().(*types.Interface); ok {
			for method := range iface.Methods() {
				for _, file := range pkg.Syntax {
					ast.Inspect(
						file,
						func(n ast.Node) bool {
							td, ok := n.(*ast.TypeSpec)
							if !ok || td.Name.Name != SERVER_INTERFACE {
								return true
							}

							if ifaceType, ok := td.Type.(*ast.InterfaceType); ok {
								for _, f := range ifaceType.Methods.List {
									if len(f.Names) > 0 && f.Names[0].Name == method.Name() {
										callback(method, f.Doc, pkg)
									}
								}
							}
							return false
						})
				}
			}
		}
	}

	return nil
}

func (p *OpenAPIProject) Validate(ctx context.Context) error {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromFile(p.specPath)
	if err != nil {
		return err
	}

	return doc.Validate(ctx)
}

func (p *OpenAPIProject) runOAPICodegen(ctx context.Context) error {
	command := "oapi-codegen"
	args := []string{"-config", p.oapiCfgPath, p.specPath}
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Dir = p.outputDir

	return cmd.Run()
}

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

const (
	OAPI_FOLDER         = "oapi"
	OAPI_GENERATED_FILE = "server.gen.go"
)

type option func(*OpenAPIProject)

// WithLogger sets the logger for the OpenAPIProject.
func WithLogger(logger *slog.Logger) option {
	return func(p *OpenAPIProject) {
		p.logger = logger
	}
}

// New creates a new OpenAPIProject with the given output directory and options.
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

// Materialize generates the project structure and code based on the OpenAPI specification.
func (p *OpenAPIProject) Materialize(ctx context.Context) error {
	// Copy OpenAPI spec and config files to output directory

	// Run oapi-codegen to generate server code
	if err := p.runOAPICodegen(ctx); err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Join(p.outputDir, "handlers"), 0755); err != nil {
		return err
	}
	cfg := &packages.Config{
		Mode: packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedImports,
	}

	pkgs, err := packages.Load(cfg, filepath.Join(p.outputDir, OAPI_GENERATED_FILE))
	if err != nil {
		return err
	}

	// Implement interface and create every handler from the generated file
	if err := ExtractServerInterface(pkgs, p.MaterializeHandler); err != nil {
		return err
	}

	// Use SQLite to handle storage operations (maybe add other later?)

	return nil
}

// MaterializeHandler generates a handler file for the given method of the ServerInterface.
func (p *OpenAPIProject) MaterializeHandler(method *types.Func, cg *ast.CommentGroup, pkg *packages.Package) {
	var out strings.Builder
	out.WriteString("// This file is auto-generated.\n")
	out.WriteString(fmt.Sprintf("//\n// Handler for %s method\n", method.Name()))

	if cg != nil {
		out.WriteString("//\n// DESCRIPTION:\n")
		for _, c := range cg.List {
			out.WriteString(c.Text + "\n")
		}
	}

	params, _ := AnalyzeMethodParams(method, pkg)

	out.WriteString("//\n// PARAMETERS:\n")
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

	file := filepath.Join(p.outputDir, "handlers", PascalToKebab(method.Name())+".go")
	_ = os.WriteFile(file, []byte(out.String()), 0644)
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

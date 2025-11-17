package openapiproject

import (
	"context"
	"fmt"
	"go/ast"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/baudii/ada-ai/internal/core/project"
	"github.com/getkin/kin-openapi/openapi3"
	"golang.org/x/tools/go/packages"
)

type OpenAPIProject struct {
	logger      *slog.Logger
	projectName string
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

// WithProjectName sets the project name for the OpenAPIProject.
func WithProjectName(name string) option {
	return func(p *OpenAPIProject) {
		p.projectName = name
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
func (p *OpenAPIProject) Materialize(ctx context.Context, hfile project.FileHandler, hfold project.FolderHandler) error {
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
	if err := ExtractServerInterface(pkgs, p.MaterializeHandler, hfile); err != nil {
		return err
	}

	// Use SQLite to handle storage operations (maybe add other later?)

	return nil
}

// MaterializeHandler generates a handler file for the given method of the ServerInterface.
func (p *OpenAPIProject) MaterializeHandler(
	methodInfo MethodInfo,
	comments *ast.CommentGroup,
	hfile project.FileHandler,
) {
	var out *strings.Builder = &strings.Builder{}
	WriteComments(out, comments, "", methodInfo.Params)
	// out.WriteString(`
	// package handlers

	// import (
	// 	"net/http"
	// )

	// func ` + method.Name() + `() {

	// }
	// `)
	file := filepath.Join(p.outputDir, "handlers", PascalToKebab(methodInfo.Name)+".go")
	hfile(file)
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

	cmd := exec.CommandContext(ctx, command, "-config", p.oapiCfgPath, p.specPath)
	cmd.Dir = p.outputDir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("run oapi-codegen: %w", err)
	}

	cmd = exec.CommandContext(ctx, "go", "mod", "init", p.projectName)
	cmd.Dir = p.outputDir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("initialize go module: %w", err)
	}

	return nil
}

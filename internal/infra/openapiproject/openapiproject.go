package openapiproject

import (
	"context"
	"fmt"
	"go/ast"
	"log/slog"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/baudii/ada-ai/internal/core/project"
	"github.com/getkin/kin-openapi/openapi3"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/imports"
)

type OpenAPIProject struct {
	logger      *slog.Logger
	projectName string
	specPath    string
	oapiCfgPath string
	outputDir   string
	moduleName  string
}

const (
	OAPI_GENERATED_FILE = "server.gen.go"
)

// Constants for folder names.
const (
	GENERATED_FOLDER = "generated"
	INTERNAL_FOLDER  = "internal"
	CONFIG_FOLDER    = "config"
	API_FOLDER       = "api"
	HANDLERS_FOLDER  = "handlers"
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

func WithSpec(specFile string) option {
	return func(p *OpenAPIProject) {
		p.specPath = filepath.Join(p.outputDir, API_FOLDER, specFile)
	}
}

// New creates a new OpenAPIProject with the given output directory and options.
func New(outputDir string, opts ...option) *OpenAPIProject {
	oapiCfgPath := filepath.Join(outputDir, CONFIG_FOLDER, "oapi.cfg.yaml")
	project := &OpenAPIProject{
		outputDir:   outputDir,
		oapiCfgPath: oapiCfgPath,
	}
	for _, opt := range opts {
		opt(project)
	}
	return project
}

// Materialize generates the project structure and code based on the OpenAPI specification.
func (o *OpenAPIProject) Materialize(ctx context.Context, hfile project.FileHandler, hfold project.FolderHandler) error {
	// Copy OpenAPI spec and config files to output directory

	// Run oapi-codegen to generate server code
	if err := o.runOAPICodegen(ctx); err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Join(o.outputDir, INTERNAL_FOLDER, HANDLERS_FOLDER), 0755); err != nil {
		return err
	}

	// Implement interface and create every handler from the generated file
	if err := o.MaterializeHandlers(hfile); err != nil {
		return err
	}

	// Use SQLite to handle storage operations (maybe add other DBMS later?)

	return nil
}

func FormatParameters(params []ParamInfo) []string {
	var res []string
	for _, p := range params {
		res = append(res, fmt.Sprintf("%s %s", p.Name, p.Type))
	}
	return res
}

// MaterializeHandlers reads the generated server code and extracts method information
// from the ServerInterface.
func (o *OpenAPIProject) MaterializeHandlers(hfile project.FileHandler) error {
	cfg := &packages.Config{
		Mode: packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedImports,
		Dir:  o.outputDir,
	}

	pkgs, err := packages.Load(cfg, fmt.Sprintf("%s/%s/%s", o.moduleName, INTERNAL_FOLDER, GENERATED_FOLDER))
	if err != nil {
		return fmt.Errorf("load package: %w", err)
	}

	processServerInterfaceMethods(o.logger, pkgs, hfile, o.MaterializeHandler)
	return nil
}

// MaterializeHandler generates a handler file for the given method of the ServerInterface.
func (o *OpenAPIProject) MaterializeHandler(
	methodInfo MethodInfo,
	comments *ast.CommentGroup,
	hfile project.FileHandler,
	pkg *packages.Package,
) error {
	var out *strings.Builder = &strings.Builder{}
	writeComments(out, comments, methodInfo, pkg)
	o.writeBody(out, methodInfo)

	snakeName := PascalToSnake(methodInfo.Name)

	spl := strings.Split(snakeName, "_")
	httpMethod := spl[0]
	spl = spl[1:]
	folderForFile := filepath.Join(o.outputDir, INTERNAL_FOLDER, HANDLERS_FOLDER, filepath.Join(spl...))

	if err := os.MkdirAll(folderForFile, 0755); err != nil {
		return fmt.Errorf("failed to create handler folder %s: %w", folderForFile, err)
	}

	file := filepath.Join(folderForFile, httpMethod+".go")
	formatted, err := imports.Process(file, []byte(out.String()), nil)
	if err != nil {
		o.logger.Error("file formatting finished with error", "file", file, "error", err)
	}

	err = os.WriteFile(file, formatted, 0644)
	if err != nil {
		return fmt.Errorf("failed to write generated handler %s: %w", file, err)
	}

	return hfile(file)
}

func (o *OpenAPIProject) writeBody(out *strings.Builder, methodInfo MethodInfo) {
	out.WriteString("package handlers\n\n")
	out.WriteString("import (\n")
	for alias, path := range methodInfo.Imports {
		pkgName := path[strings.LastIndex(path, "/")+1:]
		if alias == pkgName {
			fmt.Fprintf(out, "%q\n", path)
		} else {
			fmt.Fprintf(out, "%s %q\n", alias, path)
		}
	}

	out.WriteString(")\n\n")
	out.WriteString("func " + methodInfo.Name + "(")
	out.WriteString(strings.Join(FormatParameters(methodInfo.Params), ", "))
	out.WriteString(") {\n")
	out.WriteString("// WRITE YOUR CODE HERE\n")
	out.WriteString("}\n")
}

func (o *OpenAPIProject) Validate(ctx context.Context) error {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromFile(o.specPath)
	if err != nil {
		return err
	}

	return doc.Validate(ctx)
}

func (o *OpenAPIProject) runOAPICodegen(ctx context.Context) error {
	err := os.MkdirAll(filepath.Join(o.outputDir, INTERNAL_FOLDER, GENERATED_FOLDER), 0755)
	if err != nil {
		return err
	}

	command := "oapi-codegen"
	cmd := exec.CommandContext(ctx, command, "-config", o.oapiCfgPath, o.specPath)
	cmd.Dir = path.Join(o.outputDir, INTERNAL_FOLDER, GENERATED_FOLDER)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to run %q in %s: %w output: %s", cmd.String(), cmd.Dir, err, output)
	}

	cmd = exec.CommandContext(ctx, "go", "mod", "init", o.projectName)
	cmd.Dir = o.outputDir
	if output, err := cmd.CombinedOutput(); err != nil {
		if !strings.Contains(string(output), "go.mod already exists") {
			return fmt.Errorf("failed to run %q in %s: %w output: %s", cmd.String(), cmd.Dir, err, output)
		}
	}

	cmd = exec.CommandContext(ctx, "go", "mod", "tidy")
	cmd.Dir = o.outputDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to run %q in %s: %w output: %s", cmd.String(), cmd.Dir, err, output)
	}
	cmd = exec.Command("go", "list", "-m")
	cmd.Dir = o.outputDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to run %q in %s: %w output: %s", cmd.String(), cmd.Dir, err, output)
	}

	o.moduleName = strings.TrimSpace(string(output))
	return nil
}

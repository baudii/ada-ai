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

func WithSpec(specFile string) option {
	return func(p *OpenAPIProject) {
		p.specPath = filepath.Join(p.outputDir, OAPI_FOLDER, specFile)
	}
}

// New creates a new OpenAPIProject with the given output directory and options.
func New(outputDir string, opts ...option) *OpenAPIProject {
	oapiCfgPath := filepath.Join(outputDir, OAPI_FOLDER, "cfg.yaml")
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

	if err := os.MkdirAll(filepath.Join(o.outputDir, "handlers"), 0755); err != nil {
		return err
	}

	// Implement interface and create every handler from the generated file
	if err := o.MaterializeHandlers(o.MaterializeHandler, hfile); err != nil {
		return err
	}

	// Use SQLite to handle storage operations (maybe add other later?)

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
func (o *OpenAPIProject) MaterializeHandlers(
	callback func(MethodInfo, *ast.CommentGroup, project.FileHandler, *packages.Package),
	hfile project.FileHandler,
) error {
	cfg := &packages.Config{
		Mode: packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedImports,
		Dir:  o.outputDir,
	}

	pkgs, err := packages.Load(cfg, o.outputDir)
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
										methodInfo := analyzeMethodParams(method, pkg)
										callback(methodInfo, f.Doc, hfile, pkg)
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

// MaterializeHandler generates a handler file for the given method of the ServerInterface.
func (o *OpenAPIProject) MaterializeHandler(
	methodInfo MethodInfo,
	comments *ast.CommentGroup,
	hfile project.FileHandler,
	pkg *packages.Package,
) {
	var out *strings.Builder = &strings.Builder{}
	writeComments(out, comments, methodInfo, pkg)
	o.writeBody(out, methodInfo)

	file := filepath.Join(o.outputDir, "handlers", PascalToKebab(methodInfo.Name)+".go")
	opts := &imports.Options{
		Comments:  true,
		TabIndent: true,
		TabWidth:  8,
	}

	formatted, err := imports.Process(file, []byte(out.String()), opts)
	if err != nil {
		o.logger.Error("failed to format generated handler", "file", file, "error", err)
	}

	_ = os.WriteFile(file, formatted, 0644)
	hfile(file)
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
	command := "oapi-codegen"
	cmd := exec.CommandContext(ctx, command, "-config", o.oapiCfgPath, o.specPath)
	cmd.Dir = o.outputDir
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
	return nil
}

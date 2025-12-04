package openapiproject

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/baudii/ada-ai/internal/app"
	"github.com/baudii/ada-ai/internal/infra/folders"
	"github.com/baudii/ada-ai/internal/infra/fsutils"
	"github.com/getkin/kin-openapi/openapi3"
	"golang.org/x/tools/imports"
)

type OpenAPIProject struct {
	logger      *slog.Logger
	app         *app.App
	spec        *openapi3.T
	projectName string
	specPath    string
	outputDir   string
	moduleName  string
}

const filler = "handler-generator"

var (
	commentGroupRegex = regexp.MustCompile(`.*\(([^()\r\n]*)\)`)
	whiteSpaceRegex   = regexp.MustCompile(`[\t\s]+`)
	urlParamRegex     = regexp.MustCompile(`\{([^}]+)\}`)
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

// WithSpec sets the OpenAPI specification file path for the OpenAPIProject.
func WithSpec(specFile string) option {
	return func(p *OpenAPIProject) {
		p.specPath = filepath.Join(p.outputDir, SPEC_FOLDER, specFile)
	}
}

// WithApp sets the application context for the OpenAPIProject.
func WithApp(a app.App) option {
	return func(p *OpenAPIProject) {
		p.app = &a
	}
}

// New creates a new OpenAPIProject with the given output directory and options.
func New(outputDir string, opts ...option) *OpenAPIProject {
	project := &OpenAPIProject{
		outputDir: outputDir,
	}
	for _, opt := range opts {
		opt(project)
	}
	return project
}

// Materialize generates the project structure and code based on the OpenAPI specification.
func (o *OpenAPIProject) Materialize(ctx context.Context, openapi string) error {
	// Copy OpenAPI spec and config files to output directory
	if err := o.MaterializeSpec(ctx, openapi); err != nil {
		return err
	}

	// Run oapi-codegen to generate server code
	if err := o.runOAPICodegen(ctx); err != nil {
		return err
	}

	// Implement interface and create every handler from the generated file
	if err := o.MaterializeHandlers(ctx); err != nil {
		return err
	}

	// Use SQLite to handle storage operations (maybe add other DBMS later?)
	return nil
}

// MaterializeSpec writes the OpenAPI specification to the spec folder and validates it.
func (o *OpenAPIProject) MaterializeSpec(ctx context.Context, openapi string) error {
	json := strings.HasPrefix(strings.TrimSpace(openapi), "{")
	specFile := "openapi.yaml"
	if json {
		specFile = "openapi.json"
	}

	specPath := filepath.Join(o.SpecFolder(), specFile)
	if err := os.WriteFile(specPath, []byte(openapi), 0644); err != nil {
		return err
	}

	o.specPath = filepath.Join(o.SpecFolder(), specFile)
	return o.Validate(ctx, openapi)
}

// PrepareOutputDir creates the necessary directories and copies config files.
func (o *OpenAPIProject) PrepareOutputDir() error {
	if err := os.MkdirAll(o.SpecFolder(), 0755); err != nil {
		return err
	}

	outputConfigsFolder := o.ConfigsFolder()
	if err := os.MkdirAll(outputConfigsFolder, 0755); err != nil {
		return err
	}

	cfgPath := o.OAPIConfigFilePath()
	if err := fsutils.CopyFile(filepath.Join(folders.Config, OAPI_CFG_FILE), cfgPath); err != nil {
		return err
	}

	if err := os.MkdirAll(o.HandlersFolder(), 0755); err != nil {
		return err
	}

	if err := os.MkdirAll(o.ModelsFolder(), 0755); err != nil {
		return err
	}

	return nil
}

func (o *OpenAPIProject) Validate(ctx context.Context, openapi string) error {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(openapi))
	if err != nil {
		return err
	}

	if err := doc.Validate(ctx); err != nil {
		return err
	}

	o.spec = doc
	return nil
}

func FormatParameters(params []ParamInfo) []string {
	var res []string
	for _, p := range params {
		res = append(res, fmt.Sprintf("%s %s", p.Name, p.Type))
	}

	return res
}

func (o *OpenAPIProject) runOAPICodegen(ctx context.Context) error {
	err := os.MkdirAll(o.APIFolder(), 0755)
	if err != nil {
		return err
	}

	command := "oapi-codegen"
	cmd := exec.CommandContext(ctx, command, "-config", o.OAPIConfigFilePath(), o.specPath)
	cmd.Dir = o.APIFolder()
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

func (o *OpenAPIProject) EnsureWorking() error {
	cmd := exec.Command("go", "build", "./...")
	cmd.Dir = o.outputDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("build error: %w output: %s", err, string(output))
	}

	cmd = exec.Command("go", "mod", "tidy")
	cmd.Dir = o.outputDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("mod tidy error: %w output: %s", err, string(output))
	}
	return nil
}

func formatAndWrite(content string, path string) error {
	options := &imports.Options{
		Comments:   true,
		TabIndent:  true,
		TabWidth:   8,
		FormatOnly: false,
	}
	formatted, err := imports.Process(path, []byte(content), options)
	if err != nil {
		return fmt.Errorf("format file %s: %w", path, err)
	}

	err = os.WriteFile(path, formatted, 0o644)
	if err != nil {
		return err
	}
	return nil
}

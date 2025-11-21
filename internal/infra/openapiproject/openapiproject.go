package openapiproject

import (
	"bufio"
	"context"
	"fmt"
	"go/ast"
	"log/slog"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/baudii/ada-ai/internal/app"
	"github.com/baudii/ada-ai/internal/infra/folders"
	"github.com/baudii/ada-ai/internal/infra/fsutils"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/mohae/deepcopy"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/imports"
)

type OpenAPIProject struct {
	logger      *slog.Logger
	app         *app.App
	spec        *openapi3.T
	projectName string
	specPath    string
	oapiCfgPath string
	outputDir   string
	moduleName  string
}

const (
	filler = "handler-generator"
)

const (
	OAPI_GENERATED_FILE = "server.gen.go"
	OAPI_CFG_FILE       = "oapi.cfg.yaml"
)

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

func WithSpec(specFile string) option {
	return func(p *OpenAPIProject) {
		p.specPath = filepath.Join(p.outputDir, SPEC_FOLDER, specFile)
	}
}

func WithApp(a app.App) option {
	return func(p *OpenAPIProject) {
		p.app = &a
	}
}

// New creates a new OpenAPIProject with the given output directory and options.
func New(outputDir string, opts ...option) *OpenAPIProject {
	oapiCfgPath := filepath.Join(outputDir, CONFIGS_FOLDER, OAPI_CFG_FILE)
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

func (o *OpenAPIProject) MaterializeSpec(ctx context.Context, openapi string) error {
	json := strings.HasPrefix(strings.TrimSpace(openapi), "{")
	specFile := "openapi.yaml"
	if json {
		specFile = "openapi.json"
	}

	specPath := filepath.Join(o.outputDir, SPEC_FOLDER, specFile)
	if err := os.WriteFile(specPath, []byte(openapi), 0644); err != nil {
		return err
	}

	o.specPath = filepath.Join(o.outputDir, SPEC_FOLDER, specFile)
	return o.Validate(ctx, openapi)
}

func (o *OpenAPIProject) PrepareOutputDir() error {
	if err := os.MkdirAll(filepath.Join(o.outputDir, SPEC_FOLDER), 0755); err != nil {
		return err
	}

	outputConfigsFolder := filepath.Join(o.outputDir, CONFIGS_FOLDER)
	if err := os.MkdirAll(outputConfigsFolder, 0755); err != nil {
		return err
	}

	cfgPath := filepath.Join(outputConfigsFolder, OAPI_CFG_FILE)
	if err := fsutils.CopyFile(filepath.Join(folders.Config, OAPI_CFG_FILE), cfgPath); err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Join(o.outputDir, INTERNAL_FOLDER, HANDLERS_FOLDER), 0755); err != nil {
		return err
	}

	o.oapiCfgPath = cfgPath
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

// MaterializeHandlers reads the generated server code and extracts method information
// from the ServerInterface.
func (o *OpenAPIProject) MaterializeHandlers(ctx context.Context) error {
	cfg := &packages.Config{
		Mode: packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedImports,
		Dir:  o.outputDir,
	}

	pkgs, err := packages.Load(cfg, fmt.Sprintf("%s/%s/%s", o.moduleName, INTERNAL_FOLDER, API_FOLDER))
	if err != nil {
		return fmt.Errorf("load package: %w", err)
	}

	if err := o.createServerFile(); err != nil {
		return err
	}

	processServerInterfaceMethods(ctx, o.logger, pkgs, o.MaterializeHandler)
	return nil
}

// MaterializeHandler generates a handler file for the given method of the ServerInterface.
func (o *OpenAPIProject) MaterializeHandler(
	ctx context.Context,
	methodInfo MethodInfo,
	comments *ast.CommentGroup,
	pkg *packages.Package,
) error {
	var out *strings.Builder = &strings.Builder{}
	fileName, err := o.ParseComments(comments)
	if err != nil {
		return err
	}
	writeComments(out, comments, methodInfo, pkg)
	o.writeBody(out, methodInfo, "handlers")
	err = o.createHandler(out, fileName)
	if err != nil {
		return fmt.Errorf("create handler: %w", err)
	}

	retryCount := 0
	interfaces, err := o.ParseProjectInterfaces()
	if err != nil {
		return fmt.Errorf("parse project interfaces: %w", err)
	}
	reserveCopy := deepcopy.Copy(interfaces.m).(map[string]*ProjectInterface)

	for retryCount < 3 {
		response, err := o.app.SendInstructions(ctx, filler, []any{out.String(), interfaces.s})
		if err != nil {
			return fmt.Errorf("send instructions: %w", err)
		}
		o.logger.Debug(string(response))
		llmResponseObj, err := o.ProcessResponse(ctx, response, interfaces.m)
		if err != nil {
			retryCount++
			o.logger.Warn("retrying to create handler due to error during response processing", "error", err, "retryCount", retryCount)
			continue
		}

		if err = o.WriteProjectInterfaces(llmResponseObj.interfacesDescriptions); err != nil {
			return fmt.Errorf("write project interfaces: %w", err)
		}

		outWithFunction := o.insertFunction(*out, llmResponseObj.functionBody)
		err = o.createHandler(&outWithFunction, fileName)
		if err != nil {
			if err := o.WriteProjectInterfaces(reserveCopy); err != nil {
				return fmt.Errorf("restore project interfaces: %w", err)
			}
			retryCount++
			o.logger.Warn("retrying to create handler due to error during handler creation", "error", err, "retryCount", retryCount)
			continue
		}

		o.logger.Info("handler created successfully", "method", methodInfo.Name)
		break
	}

	return err
}

func (o *OpenAPIProject) insertFunction(out strings.Builder, functionContent string) strings.Builder {
	copy := strings.Builder{}
	reader := strings.NewReader(out.String())
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "// WRITE YOUR CODE HERE") {
			copy.WriteString(functionContent)
		} else {
			copy.WriteString(line + "\n")
		}
	}
	return copy
}

func (o *OpenAPIProject) ParseComments(comments *ast.CommentGroup) (string, error) {
	var fileName string

	for _, p := range comments.List {
		matches := commentGroupRegex.FindStringSubmatch(p.Text)
		if len(matches) > 1 {
			matches = whiteSpaceRegex.Split(matches[1], -1)
			if len(matches) > 0 {
				method := matches[0]
				url := matches[1]
				resources := strings.Split(url, "/")
				if len(resources) < 1 {
					return "", fmt.Errorf("failed to process URL: %s", url)
				}

				lastParam := urlParamRegex.FindStringSubmatch(resources[len(resources)-1])
				if lastParam != nil {
					out := []string{}
					for i := len(resources) - 1; i >= 0; i-- {
						param := urlParamRegex.FindStringSubmatch(resources[i])
						if len(param) > 0 {
							out = append(out, "_by_"+param[1])
						}
						if !urlParamRegex.MatchString(resources[i]) {
							out = append(out, resources[i])
							break
						}
					}
					res := strings.Builder{}
					for i := len(out) - 1; i >= 0; i-- {
						res.WriteString(out[i])
					}
					fileName = strings.ToLower(method + "_" + res.String())
				} else {
					fileName = strings.ToLower(method + "_" + resources[len(resources)-1])
				}
			}
		}
	}
	return fileName, nil
}

func (o *OpenAPIProject) createHandler(out *strings.Builder, fileName string) error {

	folderForFile := filepath.Join(o.outputDir, INTERNAL_FOLDER, HANDLERS_FOLDER)

	if err := os.MkdirAll(folderForFile, 0755); err != nil {
		return fmt.Errorf("failed to create handler folder %s: %w", folderForFile, err)
	}

	path := filepath.Join(folderForFile, fileName+".go")

	return formatAndWrite(out, path)
}

func (o *OpenAPIProject) writeBody(out *strings.Builder, methodInfo MethodInfo, packageName string) {
	out.WriteString("package " + packageName + "\n\n")
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
	out.WriteString("func (s *Server) " + methodInfo.Name + "(")
	out.WriteString(strings.Join(FormatParameters(methodInfo.Params), ", "))
	out.WriteString(") {\n")
	out.WriteString("// WRITE YOUR CODE HERE\n")
	out.WriteString("}\n")
}

func (o *OpenAPIProject) createServerFile() error {
	path := o.ServerFilePath()
	out := &strings.Builder{}
	out.WriteString(HEADER_COMMENT)
	out.WriteString(`package handlers

type option func(*Server)

// NewServer creates a new Server instance with the provided options.
func NewServer(opts ...option) *Server {
	s := &Server{}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Server represents the server handling API requests.
type Server struct{
}`)

	return formatAndWrite(out, path)
}

func (o *OpenAPIProject) runOAPICodegen(ctx context.Context) error {
	err := os.MkdirAll(filepath.Join(o.outputDir, INTERNAL_FOLDER, API_FOLDER), 0755)
	if err != nil {
		return err
	}

	command := "oapi-codegen"
	cmd := exec.CommandContext(ctx, command, "-config", o.oapiCfgPath, o.specPath)
	cmd.Dir = path.Join(o.outputDir, INTERNAL_FOLDER, API_FOLDER)
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

func formatAndWrite(out *strings.Builder, path string) error {
	content := out.String()
	formatted, err := imports.Process(path, []byte(content), nil)
	if err != nil {
		return fmt.Errorf("format file %s: %w", path, err)
	}

	err = os.WriteFile(path, formatted, 0o644)
	if err != nil {
		return err
	}
	return nil
}

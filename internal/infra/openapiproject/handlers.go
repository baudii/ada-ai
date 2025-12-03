package openapiproject

import (
	"bufio"
	"context"
	"fmt"
	"go/ast"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/baudii/ada-ai/internal/core/aigen"
	"golang.org/x/tools/go/packages"
)

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
	var out strings.Builder
	fileName, mainResource, err := o.ParseComments(comments)
	if err != nil {
		return err
	}
	writeComments(&out, comments, methodInfo, pkg)
	o.writeBody(&out, methodInfo, "handlers")
	err = o.createHandler(&out, fileName)
	if err != nil {
		return fmt.Errorf("create handler: %w", err)
	}

	retryCount := 0
	errors := make(map[string]string, 0)
	for retryCount < 3 {
		interfaces, err := o.ParseProjectInterfaces()
		if err != nil {
			return fmt.Errorf("parse project interfaces: %w", err)
		}

		existingModels, err := o.ParseExistingModels(mainResource)
		if err != nil {
			return fmt.Errorf("parse existing models: %w", err)
		}

		extractedModels, err := Extract[*ast.StructType](filepath.Join(o.ModelsFolder(), mainResource+".go"), "")
		if err != nil {
			return fmt.Errorf("extract models: %w", err)
		}
		p := o.InterfacesFilePath()
		extractedInterfaces, err := Extract[*ast.InterfaceType](p, "")
		if err != nil {
			return fmt.Errorf("extract interfaces: %w", err)
		}

		serverStruct, err := Extract[*ast.StructType](filepath.Join(o.HandlersFolder(), "server.go"), "Server")
		if err != nil {
			return fmt.Errorf("extract server struct: %w", err)
		}

		unifiedContext := fmt.Sprintf(
			"// interfaces.go\n%v\n// models.go\n%v\n// server.go\n%v\n",
			extractedInterfaces, extractedModels, serverStruct)

		errorCtx := ""
		if len(errors) > 0 {
			joined := []string{}
			for _, v := range errors {
				joined = append(joined, v)
			}
			errorCtx = "\n\nIMPORTANT: THIS IS ATTEMPT #" + strconv.Itoa(retryCount) + ". CONSIDER PREVIOUS ERRORS CAREFULLY:\n" + strings.Join(joined, "\n") + "\n"
		}

		response, err := o.app.SendInstructions(ctx, filler, []any{
			out.String(),
			unifiedContext,
			errorCtx,
		}, aigen.WithTemperature(0.2))
		if err != nil {
			return fmt.Errorf("send instructions: %w", err)
		}
		o.logger.Debug(string(response))
		llmResponseObj, err := o.ProcessResponse(ctx, response, interfaces.m, existingModels)
		if err != nil {
			errors["process_response"] = fmt.Sprintf("- process response error: %v", err)
			retryCount++
			o.logger.Warn("failed to process response. retrying...", "error", err, "retryCount", retryCount)
			continue
		}
		delete(errors, "process_response")

		if err = o.InsertFieldsToServer(llmResponseObj); err != nil {
			errors["insert_fields_to_server"] = fmt.Sprintf("- insert fields to server error: %v", err)
			retryCount++
			o.logger.Warn("failed to insert new fields to server. retrying...", "error", err, "retryCount", retryCount)
			continue
		}
		delete(errors, "insert_fields_to_server")

		if err = o.WriteHandlerModels(mainResource, llmResponseObj.newModels); err != nil {
			return fmt.Errorf("write project models: %w", err)
		}

		if err = o.WriteProjectInterfaces(llmResponseObj.interfacesDescriptions); err != nil {
			return fmt.Errorf("write project interfaces: %w", err)
		}

		outWithFunction := o.insertFunctionAndModelImport(out.String(), llmResponseObj.functionBody)
		err = o.createHandler(&outWithFunction, fileName)
		if err != nil {
			errors["create_handler"] = fmt.Sprintf("- create handler error: %v", err)
			retryCount++
			o.logger.Warn("failed to create handler. retrying...", "error", err, "retryCount", retryCount)
			continue
		}
		delete(errors, "create_handler")

		cmd := exec.Command("go", "build", "./...")
		cmd.Dir = o.outputDir
		if output, err := cmd.CombinedOutput(); err != nil {
			errors["build"] = fmt.Sprintf("- build error: %v, output: %s", err, string(output))
			retryCount++
			o.logger.Warn("failed to build project. retrying...", "error", err, "output", string(output), "retryCount", retryCount)
			continue
		}

		cmd = exec.Command("go", "mod", "tidy")
		cmd.Dir = o.outputDir
		if output, err := cmd.CombinedOutput(); err != nil {
			errors["mod_tidy"] = fmt.Sprintf("- mod tidy error: %v, output: %s", err, string(output))
			retryCount++
			o.logger.Warn("failed to run go mod tidy. retrying...", "error", err, "output", string(output), "retryCount", retryCount)
			continue
		}

		break
	}

	o.logger.Info("handler created successfully", "method", methodInfo.Name)
	return nil
}

func (o *OpenAPIProject) insertFunctionAndModelImport(content string, functionContent string) strings.Builder {
	copy := strings.Builder{}
	reader := strings.NewReader(content)
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "// WRITE YOUR CODE HERE") {
			copy.WriteString(functionContent)
		} else if strings.Contains(line, "package handlers") {
			copy.WriteString(line + "\n\n")
			copy.WriteString("import \"" + o.moduleName + "/internal/models\"\n\n")
		} else {
			copy.WriteString(line + "\n")
		}
	}
	return copy
}

func (o *OpenAPIProject) createHandler(out *strings.Builder, fileName string) error {
	folderForFile := o.HandlersFolder()

	if err := os.MkdirAll(folderForFile, 0755); err != nil {
		return fmt.Errorf("failed to create handler folder %s: %w", folderForFile, err)
	}

	path := filepath.Join(folderForFile, fileName+".go")

	return formatAndWrite(out.String(), path)
}

func (o *OpenAPIProject) createServerFile() error {
	path := o.ServerFilePath()
	out := strings.Builder{}
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

	return formatAndWrite(out.String(), path)
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

func (o *OpenAPIProject) ParseComments(comments *ast.CommentGroup) (string, string, error) {
	var fileName string
	var mainResource string

	for _, p := range comments.List {
		matches := commentGroupRegex.FindStringSubmatch(p.Text)
		if len(matches) > 1 {
			matches = whiteSpaceRegex.Split(matches[1], -1)
			if len(matches) > 0 {
				method := matches[0]
				url := matches[1]
				resources := strings.Split(url, "/")
				if len(resources) < 1 {
					return "", "", fmt.Errorf("failed to process URL: %s", url)
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
							mainResource = resources[i]
							break
						}
					}
					if mainResource == "" {
						return "", "", fmt.Errorf("failed to determine main resource from URL: %s", url)
					}
					res := strings.Builder{}
					for i := len(out) - 1; i >= 0; i-- {
						res.WriteString(out[i])
					}
					fileName = strings.ToLower(method + "_" + res.String())
				} else {
					mainResource = resources[len(resources)-1]
					fileName = strings.ToLower(method + "_" + resources[len(resources)-1])
				}
			}
		}
	}
	return fileName, mainResource, nil
}

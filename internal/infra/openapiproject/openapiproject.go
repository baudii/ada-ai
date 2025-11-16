package openapiproject

import (
	"context"
	"os/exec"
	"path/filepath"

	"github.com/getkin/kin-openapi/openapi3"
)

type OpenAPIProject struct {
	specPath    string
	oapiCfgPath string
	outputDir   string
}

const OAPI_FOLDER = "oapi"

func New(outputDir string) *OpenAPIProject {
	specPath := filepath.Join(outputDir, OAPI_FOLDER, "openapi.json")
	oapiCfgPath := filepath.Join(outputDir, OAPI_FOLDER, "cfg.yaml")
	return &OpenAPIProject{
		specPath:    specPath,
		outputDir:   outputDir,
		oapiCfgPath: oapiCfgPath,
	}
}

func (p *OpenAPIProject) Materialize(ctx context.Context) error {
	if err := p.runOapiCodegen(ctx); err != nil {
		return err
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

func (p *OpenAPIProject) runOapiCodegen(ctx context.Context) error {
	command := "oapi-codegen"
	args := []string{"-config", p.oapiCfgPath, p.specPath}
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Dir = p.outputDir

	return cmd.Run()
}

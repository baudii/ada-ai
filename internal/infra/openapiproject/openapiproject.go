package openapiproject

import (
	"context"
	"os/exec"

	"github.com/getkin/kin-openapi/openapi3"
)

type OpenAPIProject struct {
	specPath    string
	oapiCfgPath string
	outputDir   string
}

func New(specPath, oapiCfgPath, outputDir string) *OpenAPIProject {
	return &OpenAPIProject{
		specPath:    specPath,
		outputDir:   outputDir,
		oapiCfgPath: oapiCfgPath,
	}
}

func (p *OpenAPIProject) GenerateProject() error {
	if err := p.runOapiCodegen(); err != nil {
		return err
	}

	return nil
}

func (p *OpenAPIProject) ValidateSpec(ctx context.Context) error {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromFile(p.specPath)
	if err != nil {
		return err
	}

	return doc.Validate(ctx)
}

func (p *OpenAPIProject) runOapiCodegen() error {
	command := "oapi-codegen"
	args := []string{"-config", p.oapiCfgPath, p.specPath}
	cmd := exec.Command(command, args...)
	cmd.Dir = p.outputDir

	return cmd.Run()
}

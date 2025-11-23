package openapiproject

import "path/filepath"

// Constants for folder names.
const (
	API_FOLDER      = "api"
	INTERNAL_FOLDER = "internal"
	CONFIGS_FOLDER  = "configs"
	SPEC_FOLDER     = "spec"
	HANDLERS_FOLDER = "handlers"
	MODELS_FOLDER   = "models"
)

const (
	OAPI_GENERATED_FILE = "server.gen.go"
	OAPI_CFG_FILE       = "oapi.cfg.yaml"
	SERVER_FILE         = "server.go"
	INTERFACES_FILE     = "interfaces.go"
)

func (o *OpenAPIProject) InterfacesFilePath() string {
	return filepath.Join(o.HandlersFolder(), INTERFACES_FILE)
}

func (o *OpenAPIProject) ServerFilePath() string {
	return filepath.Join(o.HandlersFolder(), SERVER_FILE)
}

func (o *OpenAPIProject) OAPIConfigFilePath() string {
	return filepath.Join(o.ConfigsFolder(), OAPI_CFG_FILE)
}

func (o *OpenAPIProject) ConfigsFolder() string {
	return filepath.Join(o.outputDir, CONFIGS_FOLDER)
}

func (o *OpenAPIProject) HandlersFolder() string {
	return filepath.Join(o.outputDir, INTERNAL_FOLDER, HANDLERS_FOLDER)
}

func (o *OpenAPIProject) ModelsFolder() string {
	return filepath.Join(o.outputDir, INTERNAL_FOLDER, MODELS_FOLDER)
}

func (o *OpenAPIProject) APIFolder() string {
	return filepath.Join(o.outputDir, INTERNAL_FOLDER, API_FOLDER)
}

func (o *OpenAPIProject) SpecFolder() string {
	return filepath.Join(o.outputDir, SPEC_FOLDER)
}

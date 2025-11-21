package openapiproject

import "path/filepath"

// Constants for folder names.
const (
	GENERATED_FOLDER = "api"
	INTERNAL_FOLDER  = "internal"
	CONFIGS_FOLDER   = "configs"
	API_FOLDER       = "api"
	HANDLERS_FOLDER  = "handlers"
)

func (o *OpenAPIProject) InterfacesFilePath() string {
	return filepath.Join(o.outputDir, INTERNAL_FOLDER, HANDLERS_FOLDER, "interfaces.go")
}

func (o *OpenAPIProject) ServerFilePath() string {
	return filepath.Join(o.outputDir, INTERNAL_FOLDER, HANDLERS_FOLDER, "server.go")
}

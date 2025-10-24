package folders

import (
	"os"
	"path/filepath"
)

// Define the path to the named folders. Used for centrilized access to the files that contain data.
var (
	Artifacts = "artifacts"

	Projects string
	Config   string
	AiConfig string
	Prompts  string
)

const ProjectStructurePrompt = "project-structure-template.txt"

func init() {
	Artifacts = FromExecutable("", os.Executable)

	Projects = filepath.Join(Artifacts, ".projects")
	Prompts = filepath.Join(Artifacts, "prompts")

	Config = filepath.Join(Artifacts, "configs")
	AiConfig = filepath.Join(Config, "ai")
}

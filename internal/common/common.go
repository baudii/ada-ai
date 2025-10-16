package common

import (
	"os"
	"path/filepath"

	"github.com/baudii/ada-ai/pkg/utils"
)

// Define the path to the named folders. Used for centrilized access to the files that contain data.
var (
	Artifacts = "artifacts"

	ProjectsPath = filepath.Join(Artifacts, ".projects")

	DataPath      = filepath.Join(Artifacts, "data")
	ConfigPath    = filepath.Join(DataPath, "configuration")
	DebuggingPath = filepath.Join(DataPath, "debugging")
	PromptsPath   = filepath.Join(DataPath, "prompts")
)

const ProjectStructurePrompt = "project-structure-template.txt"

func init() {
	Artifacts = utils.AbsolutePath("", os.Executable)
	ProjectsPath = filepath.Join(Artifacts, ".projects")

	DataPath = filepath.Join(Artifacts, "data")
	ConfigPath = filepath.Join(DataPath, "configuration")
	DebuggingPath = filepath.Join(DataPath, "debugging")
	PromptsPath = filepath.Join(DataPath, "prompts")
}

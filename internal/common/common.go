package common

import (
	"os"
	"path/filepath"

	"github.com/baudii/ada-ai/pkg/pathx"
)

// Define the path to the named folders. Used for centrilized access to the files that contain data.
var (
	Artifacts = "artifacts"

	ProjectsPath  string
	DataPath      string
	ConfigPath    string
	AiConfigPath  string
	DebuggingPath string
	PromptsPath   string
)

const ProjectStructurePrompt = "project-structure-template.txt"

func init() {
	Artifacts = pathx.FromExecutable("", os.Executable)

	ProjectsPath = filepath.Join(Artifacts, ".projects")
	DataPath = filepath.Join(Artifacts, "data")
	ConfigPath = filepath.Join(DataPath, "configuration")
	AiConfigPath = filepath.Join(ConfigPath, "ai")
	DebuggingPath = filepath.Join(DataPath, "debugging")
	PromptsPath = filepath.Join(DataPath, "prompts")
}

package folders

import (
	"os"
	"path/filepath"

	"github.com/baudii/ada-ai/pkg/pathx"
)

// Define the path to the named folders. Used for centrilized access to the files that contain data.
var (
	Artifacts = "artifacts"

	Projects  string
	Data      string
	Config    string
	AiConfig  string
	Debugging string
	Prompts   string
)

const ProjectStructurePrompt = "project-structure-template.txt"

func init() {
	Artifacts = pathx.FromExecutable("", os.Executable)

	Projects = filepath.Join(Artifacts, ".projects")
	Data = filepath.Join(Artifacts, "data")
	Config = filepath.Join(Data, "configuration")
	AiConfig = filepath.Join(Config, "ai")
	Debugging = filepath.Join(Data, "debugging")
	Prompts = filepath.Join(Data, "prompts")
}

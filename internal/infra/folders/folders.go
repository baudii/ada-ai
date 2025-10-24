package folders

import (
    "os"
    "path/filepath"
)

var (
    // Artifacts is the root folder that contains app-generated artifacts.
    Artifacts = "artifacts"

    // Projects is the folder that stores user project data under Artifacts.
    Projects string
    // Config is the folder that stores configuration files under Artifacts.
    Config   string
    // AiConfig is the folder that stores AI provider configuration under Config.
    AiConfig string
    // Prompts is the folder that stores prompt templates under Artifacts.
    Prompts  string
)

// ProjectStructurePrompt is the filename of the project structure prompt template.
const ProjectStructurePrompt = "project-structure-template.txt"

func init() {
    Artifacts = FromExecutable("", os.Executable)

    Projects = filepath.Join(Artifacts, ".projects")
    Prompts = filepath.Join(Artifacts, "prompts")

    Config = filepath.Join(Artifacts, "configs")
    AiConfig = filepath.Join(Config, "ai")
}

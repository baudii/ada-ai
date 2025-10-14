package cli

import (
	"log/slog"
	"path/filepath"

	"github.com/baudii/ada-ai/internal/adacore"
	"github.com/baudii/ada-ai/internal/app"
	"github.com/baudii/ada-ai/internal/common"
	"github.com/baudii/ada-ai/pkg/utils"
)

type cliApp struct{}

// New creates a new instance of the CLI application that implements the Runner interface.
func New() app.Runner {
	return cliApp{}
}

// Projdata retrieves project data from a JSON file or prompts the user for input
// if the file does not exist or cannot be parsed. It sends the project data
// through the provided channel and closes the channel when done.
func (cliApp) Projdata(c chan adacore.ProjectData) {
	defer close(c)
	path := filepath.Join(common.DataPath, "user_data.json")
	projectData, err := utils.ParseJSONFile[adacore.ProjectData](path)
	if err != nil {
		username := utils.ReadInput("Provide nickname")
		projname := utils.ReadInput("Provide project name")
		plang := utils.ReadInput("Provide programming language (go, python, js, etc)")
		summary := utils.ReadInput("Provide a short summary of the project")
		projectData = &adacore.ProjectData{
			UserName: username,
			ProjName: projname,
			Language: plang,
			Summary:  summary,
		}

		err = utils.SaveJSONToFile(projectData, path)
		if err != nil {
			slog.Error("failed to save project data", "error", err)
		}
	}

	c <- *projectData
}

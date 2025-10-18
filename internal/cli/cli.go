package cli

import (
	"log/slog"
	"path/filepath"

	"github.com/baudii/ada-ai/internal/adacore"
	"github.com/baudii/ada-ai/internal/app"
	"github.com/baudii/ada-ai/internal/common"
	"github.com/baudii/ada-ai/pkg/utils"
)

type cliApp struct {
	read func(string) string
}

// New creates a new instance of the CLI application that implements the Runner interface.
func New() app.Runner {
	return cliApp{utils.ReadInput}
}

// Projdata retrieves project data from a JSON file or prompts the user for input
// if the file does not exist or cannot be parsed. It sends the project data
// through the provided channel and closes the channel when done.
func (cli cliApp) Projdata(ch chan adacore.ProjectData) {
	path := filepath.Join(common.DataPath, "user_data.json")
	defer close(ch)
	projectData, err := utils.ParseJSONFile[adacore.ProjectData](path)
	if err != nil {
		username := cli.read("Provide nickname")
		projname := cli.read("Provide project name")
		plang := cli.read("Provide programming language (go, python, js, etc)")
		summary := cli.read("Provide a short summary of the project")
		projectData = &adacore.ProjectData{
			UserName: username,
			ProjName: projname,
			Language: plang,
			Summary:  summary,
		}

	}

	err = utils.SaveJSONToFile(projectData, path)
	if err != nil {
		slog.Error("failed to save project data to file", "path", path, "error", err)
	}

	ch <- *projectData
}

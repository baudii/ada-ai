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

func New() app.Runner {
	return cliApp{}
}

func (cliApp) Options(c chan *adacore.ProjectData) {
	defer close(c)
	path := filepath.Join(common.DataPath, "user_data.json")
	projectData, err := utils.ParseJSONFile[adacore.ProjectData](path)
	if err != nil {
		username := utils.ReadInput("Provide nickname")
		projname := utils.ReadInput("Provide project name")
		projectData = &adacore.ProjectData{UserName: username, ProjName: projname}
		err = utils.SaveJSONToFile(projectData, path)
		if err != nil {
			slog.Error("failed to save project data", "error", err)
		}
	}

	c <- projectData
}

package adacore

import (
	"fmt"
	"path/filepath"

	"github.com/baudii/ada-ai/internal/common"
	"github.com/baudii/ada-ai/pkg/utils"
)

const ProjectStrucutreFile string = "project-structure.json"

type projectContext struct {
	UserName string  `json:"userName"`
	ProjName string  `json:"projName"`
	Proj     Project `json:"-"` //TODO: Move inside Ada struct
}

// AddProjCtx sets the project context for the Ada instance with the provided
// username and project name.
func (ada *Ada) AddProjCtx(userName string, projName string) {
	ada.Ctx = &projectContext{UserName: userName, ProjName: projName}
}

// SaveCtx saves the current project context to the user data file.
// If no context is set, it returns an error.
func (ada *Ada) SaveCtx() error {
	if ada.Ctx == nil {
		return fmt.Errorf("no project context to save")
	}
	return utils.SaveJSONToFile(ada.Ctx, userDataPath)
}

// ResolveProjectPath resolves a project root folder path, based using
// current project and user context.
//
// This method returns relative path.
func (ada *Ada) ResolveProjectPath() string {
	return filepath.Join(common.Artifacts, ada.cfg.ProjRoot, ada.Ctx.UserName, ada.Ctx.ProjName)
}

// SetWorkspace adds a descriptor of the project, that defines the way
// project will be saved and manipulated.
func (ada *Ada) SetWorkspace(proj Project) {
	ada.Ctx.Proj = proj
}

// Materialize calls Materialize for currently selected project.
func (ada *Ada) Materialize() {
	ada.Ctx.Proj.Materialize()
}

package cli

import (
	"os"
	"path/filepath"

	"github.com/baudii/ada-ai/internal/adacore"
	"github.com/baudii/ada-ai/internal/common"
	"github.com/baudii/ada-ai/pkg/utils"
)

const desc string = "An AI-powered app that suggests recipes based on the ingredients you already have at home"

var (
	Debug      bool
	DebugStage int
)

var debugCfg adacore.Config = adacore.Config{
	ProjRoot:        ".projects",
	ReflectionDepth: 3,
	Timeout:         "3m",
}

var (
	debugUserName = "baudii"
	debugProjName = "pantrypal"
)

func enableDebugging() {
	switch DebugStage {
	case 0:
		debugProjectStructure()
	default:
		debugStage()
	}
}

func debugStage() {
	ai := common.RegisterOllama()
	ada := adacore.New(ai, &debugCfg)
	ada.AddProjCtx(debugUserName, debugProjName)

	utils.ReadInput("Press Enter to continue")
	template := ada.GetTemplate(DebugStage, desc)
	data, err := ada.SendWithReflection(template)
	if err != nil {
		panic(err)
	}

	adacore.Print(data)
	ada.EnsureSaved(data)
	ada.Materialize(data)
}

func debugProjectStructure() {
	r := filepath.Join(common.DebuggingPath, "structure-unparsed.txt")
	path := utils.GetAbsolutePath(r)
	f, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	if f, err = utils.TrimJSON(string(f)); err != nil {
		panic(err)
	}

	adacore.Print(f)
}

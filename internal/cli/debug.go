package cli

import (
	"os"
	"path/filepath"

	"github.com/baudii/ada-ai/internal/ada"
	"github.com/baudii/ada-ai/internal/common"
	"github.com/baudii/ada-ai/pkg/utils"
)

const desc string = "An AI-powered app that suggests recipes based on the ingredients you already have at home"

var (
	Debug      bool
	DebugStage int
)

var debugCfg ada.Config = ada.Config{
	ProjRoot:        ".projects",
	ProjName:        "pantrypal",
	UserName:        "baudiis",
	ReflectionDepth: 3,
	Timeout:         "3m",
}

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
	ada.Init(ai, &debugCfg)

	utils.ReadInput("Press Enter to continue")
	template := ada.GetTemplate(DebugStage, desc)
	data, err := ada.SendWithReflection(template)
	if err != nil {
		panic(err)
	}

	ada.Print(data)
}

func debugProjectStructure() {
	r := filepath.Join(common.DebuggingPath, "structure-unparsed.txt")
	path := utils.GetAbsolutePath(r)
	f, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	data := string(f)
	if f, err = utils.TrimJSON(data); err != nil {
		panic(err)
	}

	ada.Print(f)
}

package ada

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/baudii/ada-ai/pkg/utils"
)

const desc string = "An AI-powered app that suggests recipes based on the ingredients you already have at home"

var debugCfg config = config{
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
	case 1:
		debugStage1Reflection()
	}
}

func debugStage1Reflection() {
	cfg = &debugCfg
	var err error
	utils.ReadInput("Press Enter to continue")
	if err = processInput(desc); err != nil {
		panic(err)
	}
}

func debugProjectStructure() {
	cfg = &debugCfg
	r := filepath.Join("diagnostics", "structure-unparsed.txt")
	path := utils.GetAbsolutePath(r)
	f, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	data := string(f)
	if data, err = tidy(data); err != nil {
		panic(err)
	}

	if err = saveProjectStructure(data); err != nil {
		var nfErr *ProjExist
		if !errors.As(err, &nfErr) {
			panic(err)
		}
	}

	node, err := unmarshalStructure(data)
	if err != nil {
		panic(err)
	}

	node.printTree()
	node.materialize(projRoot)
}

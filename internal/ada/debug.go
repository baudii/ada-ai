package ada

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/baudii/ada-ai/pkg/utils"
)

func debugProjectStructure() {
	r := filepath.Join("diagnostics", "structure-unparsed.txt")
	cfg.ProjName = "airline-price-analyzer"
	cfg.UserName = "baudii"
	cfg.ProjRoot = ".projects"
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

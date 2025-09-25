package ada

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/baudii/ada-ai/internal/llm"
	"github.com/baudii/ada-ai/internal/utils"
)

type config struct {
	ProjRoot string `json:"projRoot"`
	UserName string `json:"userName"`
	ProjName string `json:"projName"`
}

var cfg config
var cfgPath string
var ai llm.LLM

func Run(debugMode bool) {
	if debugMode {
		debugProjectStructure()
		return
	}

	relativePath := filepath.Join("cfg", "ada.json")
	cfgPath = utils.GetAbsolutePath(relativePath)
	cfg = utils.ParseJsonFile[config](cfgPath)

	var err error
	if ai, err = llm.Resolve(); err != nil {
		panic(err)
	}

	if cfg.UserName == "" {
		cfg.UserName = utils.ReadInput("Provide nickname")
		utils.SaveJsonToFile(cfg, cfgPath)
	}

	if cfg.ProjName == "" {
		cfg.ProjName = utils.ReadInput("Provide project name")
		utils.SaveJsonToFile(cfg, cfgPath)
	}

	input := utils.ReadInput("Provide project description")
	if err = processInput(input); err != nil {
		fmt.Printf("Something went wrong when processing the request: %v\n", err)
	}
}

func processInput(input string) error {
	response, err := sendReqWithTemplate(1, cfg.ProjName, input)
	if err != nil {
		return err
	}

	data := string(response.Message.Content)
	if data, err = tidy(data); err != nil {
		panic(err)
	}

	err = saveProjectStructure(response.Message.Content)
	if err != nil {
		fmt.Println(err)
	}

	node, err := unmarshalStructure(data)
	if err != nil {
		return err
	}

	node.printTree()
	return nil
}

func sendReqWithTemplate(step int, input ...string) (*llm.Response, error) {
	fileName := fmt.Sprintf("prompts/step%v-template.txt", step)
	template, err := os.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to parse a file %v", fileName)
	}

	prompt := fmt.Sprintf(string(template), utils.ToAnySlice(input)...)
	fmt.Println(prompt)
	res, err := ai.SendMessage(prompt)
	if err != nil {
		return nil, err
	}

	return res, nil
}

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

	fmt.Println(projRoot)
	node.printTree()
	node.materialize(projRoot)
}

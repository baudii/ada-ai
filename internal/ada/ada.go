package ada

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/baudii/ada-ai/internal"
	"github.com/baudii/ada-ai/internal/llm"
)

type config struct {
	ProjRoot string `json:"projRoot"`
	UserName string `json:"userName"`
	ProjName string `json:"projName"`
}

var cfg config
var cfgPath string
var Debug bool = false
var ai llm.LLM

func Run() {
	relativePath := filepath.Join("cfg", "ada.json")
	cfgPath = internal.GetAbsolutePath(relativePath)
	cfg = internal.ParseJsonFile[config](cfgPath)

	var err error
	ai, err = llm.Resolve()
	if err != nil {
		panic(err)
	}

	var input string
	for {
		if cfg.UserName == "" {
			cfg.UserName = internal.ReadInput("Provide nickname")
			internal.SaveJsonToFile(cfg, cfgPath)
		}

		if cfg.ProjName == "" {
			cfg.ProjName = internal.ReadInput("Provide project name")
			internal.SaveJsonToFile(cfg, cfgPath)
		}
		createDirectory()
		if projectExist() {
			err := parseStructure()
			if err != nil {
				fmt.Println(err)
			}
			internal.ReadInput("Press any key to continue...")
			continue
		}

		input = internal.ReadInput("Provide project description")
		if input == "" && Debug {
			input = "Build a web application that can automatically scan vacancies and send job applications to employers"
		}
		err = processInput(input)
		if err != nil {
			fmt.Printf("Something went wrong when processing the request: %v\n", err)
		}
	}
}

func processInput(input string) error {
	response, err := sendReqWithTemplate(input, 1)
	if err != nil {
		return err
	}

	err = saveProjectStructure(response.Message.Content)
	if err != nil {
		fmt.Println(err)
	}

	err = parseStructure()
	if err != nil {
		fmt.Println(err)
	}

	return nil
}

func sendReqWithTemplate(input string, step int) (*llm.Response, error) {
	fileName := fmt.Sprintf("prompts/step%v-template.txt", step)
	template, err := os.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to parse a file %v", fileName)
	}

	prompt := fmt.Sprintf(string(template), input)
	res, err := ai.SendMessage(prompt)
	if err != nil {
		return nil, err
	}

	return res, nil
}

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
var step1 string

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
		input = internal.ReadInput()
		err = processInput(input)
		if err != nil {
			fmt.Printf("Something went wrong when processing the request: %v\n", err)
		}
	}
}

func processInput(input string) error {
	template, err := getTemplate(1)
	if err != nil {
		return err
	}

	prompt := fmt.Sprintf(template, input)
	res, err := ai.SendMessage(prompt)
	if err != nil {
		return err
	}

	fmt.Println(res)
	return nil
}

func getTemplate(step int) (string, error) {
	fileName := fmt.Sprintf("prompts/step%v-template.txt", step)
	template, err := os.ReadFile(fileName)
	if err != nil {
		return "", fmt.Errorf("failed to parse a file %v", fileName)
	}

	return string(template), nil
}

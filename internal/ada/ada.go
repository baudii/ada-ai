package ada

import (
	"fmt"
	"os"

	"github.com/baudii/ada-ai/internal"
	"github.com/baudii/ada-ai/internal/llm"
)

var Debug bool = false
var ai llm.LLM
var step1 string

func Run() {
	cfg := internal.ParseJsonFile[llm.Config]("cfg/llm.json")
	var err error
	ai, err = llm.Resolve(cfg)
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

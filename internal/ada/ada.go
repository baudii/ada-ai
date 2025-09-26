package ada

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/baudii/ada-ai/pkg/llm"
	"github.com/baudii/ada-ai/pkg/utils"
)

type config struct {
	Timeout         string `json:"requestTimeout"`
	ReflectionDepth int    `json:"reflectionDepth"`
	ProjRoot        string `json:"projRoot"`
	UserName        string `json:"userName"`
	ProjName        string `json:"projName"`
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
	prompt := getPromptFromTemplate("step1-template.txt", cfg.ProjName, input)
	response, err := sendReqWithTemplate(prompt)
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

func sendReqWithTemplate(prompt string) (*llm.Response, error) {
	dur, err := time.ParseDuration(cfg.Timeout)
	if err != nil {
		dur = time.Minute * 3
		fmt.Printf("Failed to parse duration: '%v'. Using default: %v", cfg.Timeout, dur)
	}

	ctx, cf := context.WithTimeout(context.Background(), dur)
	res, err := ai.SendMessage(prompt, ctx, nil)
	cf()
	if err != nil {
		return nil, err
	}

	return res, nil
}

func getPromptFromTemplate(fileName string, input ...any) string {
	fileName = utils.GetAbsolutePath(filepath.Join("prompts", fileName))
	template, err := os.ReadFile(fileName)
	if err != nil {
		panic(err)
	}

	prompt := fmt.Sprintf(string(template), input...)
	return prompt
}

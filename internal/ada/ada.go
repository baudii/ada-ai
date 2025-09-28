package ada

import (
	"context"
	"fmt"
	"log/slog"
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

var cfg *config
var cfgPath string
var ai llm.LLM

var defaultCfg config = config{
	ProjRoot: ".projects",
	Timeout:  "3m",
}

func Run(debugMode bool) {
	if debugMode {
		debugProjectStructure()
		return
	}

	var err error
	if ai, err = llm.Resolve(); err != nil {
		panic(err)
	}

	relativePath := filepath.Join("cfg", "ada.json")
	cfgPath = utils.GetAbsolutePath(relativePath)
	cfg, err = utils.ParseJsonConfigWithLocal[config](cfgPath)
	if err != nil {
		cfg = &defaultCfg
	}

	if cfg.UserName == "" {
		cfg.UserName = utils.ReadInput("Provide nickname")
		utils.SaveJsonToFile(cfg, cfgPath)
	}
	slog.Info("recognized username", "username", cfg.UserName)

	if cfg.ProjName == "" {
		cfg.ProjName = utils.ReadInput("Provide project name")
		utils.SaveJsonToFile(cfg, cfgPath)
	}
	slog.Info("recognized project", "projname", cfg.ProjName)

	input := utils.ReadInput("Provide project description")
	if err = processInput(input); err != nil {
		slog.Error("something went wrong when processing the request", "error", err)
	}
}

func processInput(input string) error {
	prompt := getPromptFromTemplate("step1-template.txt", cfg.ProjName, input)
	msgs, err := sendRequest(prompt, nil)
	if err != nil {
		return err
	}

	response := msgs[len(msgs)-1]
	data := string(response.Content)
	if data, err = tidy(data); err != nil {
		panic(err)
	}

	err = saveProjectStructure(response.Content)
	if err != nil {
		slog.Error("error occurred when saving project structure", "error", err)
	}

	node, err := unmarshalStructure(data)
	if err != nil {
		return err
	}

	node.printTree()
	return nil
}

func sendRequest(prompt string, msgs []llm.Message) ([]llm.Message, error) {
	dur, err := time.ParseDuration(cfg.Timeout)
	if err != nil {
		dur = time.Minute * 3
		slog.Warn("failed to parse duration from config: using default", "duration", cfg.Timeout, "default", dur)
	}

	if msgs == nil {
		msgs = []llm.Message{}
	}

	msgs = append(msgs, llm.Message{
		Role:    llm.RoleUser,
		Content: prompt,
	})

	ctx, cf := context.WithTimeout(context.Background(), dur)
	res, err := ai.SendMessage(msgs, ctx)
	cf()
	return res, err
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

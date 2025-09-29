package ada

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/baudii/ada-ai/pkg/utils"
	"github.com/tmc/langchaingo/llms"
)

type config struct {
	Timeout         string `json:"requestTimeout"`
	ReflectionDepth int    `json:"reflectionDepth"`
	ProjRoot        string `json:"projRoot"`
	UserName        string `json:"userName"`
	ProjName        string `json:"projName"`
}

var (
	Debug      bool
	DebugStage int
)

var cfg *config
var cfgPath string
var ai llms.Model

var defaultCfg config = config{
	ProjRoot: ".projects",
	Timeout:  "3m",
}

func Run(model llms.Model) {
	ai = model
	if Debug {
		enableDebugging()
		return
	}

	var err error
	relativePath := filepath.Join("data", "configuration", "ada.json")
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
	resp, err := sendRequest(prompt, nil)
	if err != nil {
		return err
	}

	data := resp.Choices[0].Content
	if data, err = tidy(data); err != nil {
		panic(err)
	}

	if err = Improve(&prompt, &data); err != nil {
		slog.Error("failed to improve", "error", err)
	}

	slog.Info("finished improve")
	var r string
	r, err = tidy(data)
	if err != nil {
		slog.Info("failed to clean the json", "error", err)
	}

	err = saveProjectStructure(r)
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

func sendRequest(prompt string, msgs []llms.MessageContent) (*llms.ContentResponse, error) {
	dur, err := time.ParseDuration(cfg.Timeout)
	if err != nil {
		dur = time.Minute * 3
		slog.Warn("failed to parse duration from config: using default", "duration", cfg.Timeout, "default", dur)
	}

	msgs = append(msgs,
		llms.TextParts(llms.ChatMessageTypeSystem, prompt),
	)

	ctx, cf := context.WithTimeout(context.Background(), dur)
	res, err := ai.GenerateContent(ctx, msgs, llms.WithJSONMode())
	cf()
	return res, err
}

func getPromptFromTemplate(fileName string, input ...any) string {
	fileName = utils.GetAbsolutePath(filepath.Join("data", "prompts", fileName))
	template, err := os.ReadFile(fileName)
	if err != nil {
		panic(err)
	}

	prompt := fmt.Sprintf(string(template), input...)
	return prompt
}

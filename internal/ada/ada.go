package ada

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/baudii/ada-ai/internal/common"
	"github.com/baudii/ada-ai/pkg/utils"
	"github.com/tmc/langchaingo/llms"
)

type Config struct {
	Timeout         string `json:"requestTimeout"`
	ReflectionDepth int    `json:"reflectionDepth"`
	ProjRoot        string `json:"projRoot"`
	UserName        string `json:"userName"`
	ProjName        string `json:"projName"`
}

var cfg *Config
var ai llms.Model

func Init(model llms.Model, c *Config) {
	cfg = c
	ai = model
}

func SendWithReflection(prompt string) ([]byte, error) {
	resp, err := GenerateJSON(prompt, nil)
	if err != nil {
		return nil, err
	}

	data := resp.Choices[0].Content
	if err = Improve(&prompt, &data); err != nil {
		slog.Error("failed to improve", "error", err)
	}

	slog.Info("finished improve")
	return utils.TrimJSON(data)
}

func EnsureSaved(data []byte) error {
	var err error
	if err = saveProjectStructure(data); err != nil {
		var nfErr *ProjExist
		if !errors.As(err, &nfErr) {
			return err
		}
	}

	return nil
}

func Print(data []byte) error {
	node, err := parseNode(data)
	if err != nil {
		return err
	}

	printTree(node, "", true)
	return nil
}

func Materialize(data []byte) error {
	if err := EnsureSaved(data); err != nil {
		slog.Error("failed to save data", "error", err)
	}

	node, err := parseNode(data)
	if err != nil {
		return err
	}

	node.materialize(projRoot)
	return nil
}

func GetTemplate(step int, input string) string {
	return getPromptFromTemplate(fmt.Sprintf("step%d-template.txt", step), cfg.ProjName, input)
}

func GenerateJSON(prompt string, msgs []llms.MessageContent) (*llms.ContentResponse, error) {
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
	fileName = utils.GetAbsolutePath(filepath.Join(common.PromptsPath, fileName))
	template, err := os.ReadFile(fileName)
	if err != nil {
		panic(err)
	}

	prompt := fmt.Sprintf(string(template), input...)
	return prompt
}

package adacore

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/baudii/ada-ai/internal/common"
	"github.com/baudii/ada-ai/pkg/utils"
	"github.com/tmc/langchaingo/llms"
)

var UserDataPath = filepath.Join(common.DataPath, "user_data.json")

type Config struct {
	Timeout         string `json:"requestTimeout"`
	ReflectionDepth int    `json:"reflectionDepth"`
	ProjRoot        string `json:"projRoot"`
}

type projectContext struct {
	UserName     string `json:"userName"`
	ProjName     string `json:"projName"`
	projRoot     string
	projDescFile string
}

type Ada struct {
	cfg *Config
	ai  llms.Model
	Ctx *projectContext
}

func New(ai llms.Model, cfg *Config) *Ada {
	pc, err := utils.ParseJSONFile[projectContext](utils.GetAbsolutePath(UserDataPath))
	if err != nil {
		return &Ada{cfg: cfg, ai: ai}
	}
	return &Ada{cfg: cfg, ai: ai, Ctx: pc}
}

func (ada *Ada) AddProjCtx(userName string, projName string) {
	ada.Ctx = &projectContext{UserName: userName, ProjName: projName}
}

func (ada *Ada) SendWithReflection(prompt string) ([]byte, error) {
	resp, err := ada.GenerateJSON(prompt, nil)
	if err != nil {
		return nil, err
	}

	data := resp.Choices[0].Content
	if err = ada.ReflectImprove(&prompt, &data); err != nil {
		slog.Error("failed to improve", "error", err)
	}

	slog.Info("finished improve")
	return utils.TrimJSON(data)
}

func (ada *Ada) GetTemplate(step int, input string) string {
	return getPromptFromTemplate(fmt.Sprintf("step%d-template.txt", step), ada.Ctx.ProjName, input)
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

func (ada *Ada) GenerateJSON(prompt string, msgs []llms.MessageContent) (*llms.ContentResponse, error) {
	dur, err := time.ParseDuration(ada.cfg.Timeout)
	if err != nil {
		dur = time.Minute * 3
		slog.Warn("failed to parse duration from config: using default", "duration", ada.cfg.Timeout, "default", dur)
	}

	msgs = append(msgs,
		llms.TextParts(llms.ChatMessageTypeSystem, prompt),
	)

	ctx, cf := context.WithTimeout(context.Background(), dur)
	res, err := ada.ai.GenerateContent(ctx, msgs, llms.WithJSONMode())
	cf()
	return res, err
}

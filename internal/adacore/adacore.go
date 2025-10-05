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

type promptTemplate string

const (
	ReflectPromptFile promptTemplate = "reflect-template.txt"
	Step1PromptFile   promptTemplate = "step1-template.txt"
)

var userDataPath = utils.GetAbsolutePath(filepath.Join(common.DataPath, "user_data.json"))

type Config struct {
	Timeout         string `json:"requestTimeout"`
	ReflectionDepth int    `json:"reflectionDepth"`
	ProjRoot        string `json:"projRoot"`
}

type Ada struct {
	cfg *Config
	ai  llms.Model
	Ctx *projectContext
}

type Project interface {
	Materialize() error
	Structure() map[string]any
}

// New initializes a new Ada instance with the provided LLM model and configuration.
// If a user context is found in the user data file, it is loaded and associated with
// the Ada instance. Otherwise, the context remains nil and expected to be set later
// using AddProjCtx.
func New(ai llms.Model, cfg *Config) *Ada {
	pc, err := utils.ParseJSONFile[projectContext](userDataPath)
	if err != nil {
		return &Ada{cfg: cfg, ai: ai}
	}
	return &Ada{cfg: cfg, ai: ai, Ctx: pc}
}

// GenerateJSON sends a prompt along with a series of messages to the LLM
// and expects a JSON response. It uses the timeout specified in the Ada configuration.
// If the timeout is invalid, it defaults to 3 minutes.
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

// GetPromptFromTemplate reads a prompt template file and formats it with the provided input.
// It returns the formatted prompt string or an error if the file cannot be read.
func GetPromptFromTemplate(file promptTemplate, input ...any) (string, error) {
	fileName := utils.GetAbsolutePath(filepath.Join(common.PromptsPath, string(file)))
	template, err := os.ReadFile(fileName)
	if err != nil {
		return "", fmt.Errorf("failed to read prompt template file %q: %w", fileName, err)
	}

	return fmt.Sprintf(string(template), input...), nil
}

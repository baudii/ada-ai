package ada

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/tmc/langchaingo/llms"
)

const (
	reflectPrompt      = "reflect-template.txt"
	reflectShortPrompt = "reflect-template-short.txt"
	improvePrompt      = "improve-template.txt"
)

// Ada is the main struct for Ada AI workflow, encapsulating LLM models,
// configuration options, and project context.
type Ada struct {
	ai          llms.Model
	Timeout     string
	PromptsRoot string
	Reflection  ReflectConfig
}

// GenerateWithSys sends a prompt with a system message to the LLM and expects a response.
// It calls GenerateContent with the provided system prompt.
func (ada *Ada) GenerateWithSys(ctx context.Context, sys, user string, callOptions ...llms.CallOption) (*llms.ContentResponse, error) {
	msgs := []llms.MessageContent{llms.TextParts(llms.ChatMessageTypeSystem, sys)}
	return ada.GenerateContent(ctx, user, msgs, callOptions...)
}

// GenerateContent sends a prompt along with a series of messages to the LLM
// and expects a response. It uses the options and timeout specified in the
// Ada configuration. If the timeout is invalid, it defaults to 3 minutes.
func (ada *Ada) GenerateContent(ctx context.Context, prompt string, msgs []llms.MessageContent, callOptions ...llms.CallOption) (*llms.ContentResponse, error) {
	dur, err := time.ParseDuration(ada.Timeout)
	if err != nil {
		dur = time.Minute * 3
		slog.Warn("failed to parse duration from config: using default", "duration", ada.Timeout, "default", dur)
	}

	msgs = append(msgs, llms.TextParts(llms.ChatMessageTypeHuman, prompt))
	ctx, cancel := context.WithTimeout(ctx, dur)
	res, err := ada.ai.GenerateContent(ctx, msgs, callOptions...)
	cancel()
	return res, err
}

// PromptFromTemplate reads a prompt template file and formats it with the provided input.
// It returns the formatted prompt string or an error if the file cannot be read.
func (ada *Ada) PromptFromTemplate(filename string, input ...any) (string, error) {
	file := ada.prompt(filename)
	template, err := os.ReadFile(file)
	if err != nil {
		return "", fmt.Errorf("read template: %w", err)
	}

	return fmt.Sprintf(string(template), input...), nil
}

func (ada *Ada) prompt(filename string) string {
	return filepath.Join(ada.PromptsRoot, filename)
}

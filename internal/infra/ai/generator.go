package ai

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/baudii/ada-ai/internal/core/gen"
	"github.com/tmc/langchaingo/llms"
)

// GenerateWithSys sends a prompt with a system message to the LLM and expects a response.
// It calls GenerateContent with the provided system prompt.
func (g *generator) GenerateWithSys(ctx context.Context, sys, user string, opts ...gen.Option) (string, error) {
	msgs := []llms.MessageContent{llms.TextParts(llms.ChatMessageTypeSystem, sys)}
	callOptions := ToCallOptions(opts...)
	resp, err := g.GenerateContent(ctx, user, msgs, callOptions...)
	if err != nil {
		return "", err
	}

	if resp == nil || len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from LLM")
	}

	return resp.Choices[0].Content, nil
}

// GenerateContent sends a prompt along with a series of messages to the LLM
// and expects a response. It uses the options and timeout specified in the
// Ada configuration. If the timeout is invalid, it defaults to 3 minutes.
func (g *generator) GenerateContent(ctx context.Context, prompt string, msgs []llms.MessageContent, callOptions ...llms.CallOption) (*llms.ContentResponse, error) {
	dur, err := time.ParseDuration(g.Timeout)
	if err != nil {
		dur = time.Minute * 3
		slog.Warn("failed to parse duration from config: using default", "duration", g.Timeout, "default", dur)
	}

	msgs = append(msgs, llms.TextParts(llms.ChatMessageTypeHuman, prompt))
	ctx, cancel := context.WithTimeout(ctx, dur)
	res, err := g.ai.GenerateContent(ctx, msgs, callOptions...)
	cancel()
	return res, err
}

// BuildPrompt reads a prompt template file and formats it with the provided input.
// It returns the formatted prompt string or an error if the file cannot be read.
func (g *generator) BuildPrompt(filename string, input ...any) (string, error) {
	file := filepath.Join(g.PromptsRoot, filename)
	template, err := os.ReadFile(file)
	if err != nil {
		return "", fmt.Errorf("read template: %w", err)
	}

	return fmt.Sprintf(string(template), input...), nil
}

// ToCallOptions converts generation options into LLM call options.
func ToCallOptions(opts ...gen.Option) []llms.CallOption {
	params := &gen.Params{}
	for _, o := range opts {
		o(params)
	}

	var callOptions []llms.CallOption
	if params.Temperature != nil {
		callOptions = append(callOptions, llms.WithTemperature(*params.Temperature))
	}
	if params.JSONMode != nil {
		callOptions = append(callOptions, llms.WithJSONMode())
	}
	return callOptions
}

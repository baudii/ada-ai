package ada

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"

	"github.com/baudii/ada-ai/pkg/dilog"
	"github.com/tmc/langchaingo/llms"
)

type Score struct {
	Relevance    float32 `json:"relevance"`
	Accuracy     float32 `json:"accuracy"`
	Completeness float32 `json:"completeness"`
}

type Reflection struct {
	Scores      Score    `json:"scores"`
	Suggestions []string `json:"improvement_suggestions"`
}

func (r Reflection) Avg() float32 {
	s := r.Scores
	return (s.Relevance + s.Accuracy + s.Completeness) / 3
}

func Improve(request *string, response *string) error {
	var (
		reflection *Reflection
		err        error
		prompt     string

		bestScore float32               = math.SmallestNonzeroFloat32
		bestAns   *string               = response
		curAns    *string               = response
		threshold float32               = 0.95
		msgs      []llms.MessageContent = make([]llms.MessageContent, 3)
	)

	for i := 0; i < cfg.ReflectionDepth; i++ {
		slog.Debug("reflecting", "attempt", i+1, "total", cfg.ReflectionDepth)
		prompt = getPromptFromTemplate("reflect-template.txt", *request, *curAns)
		dilog.WritelnToDw("prompt:\n" + prompt)
		reflection, err = reflect(&prompt)
		if err != nil {
			slog.Error("error occurred during reflection", "error", err)
			continue
		}

		avg := reflection.Avg()
		slog.Debug("reflection completed", "reflect", reflection, "avg", avg)
		if avg > threshold {
			slog.Debug("threshold met")
			*response = *bestAns
			return nil
		}

		if avg > bestScore {
			slog.Debug("found new best score")
			bestScore = reflection.Avg()
			bestAns = curAns
		}

		msgs[0] = llms.TextParts(llms.ChatMessageTypeHuman, prompt)
		msgs[1] = llms.TextParts(llms.ChatMessageTypeAI, fmt.Sprintf("**RESPONSE_EVALUATION**: %v", reflection))
		curAns, err = improveResponse(msgs)
		if err != nil {
			*response = *bestAns
			return err
		}
	}

	if cfg.ReflectionDepth <= 0 {
		return fmt.Errorf("reflection omitted: reflection depth is set to %v", cfg.ReflectionDepth)
	}

	*response = *bestAns
	return fmt.Errorf("failed to improve the response to threshold %v - setting the best score: %v", threshold, bestScore)
}

func improveResponse(msgs []llms.MessageContent) (*string, error) {
	msgs[2] = llms.TextParts(
		llms.ChatMessageTypeHuman,
		"From the given evaluation and suggestions provide an improved version of **MODEL_RESPONSE**.",
	)

	var err error
	resp, err := ai.GenerateContent(context.Background(), msgs, llms.WithJSONMode())
	if err != nil {
		return nil, err
	}

	return &resp.Choices[0].Content, nil
}

func reflect(prompt *string) (*Reflection, error) {
	var err error
	resp, err := sendRequest(*prompt, []llms.MessageContent{})
	if err != nil {
		return nil, err
	}

	var js string
	js, err = tidy(resp.Choices[0].Content)
	if err != nil {
		return nil, fmt.Errorf("error occurred in reflect() when cleaning up the response: %v", err)
	}

	var r Reflection
	err = json.Unmarshal([]byte(js), &r)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

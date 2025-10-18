package adacore

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"

	"github.com/baudii/ada-ai/pkg/utils"
	"github.com/tmc/langchaingo/llms"
)

// ReflectConfig defines configuration parameters for the reflection algorithm.
//
// Depth controls how many maximum iterations of reflection will be performed.
// Threshold sets the minimum score (0.0-1.0) required to early break from
// reflection cycle.
type ReflectConfig struct {
	Depth      int     `json:"reflectionDepth"`
	Threshhold float32 `json:"reflectionThreshold"`
}

type score struct {
	Relevance    float32 `json:"relevance"`
	Accuracy     float32 `json:"accuracy"`
	Completeness float32 `json:"completeness"`
}

type reflection struct {
	Scores      score    `json:"scores"`
	Suggestions []string `json:"improvement_suggestions"`
}

type eval struct {
	ans   string
	score float32
}

type templates struct {
	reflect      string
	shortReflect string
	improve      string
}

// SendReflect sends a request to the LLM and then performs a
// reflection and improvement of this response based on the configured
// parameters.
func (ada *Ada) SendReflect(prompt string) ([]byte, error) {
	resp, err := ada.GenerateContent(context.Background(), prompt, nil)
	if err != nil {
		return nil, err
	}

	data := resp.Choices[0].Content
	res, err := ada.Improve(prompt, data)
	if err != nil {
		slog.Error(err.Error())
	}
	if res != nil {
		data = res.ans
	}

	slog.Info("finished improve")
	return utils.TrimJSON(data)
}

// Improve takes an initial response and iteratively reflects on it
// and attempts to improve it based on the reflection results. The process
// continues until the reflection score meets or exceeds the configured
// threshold or the maximum number of reflection cycles is reached.
func (ada *Ada) Improve(request string, response string) (*eval, error) {
	var (
		res    = &eval{response, float32(math.Inf(-1))}
		curAns = response
		msgs   = make([]llms.MessageContent, 5)
	)

	templates, err := ada.loadTemplates()
	if err != nil {
		return res, fmt.Errorf("load templates: %w", err)
	}

	for i := 0; i < ada.Opts.Reflection.Depth; i++ {
		slog.Debug("reflect cycle start", "attempt", i+1, "depth", ada.Opts.Reflection.Depth)
		msgs = append(msgs, llms.TextParts(llms.ChatMessageTypeHuman, request))
		msgs = append(msgs, llms.TextParts(llms.ChatMessageTypeAI, curAns))
		reflection, err := ada.Reflect(templates.reflect, msgs)
		if err != nil {
			slog.Error("error occurred during reflection", "error", err)
			continue
		}

		avg := reflection.avg()
		slog.Debug("reflection completed", "reflect", reflection, "avg", avg)
		if avg > res.score {
			slog.Debug("found new best score")
			res.score = avg
			res.ans = curAns
		}

		if res.score > ada.Opts.Reflection.Threshhold {
			slog.Debug("threshold met")
			return res, nil
		}

		msgs = append(msgs, llms.TextParts(llms.ChatMessageTypeHuman, templates.shortReflect))
		msgs = append(msgs, llms.TextParts(llms.ChatMessageTypeAI, fmt.Sprintf("%v", reflection)))
		resp, err := ada.GenerateContent(context.Background(), templates.improve, msgs)
		if err != nil {
			return res, fmt.Errorf("generate improved prompt: %w", err)
		}

		msgs = msgs[:0]
		msgs = append(msgs, llms.TextParts(llms.ChatMessageTypeHuman, request))
		msgs = append(msgs, llms.TextParts(llms.ChatMessageTypeAI, curAns))
		msgs = append(msgs, llms.TextParts(llms.ChatMessageTypeHuman, fmt.Sprintf("%v", reflection)))
		curAns = resp.Choices[0].Content
	}

	if ada.Opts.Reflection.Depth <= 0 {
		return res, fmt.Errorf("reflection depth is set to %v", ada.Opts.Reflection.Depth)
	}

	slog.Error("failed to improve", "best", res.score, "threshold", ada.Opts.Reflection.Threshhold)
	return res, nil
}

// Reflect sends the current conversation context to the LLM along with
// a reflection prompt, asking the model to evaluate the previous response
// and provide scores and improvement suggestions.
func (ada *Ada) Reflect(reflectPrompt string, msgs []llms.MessageContent) (*reflection, error) {
	resp, err := ada.GenerateContent(context.Background(), reflectPrompt, msgs)
	if err != nil {
		return nil, fmt.Errorf("reflect: %w", err)
	}

	var js []byte
	js, err = utils.TrimJSON(resp.Choices[0].Content)
	if err != nil {
		return nil, fmt.Errorf("trim reflect: %w", err)
	}

	var r reflection
	err = json.Unmarshal(js, &r)
	if err != nil {
		return nil, fmt.Errorf("unmarshal reflect: %w", err)
	}

	return &r, nil
}

func (ada *Ada) loadTemplates() (*templates, error) {
	reflectTemplate, err := ada.PromptFromTemplate(reflectPrompt)
	if err != nil {
		return nil, err
	}

	sReflectTemplate, err := ada.PromptFromTemplate(reflectShortPrompt)
	if err != nil {
		return nil, err
	}

	improveTemplate, err := ada.PromptFromTemplate(improvePrompt)
	if err != nil {
		return nil, err
	}

	return &templates{reflectTemplate, sReflectTemplate, improveTemplate}, nil
}

func (r reflection) avg() float32 {
	s := r.Scores
	return (s.Relevance + s.Accuracy + s.Completeness) / 3
}

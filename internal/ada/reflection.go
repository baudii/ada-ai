package ada

import (
	"encoding/json"
	"log/slog"
	"math"
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

func Improve(prompt *string, response *string) {
	var (
		r       *Reflection
		bestSc  float32 = math.SmallestNonzeroFloat32
		bestAns *string = response
		ans     *string = response
		err     error
	)

	for i := 0; i < cfg.ReflectionDepth; i++ {
		r, err = reflect(prompt, ans)
		avg := r.Avg()
		if err != nil {
			slog.Error("error occurred during reflection", "error", err)
			continue
		}
		if avg > 8 {
			*response = *bestAns
			return
		}

		if avg > bestSc {
			bestSc = r.Avg()
			bestAns = ans
		}

	}
}

func reflect(request *string, response *string) (*Reflection, error) {
	prompt := getPromptFromTemplate("reflect-template.txt", *request, *response)
	resp, err := sendRequest(prompt, nil)
	if err != nil {
		return nil, err
	}

	var r Reflection
	err = json.Unmarshal([]byte(resp.Choices[0].Content), &r)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

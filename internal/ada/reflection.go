package ada

import (
	"encoding/json"
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

func Reflect(request *string, response *string) (*Reflection, error) {
	prompt := getPromptFromTemplate("reflect-template.txt", *request, *response)
	rfl, err := sendReqWithTemplate(prompt)
	if err != nil {
		return nil, err
	}

	var r Reflection
	err = json.Unmarshal([]byte(rfl.Message.Content), &r)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

func (r Score) Avg() float32 {
	return (r.Relevance + r.Accuracy + r.Completeness) / 3
}

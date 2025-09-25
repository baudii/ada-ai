package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/baudii/ada-ai/pkg/llm"
)

type Ollama struct {
	url    string
	model  string
	stream bool
}

const defaultOllamaURL = "http://localhost:11434/api/chat"

func Init() {
	llm.Register("ollama", func(cfg map[string]any) (llm.LLM, error) {
		url, ok := cfg["url"].(string)
		if url == "" || !ok {
			url = defaultOllamaURL
		}

		model, ok := cfg["model"].(string)
		if !ok {
			model = "gemma3:4b"
		}

		stream, _ := cfg["stream"].(bool)

		return &Ollama{
			url:    url,
			model:  model,
			stream: stream,
		}, nil
	})
}

func (ol Ollama) SendMessage(prompt string, ctx context.Context) (*llm.Response, error) {
	ollamaReq := llm.Request{
		Model: ol.model,
		Messages: []llm.Message{{
			Role:    llm.RoleUser,
			Content: prompt,
		}},
		Stream: ol.stream,
	}
	js, err := json.Marshal(&ollamaReq)
	if err != nil {
		return nil, err
	}
	client := http.Client{}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, ol.url, bytes.NewReader(js))
	if err != nil {
		return nil, err
	}
	httpResp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()
	ollamaResp := llm.Response{}
	err = json.NewDecoder(httpResp.Body).Decode(&ollamaResp)
	return &ollamaResp, err
}

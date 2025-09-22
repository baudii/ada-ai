package ollama

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/baudii/ada-ai/internal/llm"
)

type Ollama struct {
	url    string
	model  string
	stream bool
}

const defaultOllamaURL = "http://localhost:11434/api/chat"

func init() {
	fmt.Println("Registering ollama")
	llm.Register("ollama", func(cfg map[string]any) (llm.LLM, error) {
		url, ok := cfg["url"].(string)
		if url == "" || !ok {
			url = defaultOllamaURL
		}

		model, ok := cfg["model"].(string)
		if !ok {
			model = "gemma3:4b"
		}

		stream, ok := cfg["stream"].(bool)
		if !stream || !ok {
			stream = false
		}

		return &Ollama{
			url:    url,
			model:  model,
			stream: stream,
		}, nil
	})
}

func (ol Ollama) SendMessage(prompt string) (*llm.Response, error) {
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
	httpReq, err := http.NewRequest(http.MethodPost, ol.url, bytes.NewReader(js))
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

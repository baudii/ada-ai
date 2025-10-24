package ai

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/llms"
)

type mockLLM struct {
	generate func(...any) (*llms.ContentResponse, error)
}

var defaultMock = func(...any) (*llms.ContentResponse, error) {
	return &llms.ContentResponse{Choices: []*llms.ContentChoice{{Content: "some response"}}}, nil
}

func (m *mockLLM) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	return prompt, nil
}

func (m *mockLLM) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	return m.generate(messages)
}

func TestNew(t *testing.T) {
	t.Parallel()
	ai := &mockLLM{}
	tests := []struct {
		name        string
		promptsRoot string
		timeout     string
		callOpts    *llms.CallOptions
		expected    *generator
	}{
		{
			name:        "default",
			promptsRoot: "",
			timeout:     "3m",
			callOpts:    &llms.CallOptions{},
			expected:    &generator{ai: ai, Timeout: "3m"},
		},
		{
			promptsRoot: "custom/prompts",
			timeout:     "5m",
			callOpts:    &llms.CallOptions{Temperature: 0.7},
			expected:    &generator{ai: ai, Timeout: "5m", PromptsRoot: "custom/prompts"}},
		{
			promptsRoot: "/absolute/path",
			timeout:     "10m",
			callOpts:    &llms.CallOptions{MaxTokens: 1000},
			expected:    &generator{ai: ai, Timeout: "10m", PromptsRoot: "/absolute/path"}},
	}
	for i, v := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			ada := NewGenerator(ai, WithPromptsRoot(v.promptsRoot), WithTimeout(v.timeout))
			assert.Equal(t, v.expected, ada)
		})
	}
}

func TestGenerateWithSys(t *testing.T) {
	sysp, usp := "system prompt", "user prompt"
	m := &mockLLM{func(a ...any) (*llms.ContentResponse, error) {
		msgs, ok := a[0].([]llms.MessageContent)
		if !ok || len(msgs) < 2 || len(msgs[0].Parts) == 0 || len(msgs[1].Parts) == 0 {
			return nil, errors.New("wrong length or type of messages")
		}
		a1, ok1 := msgs[0].Parts[0].(llms.TextContent)
		a2, ok2 := msgs[1].Parts[0].(llms.TextContent)
		if !ok1 || !ok2 {
			return nil, errors.New("invalid messages")
		}
		if !strings.Contains(a1.Text, sysp) || !strings.Contains(a2.Text, usp) {
			return nil, errors.New("invalid messages")
		}
		return defaultMock(a...)
	}}
	ada := NewGenerator(m, WithTimeout("1m"))
	res, err := ada.GenerateWithSys(context.Background(), sysp, usp, true)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "some response", res)
}

func TestGenerateWithSys_Errors(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		mockFunc func(...any) (*llms.ContentResponse, error)
		err      string
	}{
		{
			name:     "llm error",
			mockFunc: func(a ...any) (*llms.ContentResponse, error) { return nil, assert.AnError },
			err:      assert.AnError.Error(),
		},
		{
			name: "no response",
			mockFunc: func(a ...any) (*llms.ContentResponse, error) {
				return &llms.ContentResponse{Choices: []*llms.ContentChoice{}}, nil
			},
			err: "no response from LLM",
		},
		{
			name: "nil response with no error",
			mockFunc: func(a ...any) (*llms.ContentResponse, error) {
				return &llms.ContentResponse{Choices: nil}, nil
			},
			err: "no response from LLM",
		},
	}

	for _, v := range tests {
		t.Run(v.name, func(t *testing.T) {
			m := &mockLLM{v.mockFunc}
			ada := NewGenerator(m, WithTimeout("1m"))
			_, err := ada.GenerateWithSys(context.Background(), "", "", true)
			assert.ErrorContains(t, err, v.err)
		})
	}
}

func TestGenerateJSON(t *testing.T) {
	t.Parallel()
	tests := []struct {
		timeout  string
		hasError bool
		mockFunc func(...any) (*llms.ContentResponse, error)
	}{
		{"1m", false, defaultMock},
		{"invalid", true, func(...any) (*llms.ContentResponse, error) {
			return nil, errors.New("context deadline exceeded")
		}},
	}

	for i, v := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			ai := &mockLLM{v.mockFunc}
			ada := NewGenerator(ai, WithTimeout(v.timeout))
			res, err := ada.GenerateContent(context.Background(), "test prompt", []llms.MessageContent{})
			assert.Equal(t, v.hasError, err != nil)
			if !v.hasError {
				assert.NotNil(t, res)
				assert.Equal(t, "some response", res.Choices[0].Content)
			}
		})
	}
}

// func TestResolveProjectPath(t *testing.T) {
// 	t.Parallel()
// 	tests := []struct {
// 		projRoot string
// 		username string
// 		projname string
// 	}{
// 		{"projroot", "username", "projname"},
// 	}

// 	for i, v := range tests {
// 		t.Run(strconv.Itoa(i), func(t *testing.T) {
// 			cfg := Options{ProjectsRoot: v.projRoot}
// 			ai := &mockLLM{}
// 			ada := New(ai, WithOptions(cfg))
// 			ada.AddProjectData(ProjectData{UserName: v.username, ProjName: v.projname})
// 			res := ada.projectPath()
// 			assert.Contains(t, res, v.username)
// 			assert.Contains(t, res, v.projRoot)
// 			assert.Contains(t, res, v.projname)
// 		})
// 	}
// }

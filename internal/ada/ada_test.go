package ada

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
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
		promptsRoot string
		timeout     string
		reflection  ReflectConfig
		callOpts    *llms.CallOptions
		expected    *Ada
	}{
		{"", "3m", ReflectConfig{}, &llms.CallOptions{},
			&Ada{ai: ai, Timeout: "3m", Reflection: ReflectConfig{}}},
		{"custom/prompts", "5m", ReflectConfig{Depth: 2, Threshhold: 0.9}, &llms.CallOptions{Temperature: 0.7},
			&Ada{ai: ai, Timeout: "5m", PromptsRoot: "custom/prompts", Reflection: ReflectConfig{Depth: 2, Threshhold: 0.9}}},
		{"/absolute/path", "10m", ReflectConfig{Depth: 4, Threshhold: 0.8}, &llms.CallOptions{MaxTokens: 1000},
			&Ada{ai: ai, Timeout: "10m", PromptsRoot: "/absolute/path", Reflection: ReflectConfig{Depth: 4, Threshhold: 0.8}}},
	}
	for i, v := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			ada := New(ai, WithPromptsRoot(v.promptsRoot), WithTimeout(v.timeout), WithReflection(v.reflection))
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
	ada := New(m, WithTimeout("1m"))
	res, err := ada.GenerateWithSys(context.Background(), sysp, usp)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "some response", res.Choices[0].Content)
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
			ada := New(ai, WithTimeout(v.timeout))
			res, err := ada.GenerateContent(context.Background(), "test prompt", []llms.MessageContent{})
			assert.Equal(t, v.hasError, err != nil)
			if !v.hasError {
				assert.NotNil(t, res)
				assert.Equal(t, "some response", res.Choices[0].Content)
			}
		})
	}
}

func TestPromptFromTemplate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		template string
		args     []any
		hasErr   bool
	}{
		{"some content", []any{}, false},
		{"some content %v", []any{"val1"}, false},
		{"some content %s, %q", []any{"val1", "val2"}, false},
		{"", []any{}, true},
	}

	file := "test_template.txt"

	for i, v := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			tempdir := t.TempDir()
			ada := New(&mockLLM{}, WithPromptsRoot(tempdir))
			if !v.hasErr {
				err := os.WriteFile(filepath.Join(tempdir, file), []byte(v.template), 0644)
				require.NoError(t, err)
			}
			res, err := ada.BuildPrompt(file, v.args...)
			assert.Equal(t, v.hasErr, err != nil)
			assert.Equal(t, fmt.Sprintf(v.template, v.args...), res)
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

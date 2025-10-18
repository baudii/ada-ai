package adacore

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
	tests := []struct {
		options      *Options
		projData     ProjectData
		expectedRoot string
		expectedPD   ProjectData
	}{
		{&Options{ProjectsRoot: "root"}, ProjectData{}, "root", ProjectData{UserName: "unknown_user", ProjName: "project_"}},
		{&Options{Timeout: "a", Reflection: ReflectConfig{1, 2}, ModelCall: llms.CallOptions{Temperature: 0.4, JSONMode: true}}, ProjectData{UserName: "name", ProjName: "proj"}, "", ProjectData{UserName: "name", ProjName: "proj"}},
		{&Options{ProjectsRoot: "root"}, ProjectData{UserName: "name", ProjName: "proj"}, "root", ProjectData{UserName: "name", ProjName: "proj"}},
	}
	ai := &mockLLM{}
	for i, v := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			ada := New(ai, WithOptions(*v.options))
			ada.AddProjectData(v.projData)
			assert.Equal(t, *v.options, ada.Opts)
			assert.Equal(t, v.expectedRoot, ada.Opts.ProjectsRoot)
			assert.Equal(t, ada.Projdata.UserName, v.expectedPD.UserName)
			assert.Contains(t, ada.Projdata.ProjName, v.expectedPD.ProjName)
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
	ada := New(m, WithOptions(Options{Timeout: "1m"}))
	res, err := ada.GenerateWithSys(context.Background(), "main", sysp, usp)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "some response", res.Choices[0].Content)
}

func TestGenerateJSON(t *testing.T) {
	t.Parallel()
	tests := []struct {
		cfg      Options
		hasError bool
		mockFunc func(...any) (*llms.ContentResponse, error)
	}{
		{Options{Timeout: "1m"}, false, defaultMock},
		{Options{Timeout: "invalid"}, true, func(...any) (*llms.ContentResponse, error) {
			return nil, errors.New("context deadline exceeded")
		}},
	}

	for i, v := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			ai := &mockLLM{v.mockFunc}
			ada := New(ai, WithOptions(v.cfg))
			res, err := ada.GenerateContent(context.Background(), "main", "test prompt", []llms.MessageContent{})
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
			ada := New(&mockLLM{}, WithOptions(Options{PromptsRoot: tempdir}))
			if !v.hasErr {
				err := os.WriteFile(filepath.Join(tempdir, file), []byte(v.template), 0644)
				require.NoError(t, err)
			}
			res, err := ada.PromptFromTemplate(file, v.args...)
			assert.Equal(t, v.hasErr, err != nil)
			assert.Equal(t, fmt.Sprintf(v.template, v.args...), res)
		})
	}
}

func TestResolveProjectPath(t *testing.T) {
	t.Parallel()
	tests := []struct {
		projRoot string
		username string
		projname string
	}{
		{"projroot", "username", "projname"},
	}

	for i, v := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			cfg := Options{ProjectsRoot: v.projRoot}
			ai := &mockLLM{}
			ada := New(ai, WithOptions(cfg))
			ada.AddProjectData(ProjectData{UserName: v.username, ProjName: v.projname})
			res := ada.ResolveProjectPath()
			assert.Contains(t, res, v.username)
			assert.Contains(t, res, v.projRoot)
			assert.Contains(t, res, v.projname)
		})
	}
}

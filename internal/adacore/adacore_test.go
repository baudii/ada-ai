package adacore

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
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
		{&Options{ProjectsRoot: "root"}, ProjectData{"unknown_user", "project_"}, "root", ProjectData{"unknown_user", "project_"}},
		{&Options{Timeout: "a", Reflection: ReflectConfig{1, 2}, ModelCallOpts: llms.CallOptions{Temperature: 0.4, JSONMode: true}}, ProjectData{"name", "proj"}, "", ProjectData{"name", "proj"}},
		{&Options{ProjectsRoot: "root"}, ProjectData{"name", "proj"}, "root", ProjectData{"name", "proj"}},
	}
	ai := &mockLLM{}
	for i, v := range tests {
		ada := New(ai, WithOptions(*v.options))
		ada.AddProjectData(v.projData)
		assert.Nil(t, ada.Proj)
		if v.options == nil {
			assert.Equal(t, options, ada.Opts, "test: %v", i)
		} else {
			assert.Equal(t, *v.options, ada.Opts, "test: %v", i)
		}
		assert.Equal(t, v.expectedRoot, ada.Opts.ProjectsRoot, "test: %v", i)
		assert.Equal(t, ada.Project.UserName, v.expectedPD.UserName, "test: %v", i)
		assert.Contains(t, ada.Project.ProjName, v.expectedPD.ProjName, "test: %v", i)
	}
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
		ai := &mockLLM{v.mockFunc}
		ada := New(ai, WithOptions(v.cfg))
		res, err := ada.GenerateContent("test prompt", []llms.MessageContent{})
		assert.Equal(t, v.hasError, err != nil, "test: %v", i)
		if !v.hasError {
			assert.NotNil(t, res, "test: %v", i)
			assert.Equal(t, "some response", res.Choices[0].Content, "test: %v", i)
		}
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
		tempdir := t.TempDir()
		ada := New(&mockLLM{}, WithOptions(Options{PromptsRoot: tempdir}))
		if !v.hasErr {
			err := os.WriteFile(filepath.Join(tempdir, file), []byte(v.template), 0644)
			require.NoError(t, err, "test: %v", i)
		}
		res, err := ada.PromptFromTemplate(file, v.args...)
		assert.Equal(t, v.hasErr, err != nil, "test: %v", i)
		assert.Equal(t, fmt.Sprintf(v.template, v.args...), res, "test: %v", i)
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
		cfg := Options{ProjectsRoot: v.projRoot}
		ai := &mockLLM{}
		ada := New(ai, WithOptions(cfg))
		ada.AddProjectData(ProjectData{UserName: v.username, ProjName: v.projname})
		res := ada.ResolveProjectPath()
		assert.Contains(t, res, v.username, "test: %v", i)
		assert.Contains(t, res, v.projRoot, "test: %v", i)
		assert.Contains(t, res, v.projname, "test: %v", i)
	}
}

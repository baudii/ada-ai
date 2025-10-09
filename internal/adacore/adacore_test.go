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
		cfg          *Config
		opts         []SessionOption
		expectedRoot string
		expectedPD   ProjectData
	}{
		{nil, []SessionOption{WithProjectsRoot("root")}, "root", ProjectData{}},
		{&Config{"a", ReflectConfig{1, 2}}, []SessionOption{WithProjectData(ProjectData{"name", "proj"})}, "", ProjectData{"name", "proj"}},
		{&Config{}, []SessionOption{WithProjectsRoot("root"), WithProjectData(ProjectData{"name", "proj"})}, "root", ProjectData{"name", "proj"}},
	}
	ai := &mockLLM{}
	for _, v := range tests {
		ada := New(ai, v.cfg, v.opts...)
		if v.cfg == nil {
			assert.Equal(t, &defaultCfg, ada.Cfg, "test: %v", v)
		} else {
			assert.Equal(t, v.cfg, ada.Cfg, "test: %v", v)
		}
		assert.Equal(t, v.expectedRoot, ada.Session.projectsRoot, "test: %v", v)
		assert.Equal(t, v.expectedPD, ada.Session.Project, "test: %v", v)
	}
}

func TestGenerateJSON(t *testing.T) {
	t.Parallel()
	tests := []struct {
		cfg      Config
		hasError bool
		mockFunc func(...any) (*llms.ContentResponse, error)
	}{
		{Config{Timeout: "1m"}, false, defaultMock},
		{Config{Timeout: "invalid"}, true, func(...any) (*llms.ContentResponse, error) {
			return nil, errors.New("context deadline exceeded")
		}},
	}

	for _, v := range tests {
		ai := &mockLLM{v.mockFunc}
		ada := New(ai, &v.cfg)
		res, err := ada.GenerateJSON("test prompt", []llms.MessageContent{})
		assert.Equal(t, v.hasError, err != nil, "test: %v", v)
		if !v.hasError {
			assert.NotNil(t, res, "test: %v", v)
			assert.Equal(t, "some response", res.Choices[0].Content, "test: %v", v)
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

	for _, v := range tests {
		tempdir := t.TempDir()
		ada := New(&mockLLM{}, nil, WithPromptsRoot(tempdir))
		if !v.hasErr {
			err := os.WriteFile(filepath.Join(tempdir, file), []byte(v.template), 0644)
			require.NoError(t, err, "test: %v", v)
		}
		res, err := ada.PromptFromTemplate(file, v.args...)
		assert.Equal(t, v.hasErr, err != nil, "test: %v", v)
		assert.Equal(t, fmt.Sprintf(v.template, v.args...), res, "test: %v", v)
	}
}

func TestResolveProjectPath(t *testing.T) {
	// TODO: Add tests
	t.Parallel()
	tests := []struct {
		root     string
		projRoot string
		username string
		projname string
	}{
		{},
	}

	for _, v := range tests {
		cfg := &Config{}
		ai := &mockLLM{}
		ada := New(ai, cfg, WithProjectsRoot(v.root), WithProjectData(ProjectData{UserName: v.username, ProjName: v.projname}))
		res := ada.ResolveProjectPath()
		assert.Contains(t, res, v.root, "test: %v", v)
		assert.Contains(t, res, v.username, "test: %v", v)
		assert.Contains(t, res, v.projRoot, "test: %v", v)
		assert.Contains(t, res, v.projname, "test: %v", v)
	}
}

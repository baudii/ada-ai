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
	generate func() (*llms.ContentResponse, error)
}

var defaultMock = func() (*llms.ContentResponse, error) {
	return &llms.ContentResponse{Choices: []*llms.ContentChoice{{Content: "some response"}}}, nil
}

func (m *mockLLM) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	return prompt, nil
}

func (m *mockLLM) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	return m.generate()
}

func TestNew(t *testing.T) {
	t.Parallel()
	tests := []struct {
		cfg          *Config
		opts         []SessionOption
		expectedRoot string
		expectedPD   ProjectData
	}{
		{nil, []SessionOption{WithRoot("root")}, "root", ProjectData{}},
		{&Config{"a", "b", ReflectConfig{1, 2}}, []SessionOption{WithProjectData(ProjectData{"name", "proj"})}, "", ProjectData{"name", "proj"}},
		{&Config{}, []SessionOption{WithRoot("root"), WithProjectData(ProjectData{"name", "proj"})}, "root", ProjectData{"name", "proj"}},
	}
	ai := &mockLLM{}
	for _, v := range tests {
		ada := New(ai, v.cfg, v.opts...)
		if v.cfg == nil {
			assert.Equal(t, &defaultCfg, ada.Cfg)
		} else {
			assert.Equal(t, v.cfg, ada.Cfg)
		}
		assert.Equal(t, v.expectedRoot, ada.Session.root)
		assert.Equal(t, v.expectedPD, ada.Session.Project)
	}
}

func TestGenerateJSON(t *testing.T) {
	tests := []struct {
		cfg      Config
		hasError bool
		mockFunc func() (*llms.ContentResponse, error)
	}{
		{Config{Timeout: "1m"}, false, defaultMock},
		{Config{Timeout: "invalid"}, true, func() (*llms.ContentResponse, error) {
			return nil, errors.New("context deadline exceeded")
		}},
	}

	for _, v := range tests {
		ai := &mockLLM{v.mockFunc}
		ada := New(ai, &v.cfg)
		res, err := ada.GenerateJSON("test prompt", []llms.MessageContent{})
		assert.Equal(t, v.hasError, err != nil)
		if !v.hasError {
			assert.NotNil(t, res)
			assert.Equal(t, "some response", res.Choices[0].Content)
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
		ada := New(&mockLLM{}, nil, WithRoot(tempdir))
		promptsDir := filepath.Join(tempdir, promptsFolder)
		promptFile := filepath.Join(promptsDir, file)
		if !v.hasErr {
			err := os.MkdirAll(promptsDir, 0755)
			require.NoError(t, err)
			err = os.WriteFile(promptFile, []byte(v.template), 0644)
			require.NoError(t, err)
		}
		res, err := ada.PromptFromTemplate(file, v.args...)
		assert.Equal(t, v.hasErr, err != nil)
		assert.Equal(t, fmt.Sprintf(v.template, v.args...), res)
	}
}

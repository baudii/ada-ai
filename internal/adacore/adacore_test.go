package adacore

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/baudii/ada-ai/internal/common"
	"github.com/baudii/ada-ai/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/llms"
)

type mockLLM struct{}

var generateContentMock = func() (*llms.ContentResponse, error) {
	return &llms.ContentResponse{Choices: []*llms.ContentChoice{{Content: "some response"}}}, nil
}

func (m *mockLLM) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	return prompt, nil
}

func (m *mockLLM) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	return generateContentMock()
}

func TestNew(t *testing.T) {
	origDataPath := userDataPath
	t.Cleanup(func() {
		userDataPath = origDataPath
	})

	tests := []struct {
		creatUserData bool
	}{
		{true},
		{false},
	}
	for _, v := range tests {
		userDataPath = filepath.Join(t.TempDir(), "user_data.json")
		if v.creatUserData {
			os.WriteFile(userDataPath, []byte(`{ "userName": "mockedun", "projName": "mockedpn" }`), 0644)
		}
		ai := &mockLLM{}
		cfg := &Config{}
		ada := New(ai, cfg)
		assert.Equal(t, ai, ada.ai)
		assert.Equal(t, cfg, ada.cfg)
		if v.creatUserData {
			assert.Equal(t, ada.Ctx.UserName, "mockedun")
			assert.Equal(t, ada.Ctx.ProjName, "mockedpn")
		} else {
			assert.Nil(t, ada.Ctx)
		}
	}
}

func TestAddProjCtx(t *testing.T) {
	ai := &mockLLM{}
	cfg := &Config{}
	ada := New(ai, cfg)
	ada.AddProjCtx("testuser", "testproj")
	assert.NotNil(t, ada.Ctx)
	assert.Equal(t, "testuser", ada.Ctx.UserName)
	assert.Equal(t, "testproj", ada.Ctx.ProjName)
}

func TestSaveCtx(t *testing.T) {
	origDataPath := userDataPath
	t.Cleanup(func() {
		userDataPath = origDataPath
	})
	tests := []struct {
		add bool
	}{
		{true},
		{false},
	}
	for _, v := range tests {
		userDataPath = filepath.Join(t.TempDir(), "user_data.json")
		ai := &mockLLM{}
		cfg := &Config{}
		ada := New(ai, cfg)
		if v.add {
			ada.AddProjCtx("testuser", "testproj")
		}
		err := ada.SaveCtx()
		if v.add {
			assert.NoError(t, err)
			savedCtx, err := utils.ParseJSONFile[projectContext](userDataPath)
			assert.NoError(t, err)
			assert.Equal(t, "testuser", savedCtx.UserName)
			assert.Equal(t, "testproj", savedCtx.ProjName)
		} else {
			assert.Error(t, err)
		}
	}
}

func TestGenerateJSON(t *testing.T) {
	origfunc := generateContentMock
	t.Cleanup(func() {
		generateContentMock = origfunc
	})

	tests := []struct {
		cfg      Config
		fails    bool
		mockFunc func() (*llms.ContentResponse, error)
	}{
		{Config{Timeout: "1m"}, false, generateContentMock},
		{Config{Timeout: "invalid"}, true, func() (*llms.ContentResponse, error) {
			return nil, errors.New("context deadline exceeded")
		}},
	}

	for _, v := range tests {
		ai := &mockLLM{}
		ada := New(ai, &v.cfg)
		generateContentMock = v.mockFunc
		res, err := ada.GenerateJSON("test prompt", []llms.MessageContent{})
		assert.Equal(t, v.fails, err != nil)
		if !v.fails {
			assert.NotNil(t, res)
			assert.Equal(t, "some response", res.Choices[0].Content)
		}
	}
}

func TestGetPromptFromTemplate(t *testing.T) {
	o := utils.Executable
	t.Cleanup(func() {
		utils.Executable = o
	})

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

	for _, v := range tests {
		filename := "test_template.txt"
		tempdir := t.TempDir()
		promptsDir := filepath.Join(tempdir, common.PromptsPath)
		utils.Executable = func() (string, error) {
			return filepath.Join(tempdir, "mocked.exe"), nil
		}
		if !v.hasErr {
			err := os.MkdirAll(promptsDir, 0755)
			require.NoError(t, err)
			err = os.WriteFile(filepath.Join(promptsDir, filename), []byte(v.template), 0644)
			require.NoError(t, err)
		}
		res, err := PromptFromTemplate(promptTemplate(filename), v.args...)
		assert.Equal(t, v.hasErr, err != nil)
		assert.Equal(t, fmt.Sprintf(v.template, v.args...), res)
	}
}

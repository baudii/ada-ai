package app_test

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/baudii/ada-ai/internal/app"
	"github.com/baudii/ada-ai/internal/project"
	"github.com/stretchr/testify/assert"
	"github.com/tmc/langchaingo/llms"
)

// mockMaterializer is a mock implementation of a materializer for testing purposes.
type mockMaterializer struct {
	fileArg string
	err     error
}

func (m *mockMaterializer) Materialize(hfile, hfold project.Handler) error {
	if m.err != nil {
		return m.err
	}
	return hfile(m.fileArg)
}

// mockNavigator is a mock implementation of a navigation store for testing purposes.
type mockNavigator struct {
	loadNavErr error
	addNavErr  error
	contentErr error
}

func (m *mockNavigator) LoadNav(key string) ([]byte, error) {
	return nil, m.loadNavErr
}
func (m *mockNavigator) AddNav(key, ext string, content []byte) error {
	return m.addNavErr
}
func (m *mockNavigator) Content(key string) (string, error) {
	return "", m.contentErr
}

// mockGen is a mock implementation of a generator for testing purposes.
type mockGen struct {
	promptStr     string
	promptErr     error
	resp          *llms.ContentResponse
	genErr        error
	failCondition func(string) bool
}

func (m *mockGen) BuildPrompt(p string, args ...any) (string, error) {
	if m.failCondition != nil && m.failCondition(p) {
		return "", assert.AnError
	}
	return m.promptStr, m.promptErr
}

func (m *mockGen) GenerateWithSys(ctx context.Context, sys, user string, callOptions ...llms.CallOption) (*llms.ContentResponse, error) {
	return m.resp, m.genErr
}

func TestRun_Success(t *testing.T) {
	t.Parallel()
	a := app.New(
		app.WithGen(&mockGen{resp: &llms.ContentResponse{Choices: []*llms.ContentChoice{{Content: "generated content"}}}}),
		app.WithProjectData(app.ProjectData{}),
		app.WithNavigator(&mockNavigator{loadNavErr: assert.AnError}),
		app.WithNavNames([]string{project.Structure}),
		app.WithMaterializer(&mockMaterializer{fileArg: filepath.Join(t.TempDir(), "file.txt")}),
	)
	err := a.Run(context.Background())
	assert.NoError(t, err)
}

func TestRun_Fails(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		addNavErr   error
		matErr      error
		expectedErr string
	}{
		{
			name:        "fail: add nav",
			addNavErr:   assert.AnError,
			expectedErr: "add nav file",
		},
		{
			name:        "fail: add nav, fail materialize",
			addNavErr:   assert.AnError,
			matErr:      assert.AnError,
			expectedErr: "add nav file",
		},
		{
			name:        "fail: materialize",
			matErr:      assert.AnError,
			expectedErr: "materialize project",
		},
	}

	for _, v := range tests {
		t.Run(v.name, func(t *testing.T) {
			t.Parallel()
			a := app.New(
				app.WithGen(&mockGen{resp: &llms.ContentResponse{Choices: []*llms.ContentChoice{{Content: "generated content"}}}}),
				app.WithNavigator(&mockNavigator{addNavErr: v.addNavErr}),
				app.WithNavNames([]string{"nav1"}),
				app.WithMaterializer(&mockMaterializer{fileArg: filepath.Join(t.TempDir(), "file.txt"), err: v.matErr}),
			)
			err := a.Run(context.Background())
			if v.expectedErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), v.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAddNavs_Fails(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		navName     string
		contentErr  error
		expectedErr string
	}{
		{
			name:        "fail: unknown nav name",
			navName:     "unknown_nav",
			expectedErr: "no args for nav",
		},
		{
			name:        "fail: content error",
			navName:     project.Structure,
			contentErr:  assert.AnError,
			expectedErr: "inject nav content",
		},
	}

	for _, v := range tests {
		t.Run(v.name, func(t *testing.T) {
			t.Parallel()
			a := app.New(
				app.WithGen(&mockGen{resp: &llms.ContentResponse{Choices: []*llms.ContentChoice{{Content: "generated content"}}}}),
				app.WithNavigator(&mockNavigator{contentErr: v.contentErr, loadNavErr: assert.AnError}),
				app.WithNavNames([]string{v.navName}),
			)
			err := a.AddNavs(context.Background())
			if v.expectedErr != "" {
				assert.Error(t, err)
				assert.ErrorContains(t, err, v.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMaterializeProject(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		contentErr    error
		failCondition func(string) bool
		genErr        error
		expectedErr   string
	}{
		{
			name:        "fail: content",
			contentErr:  assert.AnError,
			expectedErr: "inject nav content",
		},
		{
			name:          "fail: build prompt system",
			failCondition: func(p string) bool { return strings.Contains(p, "system") },
			expectedErr:   "system and human prompts",
		},
		{
			name:          "fail: build prompt human",
			failCondition: func(p string) bool { return strings.Contains(p, "human") },
			expectedErr:   "system and human prompts",
		},
		{
			name:        "fail: generate with sys",
			genErr:      assert.AnError,
			expectedErr: "generate with sys",
		},
		{
			name:        "fail: write file",
			expectedErr: "write file",
		},
	}

	for _, v := range tests {
		t.Run(v.name, func(t *testing.T) {
			t.Parallel()
			a := app.New(
				app.WithGen(&mockGen{
					resp:          &llms.ContentResponse{Choices: []*llms.ContentChoice{{Content: "generated content"}}},
					failCondition: v.failCondition,
					genErr:        v.genErr,
				}),
				app.WithNavigator(&mockNavigator{contentErr: v.contentErr}),
				app.WithNavNames([]string{"nav1"}),
				app.WithMaterializer(&mockMaterializer{fileArg: t.TempDir()}),
			)
			err := a.MaterializeProject(context.Background())
			assert.ErrorContains(t, err, v.expectedErr)
		})
	}
}

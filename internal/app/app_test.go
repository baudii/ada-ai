package app_test

import (
	"context"
	"io"
	"log/slog"
	"path/filepath"
	"strings"
	"testing"

	"github.com/baudii/ada-ai/internal/app"
	"github.com/baudii/ada-ai/internal/core/gen"
	"github.com/baudii/ada-ai/internal/core/project"
	"github.com/stretchr/testify/assert"
)

func MockLogger(w io.Writer) *slog.Logger {
	return slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{}))
}

// mockMaterializer is a mock implementation of a materializer for testing purposes.
type mockMaterializer struct {
	fileArg string
	err     error
	fileErr error
}

func (m *mockMaterializer) Materialize(hfile project.FileHandler, hfold project.FolderHandler) error {
	if m.err != nil {
		return m.err
	}
	if hfile != nil {
		return hfile(m.fileArg)
	}
	return nil
}

func (m *mockMaterializer) HandleFile(name string, content []byte) error {
	return m.fileErr
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
func (m *mockNavigator) NavContent(key string) (string, error) {
	return "", m.contentErr
}

// mockGen is a mock implementation of a generator for testing purposes.
type mockGen struct {
	promptStr     string
	promptErr     error
	resp          string
	genErr        error
	failCondition func(string) bool
}

func (m *mockGen) BuildPrompt(p string, args ...any) (string, error) {
	if m.failCondition != nil && m.failCondition(p) {
		return "", assert.AnError
	}
	return m.promptStr, m.promptErr
}

func (m *mockGen) GenerateWithSys(ctx context.Context, sys, user string, opts ...gen.Option) (string, error) {
	return m.resp, m.genErr
}

func TestRun_Success(t *testing.T) {
	t.Parallel()
	a := app.New(
		app.WithGenerator(&mockGen{resp: "generated content"}),
		app.WithProjectData(project.Context{}),
		app.WithNavigator(&mockNavigator{loadNavErr: assert.AnError}),
		app.WithNavNames([]string{project.Structure}),
		app.WithMaterializer(&mockMaterializer{fileArg: filepath.Join(t.TempDir(), "file.txt")}),
		app.WithLogger(MockLogger(io.Discard)),
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
				app.WithGenerator(&mockGen{resp: "generated content"}),
				app.WithNavigator(&mockNavigator{addNavErr: v.addNavErr}),
				app.WithNavNames([]string{"nav1"}),
				app.WithMaterializer(&mockMaterializer{fileArg: filepath.Join(t.TempDir(), "file.txt"), err: v.matErr}),
				app.WithLogger(MockLogger(io.Discard)),
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
				app.WithGenerator(&mockGen{resp: "generated content"}),
				app.WithNavigator(&mockNavigator{contentErr: v.contentErr, loadNavErr: assert.AnError}),
				app.WithNavNames([]string{v.navName}),
				app.WithLogger(MockLogger(io.Discard)),
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
		fileErr       error
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
			fileErr:     assert.AnError,
		},
	}

	for _, v := range tests {
		t.Run(v.name, func(t *testing.T) {
			t.Parallel()
			a := app.New(
				app.WithGenerator(&mockGen{
					resp:          "generated content",
					failCondition: v.failCondition,
					genErr:        v.genErr,
				}),
				app.WithNavigator(&mockNavigator{contentErr: v.contentErr}),
				app.WithNavNames([]string{"nav1"}),
				app.WithMaterializer(&mockMaterializer{fileArg: t.TempDir(), fileErr: v.fileErr}),
				app.WithLogger(MockLogger(io.Discard)),
			)
			err := a.MaterializeProject(context.Background())
			assert.ErrorContains(t, err, v.expectedErr)
		})
	}
}

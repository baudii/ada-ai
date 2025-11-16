package app_test

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/baudii/ada-ai/internal/app"
	"github.com/baudii/ada-ai/internal/core/gen"
	"github.com/baudii/ada-ai/internal/core/project"
	"github.com/stretchr/testify/assert"
)

func MockLogger(w io.Writer) *slog.Logger {
	return slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{}))
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
		app.WithNavNames([]string{project.Structure}),
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
				app.WithNavNames([]string{"nav1"}),
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

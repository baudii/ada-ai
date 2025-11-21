package app_test

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/baudii/ada-ai/internal/app"
	"github.com/baudii/ada-ai/internal/core/gen"
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

type mockMaterializer struct {
	matErr     error
	prepareErr error
}

func (m *mockMaterializer) PrepareOutputDir() error {
	return m.prepareErr
}

func (m *mockMaterializer) Materialize(ctx context.Context, openapi string) error {
	return m.matErr
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

func TestGenerateProject_Success(t *testing.T) {
	t.Parallel()
	a := app.New(
		app.WithGenerator(&mockGen{resp: "generated content"}),
		app.WithDegree(1),
		app.WithLogger(MockLogger(io.Discard)),
	)
	err := a.GenerateProject(context.Background(), &mockMaterializer{})
	assert.NoError(t, err)
}

func TestRun_Fails(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		prepareErr  error
		matErr      error
		expectedErr string
	}{
		{
			name:        "fail: prepare output dir",
			prepareErr:  assert.AnError,
			expectedErr: "prepare output dir",
		},
		{
			name:        "fail: prepare output dir, fail materialize",
			prepareErr:  assert.AnError,
			matErr:      assert.AnError,
			expectedErr: "prepare output dir",
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
				app.WithLogger(MockLogger(io.Discard)),
			)
			err := a.GenerateProject(context.Background(), &mockMaterializer{
				matErr:     v.matErr,
				prepareErr: v.prepareErr,
			})
			if v.expectedErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), v.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

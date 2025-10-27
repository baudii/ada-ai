package cli

import (
	"bytes"
	"io"
	"log/slog"
	"testing"

	"github.com/baudii/ada-ai/internal/core/project"
	"github.com/stretchr/testify/assert"
)

func MockLogger(b io.Writer) *slog.Logger {
	return slog.New(slog.NewTextHandler(b, &slog.HandlerOptions{}))
}

func TestNew(t *testing.T) {
	t.Parallel()
	c := New(WithLogger(nil))
	assert.NotNil(t, c)
}

func TestProjData(t *testing.T) {
	t.Parallel()
	expected := project.Context{
		UserName: "testuser",
		Name:     "testproject",
		Language: "go",
		Summary:  "This is a test project.",
	}
	c := cliApp{read: func(prompt string) string {
		switch prompt {
		case "Provide nickname":
			return expected.UserName
		case "Provide project name":
			return expected.Name
		case "Provide programming language (go, python, js, etc)":
			return expected.Language
		case "Provide a short summary of the project":
			return expected.Summary
		default:
			return ""
		}
	}, save: func(any, string) error {
		return nil
	}, logger: MockLogger(io.Discard)}
	pd := c.GetProjectData(t.TempDir())
	assert.Equal(t, expected, pd)
}

func TestProjData_FailSave(t *testing.T) {
	t.Parallel()
	b := &bytes.Buffer{}
	c := cliApp{
		read: func(prompt string) string {
			return "test"
		},
		save: func(any, string) error {
			return assert.AnError
		},
		logger: MockLogger(b),
	}
	pd := c.GetProjectData(t.TempDir())
	assert.NotNil(t, pd)
	assert.NotEmpty(t, b.String())
}

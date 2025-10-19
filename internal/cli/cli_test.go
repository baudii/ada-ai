package cli

import (
	"testing"

	"github.com/baudii/ada-ai/internal/app"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Parallel()
	c := New()
	assert.NotNil(t, c)
	assert.Implements(t, (*app.Runner)(nil), c)
}

func TestProjData(t *testing.T) {
	t.Parallel()
	expected := app.ProjectData{
		UserName: "testuser",
		ProjName: "testproject",
		Language: "go",
		Summary:  "This is a test project.",
	}
	c := cliApp{read: func(prompt string) string {
		switch prompt {
		case "Provide nickname":
			return expected.UserName
		case "Provide project name":
			return expected.ProjName
		case "Provide programming language (go, python, js, etc)":
			return expected.Language
		case "Provide a short summary of the project":
			return expected.Summary
		default:
			return ""
		}
	}}
	pd := c.ReadProjdata()
	assert.Equal(t, expected, pd)
}

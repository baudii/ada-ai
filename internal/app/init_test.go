package app_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/baudii/ada-ai/internal/ada"
	"github.com/baudii/ada-ai/internal/app"
	"github.com/baudii/ada-ai/internal/common"
	"github.com/baudii/ada-ai/internal/project/folder"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Parallel()
	a := app.New(
		app.WithDegree(5),
		app.WithProject(nil),
		app.WithMode(folder.CreateNew),
		app.WithGen(nil),
		app.WithDirProvider(nil),
		app.WithProjectData(app.ProjectData{}),
		app.WithOptions(&app.Options{}),
		app.WithNavNames([]string{"a", "b"}),
		app.WithNavHandler(nil),
	)
	require.NotNil(t, a)
}

func TestParseAppOptions(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		filename string
		contents string
		err      string
		expected *app.Options
	}{
		{"non-existent path", "non-existent-path", `
			{
				"requestTimeout": "5s",
				"projectsRoot": "testdata/projects",
				"promptsRoot": "testdata/prompts",
				"reflection": {
					"depth": 1,
					"threshhold": 0.4
				}
			}`, "parse ada options", &app.Options{
			Timeout:      "5s",
			ProjectsRoot: "testdata/projects",
			PromptsRoot:  "testdata/prompts",
			Reflection:   ada.ReflectConfig{Depth: 1, Threshhold: 0.4},
		}},
		{"valid config", app.ConfigFile, `{}`, "", &app.Options{
			ProjectsRoot: common.ProjectsPath,
			PromptsRoot:  common.PromptsPath,
		}},
	}

	for _, v := range tests {
		t.Run(v.name, func(t *testing.T) {
			t.Parallel()
			d := t.TempDir()
			err := os.WriteFile(filepath.Join(d, v.filename), []byte(v.contents), 0o644)
			require.NoError(t, err)
			opts, err := app.ParseAppOptions(d)
			if v.err != "" {
				assert.ErrorContains(t, err, v.err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, v.expected, opts)
			}
		})
	}
}

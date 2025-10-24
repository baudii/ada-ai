package app_test

import (
	"testing"

	"github.com/baudii/ada-ai/internal/app"
	"github.com/baudii/ada-ai/internal/core/project"
	"github.com/baudii/ada-ai/internal/core/seqdir"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Parallel()
	a := app.New(
		app.WithDegree(5),
		app.WithProject(nil),
		app.WithMode(seqdir.CreateNew),
		app.WithGenerator(nil),
		app.WithProjectData(project.Data{}),
		app.WithOptions(app.Options{}),
		app.WithNavNames([]string{"a", "b"}),
		app.WithNavigator(nil),
		app.WithMaterializer(nil),
	)
	require.NotNil(t, a)
}

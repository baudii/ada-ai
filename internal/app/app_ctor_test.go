package app_test

import (
	"testing"

	"github.com/baudii/ada-ai/internal/app"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Parallel()
	a := app.New(
		app.WithDegree(5),
		app.WithGenerator(nil),
		app.WithOptions(app.Options{}),
		app.WithLogger(nil),
	)
	require.NotNil(t, a)
}

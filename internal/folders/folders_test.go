package folders

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	t.Parallel()
	assert.True(t, filepath.IsAbs(Artifacts), "expected %q to be absolute path", Artifacts)
}

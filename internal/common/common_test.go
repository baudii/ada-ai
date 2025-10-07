package common

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	t.Parallel()
	assert.True(t, filepath.IsAbs(Artifacts))
}

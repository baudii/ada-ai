package adacore

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestError(t *testing.T) {
	t.Parallel()
	tests := []struct {
		err      error
		expected string
	}{
		{&ProjExist{Path: "/some/path"}, "project already exists at /some/path"},
	}
	for _, v := range tests {
		assert.Equal(t, v.expected, v.err.Error(), "test: %v", v)
	}
}

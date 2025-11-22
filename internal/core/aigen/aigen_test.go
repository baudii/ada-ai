package aigen_test

import (
	"testing"

	"github.com/baudii/ada-ai/internal/core/aigen"
	"github.com/stretchr/testify/assert"
)

func TestNewParams(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		opts     []aigen.Option
		expected *aigen.Params
	}{
		{
			name:     "no options",
			opts:     []aigen.Option{},
			expected: &aigen.Params{},
		},
		{
			name:     "with temperature",
			opts:     []aigen.Option{aigen.WithTemperature(0.5)},
			expected: &aigen.Params{Temperature: floatPtr(0.5)},
		},
		{
			name:     "with JSON mode",
			opts:     []aigen.Option{aigen.WithJSONMode()},
			expected: &aigen.Params{JSONMode: boolPtr(true)},
		},
		{
			name:     "with both options",
			opts:     []aigen.Option{aigen.WithTemperature(0.7), aigen.WithJSONMode()},
			expected: &aigen.Params{Temperature: floatPtr(0.7), JSONMode: boolPtr(true)},
		},
	}
	for _, v := range tests {
		t.Run(v.name, func(t *testing.T) {
			got := &aigen.Params{}
			for _, opt := range v.opts {
				opt(got)
			}

			assert.Equal(t, v.expected, got)
		})
	}
}

func floatPtr(f float64) *float64 {
	return &f
}

func boolPtr(b bool) *bool {
	return &b
}

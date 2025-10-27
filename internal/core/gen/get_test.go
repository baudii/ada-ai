package gen_test

import (
	"testing"

	"github.com/baudii/ada-ai/internal/core/gen"
	"github.com/stretchr/testify/assert"
)

func TestNewParams(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		opts     []gen.Option
		expected *gen.Params
	}{
		{
			name:     "no options",
			opts:     []gen.Option{},
			expected: &gen.Params{},
		},
		{
			name:     "with temperature",
			opts:     []gen.Option{gen.WithTemperature(0.5)},
			expected: &gen.Params{Temperature: floatPtr(0.5)},
		},
		{
			name:     "with JSON mode",
			opts:     []gen.Option{gen.WithJSONMode()},
			expected: &gen.Params{JSONMode: boolPtr(true)},
		},
		{
			name:     "with both options",
			opts:     []gen.Option{gen.WithTemperature(0.7), gen.WithJSONMode()},
			expected: &gen.Params{Temperature: floatPtr(0.7), JSONMode: boolPtr(true)},
		},
	}
	for _, v := range tests {
		t.Run(v.name, func(t *testing.T) {
			got := &gen.Params{}
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

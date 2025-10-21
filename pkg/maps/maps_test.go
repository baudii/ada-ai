package maps_test

import (
	"strconv"
	"testing"

	"github.com/baudii/ada-ai/pkg/maps"
	"github.com/stretchr/testify/assert"
)

func TestMerge(t *testing.T) {
	t.Parallel()
	tests := []struct {
		dst      map[string]any
		src      map[string]any
		expected map[string]any
	}{
		{nil, map[string]any{"a": nil}, nil},
		{map[string]any{"b": "s"}, nil, map[string]any{"b": "s"}},
		{map[string]any{"a": "s"}, map[string]any{"a": nil}, map[string]any{"a": nil}},
		{map[string]any{"a": "s"}, map[string]any{"b": nil}, map[string]any{"a": "s", "b": nil}},
		{
			map[string]any{
				"nil1": nil,
				"int":  12,
				"e": map[string]any{
					"nil2":     nil,
					"ovwrtSt":  "st",
					"ovwrtInt": 12,
				},
			},
			map[string]any{"e": map[string]any{
				"nil2":     "not nil",
				"ovwrtSt":  12,
				"ovwrtInt": "st",
				"new":      "new",
			}},
			map[string]any{
				"nil1": nil,
				"int":  12,
				"e": map[string]any{
					"nil2":     "not nil",
					"ovwrtSt":  12,
					"ovwrtInt": "st",
					"new":      "new",
				},
			},
		},
	}

	for i, v := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			maps.Merge(v.dst, v.src)
			assert.Equal(t, v.dst, v.expected)
		})
	}
}

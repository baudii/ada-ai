package markdown_test

import (
	"strconv"
	"testing"

	"github.com/baudii/ada-ai/pkg/markdown"
	"github.com/stretchr/testify/assert"
)

func TestUlistMd(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    []any
		expected string
	}{
		{[]any{"one"}, " * one"},
		{[]any{"one", 1, true}, " * one\n * 1\n * true"},
		{[]any{}, ""},
		{nil, ""},
	}
	for i, v := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			assert.Equal(t, v.expected, markdown.Ulist(v.input...))
		})
	}
}

func TestStringify(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    any
		prefix   string
		expected string
		panics   bool
	}{
		{
			name: "struct with fields",
			input: struct {
				A string
				B int
				C bool
				D float32
				E []int
			}{"foo", 42, true, 3.14, []int{1, 2, 3}},
			prefix:   "- ",
			expected: "- A: foo\n- B: 42\n- C: true\n- D: 3.14\n- E: [1 2 3]",
		},
		{
			name:   "fail: not struct",
			input:  []int{1, 2, 3},
			prefix: "- ",
			panics: true,
		},
	}

	for i, v := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			if v.panics {
				assert.Panics(t, func() { markdown.Stringify(v.input, v.prefix) })
			} else {
				assert.Equal(t, v.expected, markdown.Stringify(v.input, v.prefix))
			}
		})
	}
}

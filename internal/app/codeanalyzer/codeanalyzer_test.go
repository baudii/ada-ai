package codeanalyzer_test

import (
	"go/ast"
	"testing"

	"github.com/baudii/ada-ai/internal/app/codeanalyzer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var src = `package demo

import (
	"go/ast"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "path/to/your/module"
)

// Some documentation comment
type A interface {
	// Some method
	Foo()
}

// Some other comment
type B struct {
	// Actual doc comment
	X int // X coordinate
	Y string
}

type C interface {
	Bar()
}

type D int
`

func TestExtractTypesOf(t *testing.T) {

	tests := []struct {
		name       string
		kind       string
		nameFilter string
		expected   string
	}{
		{
			name: "interfaces only",
			kind: "interface",
			expected: `// Some documentation comment
type A interface {
	// Some method
	Foo()
}

type C interface {
	Bar()
}`,
		},
		{
			name: "structs only",
			kind: "struct",
			expected: `// Some other comment
type B struct {
	// Actual doc comment
	X int // X coordinate
	Y string
}`,
		},
		{
			name: "all types",
			kind: "all",
			expected: `// Some documentation comment
type A interface {
	// Some method
	Foo()
}

// Some other comment
type B struct {
	// Actual doc comment
	X int // X coordinate
	Y string
}

type C interface {
	Bar()
}

type D int`,
		},
		{
			name:       "filter by name",
			kind:       "all",
			nameFilter: "C",
			expected: `type C interface {
	Bar()
}`,
		},
	}

	for _, v := range tests {
		t.Run(v.name, func(t *testing.T) {
			var (
				out string
				err error
			)

			switch v.kind {
			case "interface":
				out, err = codeanalyzer.ExtractTypesOf[*ast.InterfaceType](src, v.nameFilter)
			case "struct":
				out, err = codeanalyzer.ExtractTypesOf[*ast.StructType](src, v.nameFilter)
			case "all":
				out, err = codeanalyzer.ExtractTypesOf[ast.Expr](src, v.nameFilter)
			default:
				t.Fatalf("unknown kind: %s", v.kind)
			}

			require.NoError(t, err)

			assert.Equal(t, v.expected, out)
		})
	}
}

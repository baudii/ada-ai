package openapiproject_test

import (
	"testing"

	"github.com/baudii/ada-ai/internal/infra/openapiproject"
	"github.com/stretchr/testify/assert"
)

func TestInitialContent(t *testing.T) {
	iface := &openapiproject.ProjectInterface{
		Description: "Some interface comment",
		Name:        "Iface",
		Methods: []openapiproject.ProjectInterfaceMethod{
			{
				Description: "Method 1 comment",
				Name:        "Method1",
				Params:      []string{"a string", "b int"},
				Returns:     []string{"string", "error"},
			},
			{
				Description: "Method 2 comment",
				Name:        "Method2",
				Params:      []string{"a int"},
				Returns:     []string{"error"},
			},
		},
	}

	expected := `// Some interface comment
type iface struct {
}

// Method 1 comment
func (i iface) Method1(a string, b int) (string, error) {
}

// Method 2 comment
func (i iface) Method2(a int) (error) {
}`

	res := openapiproject.InitialUseCaseContent(iface)
	assert.Contains(t, res.String(), expected)
}

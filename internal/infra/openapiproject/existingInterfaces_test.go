package openapiproject

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseProjectInterfaces(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		contents string
		expected map[string]*ProjectInterface
		err      string
	}{
		{
			name: "valid interfaces",
			contents: `package handlers

// Other comment: unrelated

// UserHandler handles user-related operations
type UserHandler interface {
	// CreateUser creates a new user
	CreateUser(name string, age int) (User, error)
	// GetUser retrieves a user by ID
	GetUser(id string) (User, error)
}
type Repository interface {
	Save(data string) error
}
`,
			expected: map[string]*ProjectInterface{
				"UserHandler": {
					Name:        "UserHandler",
					Description: "UserHandler handles user-related operations",
					Methods: []ProjectInterfaceMethod{
						{
							Name:        "CreateUser",
							Description: "CreateUser creates a new user",
							Params: []string{
								"name string",
								"age int",
							},
							Returns: []string{
								"User",
								"error",
							},
						},
						{
							Name:        "GetUser",
							Description: "GetUser retrieves a user by ID",
							Params: []string{
								"id string",
							},
							Returns: []string{
								"User",
								"error",
							},
						},
					},
				},
				"Repository": {
					Name:        "Repository",
					Description: "",
					Methods: []ProjectInterfaceMethod{
						{
							Name:        "Save",
							Description: "",
							Params: []string{
								"data string",
							},
							Returns: []string{
								"error",
							},
						},
					},
				},
			},
		},
		{
			name: "invalid interface declaration",
			contents: `package handlers
type InvalidInterface {
	DoSomething() error
}
			}`,
			err: "parse file",
		},
		{
			name: "no interfaces",
			contents: `package handlers
type NotAnInterface struct {
	Field string
}
			`,
			expected: map[string]*ProjectInterface{},
		},
	}

	for _, v := range tests {
		t.Run(v.name, func(t *testing.T) {
			t.Parallel()
			o := New(t.TempDir())
			outInterfaces := o.InterfacesFilePath()
			err := os.MkdirAll(filepath.Dir(outInterfaces), 0o755)
			require.NoError(t, err)
			err = os.WriteFile(outInterfaces, []byte(v.contents), 0o644)
			require.NoError(t, err)

			interfaces, err := o.ParseProjectInterfaces()
			if v.err != "" {
				assert.ErrorContains(t, err, v.err)
				return
			}

			assert.NoError(t, err)
			assert.Len(t, interfaces.m, len(v.expected))
			for name, expectedInterface := range v.expected {
				actualInterface, exists := interfaces.m[name]
				assert.True(t, exists, "interface %s should exist", name)
				assert.Equal(t, *expectedInterface, *actualInterface)
			}
		})
	}
}

func TestGenerateInterfacesContent(t *testing.T) {
	t.Parallel()
	o := New(t.TempDir())
	interfaces := map[string]*ProjectInterface{
		"UserHandler": {
			Name:        "UserHandler",
			Description: "UserHandler handles user-related operations",
			Methods: []ProjectInterfaceMethod{
				{
					Name:        "CreateUser",
					Description: "CreateUser creates a new user",
					Params: []string{
						"name string",
						"age int",
					},
					Returns: []string{
						"User",
						"error",
					},
				},
				{
					Name:        "GetUser",
					Description: "GetUser retrieves a user by ID",
					Params: []string{
						"id string",
					},
					Returns: []string{
						"User",
						"error",
					},
				},
			},
		},
		"Repository": {
			Name:        "Repository",
			Description: "",
			Methods: []ProjectInterfaceMethod{
				{
					Name:        "Save",
					Description: "",
					Params: []string{
						"data string",
					},
					Returns: []string{
						"error",
					},
				},
			},
		},
	}
	expected := `type Repository interface {
	Save(data string) (error)
}

// UserHandler handles user-related operations
type UserHandler interface {
	// CreateUser creates a new user
	CreateUser(name string, age int) (User, error)
	// GetUser retrieves a user by ID
	GetUser(id string) (User, error)
}

`
	actual := o.GenerateInterfacesContent(interfaces)
	assert.Equal(t, expected, actual)
}

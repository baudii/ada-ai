package openapiproject_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/baudii/ada-ai/internal/infra/openapiproject"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseExistingModels(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected map[string]*openapiproject.ProjectModel
		hasErr   bool
	}{
		{
			name: "simple struct",
			content: `package models

const something = 42

type User struct {
	// ID is the user identifier
	ID int ` + "`json:\"id\"`" + `
	Name string ` + "`json:\"name\"`" + ` // User's name
}`,
			expected: map[string]*openapiproject.ProjectModel{
				"User": {
					Name:        "User",
					Description: "",
					Fields: []openapiproject.FieldInfo{
						{
							Name:     "ID",
							Type:     "int",
							Docs:     []string{"// ID is the user identifier"},
							Tags:     "`json:\"id\"`",
							Comments: []string{},
						},
						{
							Name:     "Name",
							Type:     "string",
							Tags:     "`json:\"name\"`",
							Comments: []string{"// User's name"},
							Docs:     []string{},
						},
					},
				},
			},
		},
		{
			name: "struct with several fields on one line",
			content: `package models
type Point struct {
	// X and Y coordinates
	X, Y int ` + "`json:\"x\" json:\"y\"`" + `
}`,
			expected: map[string]*openapiproject.ProjectModel{
				"Point": {
					Name:        "Point",
					Description: "",
					Fields: []openapiproject.FieldInfo{
						{
							Name:     "X",
							Type:     "int",
							Docs:     []string{"// X and Y coordinates"},
							Comments: []string{},
							Tags:     "`json:\"x\" json:\"y\"`",
						},
						{
							Name:     "Y",
							Type:     "int",
							Docs:     []string{"// X and Y coordinates"},
							Comments: []string{},
							Tags:     "`json:\"x\" json:\"y\"`",
						},
					},
				},
			},
		},
		{
			name: "struct without fields",
			content: `package models
type Empty struct {
}`,
			expected: map[string]*openapiproject.ProjectModel{
				"Empty": {
					Name:        "Empty",
					Description: "",
					Fields:      []openapiproject.FieldInfo{},
				},
			},
		},
		{
			name:     "file does not exist",
			content:  "",
			expected: map[string]*openapiproject.ProjectModel{},
			hasErr:   false,
		},
		{
			name: "invalid Go code",
			content: `package models
type Invalid struct {
	ID int`,
			expected: nil,
			hasErr:   true,
		},
	}

	for _, v := range tests {
		t.Run(v.name, func(t *testing.T) {
			tmp := t.TempDir()
			modelName := "addresses"
			o := openapiproject.New(tmp)
			err := os.MkdirAll(o.ModelsFolder(), 0755)
			require.NoError(t, err)
			if v.content != "" {
				err = os.WriteFile(filepath.Join(o.ModelsFolder(), modelName+".go"), []byte(v.content), 0644)
				require.NoError(t, err)
			}

			res, err := o.ParseExistingModels(modelName)
			if v.hasErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				for name, expectedInterface := range v.expected {
					resModel, exists := res[name]
					assert.True(t, exists, "interface %s should exist", name)
					assert.Equal(t, *expectedInterface, *resModel)
				}
			}
		})
	}
}

func TestGenerateModelsContent(t *testing.T) {
	models := map[string]*openapiproject.ProjectModel{
		"User": {
			Name:        "User",
			Description: "User represents a system user.",
			Fields: []openapiproject.FieldInfo{
				{
					Name:     "ID",
					Type:     "openapi_types.UUID",
					Tags:     "`json:\"id\"`",
					Docs:     []string{"// ID is the unique identifier for the user."},
					Comments: []string{},
				},
				{
					Name:     "Name",
					Type:     "string",
					Tags:     "`json:\"name\"`",
					Docs:     []string{"// Name is the name of the user."},
					Comments: []string{},
				},
			},
		},
		"Product": {
			Name:        "Product",
			Description: "Product represents an item for sale.",
			Fields: []openapiproject.FieldInfo{
				{
					Name:     "SKU",
					Type:     "string",
					Tags:     "`json:\"sku\"`",
					Docs:     []string{"// SKU is the stock keeping unit."},
					Comments: []string{"// additional comment"},
				},
				{
					Name:     "Name",
					Type:     "string",
					Tags:     "`json:\"name\"`",
					Docs:     []string{"// Name is the name of the product."},
					Comments: []string{},
				},
				{
					Name:     "Price",
					Type:     "float64",
					Tags:     "`json:\"price\"`",
					Docs:     []string{"// Price is the cost of the product."},
					Comments: []string{},
				},
			},
		},
	}
	expected := `// Product represents an item for sale.
type Product struct {
	// SKU is the stock keeping unit.
	SKU string ` + "`json:\"sku\"`" + `// additional comment
	// Name is the name of the product.
	Name string ` + "`json:\"name\"`" + `
	// Price is the cost of the product.
	Price float64 ` + "`json:\"price\"`" + `
}

// User represents a system user.
type User struct {
	// ID is the unique identifier for the user.
	ID openapi_types.UUID ` + "`json:\"id\"`" + `
	// Name is the name of the user.
	Name string ` + "`json:\"name\"`" + `
}
`
	o := openapiproject.New("some/path")
	result := o.GenerateModelsContent(models)
	assert.Equal(t, expected, result)
}

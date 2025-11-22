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
		expected map[string]openapiproject.ExistingModel
		hasErr   bool
	}{
		{
			name: "simple struct",
			content: `package models
type User struct {
	// ID is the user identifier
	ID int ` + "`json:\"id\"`" + `
	// Name is the user name
	Name string ` + "`json:\"name\"`" + `
}`,
			expected: map[string]openapiproject.ExistingModel{
				"User": {
					Name:        "User",
					Description: "",
					Fields: []openapiproject.FieldInfo{
						{
							Name: "ID",
							Type: "int",
							Doc:  "ID is the user identifier",
							Tags: "`json:\"id\"`",
						},
						{
							Name: "Name",
							Type: "string",
							Doc:  "Name is the user name",
							Tags: "`json:\"name\"`",
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
			expected: map[string]openapiproject.ExistingModel{
				"Point": {
					Name:        "Point",
					Description: "",
					Fields: []openapiproject.FieldInfo{
						{
							Name: "X",
							Type: "int",
							Doc:  "X and Y coordinates",
							Tags: "`json:\"x\" json:\"y\"`",
						},
						{
							Name: "Y",
							Type: "int",
							Doc:  "X and Y coordinates",
							Tags: "`json:\"x\" json:\"y\"`",
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
			expected: map[string]openapiproject.ExistingModel{
				"Empty": {
					Name:        "Empty",
					Description: "",
					Fields:      []openapiproject.FieldInfo{},
				},
			},
		},
	}

	for _, v := range tests {
		t.Run(v.name, func(t *testing.T) {
			tmp := t.TempDir()
			modelName := "addresses"
			o := openapiproject.New(tmp)
			err := os.MkdirAll(o.ModelsFolder(), 0755)
			require.NoError(t, err)
			err = os.WriteFile(filepath.Join(o.ModelsFolder(), modelName+".go"), []byte(v.content), 0644)
			require.NoError(t, err)

			res, err := o.ParseExistingModels(modelName)
			if v.hasErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, v.expected, res)
			}
		})
	}
}

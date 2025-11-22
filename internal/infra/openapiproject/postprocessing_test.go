package openapiproject

import (
	"go/format"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadLLMResponse(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string

		// Represents the sections of the LLM response
		functionBody          string
		addedFields           string
		interfaces            string
		addedModels           string
		interfacesDescription []string

		// Titles of each section
		functionPrefix      string
		addedFieldsPrefix   string
		interfacePrefix     string
		addedModelsPrefix   string
		interfaceDescPrefix string

		err string
	}{
		{
			name: "valid response",
			functionBody: `ctx := r.Context()

product, err := s.ProductRepository.GetProductByID(ctx, id)
if err != nil {
    http.Error(w, "failed to get product", http.StatusInternalServerError)
    return
}

w.Header().Set("Content-Type", "application/json")
if err := json.NewEncoder(w).Encode(product); err != nil {
    http.Error(w, "failed to write response", http.StatusInternalServerError)
    return
}`,
			addedFields: `- Price float64 ` + "`json:\"price\"`" + ` // Price of the product
- Description string ` + "`json:\"description\"`" + ` // Description of the product`,
			interfaces: `type ProductRepository interface {
    GetProductByID(ctx context.Context, id openapi_types.UUID) (Product, error)
}
type OrderRepository interface {
	CreateOrder(ctx context.Context, order Order) (openapi_types.UUID, error)								
}`,
			addedModels: `type Product struct {
	ID          openapi_types.UUID  ` + "`json:\"id\"`" + `
	Name        string             ` + "`json:\"name\"`" + `
	Description string             ` + "`json:\"description\"`" + `
	Price       float64            ` + "`json:\"price\"`" + `
}`,
			interfacesDescription: []string{
				`Interface: ProductRepository
Description: Provides data access methods for product entities. The GetProductsId handler requires this interface to retrieve a single product by its UUID.

Methods:
  - GetProductByID
    Signature: GetProductByID(ctx context.Context, id openapi_types.UUID) (Product, error)
    Description: Retrieves the product identified by the given UUID from the underlying storage. Returns the product and any error encountered during the lookup.`,
				`Interface: OrderRepository
Description: Manages data access methods for order entities. The CreateOrder handler relies on this interface to persist new orders.

Methods:
  - CreateOrder
    Signature: CreateOrder(ctx context.Context, order Order) (openapi_types.UUID, error)
    Description: Persists a new order in the underlying storage and returns the UUID of the created order along with any error encountered during the operation.`},
			functionPrefix:      "## function\n",
			addedFieldsPrefix:   "## added_fields\n",
			interfacePrefix:     "## interfaces\n",
			addedModelsPrefix:   "## added_models\n",
			interfaceDescPrefix: "## interfaces_description\n",
		},
		{
			name: "missing interfaces section",
			functionBody: `ctx := r.Context()

product, err := s.ProductRepository.GetProductByID(ctx, id)
if err != nil {
	http.Error(w, "failed to get product", http.StatusInternalServerError)
	return
}
`,
			interfaces:            ``,
			interfacesDescription: []string{},
			functionPrefix:        "## function\n",
			interfacePrefix:       "",
			interfaceDescPrefix:   "## interfaces_description\n",
			err:                   "failed to parse body: invalid structure",
		},
		{
			name: "none interfaces section",

			functionBody: `ctx := r.Context()`,
			addedFields:  `- Address string ` + "`json:\"address\"`" + ` // Address of the user`,
			addedModels: `type User struct {
	ID   openapi_types.UUID ` + "`json:\"id\"`" + `
	Name string            ` + "`json:\"name\"`" + `
}`,
			interfaces:            `none`,
			interfacesDescription: []string{"none"},

			functionPrefix:      "## function\n",
			addedFieldsPrefix:   "## added_fields\n",
			interfacePrefix:     "## interfaces\n",
			addedModelsPrefix:   "## added_models\n",
			interfaceDescPrefix: "## interfaces_description\n",
		},
	}

	for _, v := range tests {
		t.Run(v.name, func(t *testing.T) {
			functionBody := "```go\n" + v.functionBody + "\n```\n"
			addedFields := "```go\n" + v.addedFields + "\n```\n"
			interfaces := "```go\n" + v.interfaces + "\n```\n"
			addedModels := "```go\n" + v.addedModels + "\n```\n"
			interfacesDesc := "\n```" + strings.Join(v.interfacesDescription, "```\n```") + "\n```\n"
			body := ""
			if v.functionPrefix != "" {
				body += v.functionPrefix + functionBody
			}
			if v.addedFieldsPrefix != "" {
				body += v.addedFieldsPrefix + addedFields
			}
			if v.interfacePrefix != "" {
				body += v.interfacePrefix + interfaces
			}
			if v.addedModelsPrefix != "" {
				body += v.addedModelsPrefix + addedModels
			}
			if v.interfaceDescPrefix != "" {
				if len(v.interfacesDescription) == 0 {
					body += v.interfaceDescPrefix + "```\nnone\n```\n"
				} else {
					body += v.interfaceDescPrefix + interfacesDesc
				}
			}

			resp, err := ReadLLMResponse(body)
			if v.err != "" {
				assert.ErrorContains(t, err, v.err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, functionBody, resp.functionBody)
				assert.Equal(t, interfaces, resp.interfaces)
				assert.Equal(t, interfacesDesc, resp.interfacesDescription)
			}
		})
	}
}

func TestParseMethodSignature(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		signature       string
		expectedName    string
		expectedParams  []string
		expectedReturns []string
	}{
		{
			name:         "simple method",
			signature:    "GetUser(id string, id2 int) (User, error)",
			expectedName: "GetUser",
			expectedParams: []string{
				"id string",
				"id2 int",
			},
			expectedReturns: []string{
				"User",
				"error",
			},
		},
		{
			name:           "method with no params",
			signature:      "ListUsers() ([]User, error)",
			expectedName:   "ListUsers",
			expectedParams: []string{},
			expectedReturns: []string{
				"[]User",
				"error",
			},
		},
		{
			name:            "method with no returns",
			signature:       "Ping()",
			expectedName:    "Ping",
			expectedParams:  []string{},
			expectedReturns: []string{},
		},
		{
			name:         "method with single return and param",
			signature:    "DeleteUser(id string) error",
			expectedName: "DeleteUser",
			expectedParams: []string{
				"id string",
			},
			expectedReturns: []string{
				"error",
			},
		},
		{
			name:            "malformed signature",
			signature:       "DeleteItem)(",
			expectedName:    "",
			expectedParams:  nil,
			expectedReturns: nil,
		},
	}

	for _, v := range tests {
		t.Run(v.name, func(t *testing.T) {
			t.Parallel()
			name, params, returns := ParseMethodSignature(v.signature)
			assert.Equal(t, v.expectedName, name)
			assert.Equal(t, v.expectedParams, params)
			assert.Equal(t, v.expectedReturns, returns)
		})
	}
}

func TestInsertFieldsToServer(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		resp            *llmResponse
		expectedContent string
		expectErr       bool
	}{
		{
			name: "insert new fields",
			resp: &llmResponse{
				addedFields: "- Price float64 `json:\"price\"` // Price of the product\n" +
					"- Description string `json:\"description\"` // Description of the product",
			},
			expectedContent: `package handlers

type Server struct {
	ID          openapi_types.UUID ` + "`json:\"id\"`" + `
	Name        string             ` + "`json:\"name\"`" + `
	Price       float64            ` + "`json:\"price\"`" + `       // Price of the product
	Description string             ` + "`json:\"description\"`" + ` // Description of the product
}
`,
			expectErr: false,
		},
	}
	initialContent := `package handlers

type Server struct {
	ID   openapi_types.UUID ` + "`json:\"id\"`" + `
	Name string            ` + "`json:\"name\"`" + `
}
`
	for _, v := range tests {
		t.Run(v.name, func(t *testing.T) {
			t.Parallel()
			tmp := t.TempDir()
			o := New(tmp)
			err := os.MkdirAll(o.HandlersFolder(), 0755)
			require.NoError(t, err)
			serverPath := o.ServerFilePath()
			err = os.WriteFile(serverPath, []byte(initialContent), 0644)
			require.NoError(t, err)
			err = o.InsertFieldsToServer(v.resp)
			if v.expectErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				serverPath := o.ServerFilePath()
				content, err := os.ReadFile(serverPath)
				require.NoError(t, err)
				content, err = format.Source([]byte(content))
				require.NoError(t, err)
				assert.Equal(t, v.expectedContent, string(content))
			}
		})
	}
}

package openapiproject

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadLLMResponse(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                  string
		functionBody          string
		functionPrefix        string
		interfaces            string
		interfacePrefix       string
		interfacesDescription []string
		interfaceDescPrefix   string
		err                   string
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
			interfaces: `type ProductRepository interface {
    GetProductByID(ctx context.Context, id openapi_types.UUID) (Product, error)
}
type OrderRepository interface {
	CreateOrder(ctx context.Context, order Order) (openapi_types.UUID, error)								
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
			interfacePrefix:     "## interfaces\n",
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
			name:                  "none interfaces section",
			functionBody:          `ctx := r.Context()`,
			interfaces:            `none`,
			functionPrefix:        "## function\n",
			interfacePrefix:       "## interfaces\n",
			interfaceDescPrefix:   "## interfaces_description\n",
			interfacesDescription: []string{"none"},
		},
	}

	for _, v := range tests {
		t.Run(v.name, func(t *testing.T) {
			functionBody := "```go\n" + v.functionBody + "\n```\n"
			interfaces := "```go\n" + v.interfaces + "\n```\n"
			interfacesDesc := "\n```" + strings.Join(v.interfacesDescription, "```\n```") + "\n```\n"
			body := ""
			if v.functionPrefix != "" {
				body += v.functionPrefix + functionBody
			}
			if v.interfacePrefix != "" {
				body += v.interfacePrefix + interfaces
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

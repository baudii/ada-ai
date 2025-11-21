package app

import "context"

var s string = `openapi: 3.0.0
info:
  title: E-commerce API
  version: 1.0.0
  description: |
    Simplified e-commerce API spec including only product browsing and address management.

servers:
  - url: https://api.demo-ecommerce.com/v1
    description: Production environment
  - url: https://api.dev.demo-ecommerce.com/v1
    description: Development environment

components:
  securitySchemes:
    BearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT

  schemas:
    Product:
      type: object
      required: [id, name, price, stock, category]
      properties:
        id:
          type: string
          format: uuid
          example: eda5cbc1-a615-4da5-ae73-4a33a9acfb6a
        name:
          type: string
          example: Worry Management
        description:
          type: string
          example: Mr street sell would civil. People through shake southern force.
        price:
          type: number
          format: float
          example: 91.37
        category:
          type: string
          example: wrong
        image_url:
          type: string
          format: uri
          example: https://dummyimage.com/766x809
        stock:
          type: integer
          example: 94
        created_at:
          type: string
          format: date-time
        updated_at:
          type: string
          format: date-time

    Address:
      type: object
      required: [line1, city, state, postal_code, country]
      properties:
        line1:
          type: string
        line2:
          type: string
        city:
          type: string
        state:
          type: string
        postal_code:
          type: string
        country:
          type: string

paths:
  /products:
    get:
      summary: List all products with filters
      parameters:
        - name: category
          in: query
          schema:
            type: string
        - name: search
          in: query
          schema:
            type: string
        - name: min_price
          in: query
          schema:
            type: number
        - name: max_price
          in: query
          schema:
            type: number
      responses:
        '200':
          description: List of products
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/Product'

  /products/{id}:
    get:
      summary: Get product details by ID
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
            format: uuid
      responses:
        '200':
          description: Product details
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Product'

  /addresses:
    get:
      summary: Get your saved addresses
      security:
        - BearerAuth: []
      responses:
        '200':
          description: List of saved addresses
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/Address'

    post:
      summary: Add a new address
      security:
        - BearerAuth: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/Address'
      responses:
        '201':
          description: Address added

`

func (a *App) GenerateOpenAPISpec(ctx context.Context) (string, error) {
	// Placeholder implementation
	return s, nil
}

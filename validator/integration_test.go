package validator_test

import (
	"testing"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/validator"
	"github.com/stretchr/testify/require"
)

// TestValidatorIntegrationComprehensive tests full schema compilation and realistic validation scenarios
func TestValidatorIntegrationComprehensive(t *testing.T) {
	t.Run("User Profile Schema", func(t *testing.T) {
		// Comprehensive user profile schema with multiple constraints
		userSchema, err := schema.NewBuilder().
			Type(schema.ObjectType).
			Property("id", schema.NewBuilder().
				Type(schema.IntegerType).
				Minimum(1).MustBuild()).
			Property("username", schema.NewBuilder().
				Type(schema.StringType).
				MinLength(3).
				MaxLength(20).
				Pattern("^[a-zA-Z0-9_]+$").MustBuild()).
			Property("email", schema.NewBuilder().
				Type(schema.StringType).
				Pattern(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`).MustBuild()).
			Property("age", schema.NewBuilder().
				Type(schema.IntegerType).
				Minimum(13).
				Maximum(120).MustBuild()).
			Property("roles", schema.NewBuilder().
				Type(schema.ArrayType).
				Items(schema.NewBuilder().
					Type(schema.StringType).
					Enum("admin", "user", "moderator", "guest").MustBuild()).
				MinItems(1).
				UniqueItems(true).MustBuild()).
			Property("active", schema.NewBuilder().
				Type(schema.BooleanType).MustBuild()).
			Property("preferences", schema.NewBuilder().
				Type(schema.ObjectType).
				Property("theme", schema.NewBuilder().
					Type(schema.StringType).
					Enum("light", "dark", "auto").MustBuild()).
				Property("notifications", schema.NewBuilder().
					Type(schema.BooleanType).MustBuild()).
				AdditionalProperties(false).MustBuild()).
			Required("id", "username", "email").
			AdditionalProperties(false).
			Build()
		require.NoError(t, err)

		v, err := validator.Compile(userSchema)
		require.NoError(t, err)

		testCases := []struct {
			name    string
			user    any
			wantErr bool
			errMsg  string
		}{
			{
				name: "valid complete user",
				user: map[string]any{
					"id":       123,
					"username": "john_doe",
					"email":    "john@example.com",
					"age":      30,
					"roles":    []any{"user", "moderator"},
					"active":   true,
					"preferences": map[string]any{
						"theme":         "dark",
						"notifications": true,
					},
				},
				wantErr: false,
			},
			{
				name: "valid minimal user",
				user: map[string]any{
					"id":       1,
					"username": "jane",
					"email":    "jane@test.org",
				},
				wantErr: false,
			},
			{
				name: "missing required field",
				user: map[string]any{
					"username": "john_doe",
					"email":    "john@example.com",
					// missing id
				},
				wantErr: true,
				errMsg:  "required",
			},
			{
				name: "invalid username pattern",
				user: map[string]any{
					"id":       123,
					"username": "john-doe", // hyphens not allowed
					"email":    "john@example.com",
				},
				wantErr: true,
				errMsg:  "pattern",
			},
			{
				name: "invalid email format",
				user: map[string]any{
					"id":       123,
					"username": "john_doe",
					"email":    "not-an-email",
				},
				wantErr: true,
				errMsg:  "pattern",
			},
			{
				name: "invalid age range",
				user: map[string]any{
					"id":       123,
					"username": "john_doe",
					"email":    "john@example.com",
					"age":      12, // too young
				},
				wantErr: true,
				errMsg:  "minimum",
			},
			{
				name: "invalid role in array",
				user: map[string]any{
					"id":       123,
					"username": "john_doe",
					"email":    "john@example.com",
					"roles":    []any{"user", "invalid_role"},
				},
				wantErr: true,
				errMsg:  "enum",
			},
			{
				name: "duplicate roles",
				user: map[string]any{
					"id":       123,
					"username": "john_doe",
					"email":    "john@example.com",
					"roles":    []any{"user", "user"},
				},
				wantErr: true,
				errMsg:  "unique",
			},
			{
				name: "invalid preferences theme",
				user: map[string]any{
					"id":       123,
					"username": "john_doe",
					"email":    "john@example.com",
					"preferences": map[string]any{
						"theme": "rainbow", // invalid theme
					},
				},
				wantErr: true,
				errMsg:  "enum",
			},
			{
				name: "additional property not allowed",
				user: map[string]any{
					"id":       123,
					"username": "john_doe",
					"email":    "john@example.com",
					"extra":    "not allowed",
				},
				wantErr: true,
				errMsg:  "additional",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				err := v.Validate(tc.user)
				if tc.wantErr {
					require.Error(t, err)
					if tc.errMsg != "" {
						require.Contains(t, err.Error(), tc.errMsg)
					}
				} else {
					require.NoError(t, err)
				}
			})
		}
	})

	t.Run("Product Catalog Schema", func(t *testing.T) {
		// Complex e-commerce product schema
		productSchema, err := schema.NewBuilder().
			Type(schema.ObjectType).
			Property("id", schema.NewBuilder().
				Type(schema.StringType).
				Pattern("^PRD-[0-9]{6}$").MustBuild()).
			Property("name", schema.NewBuilder().
				Type(schema.StringType).
				MinLength(1).
				MaxLength(100).MustBuild()).
			Property("description", schema.NewBuilder().
				Type(schema.StringType).
				MaxLength(1000).MustBuild()).
			Property("price", schema.NewBuilder().
				Type(schema.NumberType).
				Minimum(0).
				MultipleOf(0.01).MustBuild()).
			Property("currency", schema.NewBuilder().
				Type(schema.StringType).
				Enum("USD", "EUR", "GBP", "JPY").MustBuild()).
			Property("category", schema.NewBuilder().
				Type(schema.StringType).
				Enum("electronics", "clothing", "books", "home", "sports").MustBuild()).
			Property("tags", schema.NewBuilder().
				Type(schema.ArrayType).
				Items(schema.NewBuilder().
					Type(schema.StringType).
					MinLength(1).
					MaxLength(50).MustBuild()).
				MaxItems(10).
				UniqueItems(true).MustBuild()).
			Property("inventory", schema.NewBuilder().
				Type(schema.ObjectType).
				Property("quantity", schema.NewBuilder().
					Type(schema.IntegerType).
					Minimum(0).MustBuild()).
				Property("warehouse", schema.NewBuilder().
					Type(schema.StringType).
					MinLength(1).MustBuild()).
				Required("quantity").MustBuild()).
			Property("active", schema.NewBuilder().
				Type(schema.BooleanType).MustBuild()).
			Required("id", "name", "price", "currency", "category").
			Build()
		require.NoError(t, err)

		v, err := validator.Compile(productSchema)
		require.NoError(t, err)

		testCases := []struct {
			name    string
			product any
			wantErr bool
			errMsg  string
		}{
			{
				name: "valid complete product",
				product: map[string]any{
					"id":          "PRD-123456",
					"name":        "Wireless Headphones",
					"description": "High-quality wireless headphones with noise cancellation",
					"price":       299.99,
					"currency":    "USD",
					"category":    "electronics",
					"tags":        []any{"audio", "wireless", "premium"},
					"inventory": map[string]any{
						"quantity":  50,
						"warehouse": "Main-01",
					},
					"active": true,
				},
				wantErr: false,
			},
			{
				name: "valid minimal product",
				product: map[string]any{
					"id":       "PRD-654321",
					"name":     "Book",
					"price":    9.99,
					"currency": "USD",
					"category": "books",
				},
				wantErr: false,
			},
			{
				name: "invalid product ID format",
				product: map[string]any{
					"id":       "INVALID-ID",
					"name":     "Product",
					"price":    10.00,
					"currency": "USD",
					"category": "electronics",
				},
				wantErr: true,
				errMsg:  "pattern",
			},
			{
				name: "invalid price - negative",
				product: map[string]any{
					"id":       "PRD-123456",
					"name":     "Product",
					"price":    -10.00,
					"currency": "USD",
					"category": "electronics",
				},
				wantErr: true,
				errMsg:  "minimum",
			},
			{
				name: "invalid price - wrong precision",
				product: map[string]any{
					"id":       "PRD-123456",
					"name":     "Product",
					"price":    10.123, // too many decimal places
					"currency": "USD",
					"category": "electronics",
				},
				wantErr: true,
				errMsg:  "multiple",
			},
			{
				name: "invalid currency",
				product: map[string]any{
					"id":       "PRD-123456",
					"name":     "Product",
					"price":    10.00,
					"currency": "BTC",
					"category": "electronics",
				},
				wantErr: true,
				errMsg:  "enum",
			},
			{
				name: "too many tags",
				product: map[string]any{
					"id":       "PRD-123456",
					"name":     "Product",
					"price":    10.00,
					"currency": "USD",
					"category": "electronics",
					"tags":     []any{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11"}, // 11 tags, max is 10
				},
				wantErr: true,
				errMsg:  "maximum items",
			},
			{
				name: "duplicate tags",
				product: map[string]any{
					"id":       "PRD-123456",
					"name":     "Product",
					"price":    10.00,
					"currency": "USD",
					"category": "electronics",
					"tags":     []any{"tag1", "tag2", "tag1"}, // duplicate
				},
				wantErr: true,
				errMsg:  "unique",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				err := v.Validate(tc.product)
				if tc.wantErr {
					require.Error(t, err)
					if tc.errMsg != "" {
						require.Contains(t, err.Error(), tc.errMsg)
					}
				} else {
					require.NoError(t, err)
				}
			})
		}
	})

	t.Run("API Response Schema with OneOf", func(t *testing.T) {
		// API response that can be either success or error
		successSchema, err := schema.NewBuilder().
			Type(schema.ObjectType).
			Property("status", schema.NewBuilder().
				Type(schema.StringType).
				Const("success").MustBuild()).
			Property("data", schema.NewBuilder().
				Type(schema.ObjectType).MustBuild()).
			Required("status", "data").
			Build()
		require.NoError(t, err)

		errorSchema, err := schema.NewBuilder().
			Type(schema.ObjectType).
			Property("status", schema.NewBuilder().
				Type(schema.StringType).
				Const("error").MustBuild()).
			Property("error", schema.NewBuilder().
				Type(schema.ObjectType).
				Property("code", schema.NewBuilder().
					Type(schema.IntegerType).MustBuild()).
				Property("message", schema.NewBuilder().
					Type(schema.StringType).MustBuild()).
				Required("code", "message").MustBuild()).
			Required("status", "error").
			Build()
		require.NoError(t, err)

		responseSchema, err := schema.NewBuilder().
			OneOf(successSchema, errorSchema).
			Build()
		require.NoError(t, err)

		v, err := validator.Compile(responseSchema)
		require.NoError(t, err)

		testCases := []struct {
			name     string
			response any
			wantErr  bool
			errMsg   string
		}{
			{
				name: "valid success response",
				response: map[string]any{
					"status": "success",
					"data": map[string]any{
						"users": []any{
							map[string]any{"id": 1, "name": "John"},
						},
					},
				},
				wantErr: false,
			},
			{
				name: "valid error response",
				response: map[string]any{
					"status": "error",
					"error": map[string]any{
						"code":    404,
						"message": "Not found",
					},
				},
				wantErr: false,
			},
			{
				name: "invalid - ambiguous response",
				response: map[string]any{
					"status": "success",
					"data":   map[string]any{},
					"error": map[string]any{
						"code":    500,
						"message": "Internal error",
					},
				},
				wantErr: true, // Matches both schemas
				errMsg:  "oneOf",
			},
			{
				name: "invalid - wrong status",
				response: map[string]any{
					"status": "pending",
					"data":   map[string]any{},
				},
				wantErr: true, // Matches neither schema
				errMsg:  "oneOf",
			},
			{
				name: "invalid - missing required field",
				response: map[string]any{
					"status": "success",
					// missing data field
				},
				wantErr: true,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				err := v.Validate(tc.response)
				if tc.wantErr {
					require.Error(t, err)
					if tc.errMsg != "" {
						require.Contains(t, err.Error(), tc.errMsg)
					}
				} else {
					require.NoError(t, err)
				}
			})
		}
	})

	t.Run("Configuration Schema with AllOf", func(t *testing.T) {
		// Configuration that must satisfy base requirements AND specific feature requirements
		baseConfigSchema, err := schema.NewBuilder().
			Type(schema.ObjectType).
			Property("appName", schema.NewBuilder().
				Type(schema.StringType).
				MinLength(1).MustBuild()).
			Property("version", schema.NewBuilder().
				Type(schema.StringType).
				Pattern(`^\d+\.\d+\.\d+$`).MustBuild()).
			Required("appName", "version").
			Build()
		require.NoError(t, err)

		databaseConfigSchema, err := schema.NewBuilder().
			Type(schema.ObjectType).
			Property("database", schema.NewBuilder().
				Type(schema.ObjectType).
				Property("host", schema.NewBuilder().
					Type(schema.StringType).
					MinLength(1).MustBuild()).
				Property("port", schema.NewBuilder().
					Type(schema.IntegerType).
					Minimum(1).
					Maximum(65535).MustBuild()).
				Property("name", schema.NewBuilder().
					Type(schema.StringType).
					MinLength(1).MustBuild()).
				Required("host", "port", "name").MustBuild()).
			Build()
		require.NoError(t, err)

		serverConfigSchema, err := schema.NewBuilder().
			Type(schema.ObjectType).
			Property("server", schema.NewBuilder().
				Type(schema.ObjectType).
				Property("port", schema.NewBuilder().
					Type(schema.IntegerType).
					Minimum(1024).
					Maximum(65535).MustBuild()).
				Property("host", schema.NewBuilder().
					Type(schema.StringType).
					Enum("localhost", "0.0.0.0").MustBuild()).
				Required("port").MustBuild()).
			Build()
		require.NoError(t, err)

		configSchema, err := schema.NewBuilder().
			AllOf(
				baseConfigSchema,
				databaseConfigSchema,
				serverConfigSchema,
			).
			Build()
		require.NoError(t, err)

		v, err := validator.Compile(configSchema)
		require.NoError(t, err)

		testCases := []struct {
			name    string
			config  any
			wantErr bool
			errMsg  string
		}{
			{
				name: "valid complete config",
				config: map[string]any{
					"appName": "MyApp",
					"version": "1.2.3",
					"database": map[string]any{
						"host": "localhost",
						"port": 5432,
						"name": "myapp_db",
					},
					"server": map[string]any{
						"host": "0.0.0.0",
						"port": 8080,
					},
				},
				wantErr: false,
			},
			{
				name: "invalid version format",
				config: map[string]any{
					"appName": "MyApp",
					"version": "v1.2.3", // invalid format
					"database": map[string]any{
						"host": "localhost",
						"port": 5432,
						"name": "myapp_db",
					},
					"server": map[string]any{
						"port": 8080,
					},
				},
				wantErr: true,
				errMsg:  "pattern",
			},
			{
				name: "missing database config",
				config: map[string]any{
					"appName": "MyApp",
					"version": "1.2.3",
					"server": map[string]any{
						"port": 8080,
					},
					// missing database
				},
				wantErr: true,
				errMsg:  "required",
			},
			{
				name: "invalid server port range",
				config: map[string]any{
					"appName": "MyApp",
					"version": "1.2.3",
					"database": map[string]any{
						"host": "localhost",
						"port": 5432,
						"name": "myapp_db",
					},
					"server": map[string]any{
						"port": 80, // below minimum for server
					},
				},
				wantErr: true,
				errMsg:  "minimum",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				err := v.Validate(tc.config)
				if tc.wantErr {
					require.Error(t, err)
					if tc.errMsg != "" {
						require.Contains(t, err.Error(), tc.errMsg)
					}
				} else {
					require.NoError(t, err)
				}
			})
		}
	})

	t.Run("Recursive Schema Structure", func(t *testing.T) {
		// Note: This test assumes recursive schema support exists
		// If not implemented yet, it will test the intended behavior

		// This would typically use $ref for recursion in a real implementation
		simpleTreeSchema, err := schema.NewBuilder().
			Type(schema.ObjectType).
			Property("name", schema.NewBuilder().
				Type(schema.StringType).
				MinLength(1).MustBuild()).
			Property("type", schema.NewBuilder().
				Type(schema.StringType).
				Enum("file", "directory").MustBuild()).
			Property("size", schema.NewBuilder().
				Type(schema.IntegerType).
				Minimum(0).MustBuild()).
			Property("children", schema.NewBuilder().
				Type(schema.ArrayType).
				Items(schema.NewBuilder().
					Type(schema.ObjectType).MustBuild()).MustBuild()). // Simplified - would be recursive reference
			Required("name", "type").
			Build()
		require.NoError(t, err)

		v, err := validator.Compile(simpleTreeSchema)
		require.NoError(t, err)

		testCases := []struct {
			name    string
			tree    any
			wantErr bool
			errMsg  string
		}{
			{
				name: "simple file",
				tree: map[string]any{
					"name": "document.txt",
					"type": "file",
					"size": 1024,
				},
				wantErr: false,
			},
			{
				name: "directory with files",
				tree: map[string]any{
					"name": "project",
					"type": "directory",
					"children": []any{
						map[string]any{
							"name": "README.md",
							"type": "file",
							"size": 512,
						},
						map[string]any{
							"name":     "src",
							"type":     "directory",
							"children": []any{},
						},
					},
				},
				wantErr: false,
			},
			{
				name: "invalid type",
				tree: map[string]any{
					"name": "item",
					"type": "link", // not in enum
				},
				wantErr: true,
				errMsg:  "enum",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				err := v.Validate(tc.tree)
				if tc.wantErr {
					require.Error(t, err)
					if tc.errMsg != "" {
						require.Contains(t, err.Error(), tc.errMsg)
					}
				} else {
					require.NoError(t, err)
				}
			})
		}
	})
}

// TestValidatorErrorMessages tests that error messages are helpful and specific
func TestValidatorErrorMessages(t *testing.T) {
	t.Run("Detailed Error Messages", func(t *testing.T) {
		userSchema, err := schema.NewBuilder().
			Type(schema.ObjectType).
			Property("name", schema.NewBuilder().
				Type(schema.StringType).
				MinLength(2).
				MaxLength(50).MustBuild()).
			Property("age", schema.NewBuilder().
				Type(schema.IntegerType).
				Minimum(0).
				Maximum(150).MustBuild()).
			Required("name").
			Build()
		require.NoError(t, err)

		v, err := validator.Compile(userSchema)
		require.NoError(t, err)

		testCases := []struct {
			name        string
			value       any
			expectedErr []string // Error message should contain these strings
		}{
			{
				name:        "missing required property",
				value:       map[string]any{"age": 25},
				expectedErr: []string{"required", "name"},
			},
			{
				name:        "string too short",
				value:       map[string]any{"name": "A", "age": 25},
				expectedErr: []string{"minLength", "2"},
			},
			{
				name:        "integer out of range",
				value:       map[string]any{"name": "John", "age": 200},
				expectedErr: []string{"maximum", "150"},
			},
			{
				name:        "wrong type",
				value:       map[string]any{"name": "John", "age": "twenty"},
				expectedErr: []string{"expected", "integer"},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				err := v.Validate(tc.value)
				require.Error(t, err)

				errMsg := err.Error()
				for _, expected := range tc.expectedErr {
					require.Contains(t, errMsg, expected,
						"Error message should contain '%s': %s", expected, errMsg)
				}
			})
		}
	})
}

// TestValidatorPerformance tests validation performance with complex schemas
func TestValidatorPerformance(t *testing.T) {
	t.Run("Large Object Validation", func(t *testing.T) {
		// Create a schema for validating large objects
		itemSchema, err := schema.NewBuilder().
			Type(schema.ObjectType).
			Property("id", schema.NewBuilder().Type(schema.IntegerType).MustBuild()).
			Property("name", schema.NewBuilder().Type(schema.StringType).MustBuild()).
			Property("value", schema.NewBuilder().Type(schema.NumberType).MustBuild()).
			Build()
		require.NoError(t, err)

		listSchema, err := schema.NewBuilder().
			Type(schema.ArrayType).
			Items(itemSchema).
			MinItems(1).
			MaxItems(1000).
			Build()
		require.NoError(t, err)

		v, err := validator.Compile(listSchema)
		require.NoError(t, err)

		// Create a large array for testing
		largeArray := make([]any, 100)
		for i := 0; i < 100; i++ {
			largeArray[i] = map[string]any{
				"id":    i,
				"name":  "Item " + string(rune(i)),
				"value": float64(i) * 1.5,
			}
		}

		// Test that validation completes without timing out
		err = v.Validate(largeArray)
		require.NoError(t, err)
	})
}

// TestValidatorEdgeCases tests edge cases and boundary conditions
func TestValidatorEdgeCases(t *testing.T) {
	t.Run("Empty Schema", func(t *testing.T) {
		s, err := schema.NewBuilder().Build() // No constraints
		require.NoError(t, err)

		v, err := validator.Compile(s)
		require.NoError(t, err)

		// Empty schema should allow anything
		testValues := []any{
			"string",
			123,
			true,
			[]any{1, 2, 3},
			map[string]any{"key": "value"},
			nil,
		}

		for _, value := range testValues {
			err = v.Validate(value)
			require.NoError(t, err, "Empty schema should allow value: %v", value)
		}
	})

	t.Run("Schema with Only Type", func(t *testing.T) {
		s, err := schema.NewBuilder().Type(schema.StringType).Build()
		require.NoError(t, err)

		v, err := validator.Compile(s)
		require.NoError(t, err)

		// Should allow any string
		require.NoError(t, v.Validate(""))
		require.NoError(t, v.Validate("any string"))
		require.NoError(t, v.Validate("very long string with lots of content"))

		// Should reject non-strings
		require.Error(t, v.Validate(123))
		require.Error(t, v.Validate(true))
		require.Error(t, v.Validate([]any{}))
	})
}

package validate

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/null-create/mcp-tls/pkg/mcp"
)

func TestValidateToolInputSchema(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name           string
		tool           *mcp.Tool
		inputArguments []byte
		expectedStatus mcp.ExecutionStatus
		expectError    bool
		errorContains  string
	}{
		{
			name: "valid input with required fields",
			tool: &mcp.Tool{
				Name: "test-tool",
				InputSchema: mustMarshalJSON(map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"name": map[string]interface{}{
							"type": "string",
						},
						"age": map[string]interface{}{
							"type": "integer",
						},
					},
					"required": []string{"name"},
				}),
			},
			inputArguments: mustMarshalJSON(map[string]interface{}{
				"name": "John",
				"age":  30,
			}),
			expectedStatus: mcp.StatusSucceeded,
			expectError:    false,
		},
		{
			name: "valid input with only required fields",
			tool: &mcp.Tool{
				Name: "test-tool",
				InputSchema: mustMarshalJSON(map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"name": map[string]interface{}{
							"type": "string",
						},
						"age": map[string]interface{}{
							"type": "integer",
						},
					},
					"required": []string{"name"},
				}),
			},
			inputArguments: mustMarshalJSON(map[string]interface{}{
				"name": "John",
			}),
			expectedStatus: mcp.StatusSucceeded,
			expectError:    false,
		},
		{
			name: "missing required field",
			tool: &mcp.Tool{
				Name: "test-tool",
				InputSchema: mustMarshalJSON(map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"name": map[string]interface{}{
							"type": "string",
						},
						"age": map[string]interface{}{
							"type": "integer",
						},
					},
					"required": []string{"name"},
				}),
			},
			inputArguments: mustMarshalJSON(map[string]interface{}{
				"age": 30,
			}),
			expectedStatus: mcp.StatusFailed,
			expectError:    true,
			errorContains:  "Input validation failed",
		},
		{
			name: "wrong type for field",
			tool: &mcp.Tool{
				Name: "test-tool",
				InputSchema: mustMarshalJSON(map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"name": map[string]interface{}{
							"type": "string",
						},
						"age": map[string]interface{}{
							"type": "integer",
						},
					},
					"required": []string{"name"},
				}),
			},
			inputArguments: mustMarshalJSON(map[string]interface{}{
				"name": "John",
				"age":  "thirty", // should be integer
			}),
			expectedStatus: mcp.StatusFailed,
			expectError:    true,
			errorContains:  "Input validation failed",
		},
		{
			name: "empty input arguments with required fields",
			tool: &mcp.Tool{
				Name: "test-tool",
				InputSchema: mustMarshalJSON(map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"name": map[string]interface{}{
							"type": "string",
						},
					},
					"required": []string{"name"},
				}),
			},
			inputArguments: []byte(`{}`),
			expectedStatus: mcp.StatusFailed,
			expectError:    true,
			errorContains:  "Input validation failed",
		},
		{
			name: "valid empty input with no required fields",
			tool: &mcp.Tool{
				Name: "test-tool",
				InputSchema: mustMarshalJSON(map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"name": map[string]interface{}{
							"type": "string",
						},
					},
				}),
			},
			inputArguments: []byte(`{}`),
			expectedStatus: mcp.StatusSucceeded,
			expectError:    false,
		},
		{
			name: "invalid JSON in input arguments",
			tool: &mcp.Tool{
				Name: "test-tool",
				InputSchema: mustMarshalJSON(map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"name": map[string]interface{}{
							"type": "string",
						},
					},
				}),
			},
			inputArguments: []byte(`{"name": "John"`), // missing closing brace
			expectedStatus: mcp.StatusError,
			expectError:    true,
			errorContains:  "internal validation error",
		},
		{
			name: "invalid schema in tool",
			tool: &mcp.Tool{
				Name:        "test-tool",
				InputSchema: []byte(`{"type": "invalid-type"}`), // invalid schema
			},
			inputArguments: mustMarshalJSON(map[string]interface{}{
				"name": "John",
			}),
			expectedStatus: mcp.StatusError,
			expectError:    true,
			errorContains:  "internal schema error",
		},
		{
			name: "no input schema defined",
			tool: &mcp.Tool{
				Name:        "test-tool",
				InputSchema: []byte{}, // empty schema
			},
			inputArguments: mustMarshalJSON(map[string]interface{}{
				"name": "John",
			}),
			expectedStatus: mcp.StatusFailed,
			expectError:    true,
			errorContains:  "no InputSchema defined",
		},
		{
			name: "complex nested schema validation",
			tool: &mcp.Tool{
				Name: "complex-tool",
				InputSchema: mustMarshalJSON(map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"user": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"name": map[string]interface{}{
									"type": "string",
								},
								"contact": map[string]interface{}{
									"type": "object",
									"properties": map[string]interface{}{
										"email": map[string]interface{}{
											"type":   "string",
											"format": "email",
										},
									},
									"required": []string{"email"},
								},
							},
							"required": []string{"name", "contact"},
						},
					},
					"required": []string{"user"},
				}),
			},
			inputArguments: mustMarshalJSON(map[string]interface{}{
				"user": map[string]interface{}{
					"name": "John Doe",
					"contact": map[string]interface{}{
						"email": "john@example.com",
					},
				},
			}),
			expectedStatus: mcp.StatusSucceeded,
			expectError:    false,
		},
		{
			name: "complex nested schema validation - missing nested required field",
			tool: &mcp.Tool{
				Name: "complex-tool",
				InputSchema: mustMarshalJSON(map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"user": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"name": map[string]interface{}{
									"type": "string",
								},
								"contact": map[string]interface{}{
									"type": "object",
									"properties": map[string]interface{}{
										"email": map[string]interface{}{
											"type":   "string",
											"format": "email",
										},
									},
									"required": []string{"email"},
								},
							},
							"required": []string{"name", "contact"},
						},
					},
					"required": []string{"user"},
				}),
			},
			inputArguments: mustMarshalJSON(map[string]interface{}{
				"user": map[string]interface{}{
					"name":    "John Doe",
					"contact": map[string]interface{}{
						// missing email
					},
				},
			}),
			expectedStatus: mcp.StatusFailed,
			expectError:    true,
			errorContains:  "Input validation failed",
		},
		{
			name: "array type validation - valid",
			tool: &mcp.Tool{
				Name: "array-tool",
				InputSchema: mustMarshalJSON(map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"items": map[string]interface{}{
							"type": "array",
							"items": map[string]interface{}{
								"type": "string",
							},
						},
					},
					"required": []string{"items"},
				}),
			},
			inputArguments: mustMarshalJSON(map[string]interface{}{
				"items": []string{"apple", "banana", "cherry"},
			}),
			expectedStatus: mcp.StatusSucceeded,
			expectError:    false,
		},
		{
			name: "array type validation - invalid item type",
			tool: &mcp.Tool{
				Name: "array-tool",
				InputSchema: mustMarshalJSON(map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"items": map[string]interface{}{
							"type": "array",
							"items": map[string]interface{}{
								"type": "string",
							},
						},
					},
					"required": []string{"items"},
				}),
			},
			inputArguments: mustMarshalJSON(map[string]interface{}{
				"items": []interface{}{"apple", 123, "cherry"}, // 123 should be string
			}),
			expectedStatus: mcp.StatusFailed,
			expectError:    true,
			errorContains:  "Input validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, err := ValidateToolInputSchema(ctx, tt.tool, tt.inputArguments)

			// Check execution status
			if status != tt.expectedStatus {
				t.Errorf("ValidateToolInputSchema() status = %v, want %v", status, tt.expectedStatus)
			}

			// Check error expectation
			if tt.expectError && err == nil {
				t.Errorf("ValidateToolInputSchema() expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("ValidateToolInputSchema() unexpected error: %v", err)
			}

			// Check error message content if specified
			if tt.expectError && err != nil && tt.errorContains != "" {
				if !containsString(err.Error(), tt.errorContains) {
					t.Errorf("ValidateToolInputSchema() error = %v, want error containing %v", err.Error(), tt.errorContains)
				}
			}
		})
	}
}

func TestValidateToolInputSchema_NilTool(t *testing.T) {
	ctx := context.Background()
	inputArgs := []byte(`{"test": "value"}`)

	// This should panic or handle gracefully - testing the behavior
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("ValidateToolInputSchema() with nil tool should panic or handle gracefully")
		}
	}()

	ValidateToolInputSchema(ctx, nil, inputArgs)
}

func TestValidateToolInputSchema_NilInputArguments(t *testing.T) {
	ctx := context.Background()
	tool := &mcp.Tool{
		Name: "test-tool",
		InputSchema: mustMarshalJSON(map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"name": map[string]interface{}{
					"type": "string",
				},
			},
		}),
	}

	status, err := ValidateToolInputSchema(ctx, tool, nil)

	// Should handle nil input gracefully
	if status != mcp.StatusError && status != mcp.StatusFailed {
		t.Errorf("ValidateToolInputSchema() with nil input should return error or failed status, got %v", status)
	}
	if err == nil {
		t.Errorf("ValidateToolInputSchema() with nil input should return an error")
	}
}

// Helper functions for tests

func mustMarshalJSON(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && stringContains(s, substr)))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

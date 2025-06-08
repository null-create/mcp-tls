package validate

import (
	"encoding/json"
	"testing"

	"github.com/null-create/mcp-tls/pkg/mcp"
)

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

func generateLargeString(size int) string {
	result := make([]byte, size)
	for i := range result {
		result[i] = 'a' + byte(i%26)
	}
	return string(result)
}

func TestValidateToolInputSchema(t *testing.T) {
	tests := []struct {
		name           string
		tool           *mcp.Tool
		inputArguments []byte
		expectedStatus ValidationStatus
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
			expectedStatus: StatusSucceeded,
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
			expectedStatus: StatusSucceeded,
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
			expectedStatus: StatusFailed,
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
			expectedStatus: StatusFailed,
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
			expectedStatus: StatusFailed,
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
			expectedStatus: StatusSucceeded,
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
			expectedStatus: StatusError,
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
			expectedStatus: StatusError,
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
			expectedStatus: StatusFailed,
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
			expectedStatus: StatusSucceeded,
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
			expectedStatus: StatusFailed,
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
			expectedStatus: StatusSucceeded,
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
			expectedStatus: StatusFailed,
			expectError:    true,
			errorContains:  "Input validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, err := ValidateToolInputSchema(tt.tool, tt.inputArguments)

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
	inputArgs := []byte(`{"test": "value"}`)

	// This should panic or handle gracefully - testing the behavior
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("ValidateToolInputSchema() with nil tool should panic or handle gracefully")
		}
	}()

	ValidateToolInputSchema(nil, inputArgs)
}

func TestValidateToolInputSchema_NilInputArguments(t *testing.T) {
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

	status, err := ValidateToolInputSchema(tool, nil)

	// Should handle nil input gracefully
	if status != StatusError && status != StatusFailed {
		t.Errorf("ValidateToolInputSchema() with nil input should return error or failed status, got %v", status)
	}
	if err == nil {
		t.Errorf("ValidateToolInputSchema() with nil input should return an error")
	}
}

func TestValidateToolOutput(t *testing.T) {
	tests := []struct {
		name           string
		tool           *mcp.Tool
		rawResult      string
		expectedStatus ValidationStatus
		expectError    bool
		errorContains  string
	}{
		{
			name: "valid output with string result",
			tool: &mcp.Tool{
				Name: "test-tool",
				OutputSchema: mustMarshalJSON(map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"message": map[string]interface{}{
							"type": "string",
						},
						"status": map[string]interface{}{
							"type": "string",
						},
					},
					"required": []string{"message"},
				}),
			},
			rawResult:      `{"message": "Hello World", "status": "success"}`,
			expectedStatus: StatusSucceeded,
			expectError:    false,
		},
		{
			name: "valid output with only required fields",
			tool: &mcp.Tool{
				Name: "test-tool",
				OutputSchema: mustMarshalJSON(map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"message": map[string]interface{}{
							"type": "string",
						},
						"status": map[string]interface{}{
							"type": "string",
						},
					},
					"required": []string{"message"},
				}),
			},
			rawResult:      `{"message": "Hello World"}`,
			expectedStatus: StatusSucceeded,
			expectError:    false,
		},
		{
			name: "missing required field in output",
			tool: &mcp.Tool{
				Name: "test-tool",
				OutputSchema: mustMarshalJSON(map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"message": map[string]interface{}{
							"type": "string",
						},
						"status": map[string]interface{}{
							"type": "string",
						},
					},
					"required": []string{"message"},
				}),
			},
			rawResult:      `{"status": "success"}`,
			expectedStatus: StatusFailed,
			expectError:    true,
			errorContains:  "output failed validation",
		},
		{
			name: "wrong type in output",
			tool: &mcp.Tool{
				Name: "test-tool",
				OutputSchema: mustMarshalJSON(map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"count": map[string]interface{}{
							"type": "integer",
						},
					},
					"required": []string{"count"},
				}),
			},
			rawResult:      `{"count": "not-a-number"}`,
			expectedStatus: StatusFailed,
			expectError:    true,
			errorContains:  "output failed validation",
		},
		{
			name: "invalid JSON in output",
			tool: &mcp.Tool{
				Name: "test-tool",
				OutputSchema: mustMarshalJSON(map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"message": map[string]interface{}{
							"type": "string",
						},
					},
				}),
			},
			rawResult:      `{"message": "Hello World"`, // missing closing brace
			expectedStatus: StatusError,
			expectError:    true,
			errorContains:  "internal output validation error",
		},
		{
			name: "invalid output schema",
			tool: &mcp.Tool{
				Name:         "test-tool",
				OutputSchema: []byte(`{"type": "invalid-type"}`), // invalid schema
			},
			rawResult:      `{"message": "Hello World"}`,
			expectedStatus: StatusError,
			expectError:    true,
			errorContains:  "internal output schema error",
		},
		{
			name: "no output schema defined - should succeed",
			tool: &mcp.Tool{
				Name:         "test-tool",
				OutputSchema: []byte{}, // empty schema
			},
			rawResult:      `{"anything": "goes"}`,
			expectedStatus: StatusSucceeded,
			expectError:    false,
		},
		{
			name: "nil output schema - should succeed",
			tool: &mcp.Tool{
				Name:         "test-tool",
				OutputSchema: nil, // nil schema
			},
			rawResult:      `{"anything": "goes"}`,
			expectedStatus: StatusSucceeded,
			expectError:    false,
		},
		{
			name: "complex nested output validation - valid",
			tool: &mcp.Tool{
				Name: "complex-tool",
				OutputSchema: mustMarshalJSON(map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"result": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"data": map[string]interface{}{
									"type": "array",
									"items": map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"id": map[string]interface{}{
												"type": "integer",
											},
											"name": map[string]interface{}{
												"type": "string",
											},
										},
										"required": []string{"id", "name"},
									},
								},
								"metadata": map[string]interface{}{
									"type": "object",
									"properties": map[string]interface{}{
										"total": map[string]interface{}{
											"type": "integer",
										},
									},
									"required": []string{"total"},
								},
							},
							"required": []string{"data", "metadata"},
						},
					},
					"required": []string{"result"},
				}),
			},
			rawResult: `{
				"result": {
					"data": [
						{"id": 1, "name": "Item 1"},
						{"id": 2, "name": "Item 2"}
					],
					"metadata": {
						"total": 2
					}
				}
			}`,
			expectedStatus: StatusSucceeded,
			expectError:    false,
		},
		{
			name: "complex nested output validation - missing nested field",
			tool: &mcp.Tool{
				Name: "complex-tool",
				OutputSchema: mustMarshalJSON(map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"result": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"data": map[string]interface{}{
									"type": "array",
									"items": map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"id": map[string]interface{}{
												"type": "integer",
											},
											"name": map[string]interface{}{
												"type": "string",
											},
										},
										"required": []string{"id", "name"},
									},
								},
								"metadata": map[string]interface{}{
									"type": "object",
									"properties": map[string]interface{}{
										"total": map[string]interface{}{
											"type": "integer",
										},
									},
									"required": []string{"total"},
								},
							},
							"required": []string{"data", "metadata"},
						},
					},
					"required": []string{"result"},
				}),
			},
			rawResult: `{
				"result": {
					"data": [
						{"id": 1, "name": "Item 1"},
						{"id": 2}
					],
					"metadata": {
						"total": 2
					}
				}
			}`,
			expectedStatus: StatusFailed,
			expectError:    true,
			errorContains:  "output failed validation",
		},
		{
			name: "array output validation - valid",
			tool: &mcp.Tool{
				Name: "array-tool",
				OutputSchema: mustMarshalJSON(map[string]interface{}{
					"type": "array",
					"items": map[string]interface{}{
						"type": "string",
					},
				}),
			},
			rawResult:      `["apple", "banana", "cherry"]`,
			expectedStatus: StatusSucceeded,
			expectError:    false,
		},
		{
			name: "array output validation - invalid item type",
			tool: &mcp.Tool{
				Name: "array-tool",
				OutputSchema: mustMarshalJSON(map[string]interface{}{
					"type": "array",
					"items": map[string]interface{}{
						"type": "string",
					},
				}),
			},
			rawResult:      `["apple", 123, "cherry"]`,
			expectedStatus: StatusFailed,
			expectError:    true,
			errorContains:  "output failed validation",
		},
		{
			name: "simple string output validation - valid",
			tool: &mcp.Tool{
				Name: "string-tool",
				OutputSchema: mustMarshalJSON(map[string]interface{}{
					"type": "string",
				}),
			},
			rawResult:      `"Hello World"`,
			expectedStatus: StatusSucceeded,
			expectError:    false,
		},
		{
			name: "simple string output validation - invalid type",
			tool: &mcp.Tool{
				Name: "string-tool",
				OutputSchema: mustMarshalJSON(map[string]interface{}{
					"type": "string",
				}),
			},
			rawResult:      `123`,
			expectedStatus: StatusFailed,
			expectError:    true,
			errorContains:  "output failed validation",
		},
		{
			name: "boolean output validation - valid",
			tool: &mcp.Tool{
				Name: "boolean-tool",
				OutputSchema: mustMarshalJSON(map[string]interface{}{
					"type": "boolean",
				}),
			},
			rawResult:      `true`,
			expectedStatus: StatusSucceeded,
			expectError:    false,
		},
		{
			name: "boolean output validation - invalid type",
			tool: &mcp.Tool{
				Name: "boolean-tool",
				OutputSchema: mustMarshalJSON(map[string]interface{}{
					"type": "boolean",
				}),
			},
			rawResult:      `"true"`,
			expectedStatus: StatusFailed,
			expectError:    true,
			errorContains:  "output failed validation",
		},
		{
			name: "number output validation - valid integer",
			tool: &mcp.Tool{
				Name: "number-tool",
				OutputSchema: mustMarshalJSON(map[string]interface{}{
					"type": "number",
				}),
			},
			rawResult:      `42`,
			expectedStatus: StatusSucceeded,
			expectError:    false,
		},
		{
			name: "number output validation - valid float",
			tool: &mcp.Tool{
				Name: "number-tool",
				OutputSchema: mustMarshalJSON(map[string]interface{}{
					"type": "number",
				}),
			},
			rawResult:      `42.5`,
			expectedStatus: StatusSucceeded,
			expectError:    false,
		},
		{
			name: "empty string output",
			tool: &mcp.Tool{
				Name: "test-tool",
				OutputSchema: mustMarshalJSON(map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"message": map[string]interface{}{
							"type": "string",
						},
					},
				}),
			},
			rawResult:      ``,
			expectedStatus: StatusError,
			expectError:    true,
			errorContains:  "internal output validation error",
		},
		{
			name: "null output validation",
			tool: &mcp.Tool{
				Name: "null-tool",
				OutputSchema: mustMarshalJSON(map[string]interface{}{
					"type": "null",
				}),
			},
			rawResult:      `null`,
			expectedStatus: StatusSucceeded,
			expectError:    false,
		},
		{
			name: "output with additional properties allowed",
			tool: &mcp.Tool{
				Name: "flexible-tool",
				OutputSchema: mustMarshalJSON(map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"required_field": map[string]interface{}{
							"type": "string",
						},
					},
					"required":             []string{"required_field"},
					"additionalProperties": true,
				}),
			},
			rawResult:      `{"required_field": "value", "extra_field": "allowed"}`,
			expectedStatus: StatusSucceeded,
			expectError:    false,
		},
		{
			name: "output with additional properties forbidden",
			tool: &mcp.Tool{
				Name: "strict-tool",
				OutputSchema: mustMarshalJSON(map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"required_field": map[string]interface{}{
							"type": "string",
						},
					},
					"required":             []string{"required_field"},
					"additionalProperties": false,
				}),
			},
			rawResult:      `{"required_field": "value", "extra_field": "not_allowed"}`,
			expectedStatus: StatusFailed,
			expectError:    true,
			errorContains:  "output failed validation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, err := ValidateToolOutput(tt.rawResult, tt.tool)

			// Check execution status
			if status != tt.expectedStatus {
				t.Errorf("ValidateToolOutput() status = %v, want %v", status, tt.expectedStatus)
			}

			// Check error expectation
			if tt.expectError && err == nil {
				t.Errorf("ValidateToolOutput() expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("ValidateToolOutput() unexpected error: %v", err)
			}

			// Check error message content if specified
			if tt.expectError && err != nil && tt.errorContains != "" {
				if !containsString(err.Error(), tt.errorContains) {
					t.Errorf("ValidateToolOutput() error = %v, want error containing %v", err.Error(), tt.errorContains)
				}
			}
		})
	}
}

func TestValidateToolOutput_NilTool(t *testing.T) {
	rawResult := `{"test": "value"}`

	// This should panic or handle gracefully - testing the behavior
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("ValidateToolOutput() with nil tool should panic or handle gracefully")
		}
	}()

	ValidateToolOutput(rawResult, nil)
}

func TestValidateToolOutput_EdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		tool           *mcp.Tool
		rawResult      string
		expectedStatus ValidationStatus
		expectError    bool
		errorContains  string
	}{
		{
			name: "very large output string",
			tool: &mcp.Tool{
				Name: "large-output-tool",
				OutputSchema: mustMarshalJSON(map[string]interface{}{
					"type": "string",
				}),
			},
			rawResult:      `"` + generateLargeString(10000) + `"`,
			expectedStatus: StatusSucceeded,
			expectError:    false,
		},
		{
			name: "unicode characters in output",
			tool: &mcp.Tool{
				Name: "unicode-tool",
				OutputSchema: mustMarshalJSON(map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"message": map[string]interface{}{
							"type": "string",
						},
					},
				}),
			},
			rawResult:      `{"message": "Hello 世界! 🌍 emoji test"}`,
			expectedStatus: StatusSucceeded,
			expectError:    false,
		},
		{
			name: "deeply nested output structure",
			tool: &mcp.Tool{
				Name: "deep-tool",
				OutputSchema: mustMarshalJSON(map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"level1": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"level2": map[string]interface{}{
									"type": "object",
									"properties": map[string]interface{}{
										"level3": map[string]interface{}{
											"type": "object",
											"properties": map[string]interface{}{
												"value": map[string]interface{}{
													"type": "string",
												},
											},
											"required": []string{"value"},
										},
									},
									"required": []string{"level3"},
								},
							},
							"required": []string{"level2"},
						},
					},
					"required": []string{"level1"},
				}),
			},
			rawResult:      `{"level1": {"level2": {"level3": {"value": "deep"}}}}`,
			expectedStatus: StatusSucceeded,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, err := ValidateToolOutput(tt.rawResult, tt.tool)

			if status != tt.expectedStatus {
				t.Errorf("ValidateToolOutput() status = %v, want %v", status, tt.expectedStatus)
			}

			if tt.expectError && err == nil {
				t.Errorf("ValidateToolOutput() expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("ValidateToolOutput() unexpected error: %v", err)
			}

			if tt.expectError && err != nil && tt.errorContains != "" {
				if !containsString(err.Error(), tt.errorContains) {
					t.Errorf("ValidateToolOutput() error = %v, want error containing %v", err.Error(), tt.errorContains)
				}
			}
		})
	}
}

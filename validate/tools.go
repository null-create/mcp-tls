package validate

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/null-create/mcp-tls/mcp"

	"github.com/xeipuuv/gojsonschema"
)

// --- Tool Schema Validation ---

// FindToolDescription retrieves the trusted tool description by name.
// In a real system, this might involve looking up in a secure registry
// and potentially verifying signatures/sources stored in SecurityMetadata.
func FindToolDescription(name string, availableTools []mcp.ToolDescription) (*mcp.ToolDescription, error) {
	for _, tool := range availableTools {
		if tool.Name == name {
			// TODO: Add verification of tool description source/integrity here
			// based on tool.SecurityMetadata if available.
			return &tool, nil // Return pointer to avoid copying large schemas
		}
	}
	return nil, fmt.Errorf("tool '%s' not found or not permitted", name)
}

// ValidateToolSchema is called by the orchestrator when an LLM requests a tool call.
func ValidateToolSchema(
	ctx context.Context,
	toolCall mcp.ToolCall,
	availableTools []mcp.ToolDescription,
) (executionStatus mcp.ExecutionStatus, execErr error) {
	toolDesc, err := FindToolDescription(toolCall.FunctionName, availableTools)
	if err != nil {
		return mcp.StatusError, fmt.Errorf("tool description lookup failed: %w", err)
	}

	// Only validate if schema is provided
	if len(toolDesc.InputSchema) > 0 {
		schemaLoader := gojsonschema.NewBytesLoader(toolDesc.InputSchema)
		documentLoader := gojsonschema.NewBytesLoader(toolCall.Arguments)
		schema, err := gojsonschema.NewSchema(schemaLoader)
		if err != nil {
			return mcp.StatusError, fmt.Errorf("internal schema error for tool '%s'", toolDesc.Name)
		}

		result, err := schema.Validate(documentLoader)
		if err != nil {
			return mcp.StatusError, fmt.Errorf("internal validation error for tool '%s'", toolDesc.Name)
		}

		if !result.Valid() {
			var validationErrors []string
			for _, desc := range result.Errors() {
				validationErrors = append(validationErrors, fmt.Sprintf("- %s", desc))
			}
			errorMsg := fmt.Sprintf("Input validation failed for tool '%s':\n%s",
				toolDesc.Name, strings.Join(validationErrors, "\n"))
			fmt.Println("SECURITY ALERT:", errorMsg)
			return mcp.StatusFailed, errors.New(errorMsg)
		}
		fmt.Printf("Input arguments for tool '%s' validated successfully.\n", toolDesc.Name)
	} else {
		return mcp.StatusFailed, fmt.Errorf("no InputSchema defined for tool '%s'", toolDesc.Name)
	}

	return mcp.StatusSucceeded, nil
}

func ValidateToolCallOutput(
	rawResult string,
	toolCall mcp.ToolCall,
	availableTools []mcp.ToolDescription,
) (mcp.ExecutionStatus, error) {
	toolDesc, err := FindToolDescription(toolCall.FunctionName, availableTools)
	if err != nil {
		return mcp.StatusError, fmt.Errorf("tool description lookup failed: %w", err)
	}

	if len(toolDesc.OutputSchema) > 0 {
		outputSchemaLoader := gojsonschema.NewBytesLoader(toolDesc.OutputSchema)
		outputDocumentLoader := gojsonschema.NewStringLoader(rawResult)

		outputSchema, err := gojsonschema.NewSchema(outputSchemaLoader)
		if err != nil {
			fmt.Printf("ERROR: Invalid OutputSchema for tool '%s': %v\n", toolDesc.Name, err)
			return mcp.StatusError, fmt.Errorf("internal output schema error for tool '%s'", toolDesc.Name)
		}

		outputResult, err := outputSchema.Validate(outputDocumentLoader)
		if err != nil {
			fmt.Printf("ERROR: Output validation process error for tool '%s': %v\n", toolDesc.Name, err)
			return mcp.StatusError, fmt.Errorf("internal output validation error for tool '%s'", toolDesc.Name)
		}

		if !outputResult.Valid() {
			var validationErrors []string
			for _, desc := range outputResult.Errors() {
				validationErrors = append(validationErrors, fmt.Sprintf("- %s", desc))
			}
			errorMsg := fmt.Sprintf("Tool '%s' output failed validation:\n%s\nRaw Output: %s",
				toolDesc.Name, strings.Join(validationErrors, "\n"), rawResult)
			fmt.Println("SECURITY ALERT:", errorMsg)
			return mcp.StatusFailed, errors.New(errorMsg)
		}
		fmt.Printf("Output content for tool '%s' validated successfully.\n", toolDesc.Name)
	}
	return mcp.StatusSucceeded, nil
}

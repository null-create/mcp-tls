package validate

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/null-create/mcp-tls/pkg/mcp"

	"github.com/xeipuuv/gojsonschema"
)

type ValidationStatus string

const (
	StatusSucceeded ValidationStatus = "succeeded"
	StatusFailed    ValidationStatus = "failed"
	StatusError     ValidationStatus = "error"
)

// FindTool retrieves the trusted tool by name from the tool registry.
// In a real system, this might involve looking up in a secure registry
// and potentially verifying signatures/sources stored in SecurityMetadata.
func FindTool(toolName string, toolManager *mcp.ToolManager) (*mcp.Tool, error) {
	tool, err := toolManager.GetTool(toolName)
	if err != nil {
		return nil, fmt.Errorf("tool '%s' not found or not permitted: %w", toolName, err)
	}
	return &tool, nil
}

// ValidateToolCall validates both the tool lookup and input arguments in one call.
// This is a convenience function that combines tool lookup and input validation.
func ValidateToolCall(
	toolName string,
	inputArguments []byte,
	toolManager *mcp.ToolManager,
) (*mcp.Tool, ValidationStatus, error) {
	// Find the tool
	foundTool, err := FindTool(toolName, toolManager)
	if err != nil {
		return nil, StatusError, fmt.Errorf("tool lookup failed: %w", err)
	}

	// Validate the input
	status, err := ValidateToolInputSchema(foundTool, inputArguments)
	if err != nil {
		return foundTool, status, err
	}

	return foundTool, status, nil
}

// ValidateToolInputSchema validates the input arguments against the tool's input schema.
func ValidateToolInputSchema(tool *mcp.Tool, inputArguments []byte) (ValidationStatus, error) {
	// Only validate if schema is provided
	if len(tool.InputSchema) > 0 {
		schemaLoader := gojsonschema.NewBytesLoader(tool.InputSchema)
		documentLoader := gojsonschema.NewBytesLoader(inputArguments)
		schema, err := gojsonschema.NewSchema(schemaLoader)
		if err != nil {
			return StatusError, fmt.Errorf("internal schema error for tool '%s'", tool.Name)
		}

		result, err := schema.Validate(documentLoader)
		if err != nil {
			return StatusError, fmt.Errorf("internal validation error for tool '%s'", tool.Name)
		}

		if !result.Valid() {
			var validationErrors []string
			for _, desc := range result.Errors() {
				validationErrors = append(validationErrors, fmt.Sprintf("- %s", desc))
			}
			errorMsg := fmt.Sprintf(
				"Input validation failed for tool '%s':\n%s",
				tool.Name, strings.Join(validationErrors, "\n"),
			)
			fmt.Println("SECURITY ALERT:", errorMsg)
			return StatusFailed, errors.New(errorMsg)
		}
		fmt.Printf("Input arguments for tool '%s' validated successfully", tool.Name)
	} else {
		return StatusFailed, fmt.Errorf("no InputSchema defined for tool '%s'", tool.Name)
	}

	return StatusSucceeded, nil
}

// ValidateToolOutput validates the tool's output against its output schema.
func ValidateToolOutput(rawResult string, tool *mcp.Tool) (ValidationStatus, error) {
	if len(tool.OutputSchema) > 0 {
		outputSchemaLoader := gojsonschema.NewBytesLoader(tool.OutputSchema)
		outputDocumentLoader := gojsonschema.NewStringLoader(rawResult)
		outputSchema, err := gojsonschema.NewSchema(outputSchemaLoader)
		if err != nil {
			fmt.Printf("ERROR: Invalid OutputSchema for tool '%s': %v\n", tool.Name, err)
			return StatusError, fmt.Errorf("internal output schema error for tool '%s'", tool.Name)
		}

		outputResult, err := outputSchema.Validate(outputDocumentLoader)
		if err != nil {
			fmt.Printf("ERROR: Output validation process error for tool '%s': %v\n", tool.Name, err)
			return StatusError, fmt.Errorf("internal output validation error for tool '%s'", tool.Name)
		}

		if !outputResult.Valid() {
			var validationErrors []string
			for _, desc := range outputResult.Errors() {
				validationErrors = append(validationErrors, fmt.Sprintf("- %s", desc))
			}
			errorMsg := fmt.Sprintf("Tool '%s' output failed validation:\n%s\nRaw Output: %s",
				tool.Name, strings.Join(validationErrors, "\n"), rawResult)
			fmt.Println("SECURITY ALERT:", errorMsg)
			return StatusFailed, errors.New(errorMsg)
		}
		fmt.Printf("Output content for tool '%s' validated successfully.\n", tool.Name)
	}
	return StatusSucceeded, nil
}

// ValidateToolDescription analyzes the tools descriptive text for hidden characters
// and potentially injected prompts
func ValidateToolDescription(toolDescription string) error {
	detections := detectHiddenUnicode(toolDescription)
	if len(detections) == 0 {
		return nil
	}
	return fmt.Errorf("ALERT: %d hidden characters detected in tool description text", len(detections))
}

// ValidateToolSecurity performs comprehensive security validation on a tool.
// This includes checksum validation, schema fingerprint validation, and description validation.
func ValidateToolSecurity(tool *mcp.Tool, toolManager *mcp.ToolManager) error {
	if err := ValidateToolDescription(tool.Description); err != nil {
		return fmt.Errorf("tool description validation failed: %w", err)
	}

	// Get the tool from registry to perform security checks (this validates checksums/signatures)
	_, err := toolManager.GetTool(tool.Name)
	if err != nil {
		return fmt.Errorf("tool security validation failed: %w", err)
	}

	return nil
}

// ValidateToolIntegrity performs integrity checks on a tool's security metadata.
func ValidateToolIntegrity(tool *mcp.Tool) error {
	// Validate checksum if present
	if tool.SecurityMetadata.Checksum != "" {
		expectedChecksum, err := GenerateToolChecksum(*tool)
		if err != nil {
			return fmt.Errorf("failed to generate checksum for validation: %w", err)
		}
		if expectedChecksum != tool.SecurityMetadata.Checksum {
			return errors.New("tool checksum validation failed - tool may have been tampered with")
		}
	}

	// Validate schema fingerprint if present
	if tool.SecurityMetadata.Signature != "" {
		expectedFingerprint, err := GenerateSchemaFingerprint(tool.InputSchema)
		if err != nil {
			return fmt.Errorf("failed to generate schema fingerprint for validation: %w", err)
		}
		if expectedFingerprint != tool.SecurityMetadata.Signature {
			return errors.New("schema fingerprint validation failed - schema may have been tampered with")
		}
	}

	return nil
}

// canonicalizeJson converts a JSON object to a canonical form for consistent hashing
func canonicalizeJson(data json.RawMessage) (json.RawMessage, error) {
	var obj any
	if err := json.Unmarshal(data, &obj); err != nil {
		return nil, err
	}

	// Sort keys and ensure consistent serialization
	canonical, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}

	return canonical, nil
}

// GenerateSchemaFingerprint creates a fingerprint of the schema using SHA-256
func GenerateSchemaFingerprint(schema json.RawMessage) (string, error) {
	canonical, err := canonicalizeJson(schema)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(canonical)
	return hex.EncodeToString(hash[:]), nil
}

// GenerateToolChecksum creates a checksum of the entire tool definition using SHA-256
func GenerateToolChecksum(tool mcp.Tool) (string, error) {
	toolCopy := tool

	data, err := json.Marshal(toolCopy)
	if err != nil {
		return "", err
	}

	// Use canonical JSON for consistent checksums
	canonical, err := canonicalizeJson(data)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(canonical)
	return hex.EncodeToString(hash[:]), nil
}

// Use canonical serialization (deterministic field order)
func CanonicalizeAndHash(tool mcp.Tool) (string, error) {
	buf := &bytes.Buffer{}
	encoder := json.NewEncoder(buf)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "")

	if err := encoder.Encode(tool); err != nil {
		return "", fmt.Errorf("failed to serialize tool: %w", err)
	}

	hash := sha256.Sum256(buf.Bytes())
	return fmt.Sprintf("%x", hash[:]), nil
}

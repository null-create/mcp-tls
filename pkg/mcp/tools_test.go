package mcp

import (
	"encoding/json"
	"testing"
)

func TestToolRegistry(t *testing.T) {
	// Create a tool registry with security enabled
	registry := NewToolRegistry(true)
	registry.SetSecurityOptions(true, true)

	// Create a test tool
	testSchema := json.RawMessage(`{
		"type": "object",
		"properties": {
			"name": {"type": "string"},
			"value": {"type": "number"}
		},
		"required": ["name", "value"]
	}`)

	tool := Tool{
		Name:        "test-tool",
		Description: "A test tool",
		InputSchema: testSchema,
	}

	// Register the tool
	if err := registry.RegisterTool(tool); err != nil {
		t.Fatalf("Failed to register tool: %v", err)
	}

	// Retrieve the tool
	retrievedTool, err := registry.GetTool("test-tool")
	if err != nil {
		t.Fatalf("Failed to get tool: %v", err)
	}

	// Verify that the checksum and fingerprint were generated
	if retrievedTool.SecurityMetadata.Checksum == "" {
		t.Error("Tool checksum was not generated")
	}

	if retrievedTool.SecurityMetadata.Signature == "" {
		t.Error("Schema signature was not generated")
	}

	// Test tool list
	toolSet := registry.ListTools()
	if len(toolSet.Tools) != 1 {
		t.Errorf("Expected 1 tool, but got %d", len(toolSet.Tools))
	}

	if !toolSet.SecurityEnabled {
		t.Error("ToolSet security is not enabled")
	}
}

func TestToolTampering(t *testing.T) {
	// Create a tool registry with security enabled
	registry := NewToolRegistry(true)
	registry.SetSecurityOptions(true, true)

	// Create a test tool
	testSchema := json.RawMessage(`{
		"type": "object",
		"properties": {
			"name": {"type": "string"},
			"value": {"type": "number"}
		},
		"required": ["name", "value"]
	}`)

	tool := Tool{
		Name:        "test-tool",
		Description: "A test tool",
		InputSchema: testSchema,
	}

	// Register the tool
	if err := registry.RegisterTool(tool); err != nil {
		t.Fatalf("Failed to register tool: %v", err)
	}

	// Get the registered tool
	registeredTool, err := registry.GetTool("test-tool")
	if err != nil {
		t.Fatalf("Failed to get tool: %v", err)
	}

	// Tamper with the tool
	tamperedTool := registeredTool
	tamperedTool.Description = "Tampered description"

	// Replace the tool with the tampered version (bypassing checksum validation)
	registry.tools["test-tool"] = tamperedTool

	// Try to get the tool again - should fail checksum validation
	_, err = registry.GetTool("test-tool")
	if err == nil {
		t.Error("Expected checksum validation to fail, but it succeeded")
	}
}

func TestSchemaFingerprint(t *testing.T) {
	// Create test schemas
	schema1 := json.RawMessage(`{"type": "object", "properties": {"a": {"type": "string"}}}`)
	schema2 := json.RawMessage(`{"type": "object", "properties": {"a": {"type": "string"}}}`)
	schema3 := json.RawMessage(`{"type": "object", "properties": {"b": {"type": "string"}}}`)

	// Generate fingerprints
	fingerprint1, err := generateSchemaFingerprint(schema1)
	if err != nil {
		t.Fatalf("Failed to generate fingerprint: %v", err)
	}

	fingerprint2, err := generateSchemaFingerprint(schema2)
	if err != nil {
		t.Fatalf("Failed to generate fingerprint: %v", err)
	}

	fingerprint3, err := generateSchemaFingerprint(schema3)
	if err != nil {
		t.Fatalf("Failed to generate fingerprint: %v", err)
	}

	// Identical schemas should have identical fingerprints
	if fingerprint1 != fingerprint2 {
		t.Error("Identical schemas produced different fingerprints")
	}

	// Different schemas should have different fingerprints
	if fingerprint1 == fingerprint3 {
		t.Error("Different schemas produced the same fingerprint")
	}
}

func TestCanonicalJson(t *testing.T) {
	// Create two JSON objects with the same content but different field order
	json1 := json.RawMessage(`{"b": 2, "a": 1}`)
	json2 := json.RawMessage(`{"a": 1, "b": 2}`)

	// Canonicalize both objects
	canonical1, err := canonicalizeJson(json1)
	if err != nil {
		t.Fatalf("Failed to canonicalize JSON: %v", err)
	}

	canonical2, err := canonicalizeJson(json2)
	if err != nil {
		t.Fatalf("Failed to canonicalize JSON: %v", err)
	}

	// Canonical forms should be identical
	if string(canonical1) != string(canonical2) {
		t.Errorf("Canonical forms differ: %s vs %s", canonical1, canonical2)
	}
}

func TestSchemaModification(t *testing.T) {
	// Create a tool registry with security enabled
	registry := NewToolRegistry(true)
	registry.SetSecurityOptions(true, true)

	// Create a test tool
	originalSchema := json.RawMessage(`{
		"type": "object",
		"properties": {
			"name": {"type": "string"},
			"value": {"type": "number"}
		},
		"required": ["name", "value"]
	}`)

	tool := Tool{
		Name:        "schema-tool",
		Description: "A tool with a schema",
		InputSchema: originalSchema,
	}

	// Register the tool
	if err := registry.RegisterTool(tool); err != nil {
		t.Fatalf("Failed to register tool: %v", err)
	}

	// Get the registered tool
	registeredTool, err := registry.GetTool("schema-tool")
	if err != nil {
		t.Fatalf("Failed to get tool: %v", err)
	}

	// Modify the schema
	modifiedSchema := json.RawMessage(`{
		"type": "object",
		"properties": {
			"name": {"type": "string"},
			"value": {"type": "number"},
			"newField": {"type": "boolean"}
		},
		"required": ["name", "value"]
	}`)

	// Update the tool with modified schema but keep original fingerprint
	modifiedTool := registeredTool
	modifiedTool.InputSchema = modifiedSchema

	// Replace the tool (bypassing validation)
	registry.tools["schema-tool"] = modifiedTool

	// Try to get the tool - should fail fingerprint validation
	_, err = registry.GetTool("schema-tool")
	if err == nil {
		t.Error("Expected schema modification to be detected, but it was not")
	}
}

func TestToolManagerLifecycle(t *testing.T) {
	manager := NewToolManager("TestServer", "1.0.0", true)

	// Register a tool
	tool := Tool{
		Name:        "lifecycle-tool",
		Description: "A tool for testing lifecycle",
		InputSchema: json.RawMessage(`{"type": "object"}`),
	}

	if err := manager.RegisterTool(tool); err != nil {
		t.Fatalf("Failed to register tool: %v", err)
	}

	// Get the tool
	retrievedTool, err := manager.GetTool("lifecycle-tool")
	if err != nil {
		t.Fatalf("Failed to get tool: %v", err)
	}

	// Verify that the tool is properly signed
	if retrievedTool.SecurityMetadata.Checksum == "" || retrievedTool.SecurityMetadata.Signature == "" {
		t.Error("Tool was not properly signed during registration")
	}

	// List all tools
	toolSet := manager.ListTools()
	if len(toolSet.Tools) != 1 {
		t.Errorf("Expected 1 tool, but got %d", len(toolSet.Tools))
	}
}

func TestToolManagerInitialization(t *testing.T) {
	manager := NewToolManager("TestServer", "1.0.0", true)

	// Create client params
	params := InitializeParams{
		ProtocolVersion: Version,
		Capabilities: ServerToolCapabilities{
			Tools: &ToolCapabilities{
				Security: &SecurityCapabilities{
					SchemaFingerprint:  true,
					ChecksumValidation: true,
				},
			},
		},
		ClientInfo: Implementation{
			Name:    "TestClient",
			Version: "1.0.0",
		},
	}

	// Initialize the server
	result := manager.HandleInitialize(params)

	// Verify the result
	if result.ProtocolVersion != Version {
		t.Errorf("Expected protocol version %s, but got %s", Version, result.ProtocolVersion)
	}

	if result.ServerInfo.Name != "TestServer" {
		t.Errorf("Expected server name TestServer, but got %s", result.ServerInfo.Name)
	}

	if !result.Capabilities.Tools.Security.SchemaFingerprint {
		t.Error("Schema fingerprint capability should be enabled")
	}

	if !result.Capabilities.Tools.Security.ChecksumValidation {
		t.Error("Checksum validation capability should be enabled")
	}
}

func TestUnsignedToolRejection(t *testing.T) {
	// Create a server with security enabled
	server := NewToolManager("TestServer", "1.0.0", true)

	// Set security options
	server.toolRegistry.SetSecurityOptions(true, true)

	// Register a tool without fingerprint and checksum
	tool := Tool{
		Name:        "unsigned-tool",
		Description: "An unsigned tool",
		InputSchema: json.RawMessage(`{"type": "object"}`),
		// Deliberately omit fingerprint and checksum
	}

	// Manually add the tool to bypass the automatic fingerprint/checksum generation
	server.toolRegistry.tools[tool.Name] = tool

	// Try to get the tool - should fail due to missing signatures
	_, err := server.GetTool("unsigned-tool")
	if err == nil {
		t.Error("Expected unsigned tool to be rejected, but it was accepted")
	}
}

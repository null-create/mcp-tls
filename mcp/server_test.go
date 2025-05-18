package mcp

import (
	"encoding/json"
	"testing"
)

func TestServerToolLifecycle(t *testing.T) {
	// Create a server with security enabled
	server := NewServer("TestServer", "1.0.0", true)

	// Register a tool
	tool := Tool{
		Name:        "lifecycle-tool",
		Description: "A tool for testing lifecycle",
		Schema:      json.RawMessage(`{"type": "object"}`),
	}

	if err := server.RegisterTool(tool); err != nil {
		t.Fatalf("Failed to register tool: %v", err)
	}

	// Get the tool
	retrievedTool, err := server.GetTool("lifecycle-tool")
	if err != nil {
		t.Fatalf("Failed to get tool: %v", err)
	}

	// Verify that the tool is properly signed
	if retrievedTool.SecurityMetadata.Checksum == "" || retrievedTool.SecurityMetadata.Signature == "" {
		t.Error("Tool was not properly signed during registration")
	}

	// List all tools
	toolSet := server.ListTools()
	if len(toolSet.Tools) != 1 {
		t.Errorf("Expected 1 tool, but got %d", len(toolSet.Tools))
	}
}

func TestServerInitialization(t *testing.T) {
	// Create a server with security enabled
	server := NewServer("TestServer", "1.0.0", true)

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
	result := server.HandleInitialize(params)

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
	server := NewServer("TestServer", "1.0.0", true)

	// Set security options
	server.toolRegistry.SetSecurityOptions(true, true)

	// Register a tool without fingerprint and checksum
	tool := Tool{
		Name:        "unsigned-tool",
		Description: "An unsigned tool",
		Schema:      json.RawMessage(`{"type": "object"}`),
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

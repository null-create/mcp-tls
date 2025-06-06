package mcp

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
)

// ToolRegistry maintains the set of trusted tools and schemas
type ToolRegistry struct {
	tools               map[string]Tool
	securityEnabled     bool
	validateChecksums   bool
	rejectUnsignedTools bool
}

// NewToolRegistry creates a new tool registry
func NewToolRegistry(securityEnabled bool) *ToolRegistry {
	return &ToolRegistry{
		tools:           make(map[string]Tool),
		securityEnabled: securityEnabled,
	}
}

// SetSecurityOptions configures the security options for the tool registry
func (tr *ToolRegistry) SetSecurityOptions(validateChecksums, rejectUnsignedTools bool) {
	tr.validateChecksums = validateChecksums
	tr.rejectUnsignedTools = rejectUnsignedTools
}

// RegisterTool adds a tool to the registry with security checks
func (tr *ToolRegistry) RegisterTool(tool Tool) error {
	if tr.securityEnabled {
		if tool.SecurityMetadata.Checksum == "" {
			checksum, err := generateToolChecksum(tool)
			if err != nil {
				return err
			}
			tool.SecurityMetadata.Checksum = checksum
		}

		if tool.SecurityMetadata.Signature == "" {
			fingerprint, err := generateSchemaFingerprint(tool.Schema)
			if err != nil {
				return err
			}
			tool.SecurityMetadata.Signature = fingerprint
		}
	}

	tr.tools[tool.Name] = tool
	return nil
}

// GetTool retrieves a tool from the registry with security validation
func (tr *ToolRegistry) GetTool(name string) (Tool, error) {
	tool, exists := tr.tools[name]
	if !exists {
		return Tool{}, fmt.Errorf("tool '%s' not found", name)
	}

	if tr.securityEnabled && tr.validateChecksums {
		expectedChecksum, err := generateToolChecksum(tool)
		if err != nil {
			return Tool{}, fmt.Errorf("failed to generate expected checksum: %v", err)
		}

		if expectedChecksum != tool.SecurityMetadata.Checksum {
			return Tool{}, errors.New("tool checksum validation failed")
		}

		expectedSignature, err := generateSchemaFingerprint(tool.Schema)
		if err != nil {
			return Tool{}, fmt.Errorf("failed to generate expected signature: %v", err)
		}

		if expectedSignature != tool.SecurityMetadata.Signature {
			return Tool{}, errors.New("schema fingerprint validation failed")
		}
	}

	if tr.securityEnabled && tr.rejectUnsignedTools && (tool.SecurityMetadata.Checksum == "" || tool.SecurityMetadata.Signature == "") {
		return Tool{}, errors.New("unsigned tool rejected")
	}

	return tool, nil
}

// ListTools returns all registered tools
func (tr *ToolRegistry) ListTools() ToolSet {
	tools := make([]Tool, 0, len(tr.tools))
	for _, tool := range tr.tools {
		tools = append(tools, tool)
	}

	// Sort tools by name for consistent ordering
	sort.Slice(tools, func(i, j int) bool {
		return tools[i].Name < tools[j].Name
	})

	return ToolSet{
		Tools:                 tools,
		SecurityEnabled:       tr.securityEnabled,
		SchemaFingerprintAlgo: "SHA-256",
		ChecksumAlgo:          "SHA-256",
	}
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

// generateSchemaFingerprint creates a fingerprint of the schema using SHA-256
func generateSchemaFingerprint(schema json.RawMessage) (string, error) {
	canonical, err := canonicalizeJson(schema)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(canonical)
	return hex.EncodeToString(hash[:]), nil
}

// generateToolChecksum creates a checksum of the entire tool definition using SHA-256
func generateToolChecksum(tool Tool) (string, error) {
	toolCopy := Tool{
		Name:        tool.Name,
		Description: tool.Description,
		Schema:      tool.Schema,
	}

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

// ToolVerificationError represents an error during tool verification
type ToolVerificationError struct {
	Message string
	Code    int
}

// Error returns the error message
func (e ToolVerificationError) Error() string {
	return e.Message
}

// ErrorCode constants for tool verification
const (
	ErrChecksumMismatch      int = 4001
	ErrFingerprintMismatch   int = 4002
	ErrUnsignedTool          int = 4003
	ErrToolNotFound          int = 4004
	ErrInvalidToolDefinition int = 4005
)

// ToolManager represents an MCP-TLS server
type ToolManager struct {
	toolRegistry *ToolRegistry
	serverInfo   Implementation
	capabilities ServerCapabilities
}

// NewToolManager creates a new MCP-TLS server tool maanger
func NewToolManager(name, version string, securityEnabled bool) *ToolManager {
	return &ToolManager{
		toolRegistry: NewToolRegistry(securityEnabled),
		serverInfo: Implementation{
			Name:    name,
			Version: version,
		},
		capabilities: ServerCapabilities{
			Tools: &ToolCapabilities{
				ListChanged: true,
				Security: &SecurityCapabilities{
					SchemaFingerprint:  securityEnabled,
					ChecksumValidation: securityEnabled,
				},
			},
		},
	}
}

// HandleInitialize processes an initialize request
func (s *ToolManager) HandleInitialize(params InitializeParams) InitializeResult {
	// Configure security settings based on client capabilities
	if params.Capabilities.Tools != nil && params.Capabilities.Tools.Security != nil {
		s.toolRegistry.SetSecurityOptions(
			params.Capabilities.Tools.Security.ChecksumValidation,
			params.Capabilities.Tools.Security.SchemaFingerprint,
		)
	}

	return InitializeResult{
		ProtocolVersion: Version,
		Capabilities:    s.capabilities,
		ServerInfo:      s.serverInfo,
	}
}

// RegisterTool adds a tool to the server's registry
func (s *ToolManager) RegisterTool(tool Tool) error {
	return s.toolRegistry.RegisterTool(tool)
}

// GetTool retrieves a tool from the server's registry
func (s *ToolManager) GetTool(name string) (Tool, error) {
	return s.toolRegistry.GetTool(name)
}

// ListTools returns all tools registered with the server
func (s *ToolManager) ListTools() ToolSet {
	return s.toolRegistry.ListTools()
}

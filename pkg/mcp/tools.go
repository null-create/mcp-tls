package mcp

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sort"
	"time"
)

// SecurityMetadata contains information used to verify the trust and integrity of components.
type SecurityMetadata struct {
	Source      string `json:"source,omitempty"`        // Origin of the data (e.g., "trusted-registry", "user-provided", "api-endpoint-v2")
	Signature   string `json:"signature,omitempty"`     // Cryptographic signature to verify authenticity/integrity (e.g., JWT, HMAC-SHA256)
	PublicKeyID string `json:"public_key_id,omitempty"` // Identifier for the key needed to verify the signature
	Version     string `json:"version,omitempty"`       // Version identifier for the tool description or other signed component
	Checksum    string `json:"checksum,omitempty"`      // Hash of the component itself (e.g., hash of the ToolDescription structure)
}

func (s *SecurityMetadata) IsEmpty() bool {
	return s.Source == "" && s.Signature == "" &&
		s.PublicKeyID == "" && s.Version == "" &&
		s.Checksum == ""
}

// ToolOption is a function that configures a Tool.
// It provides a flexible way to set various properties of a Tool using the functional options pattern.
type ToolOption func(*Tool)

// ToolInputSchema represents a trusted schema used for validation
type ToolInputSchema struct {
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties,omitempty"`
	Required   []string               `json:"required,omitempty"`
}

// Tool represents a tool definition used by MCP servers and clients
type Tool struct {
	Name             string           `json:"name"`
	Description      string           `json:"description"`
	Arguments        json.RawMessage  `json:"arguments"`
	Parameters       map[string]any   `json:"parameters"`
	InputSchema      json.RawMessage  `json:"inputSchema"`
	OutputSchema     json.RawMessage  `json:"outputSchema"`
	Annotations      ToolAnnotation   `json:"annotations"`
	SecurityMetadata SecurityMetadata `json:"secMetaData"`
}

// ToolSet represents a collection of tools with security information
type ToolSet struct {
	Tools                 []Tool `json:"tools"`
	SecurityEnabled       bool   `json:"securityEnabled"`
	SchemaFingerprintAlgo string `json:"schemaFingerprintAlgo,omitempty"`
	ChecksumAlgo          string `json:"checksumAlgo,omitempty"`
}

type ToolAnnotation struct {
	// Human-readable title for the tool
	Title string `json:"title,omitempty"`
	// If true, the tool does not modify its environment
	ReadOnlyHint bool `json:"readOnlyHint,omitempty"`
	// If true, the tool may perform destructive updates
	DestructiveHint bool `json:"destructiveHint,omitempty"`
	// If true, repeated calls with same args have no additional effect
	IdempotentHint bool `json:"idempotentHint,omitempty"`
	// If true, tool interacts with external entities
	OpenWorldHint bool `json:"openWorldHint,omitempty"`
}

// NewTool creates a new Tool with the given name and options.
// The tool will have an object-type input schema with configurable properties.
// Options are applied in order, allowing for flexible tool configuration.
func NewTool(name string, opts ...ToolOption) Tool {
	inputSchema, err := json.Marshal(ToolInputSchema{
		Type:       "object",
		Properties: make(map[string]any),
		Required:   nil, // Will be omitted from JSON if empty
	})
	if err != nil {
		log.Fatal(err)
	}

	tool := Tool{
		Name:        name,
		InputSchema: inputSchema,
		Annotations: ToolAnnotation{
			Title:           "",
			ReadOnlyHint:    false,
			DestructiveHint: true,
			IdempotentHint:  false,
			OpenWorldHint:   true,
		},
	}

	for _, opt := range opts {
		opt(&tool)
	}

	return tool
}

type ExecutionStatus string

const (
	StatusSucceeded ExecutionStatus = "succeeded"
	StatusFailed    ExecutionStatus = "failed" // Tool executed but produced an error or unwanted result
	StatusError     ExecutionStatus = "error"  // System-level error trying to execute the tool
)

type ToolDescription struct {
	Name         string          `json:"name"`
	Description  string          `json:"description"`
	InputSchema  json.RawMessage `json:"input_schema"`            // Expects JSON Schema definition here
	OutputSchema json.RawMessage `json:"output_schema,omitempty"` // Optional: Schema for the tool's result
}

// ToolValidationResult details the results of a tool validation process
type ToolValidationResult struct {
	Name     string `json:"name"`
	Checksum string `json:"checksum,omitempty"`
	Valid    bool   `json:"valid"`
	Error    string `json:"error,omitempty"`
}

// ToolRegistry maintains the set of trusted tools and schemas
// used for validation
type ToolRegistry struct {
	toolRepo            string // URL to exteral repository of trusted tools
	apiKey              string // API key to trust tool repo
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

// Configure the remote tool repo credentials
func (tr *ToolRegistry) SetRegistryCreds(url, apiKey string) {
	tr.toolRepo = url
	tr.apiKey = apiKey
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
			fingerprint, err := generateSchemaFingerprint(tool.InputSchema)
			if err != nil {
				return err
			}
			tool.SecurityMetadata.Signature = fingerprint
		}
	}
	if _, ok := tr.tools[tool.Name]; !ok {
		tr.tools[tool.Name] = tool
	}
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

		expectedSignature, err := generateSchemaFingerprint(tool.InputSchema)
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

// LoadTools retrieves all trusted tool schema definitions
// into the internal map. These definitions are not exported anywhere
// since the validator is intended to be stateless.
func (tr *ToolRegistry) LoadTools() error {
	if tr.apiKey == "" || tr.toolRepo == "" {
		return fmt.Errorf("missing tool repo credentials")
	}

	// API call to get list of trusted tool schemas
	client := http.Client{Timeout: time.Second * 3}

	req, err := http.NewRequest(http.MethodGet, tr.toolRepo, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received non-200 status: %d", resp.StatusCode)
	}

	// parse results into mcp.Tool objects and add to internal map
	var tools map[string]Tool
	if err = json.NewDecoder(resp.Body).Decode(&tools); err != nil {
		return err
	}

	tr.tools = tools

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
		InputSchema: tool.InputSchema,
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
func (t *ToolManager) RegisterTool(tool Tool) error {
	return t.toolRegistry.RegisterTool(tool)
}

// GetTool retrieves a tool from the server's registry
func (t *ToolManager) GetTool(name string) (Tool, error) {
	return t.toolRegistry.GetTool(name)
}

// ListTools returns all tools registered with the server
func (t *ToolManager) ListTools() ToolSet {
	return t.toolRegistry.ListTools()
}

// LoadTools retrieves all trusted tools from an external API
func (t *ToolManager) LoadTools() error {
	return t.toolRegistry.LoadTools()
}

// GetTools returns all tools available from the internal tool registry
func (t *ToolManager) GetTools() []Tool {
	return t.toolRegistry.ListTools().Tools
}

// SchemaFingerprint generates a hash for a given tools schema
func (t *ToolManager) SchemaFingerprint(tool *Tool) error {
	fingerPrint, err := generateSchemaFingerprint(tool.InputSchema)
	if err != nil {
		return err
	}
	tool.SecurityMetadata.Signature = fingerPrint
	return nil
}

// ToolChecksum creates a checksum of the entire tool definition using SHA-256
func (t *ToolManager) ToolChecksum(tool *Tool) error {
	checkSum, err := generateToolChecksum(*tool)
	if err != nil {
		return err
	}
	tool.SecurityMetadata.Checksum = checkSum
	return nil
}

// SecureTool adds security metadata to a tool
func SecureTool(tool *Tool) error {
	// Generate fingerprint from parameters schema
	fingerprint, err := generateSchemaFingerprint(tool.InputSchema)
	if err != nil {
		return err
	}

	// Generate checksum from parameters schema
	checksum, err := generateToolChecksum(*tool)
	if err != nil {
		return err
	}

	// Set security metadata
	tool.SecurityMetadata = SecurityMetadata{
		Signature: fingerprint,
		Checksum:  checksum,
	}

	return nil
}

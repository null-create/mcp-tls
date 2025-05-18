package mcp

// --- MCP Handshake Specific Structures ---

// Version represents the MCP-TLS protocol version
const Version = "2025-03-26"

// Capabilities Structures
type RootCapabilities struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

type SamplingCapabilities struct {
	// Empty object {} indicates support
}

type LoggingCapabilities struct {
	// Empty object {} indicates support
}

type ToolValidationCapabilities struct {
	// Empty object {} indicates support
}

type PromptCapabilities struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

type ResourceCapabilities struct {
	Subscribe   bool `json:"subscribe,omitempty"`
	ListChanged bool `json:"listChanged,omitempty"`
}

// SecurityCapabilities defines the security-related capabilities for MCP-TLS
type SecurityCapabilities struct {
	SchemaFingerprint  bool `json:"schemaFingerprint"`
	ChecksumValidation bool `json:"checksumValidation"`
}

// ClientSecurityCapabilities defines the security capabilities that a client can request
type ClientSecurityCapabilities struct {
	ValidateChecksums   bool `json:"validateChecksums"`
	RejectUnsignedTools bool `json:"rejectUnsignedTools"`
}

// ToolCapabilities defines the capabilities related to tools
type ToolCapabilities struct {
	ListChanged bool                  `json:"listChanged,omitempty"`
	Security    *SecurityCapabilities `json:"security,omitempty"`
}

// ClientToolCapabilities defines the tool capabilities that a client can request
type ClientToolCapabilities struct {
	Security *ClientSecurityCapabilities `json:"security,omitempty"`
}

// ServerToolCapabilities represents all capabilities supported by the server
type ServerToolCapabilities struct {
	Tools *ToolCapabilities `json:"tools,omitempty"`
}

// Use map for flexibility with experimental features
type ExperimentalCapabilities map[string]any

type ClientCapabilities struct {
	Roots        *RootCapabilities          `json:"roots,omitempty"`
	Sampling     *SamplingCapabilities      `json:"sampling,omitempty"`
	Experimental ExperimentalCapabilities   `json:"experimental,omitempty"`
	Security     ClientSecurityCapabilities `json:"security"`
}

func NewClientCapabilities() ClientCapabilities {
	return ClientCapabilities{
		Roots:        &RootCapabilities{ListChanged: true},
		Sampling:     &SamplingCapabilities{},
		Experimental: ExperimentalCapabilities{},
		Security:     ClientSecurityCapabilities{},
	}
}

type ServerCapabilities struct {
	Subscribe    bool                     `json:"subscribe,omitempty"`
	ListChanged  bool                     `json:"listChanged,omitempty"`
	Logging      *LoggingCapabilities     `json:"logging,omitempty"`
	Prompts      *PromptCapabilities      `json:"prompts,omitempty"`
	Resources    *ResourceCapabilities    `json:"resources,omitempty"`
	Tools        *ToolCapabilities        `json:"tools,omitempty"`
	Experimental ExperimentalCapabilities `json:"experimental,omitempty"`
	Security     SecurityCapabilities     `json:"security"`
}

// Request is a message that expects a response
// It corresponds to a method call with optional parameters.
type Request struct {
	Method string `json:"method"`
	Params any    `json:"params,omitempty"`
}

type Result struct {
	// This result property is reserved by the protocol to allow clients and
	// servers to attach additional metadata to their responses.
	Meta map[string]interface{} `json:"_meta,omitempty"`
}

// Implementation describes the name and version of an MCP implementation.
type Implementation struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// InitializeRequest is sent from the client to the server when it first
// connects, asking it to begin initialization.
type InitializeRequest struct {
	Request
	Params struct {
		// The latest version of the Model Context Protocol that the client supports.
		// The client MAY decide to support older versions as well.
		ProtocolVersion string             `json:"protocolVersion"`
		Capabilities    ClientCapabilities `json:"capabilities"`
		ClientInfo      Implementation     `json:"clientInfo"`
	} `json:"params"`
}

// InitializeResult is sent after receiving an initialize request from the
// client.
type InitializeResult struct {
	Result
	// The version of the Model Context Protocol that the server wants to use.
	// This may not match the version that the client requested. If the client cannot
	// support this version, it MUST disconnect.
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ServerCapabilities `json:"capabilities"`
	ServerInfo      Implementation     `json:"serverInfo"`
	// Instructions describing how to use the server and its features.
	//
	// This can be used by clients to improve the LLM's understanding of
	// available tools, resources, etc. It can be thought of like a "hint" to the model.
	// For example, this information MAY be added to the system prompt.
	Instructions string `json:"instructions,omitempty"`
}

// InitializeParams represents parameters for the initialize method
type InitializeParams struct {
	ProtocolVersion string                 `json:"protocolVersion"`
	Capabilities    ServerToolCapabilities `json:"capabilities"`
	ClientInfo      Implementation         `json:"clientInfo"`
}

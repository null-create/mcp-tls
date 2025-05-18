package mcp

// Server represents an MCP-TLS server
type Server struct {
	toolRegistry *ToolRegistry
	serverInfo   Implementation
	capabilities ServerCapabilities
}

// NewServer creates a new MCP-TLS server
func NewServer(name, version string, securityEnabled bool) *Server {
	return &Server{
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
func (s *Server) HandleInitialize(params InitializeParams) InitializeResult {
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
func (s *Server) RegisterTool(tool Tool) error {
	return s.toolRegistry.RegisterTool(tool)
}

// GetTool retrieves a tool from the server's registry
func (s *Server) GetTool(name string) (Tool, error) {
	return s.toolRegistry.GetTool(name)
}

// ListTools returns all tools registered with the server
func (s *Server) ListTools() ToolSet {
	return s.toolRegistry.ListTools()
}

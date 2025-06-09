package server

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/null-create/mcp-tls/pkg/mcp"
	"github.com/null-create/mcp-tls/pkg/util"
	"github.com/null-create/mcp-tls/pkg/validate"
)

const (
	proxyListenAddr  = ":9000"
	targetServerAddr = "localhost:9001"
)

type Handlers struct {
	ClientURL   string
	ServerURL   string
	toolManager *mcp.ToolManager
}

func NewHandler() Handlers {
	return Handlers{
		toolManager: mcp.NewToolManager("mcp-tls-tool-manager", "1.0.0", true),
	}
}

func (h *Handlers) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	if err := json.NewEncoder(w).Encode(`{"status":"ok"}`); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handlers) ValidateToolHandler(w http.ResponseWriter, r *http.Request) {
	var tool mcp.Tool
	if err := json.NewDecoder(r.Body).Decode(&tool); err != nil {
		util.WriteError(w, http.StatusBadRequest, "Invalid tool JSON: "+err.Error())
		return
	}

	result := h.validate(&tool)

	util.WriteJSON(w, result)
}

func (h *Handlers) ValidateToolsHandler(w http.ResponseWriter, r *http.Request) {
	var tools []mcp.Tool
	if err := json.NewDecoder(r.Body).Decode(&tools); err != nil {
		util.WriteError(w, http.StatusBadRequest, "Invalid JSON array: "+err.Error())
		return
	}

	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		results = make([]mcp.ToolValidationResult, 0, len(tools))
	)

	for _, tool := range tools {
		wg.Add(1)
		go func() {
			defer wg.Done()

			result := h.validate(&tool)

			mu.Lock()
			results = append(results, result)
			mu.Unlock()
		}()

	}
	wg.Wait()

	util.WriteJSON(w, results)
}

func (h *Handlers) validate(tool *mcp.Tool) mcp.ToolValidationResult {
	origTool, err := h.toolManager.GetTool(tool.Name)
	if err != nil {
		return mcp.ToolValidationResult{
			Name:  tool.Name,
			Valid: false,
			Error: err.Error(),
		}
	}

	if tool.SecurityMetadata.Signature != origTool.SecurityMetadata.Signature ||
		tool.SecurityMetadata.Checksum != origTool.SecurityMetadata.Checksum {
		return mcp.ToolValidationResult{
			Name:  tool.Name,
			Valid: false,
			Error: "signature or checksum mismatch",
		}
	}

	// validate tool description
	err = validate.ValidateToolDescription(tool.Description)
	if err != nil {
		return mcp.ToolValidationResult{
			Name:  tool.Name,
			Valid: false,
			Error: err.Error(),
		}
	}

	// validate tool schema
	status, err := validate.ValidateToolInputSchema(tool, tool.Arguments)
	if err != nil {
		return mcp.ToolValidationResult{
			Name:  tool.Name,
			Valid: false,
			Error: err.Error(),
		}
	}
	if status == validate.StatusFailed {
		return mcp.ToolValidationResult{
			Name:  tool.Name,
			Valid: false,
			Error: "validation failed",
		}
	}

	return mcp.ToolValidationResult{
		Name:     tool.Name,
		Valid:    true,
		Checksum: tool.SecurityMetadata.Checksum,
	}
}

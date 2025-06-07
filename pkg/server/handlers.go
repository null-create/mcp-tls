package server

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"maps"

	"github.com/null-create/mcp-tls/pkg/mcp"
	"github.com/null-create/mcp-tls/pkg/util"
)

type Handlers struct {
	TargetURL   string // url to pass requests or responses to
	toolManager *mcp.ToolManager
}

func NewHandler() Handlers {
	return Handlers{
		TargetURL:   "",
		toolManager: mcp.NewToolManager("mcp-tls-tool-manager", "1.0.0", true),
	}
}

func (h *Handlers) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	if err := json.NewEncoder(w).Encode(`{"status":"ok"}`); err != nil {
		log.Printf("ERROR: failed to encode health check response: %v", err)
	}
}

func (h *Handlers) ValidateToolHandler(w http.ResponseWriter, r *http.Request) {
	var tool mcp.Tool // or ToolDescription? This is what's used in the validate module
	if err := json.NewDecoder(r.Body).Decode(&tool); err != nil {
		util.WriteError(w, http.StatusBadRequest, "Invalid tool JSON: "+err.Error())
		return
	}

	hash, err := mcp.CanonicalizeAndHash(tool)
	if err != nil {
		util.WriteJSON(w, mcp.ToolValidationResult{
			Name:  tool.Name,
			Valid: false,
			Error: err.Error(),
		})
		return
	}

	// TODO: add validation logic

	util.WriteJSON(w, mcp.ToolValidationResult{
		Name:     tool.Name,
		Checksum: hash,
		Valid:    true,
	})
}

func (h *Handlers) ValidateToolsHandler(w http.ResponseWriter, r *http.Request) {
	var tools []mcp.Tool // or ToolDescription? This is what's used in the validate module
	if err := json.NewDecoder(r.Body).Decode(&tools); err != nil {
		util.WriteError(w, http.StatusBadRequest, "Invalid JSON array: "+err.Error())
		return
	}

	results := make([]mcp.ToolValidationResult, 0, len(tools))
	for _, tool := range tools {
		hash, err := mcp.CanonicalizeAndHash(tool)
		if err != nil {
			results = append(results, mcp.ToolValidationResult{
				Name:  tool.Name,
				Valid: false,
				Error: err.Error(),
			})
		} else {

			// TODO: validate each tool in their own goroutine

			results = append(results, mcp.ToolValidationResult{
				Name:     tool.Name,
				Valid:    true,
				Checksum: hash,
			})
		}
	}

	util.WriteJSON(w, results)
}

// ProxyHandler handles both directions of JSON-RPC over HTTP.
func (h *Handlers) ProxyHandler(w http.ResponseWriter, r *http.Request) {
	h.proxy(w, r, h.TargetURL)
}

func (h *Handlers) proxy(w http.ResponseWriter, r *http.Request, targetURL string) {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Unable to read request", http.StatusBadRequest)
		log.Println("Error reading body:", err)
		return
	}
	r.Body.Close()

	// Forward request
	req, err := http.NewRequest(r.Method, targetURL, bytes.NewReader(bodyBytes))
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		log.Println("Error creating request:", err)
		return
	}
	req.Header = r.Header.Clone()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, "MCP server unreachable", http.StatusBadGateway)
		log.Println("Error contacting MCP server:", err)
		return
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read server response", http.StatusInternalServerError)
		log.Println("Error reading server response:", err)
		return
	}

	// Copy headers and status code
	maps.Copy(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	w.Write(respBody)
}

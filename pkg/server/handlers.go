package server

import (
	"encoding/json"
	"net/http"

	"github.com/null-create/mcp-tls/pkg/mcp"
	"github.com/null-create/mcp-tls/pkg/util"
)

type ToolValidationResult struct {
	Name     string `json:"name"`
	Checksum string `json:"checksum,omitempty"`
	Valid    bool   `json:"valid"`
	Error    string `json:"error,omitempty"`
}

func ValidateToolHandler(w http.ResponseWriter, r *http.Request) {
	var tool mcp.Tool
	if err := json.NewDecoder(r.Body).Decode(&tool); err != nil {
		util.WriteError(w, http.StatusBadRequest, "Invalid tool JSON: "+err.Error())
		return
	}

	hash, err := mcp.CanonicalizeAndHash(tool)
	if err != nil {
		util.WriteJSON(w, ToolValidationResult{
			Name:  tool.Name,
			Valid: false,
			Error: err.Error(),
		})
		return
	}

	util.WriteJSON(w, ToolValidationResult{
		Name:     tool.Name,
		Checksum: hash,
		Valid:    true,
	})
}

func ValidateToolsHandler(w http.ResponseWriter, r *http.Request) {
	var tools []mcp.Tool
	if err := json.NewDecoder(r.Body).Decode(&tools); err != nil {
		util.WriteError(w, http.StatusBadRequest, "Invalid JSON array: "+err.Error())
		return
	}

	results := make([]ToolValidationResult, 0, len(tools))
	for _, tool := range tools {
		hash, err := mcp.CanonicalizeAndHash(tool)
		if err != nil {
			results = append(results, ToolValidationResult{
				Name:  tool.Name,
				Valid: false,
				Error: err.Error(),
			})
		} else {
			results = append(results, ToolValidationResult{
				Name:     tool.Name,
				Valid:    true,
				Checksum: hash,
			})
		}
	}

	util.WriteJSON(w, results)
}

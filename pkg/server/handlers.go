package server

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"

	"github.com/null-create/mcp-tls/pkg/codec"
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
		ClientURL:   "",
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

	// validate tool description
	err = validate.ValidateToolDescription(tool.Description)
	if err != nil {
		util.WriteJSON(w, mcp.ToolValidationResult{
			Name:  tool.Name,
			Valid: false,
			Error: err.Error(),
		})
		return
	}

	// validate tool schema

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

// Intercepts client-to-server and validates tool call requests
func (h *Handlers) validateAndForward(data []byte) ([]byte, error) {
	var req codec.JSONRPCRequest
	if err := json.Unmarshal(data, &req); err != nil {
		log.Println("Invalid JSON-RPC:", err)
		return nil, err
	}

	if req.Method == "tool.call" {
		var tool mcp.Tool
		if err := json.Unmarshal(req.Params, &tool); err != nil {
			log.Printf("Failed to unmarshal request params to tool description object: %v", err)
			return nil, err
		}

		status, err := validate.ValidateToolInputSchema(context.Background(), &tool, nil)
		if err != nil {
			log.Printf("Failed to validate tool schema: %v", err)
			return nil, err
		}
		if status == mcp.StatusSucceeded {
			// valid schema. validate description before passing onward
			return json.Marshal(req)
		}
	}
	return json.Marshal(codec.JSONRPCError{
		Code: codec.INVALID_REQUEST,
	})
}

func (h *Handlers) handleConnection(clientConn net.Conn) {
	defer clientConn.Close()

	serverConn, err := net.Dial("tcp", targetServerAddr)
	if err != nil {
		log.Printf("Failed to connect to MCP server: %v", err)
		return
	}
	defer serverConn.Close()

	go h.proxyStream(clientConn, serverConn, h.validateAndForward)
	h.proxyStream(serverConn, clientConn, h.passthrough)
}

// Simple passthrough for server-to-client direction
func (h *Handlers) passthrough(data []byte) ([]byte, error) {
	return data, nil
}

type toolError string

func (e toolError) Error() string { return string(e) }

func ErrInvalidTool(msg string) error { return toolError("Invalid tool call: " + msg) }

// Handles framed JSON messages over TCP (e.g., newline-delimited)
func (h *Handlers) proxyStream(src, dst net.Conn, transform func([]byte) ([]byte, error)) {
	reader := bufio.NewReader(src)
	writer := bufio.NewWriter(dst)

	for {
		line, err := reader.ReadBytes('\n') // framing logic (newline-delimited)
		if err != nil {
			if err != io.EOF {
				log.Printf("Stream read error: %v", err)
			}
			return
		}

		processed, err := transform(line)
		if err != nil {
			log.Printf("Processing error: %v", err)
			return
		}

		writer.Write(processed)
		writer.Flush()
	}
}

func (h *Handlers) proxy() {
	listener, err := net.Listen("tcp", proxyListenAddr)
	if err != nil {
		log.Fatalf("Proxy listen failed: %v", err)
	}
	log.Printf("MCP proxy listening on %s → %s", proxyListenAddr, targetServerAddr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Connection accept failed: %v", err)
			continue
		}
		go h.handleConnection(conn)
	}
}

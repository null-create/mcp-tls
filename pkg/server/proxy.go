package server

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
	"net"

	"github.com/null-create/mcp-tls/pkg/codec"
	"github.com/null-create/mcp-tls/pkg/mcp"
	"github.com/null-create/mcp-tls/pkg/validate"
)

// ---- Proxy handlers

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

		status, err := validate.ValidateToolInputSchema(&tool, tool.Arguments)
		if err != nil {
			log.Printf("Failed to validate tool schema: %v", err)
			return nil, err
		}
		// valid schema. validate description before passing onward
		if status == validate.StatusSucceeded {
			if err := validate.ValidateToolDescription(tool.Description); err != nil {
				return nil, err
			}
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

func Proxy() {
	listener, err := net.Listen("tcp", proxyListenAddr)
	if err != nil {
		log.Fatalf("Proxy listen failed: %v", err)
	}
	log.Printf("MCP proxy listening on %s → %s", proxyListenAddr, targetServerAddr)

	h := NewHandler()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Connection accept failed: %v", err)
			continue
		}
		go h.handleConnection(conn)
	}
}

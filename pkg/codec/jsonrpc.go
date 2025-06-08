package codec

import (
	"encoding/json"
	"maps"
)

// Supported JSON-RPC version
const JsonRPCVersion string = "2.0"

// JSON-RPC 2.0 standard error codes
const (
	PARSE_ERROR      = -32700
	INVALID_REQUEST  = -32600
	METHOD_NOT_FOUND = -32601
	INVALID_PARAMS   = -32602
	INTERNAL_ERROR   = -32603
)

// Generic interface for JSON RPC Messages
type JSONRPCMessage any

type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
	ID      int64           `json:"id"`
}

func (j *JSONRPCRequest) MarshalJSON() ([]byte, error) {
	b, err := json.Marshal(j)
	if err != nil {
		return nil, err
	}
	return b, nil
}

type JSONRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *JSONRPCError   `json:"error,omitempty"`
	ID      int64           `json:"id"`
}

func (j *JSONRPCResponse) MarshalJSON() ([]byte, error) {
	b, err := json.Marshal(j.Result)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func NewJSONRPCResponse() JSONRPCResponse {
	return JSONRPCResponse{
		JSONRPC: JsonRPCVersion,
	}
}

type JSONRCPNotification struct {
	JSONRPC string `json:"jsonrpc"`
	Notification
}

func (j *JSONRCPNotification) MarshalJSON() ([]byte, error) {
	b, err := json.Marshal(j)
	if err != nil {
		return nil, err
	}
	return b, nil
}

type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func (r *JSONRPCError) ErrCode() int { return r.Code }
func (r *JSONRPCError) Msg() string  { return r.Message }

type Notification struct {
	Method string             `json:"method"`
	Params NotificationParams `json:"params,omitempty"` // Often null/omitted for simple notifications
}

func (n *Notification) MarshalJSON() ([]byte, error) {
	b, err := json.Marshal(n)
	if err != nil {
		return nil, err
	}
	return b, nil
}

type NotificationParams struct {
	Meta             map[string]any `json:"_meta,omitempty"`
	AdditionalFields map[string]any `json:"-"`
}

func (n NotificationParams) MarshalJSON() ([]byte, error) {
	base := make(map[string]any)

	if n.Meta != nil {
		base["_meta"] = n.Meta
	}

	maps.Copy(base, n.AdditionalFields)

	return json.Marshal(base)
}

func (p *NotificationParams) UnmarshalJSON(data []byte) error {
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	p.AdditionalFields = raw

	if meta, ok := raw["_meta"]; ok {
		if metaMap, ok := meta.(map[string]any); ok {
			p.Meta = metaMap
		}
	}

	return nil
}

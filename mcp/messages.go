package mcp

import (
	"encoding/json"
	"time"
)

// General SSE event/data handler
type Handler func(event, data string) error

// Message Handlers work with stdio in the Transport layer
type MessageHandler func(event string, message json.RawMessage) error

// IOHandler works with stdio in the I/O layer
type IOHandler func(message json.RawMessage) error

// ExecutionStatus indicates the outcome of a tool execution attempt.
type ExecutionStatus string

// Cursor is an opaque token used to represent a cursor for pagination.
type Cursor string

/* Empty result */

// EmptyResult represents a response that indicates success but carries no data.
type EmptyResult Result

const (
	StatusSucceeded ExecutionStatus = "succeeded"
	StatusFailed    ExecutionStatus = "failed" // Tool executed but produced an error or unwanted result
	StatusError     ExecutionStatus = "error"  // System-level error trying to execute the tool
)

// Request is a message that expects a response
// It corresponds to a method call with optional parameters.
type Request struct {
	Method string `json:"method"`
	Params any    `json:"params,omitempty"`
}

/* Tools */

type ChatCompletionRequest struct {
	Model    string `json:"model"`
	Messages []struct {
		Role      string     `json:"role"`
		Content   string     `json:"content"`
		ToolCalls []ToolCall `json:"tool_calls,omitempty"`
	} `json:"messages"`
	Tools  []ToolDefinition `json:"tools,omitempty"`
	Stream bool             `json:"stream,omitempty"`
}

// CompleteRequest requests completion options for a given argument or context.
type CompleteRequest struct {
	Request
	ContextID string `json:"contextId"`
	Argument  string `json:"argument"`
}

// CompleteResult contains the list of possible completions.
type CompleteResult struct {
	Completions []string `json:"completions"`
}

// Message represents a single turn or piece of information in the interaction history.
type Message struct {
	ID         string              `json:"id"`                     // Unique identifier for this message
	Role       Role                `json:"role"`                   // Who sent this message?
	Content    string              `json:"content"`                // Text content of the message (or tool result data)
	Timestamp  time.Time           `json:"timestamp"`              // Time the message was generated
	ToolCalls  []ToolCall          `json:"tool_calls,omitempty"`   // Assistant requests to call tools (only if Role == RoleAssistant)
	ToolCallID string              `json:"tool_call_id,omitempty"` // Links a Tool Result message back to its request (only if Role == RoleTool)
	ToolResult *ToolResultMetadata `json:"tool_result,omitempty"`  // Metadata about the tool execution (only if Role == RoleTool)
	// Security Note: While content might contain sensitive data, the MCP structure itself
	// should ideally not add *new* vulnerabilities. The focus here is on tool interaction safety.
}

type Annotations struct {
	// Describes who the intended customer of this object or data is.
	//
	// It can include multiple entries to indicate content useful for multiple
	// audiences (e.g., `["user", "assistant"]`).
	Audience []Role `json:"audience,omitempty"`

	// Describes how important this data is for operating the server.
	//
	// A value of 1 means "most important," and indicates that the data is
	// effectively required, while 0 means "least important," and indicates that
	// the data is entirely optional.
	Priority float64 `json:"priority,omitempty"`
}

// Annotated is the base for objects that include optional annotations for the
// client. The client can use annotations to inform how objects are used or
// displayed
type Annotated struct {
	Annotations *Annotations `json:"annotations,omitempty"`
}

type Content interface {
	isContent()
}

// TextContent represents text provided to or from an LLM.
// It must have Type set to "text".
type TextContent struct {
	Annotated
	Type string `json:"type"` // Must be "text"
	// The text content of the message.
	Text string `json:"text"`
}

func (TextContent) isContent() {}

// ImageContent represents an image provided to or from an LLM.
// It must have Type set to "image".
type ImageContent struct {
	Annotated
	Type string `json:"type"` // Must be "image"
	// The base64-encoded image data.
	Data string `json:"data"`
	// The MIME type of the image. Different providers may support different image types.
	MIMEType string `json:"mimeType"`
}

func (ImageContent) isContent() {}

/* Pagination */

type PaginatedRequest struct {
	Request
	Params struct {
		// An opaque token representing the current pagination position.
		// If provided, the server should return results starting after this cursor.
		Cursor Cursor `json:"cursor,omitempty"`
	} `json:"params,omitempty"`
}

type PaginatedResult struct {
	Result
	// An opaque token representing the pagination position after the last
	// returned result.
	// If present, there may be more results available.
	NextCursor Cursor `json:"nextCursor,omitempty"`
}

package mcp

// ExecutionStatus indicates the outcome of a tool execution attempt.
type ExecutionStatus string

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
	Tools  []Tool `json:"tools,omitempty"`
	Stream bool   `json:"stream,omitempty"`
}

type Content interface {
	isContent()
}

package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/yourusername/perplexity-mcp-golang/internal/core/domain"
)

// MCPRequest represents a generic MCP protocol request
type MCPRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      any         `json:"id,omitempty"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// MCPResponse represents a generic MCP protocol response
type MCPResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      any         `json:"id,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *MCPError   `json:"error,omitempty"`
}

// MCPError represents an error in MCP protocol format
type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// MCP error codes
const (
	MCPErrorParseError     = -32700
	MCPErrorInvalidRequest = -32600
	MCPErrorMethodNotFound = -32601
	MCPErrorInvalidParams  = -32602
	MCPErrorInternalError  = -32603
	MCPErrorServerError    = -32000
)

// ToolCallParams represents the parameters for a tool call
type ToolCallParams struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments,omitempty"`
}

// ToolListResult represents the result of listing tools
type ToolListResult struct {
	Tools []domain.ToolInfo `json:"tools"`
}

// ToolCallResult represents the result of calling a tool
type ToolCallResult struct {
	Content   string            `json:"content"`
	IsError   bool              `json:"isError"`
	Metadata  map[string]any    `json:"metadata,omitempty"`
	Citations []domain.Citation `json:"citations,omitempty"`
}

// MCPHandler provides MCP protocol handling functionality
type MCPHandler struct {
	server *Server
	logger domain.Logger
}

// NewMCPHandler creates a new MCP protocol handler
func NewMCPHandler(server *Server, logger domain.Logger) *MCPHandler {
	return &MCPHandler{
		server: server,
		logger: logger,
	}
}

// HandleRequest processes an MCP protocol request
func (h *MCPHandler) HandleRequest(ctx context.Context, requestData []byte) ([]byte, error) {
	// Add timeout to context if not already set
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, DefaultTimeout)
		defer cancel()
	}

	h.logger.Debug("Handling MCP request", "request_size", len(requestData))

	// Parse the request
	var request MCPRequest
	if err := json.Unmarshal(requestData, &request); err != nil {
		h.logger.Error("Failed to parse MCP request", "error", err.Error())
		return h.createErrorResponse(nil, MCPErrorParseError, "Parse error", err)
	}

	// Validate JSONRPC version
	if request.JSONRPC != "2.0" {
		h.logger.Error("Invalid JSONRPC version", "version", request.JSONRPC)
		return h.createErrorResponse(request.ID, MCPErrorInvalidRequest, "Invalid JSONRPC version", nil)
	}

	h.logger.Info("Processing MCP request", "method", request.Method, "id", request.ID)

	// Route the request based on method
	var result interface{}
	var err error

	switch request.Method {
	case "tools/list":
		result, err = h.handleToolsList(ctx, request.Params)
	case "tools/call":
		result, err = h.handleToolCall(ctx, request.Params)
	default:
		h.logger.Error("Unknown MCP method", "method", request.Method)
		return h.createErrorResponse(request.ID, MCPErrorMethodNotFound, fmt.Sprintf("Method not found: %s", request.Method), nil)
	}

	// Handle execution errors
	if err != nil {
		h.logger.Error("MCP method execution failed",
			"method", request.Method,
			"error", err.Error())
		return h.createErrorResponse(request.ID, MCPErrorServerError, "Server error", err)
	}

	// Create successful response
	response := MCPResponse{
		JSONRPC: "2.0",
		ID:      request.ID,
		Result:  result,
	}

	responseData, err := json.Marshal(response)
	if err != nil {
		h.logger.Error("Failed to marshal MCP response", "error", err.Error())
		return h.createErrorResponse(request.ID, MCPErrorInternalError, "Internal error", err)
	}

	h.logger.Info("MCP request processed successfully",
		"method", request.Method,
		"response_size", len(responseData))

	return responseData, nil
}

// handleToolsList handles the tools/list method
func (h *MCPHandler) handleToolsList(ctx context.Context, params interface{}) (interface{}, error) {
	h.logger.Debug("Handling tools/list request")

	// Get tools from server
	tools, err := h.server.ListTools(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list tools: %w", err)
	}

	result := ToolListResult{
		Tools: tools,
	}

	h.logger.Info("Tools listed successfully", "count", len(tools))

	return result, nil
}

// handleToolCall handles the tools/call method
func (h *MCPHandler) handleToolCall(ctx context.Context, params interface{}) (interface{}, error) {
	h.logger.Debug("Handling tools/call request")

	// Parse parameters
	paramsData, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal params: %w", err)
	}

	var toolParams ToolCallParams
	if err := json.Unmarshal(paramsData, &toolParams); err != nil {
		return nil, fmt.Errorf("invalid tool call parameters: %w", err)
	}

	if toolParams.Name == "" {
		return nil, fmt.Errorf("%w: tool name is required", domain.ErrInvalidRequest)
	}

	h.logger.Debug("Parsed tool call parameters",
		"tool_name", toolParams.Name,
		"args_count", len(toolParams.Arguments))

	// Execute the tool
	toolResult, err := h.server.ExecuteTool(ctx, toolParams.Name, toolParams.Arguments)
	if err != nil {
		// Check if this is a tool not found error
		if err == domain.ErrToolNotFound {
			return nil, fmt.Errorf("tool not found: %s", toolParams.Name)
		}
		return nil, fmt.Errorf("tool execution failed: %w", err)
	}

	// Convert domain ToolResult to MCP ToolCallResult
	result := ToolCallResult{
		Content:   toolResult.Content,
		IsError:   toolResult.IsError,
		Metadata:  toolResult.Metadata,
		Citations: toolResult.Citations,
	}

	h.logger.Info("Tool call completed successfully",
		"tool_name", toolParams.Name,
		"result_error", result.IsError,
		"content_length", len(result.Content))

	return result, nil
}

// createErrorResponse creates an error response in MCP format
func (h *MCPHandler) createErrorResponse(id any, code int, message string, data error) ([]byte, error) {
	mcpError := &MCPError{
		Code:    code,
		Message: message,
	}

	if data != nil {
		mcpError.Data = data.Error()
	}

	response := MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error:   mcpError,
	}

	responseData, err := json.Marshal(response)
	if err != nil {
		// Fallback to a basic error response if marshaling fails
		fallbackResponse := `{"jsonrpc":"2.0","error":{"code":-32603,"message":"Internal error"}}`
		return []byte(fallbackResponse), err
	}

	return responseData, nil
}

// ValidateRequest validates an MCP request structure
func (h *MCPHandler) ValidateRequest(requestData []byte) error {
	var request MCPRequest
	if err := json.Unmarshal(requestData, &request); err != nil {
		return fmt.Errorf("%w: invalid JSON", domain.ErrMCPProtocol)
	}

	if request.JSONRPC != "2.0" {
		return fmt.Errorf("%w: invalid JSONRPC version", domain.ErrMCPProtocol)
	}

	if request.Method == "" {
		return fmt.Errorf("%w: method is required", domain.ErrMCPProtocol)
	}

	return nil
}

// GetSupportedMethods returns the list of supported MCP methods
func (h *MCPHandler) GetSupportedMethods() []string {
	return []string{
		"tools/list",
		"tools/call",
	}
}

// CreateNotification creates an MCP notification message
func (h *MCPHandler) CreateNotification(method string, params interface{}) ([]byte, error) {
	notification := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  method,
		"params":  params,
	}

	return json.Marshal(notification)
}

// LogLevel represents MCP log levels
type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

// LogParams represents parameters for logging notifications
type LogParams struct {
	Level   LogLevel `json:"level"`
	Message string   `json:"message"`
	Data    any      `json:"data,omitempty"`
}

// SendLogNotification sends a log notification via MCP
func (h *MCPHandler) SendLogNotification(level LogLevel, message string, data any) ([]byte, error) {
	params := LogParams{
		Level:   level,
		Message: message,
		Data:    data,
	}

	return h.CreateNotification("notifications/log", params)
}

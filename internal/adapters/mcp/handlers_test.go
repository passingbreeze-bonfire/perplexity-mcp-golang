package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/yourusername/perplexity-mcp-golang/internal/core/domain"
)

func TestNewMCPHandler(t *testing.T) {
	logger := &mockLogger{}
	searchUC := &mockSearchUseCase{}

	server := NewServer(logger, searchUC)
	handler := NewMCPHandler(server, logger)

	if handler == nil {
		t.Fatal("Expected handler to be created, got nil")
	}

	if handler.server != server {
		t.Error("Expected server to be set correctly")
	}

	if handler.logger != logger {
		t.Error("Expected logger to be set correctly")
	}
}

func TestMCPHandler_HandleRequest_ToolsList(t *testing.T) {
	logger := &mockLogger{}
	searchUC := &mockSearchUseCase{}

	server := NewServer(logger, searchUC)
	handler := NewMCPHandler(server, logger)
	ctx := context.Background()

	t.Run("successful tools/list", func(t *testing.T) {
		request := MCPRequest{
			JSONRPC: "2.0",
			ID:      "test-1",
			Method:  "tools/list",
		}

		requestData, err := json.Marshal(request)
		if err != nil {
			t.Fatalf("Failed to marshal request: %v", err)
		}

		responseData, err := handler.HandleRequest(ctx, requestData)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		var response MCPResponse
		err = json.Unmarshal(responseData, &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response.JSONRPC != "2.0" {
			t.Errorf("Expected JSONRPC 2.0, got %s", response.JSONRPC)
		}

		if response.ID != "test-1" {
			t.Errorf("Expected ID test-1, got %v", response.ID)
		}

		if response.Error != nil {
			t.Errorf("Expected no error, got %v", response.Error)
		}

		// Check that result contains tools
		result, ok := response.Result.(map[string]interface{})
		if !ok {
			t.Fatal("Expected result to be an object")
		}

		tools, ok := result["tools"].([]interface{})
		if !ok {
			t.Fatal("Expected tools to be an array")
		}

		// Should have 1 default tool
		if len(tools) != 1 {
			t.Errorf("Expected 1 tool, got %d", len(tools))
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		requestData := []byte("{invalid json")

		responseData, err := handler.HandleRequest(ctx, requestData)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		var response MCPResponse
		err = json.Unmarshal(responseData, &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response.Error == nil {
			t.Error("Expected error response")
		}

		if response.Error.Code != MCPErrorParseError {
			t.Errorf("Expected parse error code %d, got %d", MCPErrorParseError, response.Error.Code)
		}
	})

	t.Run("invalid JSONRPC version", func(t *testing.T) {
		request := MCPRequest{
			JSONRPC: "1.0",
			ID:      "test-2",
			Method:  "tools/list",
		}

		requestData, err := json.Marshal(request)
		if err != nil {
			t.Fatalf("Failed to marshal request: %v", err)
		}

		responseData, err := handler.HandleRequest(ctx, requestData)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		var response MCPResponse
		err = json.Unmarshal(responseData, &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response.Error == nil {
			t.Error("Expected error response")
		}

		if response.Error.Code != MCPErrorInvalidRequest {
			t.Errorf("Expected invalid request error code %d, got %d", MCPErrorInvalidRequest, response.Error.Code)
		}
	})

	t.Run("method not found", func(t *testing.T) {
		request := MCPRequest{
			JSONRPC: "2.0",
			ID:      "test-3",
			Method:  "unknown/method",
		}

		requestData, err := json.Marshal(request)
		if err != nil {
			t.Fatalf("Failed to marshal request: %v", err)
		}

		responseData, err := handler.HandleRequest(ctx, requestData)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		var response MCPResponse
		err = json.Unmarshal(responseData, &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response.Error == nil {
			t.Error("Expected error response")
		}

		if response.Error.Code != MCPErrorMethodNotFound {
			t.Errorf("Expected method not found error code %d, got %d", MCPErrorMethodNotFound, response.Error.Code)
		}
	})
}

func TestMCPHandler_HandleRequest_ToolsCall(t *testing.T) {
	logger := &mockLogger{}
	searchUC := &mockSearchUseCase{}

	server := NewServer(logger, searchUC)
	handler := NewMCPHandler(server, logger)
	ctx := context.Background()

	t.Run("successful perplexity_search call", func(t *testing.T) {
		params := ToolCallParams{
			Name: "perplexity_search",
			Arguments: map[string]any{
				"query": "test query",
			},
		}

		request := MCPRequest{
			JSONRPC: "2.0",
			ID:      "test-4",
			Method:  "tools/call",
			Params:  params,
		}

		requestData, err := json.Marshal(request)
		if err != nil {
			t.Fatalf("Failed to marshal request: %v", err)
		}

		responseData, err := handler.HandleRequest(ctx, requestData)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		var response MCPResponse
		err = json.Unmarshal(responseData, &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response.Error != nil {
			t.Errorf("Expected no error, got %v", response.Error)
		}

		// Check that result contains tool call result
		result, ok := response.Result.(map[string]interface{})
		if !ok {
			t.Fatal("Expected result to be an object")
		}

		content, ok := result["content"].(string)
		if !ok {
			t.Fatal("Expected content to be a string")
		}

		if content == "" {
			t.Error("Expected content to be non-empty")
		}

		isError, ok := result["isError"].(bool)
		if !ok {
			t.Fatal("Expected isError to be a boolean")
		}

		if isError {
			t.Error("Expected isError to be false")
		}
	})

	t.Run("tool not found", func(t *testing.T) {
		params := ToolCallParams{
			Name: "nonexistent_tool",
			Arguments: map[string]any{
				"query": "test query",
			},
		}

		request := MCPRequest{
			JSONRPC: "2.0",
			ID:      "test-5",
			Method:  "tools/call",
			Params:  params,
		}

		requestData, err := json.Marshal(request)
		if err != nil {
			t.Fatalf("Failed to marshal request: %v", err)
		}

		responseData, err := handler.HandleRequest(ctx, requestData)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		var response MCPResponse
		err = json.Unmarshal(responseData, &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response.Error == nil {
			t.Error("Expected error response")
		}

		if response.Error.Code != MCPErrorServerError {
			t.Errorf("Expected server error code %d, got %d", MCPErrorServerError, response.Error.Code)
		}
	})

	t.Run("empty tool name", func(t *testing.T) {
		params := ToolCallParams{
			Name: "",
			Arguments: map[string]any{
				"query": "test query",
			},
		}

		request := MCPRequest{
			JSONRPC: "2.0",
			ID:      "test-6",
			Method:  "tools/call",
			Params:  params,
		}

		requestData, err := json.Marshal(request)
		if err != nil {
			t.Fatalf("Failed to marshal request: %v", err)
		}

		responseData, err := handler.HandleRequest(ctx, requestData)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		var response MCPResponse
		err = json.Unmarshal(responseData, &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response.Error == nil {
			t.Error("Expected error response")
		}

		if response.Error.Code != MCPErrorServerError {
			t.Errorf("Expected server error code %d, got %d", MCPErrorServerError, response.Error.Code)
		}
	})

	t.Run("tool execution error", func(t *testing.T) {
		// Register a failing tool
		failingTool := &mockTool{
			name: "failing_tool",
			executeFunc: func(ctx context.Context, args map[string]any) (*domain.ToolResult, error) {
				return nil, errors.New("tool execution failed")
			},
		}

		err := server.RegisterTool("failing_tool", failingTool)
		if err != nil {
			t.Fatalf("Failed to register failing tool: %v", err)
		}

		params := ToolCallParams{
			Name:      "failing_tool",
			Arguments: map[string]any{},
		}

		request := MCPRequest{
			JSONRPC: "2.0",
			ID:      "test-7",
			Method:  "tools/call",
			Params:  params,
		}

		requestData, err := json.Marshal(request)
		if err != nil {
			t.Fatalf("Failed to marshal request: %v", err)
		}

		responseData, err := handler.HandleRequest(ctx, requestData)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		var response MCPResponse
		err = json.Unmarshal(responseData, &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response.Error == nil {
			t.Error("Expected error response")
		}

		if response.Error.Code != MCPErrorServerError {
			t.Errorf("Expected server error code %d, got %d", MCPErrorServerError, response.Error.Code)
		}
	})
}

func TestMCPHandler_ValidateRequest(t *testing.T) {
	logger := &mockLogger{}
	searchUC := &mockSearchUseCase{}

	server := NewServer(logger, searchUC)
	handler := NewMCPHandler(server, logger)

	t.Run("valid request", func(t *testing.T) {
		request := MCPRequest{
			JSONRPC: "2.0",
			Method:  "tools/list",
		}

		requestData, err := json.Marshal(request)
		if err != nil {
			t.Fatalf("Failed to marshal request: %v", err)
		}

		err = handler.ValidateRequest(requestData)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		requestData := []byte("{invalid json")

		err := handler.ValidateRequest(requestData)
		if err == nil {
			t.Error("Expected error for invalid JSON")
		}

		if !strings.Contains(err.Error(), "invalid JSON") {
			t.Errorf("Expected error to contain 'invalid JSON', got %s", err.Error())
		}
	})

	t.Run("invalid JSONRPC version", func(t *testing.T) {
		request := MCPRequest{
			JSONRPC: "1.0",
			Method:  "tools/list",
		}

		requestData, err := json.Marshal(request)
		if err != nil {
			t.Fatalf("Failed to marshal request: %v", err)
		}

		err = handler.ValidateRequest(requestData)
		if err == nil {
			t.Error("Expected error for invalid JSONRPC version")
		}

		if !strings.Contains(err.Error(), "invalid JSONRPC version") {
			t.Errorf("Expected error to contain 'invalid JSONRPC version', got %s", err.Error())
		}
	})

	t.Run("missing method", func(t *testing.T) {
		request := MCPRequest{
			JSONRPC: "2.0",
		}

		requestData, err := json.Marshal(request)
		if err != nil {
			t.Fatalf("Failed to marshal request: %v", err)
		}

		err = handler.ValidateRequest(requestData)
		if err == nil {
			t.Error("Expected error for missing method")
		}

		if !strings.Contains(err.Error(), "method is required") {
			t.Errorf("Expected error to contain 'method is required', got %s", err.Error())
		}
	})
}

func TestMCPHandler_GetSupportedMethods(t *testing.T) {
	logger := &mockLogger{}
	searchUC := &mockSearchUseCase{}

	server := NewServer(logger, searchUC)
	handler := NewMCPHandler(server, logger)

	methods := handler.GetSupportedMethods()

	expectedMethods := []string{"tools/list", "tools/call"}
	if len(methods) != len(expectedMethods) {
		t.Errorf("Expected %d methods, got %d", len(expectedMethods), len(methods))
	}

	for _, expectedMethod := range expectedMethods {
		found := false
		for _, method := range methods {
			if method == expectedMethod {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected method %s to be supported", expectedMethod)
		}
	}
}

func TestMCPHandler_CreateNotification(t *testing.T) {
	logger := &mockLogger{}
	searchUC := &mockSearchUseCase{}

	server := NewServer(logger, searchUC)
	handler := NewMCPHandler(server, logger)

	params := map[string]any{
		"test": "value",
	}

	notificationData, err := handler.CreateNotification("test/method", params)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	var notification map[string]interface{}
	err = json.Unmarshal(notificationData, &notification)
	if err != nil {
		t.Fatalf("Failed to unmarshal notification: %v", err)
	}

	if notification["jsonrpc"] != "2.0" {
		t.Errorf("Expected JSONRPC 2.0, got %v", notification["jsonrpc"])
	}

	if notification["method"] != "test/method" {
		t.Errorf("Expected method test/method, got %v", notification["method"])
	}

	notificationParams, ok := notification["params"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected params to be an object")
	}

	if notificationParams["test"] != "value" {
		t.Errorf("Expected params.test to be 'value', got %v", notificationParams["test"])
	}
}

func TestMCPHandler_SendLogNotification(t *testing.T) {
	logger := &mockLogger{}
	searchUC := &mockSearchUseCase{}

	server := NewServer(logger, searchUC)
	handler := NewMCPHandler(server, logger)

	data := map[string]any{
		"context": "test",
	}

	notificationData, err := handler.SendLogNotification(LogLevelInfo, "test message", data)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	var notification map[string]interface{}
	err = json.Unmarshal(notificationData, &notification)
	if err != nil {
		t.Fatalf("Failed to unmarshal notification: %v", err)
	}

	if notification["method"] != "notifications/log" {
		t.Errorf("Expected method notifications/log, got %v", notification["method"])
	}

	params, ok := notification["params"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected params to be an object")
	}

	if params["level"] != string(LogLevelInfo) {
		t.Errorf("Expected level info, got %v", params["level"])
	}

	if params["message"] != "test message" {
		t.Errorf("Expected message 'test message', got %v", params["message"])
	}
}

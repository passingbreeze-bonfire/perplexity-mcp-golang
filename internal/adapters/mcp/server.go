package mcp

import (
	"context"
	"fmt"
	"sync"

	"github.com/yourusername/perplexity-mcp-golang/internal/adapters/mcp/tools"
	"github.com/yourusername/perplexity-mcp-golang/internal/core/domain"
)

// Server implements the MCPServer interface and manages MCP tools
type Server struct {
	tools     map[string]domain.Tool
	toolsLock sync.RWMutex
	logger    domain.Logger

	// Use case for search functionality
	searchUseCase *SearchUseCaseInterface
}

// SearchUseCaseInterface defines the contract for search use case
type SearchUseCaseInterface interface {
	Execute(ctx context.Context, request domain.SearchRequest) (*domain.SearchResult, error)
	ValidateRequest(request domain.SearchRequest) error
}

// NewServer creates a new MCP server instance
func NewServer(
	logger domain.Logger,
	searchUseCase SearchUseCaseInterface,
) *Server {
	server := &Server{
		tools:         make(map[string]domain.Tool),
		logger:        logger,
		searchUseCase: &searchUseCase,
	}

	// Register search tool
	if err := server.registerSearchTool(); err != nil {
		logger.Error("Failed to register search tool", "error", err)
	}

	return server
}

// RegisterTool registers a new tool with the MCP server
func (s *Server) RegisterTool(name string, tool domain.Tool) error {
	s.toolsLock.Lock()
	defer s.toolsLock.Unlock()

	if name == "" {
		return fmt.Errorf("%w: tool name cannot be empty", domain.ErrInvalidRequest)
	}

	if tool == nil {
		return fmt.Errorf("%w: tool cannot be nil", domain.ErrInvalidRequest)
	}

	// Validate that tool name matches the tool's own name
	if tool.Name() != name {
		return fmt.Errorf("%w: tool name mismatch: registered as '%s' but tool reports name '%s'",
			domain.ErrInvalidRequest, name, tool.Name())
	}

	// Check if tool already exists
	if _, exists := s.tools[name]; exists {
		s.logger.Warn("Tool already registered, replacing", "tool_name", name)
	}

	s.tools[name] = tool
	s.logger.Info("Tool registered successfully", "tool_name", name, "description", tool.Description())

	return nil
}

// ExecuteTool executes a tool by name with the provided arguments
func (s *Server) ExecuteTool(ctx context.Context, name string, args map[string]any) (*domain.ToolResult, error) {
	// Add safety timeout if no deadline is set
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, DefaultTimeout)
		defer cancel()
	}

	s.toolsLock.RLock()
	tool, exists := s.tools[name]
	s.toolsLock.RUnlock()

	if !exists {
		s.logger.Error("Tool not found", "tool_name", name)
		return &domain.ToolResult{
			Content: fmt.Sprintf("Tool '%s' not found", name),
			IsError: true,
		}, fmt.Errorf("%w: tool '%s' not found", domain.ErrToolNotFound, name)
	}

	s.logger.Info("Executing tool", "tool_name", name, "args_count", len(args))

	// Execute the tool
	result, err := tool.Execute(ctx, args)
	if err != nil {
		s.logger.Error("Tool execution failed",
			"tool_name", name,
			"error", err.Error(),
		)

		// Return structured error result
		return &domain.ToolResult{
			Content: fmt.Sprintf("Tool execution failed: %s", err.Error()),
			IsError: true,
			Metadata: map[string]any{
				"tool_name":  name,
				"error_type": "execution_error",
			},
		}, fmt.Errorf("%w: %s", domain.ErrToolExecution, err.Error())
	}

	s.logger.Info("Tool execution completed successfully",
		"tool_name", name,
		"result_error", result.IsError,
		"result_content_length", len(result.Content),
		"metadata_keys", len(result.Metadata),
		"citations_count", len(result.Citations),
	)

	return result, nil
}

// ListTools returns information about all registered tools
func (s *Server) ListTools(ctx context.Context) ([]domain.ToolInfo, error) {
	s.toolsLock.RLock()
	defer s.toolsLock.RUnlock()

	tools := make([]domain.ToolInfo, 0, len(s.tools))

	for _, tool := range s.tools {
		toolInfo := domain.ToolInfo{
			Name:        tool.Name(),
			Description: tool.Description(),
			InputSchema: tool.InputSchema(),
		}
		tools = append(tools, toolInfo)
	}

	s.logger.Debug("Listed tools", "count", len(tools))

	return tools, nil
}

// Start initializes and starts the MCP server
func (s *Server) Start(ctx context.Context) error {
	s.logger.Info("Starting MCP server")

	// Validate that search tool is registered
	requiredTools := []string{"perplexity_search"}

	s.toolsLock.RLock()
	for _, toolName := range requiredTools {
		if _, exists := s.tools[toolName]; !exists {
			s.toolsLock.RUnlock()
			return fmt.Errorf("%w: required tool '%s' not registered", domain.ErrConfigurationError, toolName)
		}
	}
	toolCount := len(s.tools)
	s.toolsLock.RUnlock()

	s.logger.Info("MCP server started successfully",
		"tools_count", toolCount,
		"required_tools", requiredTools,
	)

	return nil
}

// registerSearchTool registers the search tool
func (s *Server) registerSearchTool() error {
	// Register search tool
	searchTool := tools.NewSearchTool(*s.searchUseCase, s.logger)
	if err := s.RegisterTool(searchTool.Name(), searchTool); err != nil {
		return fmt.Errorf("failed to register search tool: %w", err)
	}

	return nil
}

// GetToolCount returns the number of registered tools
func (s *Server) GetToolCount() int {
	s.toolsLock.RLock()
	defer s.toolsLock.RUnlock()
	return len(s.tools)
}

// HasTool checks if a tool with the given name is registered
func (s *Server) HasTool(name string) bool {
	s.toolsLock.RLock()
	defer s.toolsLock.RUnlock()
	_, exists := s.tools[name]
	return exists
}

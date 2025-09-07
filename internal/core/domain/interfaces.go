package domain

import "context"

type PerplexityClient interface {
	Search(ctx context.Context, request SearchRequest) (*SearchResult, error)
}

type MCPServer interface {
	RegisterTool(name string, tool Tool) error
	ExecuteTool(ctx context.Context, name string, args map[string]any) (*ToolResult, error)
	ListTools(ctx context.Context) ([]ToolInfo, error)
	Start(ctx context.Context) error
}

type Tool interface {
	Name() string
	Description() string
	InputSchema() map[string]any
	Execute(ctx context.Context, args map[string]any) (*ToolResult, error)
}

type ConfigProvider interface {
	GetPerplexityAPIKey() string
	GetDefaultModel() string
	GetRequestTimeout() int
	GetLogLevel() string
}

type Logger interface {
	Info(msg string, fields ...any)
	Error(msg string, fields ...any)
	Debug(msg string, fields ...any)
	Warn(msg string, fields ...any)
}

package mcp

import "time"

// DefaultTimeout is the default timeout for MCP operations
const DefaultTimeout = 30 * time.Second

// Tool names for the Perplexity MCP tools
const (
	ToolNameSearch   = "perplexity_search"
	ToolNameChat     = "perplexity_chat"
	ToolNameResearch = "perplexity_research"
)

// JSON Schema property types
const (
	JSONTypeString = "string"
	JSONTypeNumber = "number"
	JSONTypeObject = "object"
	JSONTypeArray  = "array"
	JSONTypeBool   = "boolean"
)

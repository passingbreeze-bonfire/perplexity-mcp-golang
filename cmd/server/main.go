package main

import (
	"log"
	"os"

	"github.com/mark3labs/mcp-go/server"
	"github.com/passingbreeze-bonfire/perplexity-mcp-golang/internal"
)

func main() {
	// Direct all logging to stderr
	log.SetOutput(os.Stderr)
	logger := log.New(os.Stderr, "[MAIN] ", log.LstdFlags|log.Lshortfile)

	if err := run(logger); err != nil {
		logger.Printf("error: %v", err)
		os.Exit(1)
	}
}

func run(logger *log.Logger) error {
	logger.Println("Starting Perplexity MCP Server")

	// Load configuration
	config, err := internal.NewConfig()
	if err != nil {
		return err
	}

	if err := config.Validate(); err != nil {
		return err
	}

	logger.Printf("Configuration loaded - Model: %s, Timeout: %s",
		config.DefaultModel, config.RequestTimeout)

	// Create Perplexity client
	client, err := internal.NewPerplexityClient(config.PerplexityAPIKey)
	if err != nil {
		return err
	}

	// Create MCP server
	mcpServer := server.NewMCPServer("perplexity-mcp-server", "1.0.0")

	// Register the perplexity search tool
	searchTool := internal.CreatePerplexitySearchTool(client)
	searchHandler := internal.PerplexitySearchHandler(client)
	mcpServer.AddTool(searchTool, searchHandler)

	logger.Printf("MCP server configured with 1 tool: perplexity_search")
	logger.Println("Starting MCP server on stdio")

	// Serve on stdio - blocks until stdin is closed
	return server.ServeStdio(mcpServer)
}

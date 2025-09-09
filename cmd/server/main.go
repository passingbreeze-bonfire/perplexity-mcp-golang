package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mark3labs/mcp-go/server"
	"github.com/passingbreeze/perplexity-mcp-golang/internal"
)

const (
	ShutdownTimeout = 10 * time.Second
)

func main() {
	logger := log.New(os.Stdout, "[MAIN] ", log.LstdFlags|log.Lshortfile)

	logger.Println("Starting Perplexity MCP Server")

	// Load configuration
	config, err := internal.NewConfig()
	if err != nil {
		logger.Fatalf("Failed to load config: %v", err)
	}

	if err := config.Validate(); err != nil {
		logger.Fatalf("Config validation failed: %v", err)
	}

	logger.Printf("Configuration loaded - Model: %s, Timeout: %s, HTTP: %s:%s",
		config.DefaultModel, config.RequestTimeout, config.HTTPHost, config.HTTPPort)

	// Create Perplexity client
	client, err := internal.NewPerplexityClient(config.PerplexityAPIKey)
	if err != nil {
		logger.Fatalf("Failed to create Perplexity client: %v", err)
	}

	// Create MCP server
	mcpServer := server.NewMCPServer("perplexity-mcp-server", "1.0.0")

	// Register the perplexity search tool
	searchTool := internal.CreatePerplexitySearchTool(client)
	searchHandler := internal.PerplexitySearchHandler(client)
	mcpServer.AddTool(searchTool, searchHandler)

	// Create HTTP server
	httpServer := server.NewStreamableHTTPServer(mcpServer)

	logger.Printf("MCP server configured with 1 tool: perplexity_search")

	// Run the server
	if err := run(httpServer, config, logger); err != nil {
		logger.Fatalf("Server failed: %v", err)
	}

	logger.Println("Server shutdown complete")
}

func run(httpServer *server.StreamableHTTPServer, config *internal.Config, logger *log.Logger) error {
	// Create server address
	addr := net.JoinHostPort(config.HTTPHost, config.HTTPPort)

	// Channel for server errors
	serverErrChan := make(chan error, 1)

	// Start server in goroutine
	go func() {
		logger.Printf("Starting HTTP server on %s", addr)
		if err := httpServer.Start(addr); err != nil {
			serverErrChan <- fmt.Errorf("HTTP server failed to start: %w", err)
		}
	}()

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	logger.Printf("MCP HTTP server running on %s", addr)
	logger.Println("Press Ctrl+C to shutdown gracefully")

	// Wait for shutdown signal or server error
	select {
	case sig := <-sigChan:
		logger.Printf("Received shutdown signal: %s", sig.String())

	case err := <-serverErrChan:
		logger.Printf("Server error occurred: %v", err)
		return err
	}

	logger.Println("Initiating graceful shutdown")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), ShutdownTimeout)
	defer cancel()

	// Shutdown the HTTP server
	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Printf("Error during shutdown: %v", err)
		return fmt.Errorf("graceful shutdown failed: %w", err)
	}

	logger.Println("Graceful shutdown completed successfully")
	return nil
}

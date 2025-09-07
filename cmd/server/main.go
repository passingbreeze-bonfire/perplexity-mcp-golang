package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yourusername/perplexity-mcp-golang/internal/adapters/mcp"
	"github.com/yourusername/perplexity-mcp-golang/internal/adapters/perplexity"
	"github.com/yourusername/perplexity-mcp-golang/internal/core/usecases"
	"github.com/yourusername/perplexity-mcp-golang/internal/infrastructure/config"
	"github.com/yourusername/perplexity-mcp-golang/internal/infrastructure/logger"
)

const (
	// ShutdownTimeout is the maximum time allowed for graceful shutdown
	ShutdownTimeout = 10 * time.Second
	// StartupTimeout is the maximum time allowed for server startup
	StartupTimeout = 5 * time.Second
)

func main() {
	// Create root context for the application
	ctx := context.Background()

	// Load configuration
	cfg := config.NewConfig()

	// Initialize logger
	log := logger.NewLogger(cfg.GetLogLevel(), os.Stdout)
	log.Info("Starting Perplexity MCP Server")

	// Validate configuration on startup
	if err := validateConfiguration(cfg, log); err != nil {
		log.Error("Configuration validation failed", "error", err)
		os.Exit(1)
	}

	// Wire up dependencies and start server
	if err := run(ctx, cfg, log); err != nil {
		log.Error("Server failed", "error", err)
		os.Exit(1)
	}

	log.Info("Server shutdown complete")
}

// run orchestrates the application startup and handles graceful shutdown
func run(ctx context.Context, cfg *config.Config, log *logger.Logger) error {
	// Create a context with timeout for startup operations
	startupCtx, startupCancel := context.WithTimeout(ctx, StartupTimeout)
	defer startupCancel()

	// Wire up all dependencies
	dependencies, err := wireDependencies(startupCtx, cfg, log)
	if err != nil {
		return fmt.Errorf("failed to wire dependencies: %w", err)
	}

	log.Info("Dependencies wired successfully")

	// Set up graceful shutdown handling
	shutdownCtx, shutdownCancel := context.WithCancel(ctx)
	defer shutdownCancel()

	// Channel to listen for interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start the MCP server in a goroutine
	serverErrChan := make(chan error, 1)
	go func() {
		log.Info("Starting MCP server")
		if err := dependencies.MCPServer.Start(shutdownCtx); err != nil {
			serverErrChan <- fmt.Errorf("MCP server failed to start: %w", err)
			return
		}

		log.Info("MCP server is running and ready to accept requests")
		log.Info("Press Ctrl+C to shutdown gracefully")

		// Keep server running until context is cancelled
		<-shutdownCtx.Done()
		log.Info("MCP server context cancelled, shutting down")
	}()

	// Wait for either a shutdown signal or server error
	select {
	case sig := <-sigChan:
		log.Info("Received shutdown signal", "signal", sig.String())

	case err := <-serverErrChan:
		log.Error("Server error occurred", "error", err)
		return err
	}

	// Initiate graceful shutdown
	log.Info("Initiating graceful shutdown")
	shutdownCancel()

	// Wait for shutdown to complete or timeout
	shutdownTimeout := time.NewTimer(ShutdownTimeout)
	defer shutdownTimeout.Stop()

	select {
	case <-shutdownTimeout.C:
		log.Warn("Graceful shutdown timed out, forcing exit")
		return fmt.Errorf("shutdown timeout exceeded")

	case err := <-serverErrChan:
		if err != nil {
			log.Error("Error during shutdown", "error", err)
			return err
		}
		log.Info("Graceful shutdown completed successfully")
	}

	return nil
}

// Dependencies holds all wired up dependencies
type Dependencies struct {
	Config            *config.Config
	Logger            *logger.Logger
	PerplexityClient  *perplexity.Client
	SearchUseCase     *usecases.SearchUseCase
	MCPServer         *mcp.Server
}

// wireDependencies creates and wires up all application dependencies
// following the dependency injection pattern for clean architecture
func wireDependencies(ctx context.Context, cfg *config.Config, log *logger.Logger) (*Dependencies, error) {
	deps := &Dependencies{
		Config: cfg,
		Logger: log,
	}

	log.Debug("Wiring dependencies")

	// Create Perplexity API client
	perplexityClient, err := perplexity.NewClient(
		cfg.GetPerplexityAPIKey(),
		log,
		perplexity.WithTimeout(time.Duration(cfg.GetRequestTimeout())*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Perplexity client: %w", err)
	}
	deps.PerplexityClient = perplexityClient

	log.Debug("Perplexity client created successfully")

	// Create search use case with dependency injection
	deps.SearchUseCase = usecases.NewSearchUseCase(perplexityClient, cfg, log)

	log.Debug("Use cases created successfully")

	// Create MCP server with search use case only
	deps.MCPServer = mcp.NewServer(
		log,
		deps.SearchUseCase,
	)

	log.Debug("MCP server created successfully")

	// Validate that the MCP server has the search tool
	requiredTools := []string{"perplexity_search"}
	for _, toolName := range requiredTools {
		if !deps.MCPServer.HasTool(toolName) {
			return nil, fmt.Errorf("required tool '%s' not found in MCP server", toolName)
		}
	}

	toolCount := deps.MCPServer.GetToolCount()
	log.Info("Dependency wiring completed successfully",
		"tool_count", toolCount,
		"required_tools", requiredTools,
	)

	return deps, nil
}

// validateConfiguration validates the configuration and logs any issues
func validateConfiguration(cfg *config.Config, log *logger.Logger) error {
	log.Info("Validating configuration")

	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	// Log configuration summary (without sensitive information)
	log.Info("Configuration validated successfully",
		"default_model", cfg.GetDefaultModel(),
		"request_timeout", cfg.GetRequestTimeout(),
		"log_level", cfg.GetLogLevel(),
		"api_key_configured", cfg.GetPerplexityAPIKey() != "",
	)

	return nil
}
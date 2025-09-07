package benchmark

import (
	"context"
	"testing"

	"github.com/yourusername/perplexity-mcp-golang/internal/core/domain"
	"github.com/yourusername/perplexity-mcp-golang/test/integration"
)

// BenchmarkSearchToolExecution benchmarks search tool performance (single-threaded baseline)
func BenchmarkSearchToolExecution(b *testing.B) {
	helper := integration.NewBenchmarkHelper()
	env := helper.GetEnvironment()

	ctx := context.Background()
	args := map[string]any{
		"query": "What is artificial intelligence?",
		"model": "llama-3.1-sonar-small-128k-online",
	}

	// Reset counters and start timing
	b.ResetTimer()

	// Single-threaded baseline - process items sequentially
	for i := 0; i < b.N; i++ {
		result, err := env.Server.ExecuteTool(ctx, "perplexity_search", args)
		if err != nil {
			b.Fatalf("Search tool execution failed: %v", err)
		}
		if result.IsError {
			b.Fatalf("Search tool returned error: %s", result.Content)
		}
	}
}

// BenchmarkChatToolExecution benchmarks chat tool performance (single-threaded baseline)
func BenchmarkChatToolExecution(b *testing.B) {
	helper := integration.NewBenchmarkHelper()
	env := helper.GetEnvironment()

	ctx := context.Background()
	args := map[string]any{
		"messages": []map[string]string{
			{"role": "user", "content": "Explain machine learning concepts"},
		},
		"model": "llama-3.1-sonar-small-128k-chat",
	}

	b.ResetTimer()

	// Single-threaded baseline
	for i := 0; i < b.N; i++ {
		result, err := env.Server.ExecuteTool(ctx, "perplexity_chat", args)
		if err != nil {
			b.Fatalf("Chat tool execution failed: %v", err)
		}
		if result.IsError {
			b.Fatalf("Chat tool returned error: %s", result.Content)
		}
	}
}

// BenchmarkResearchToolExecution benchmarks research tool performance (single-threaded baseline)
func BenchmarkResearchToolExecution(b *testing.B) {
	helper := integration.NewBenchmarkHelper()
	env := helper.GetEnvironment()

	ctx := context.Background()
	args := map[string]any{
		"topic":            "Climate change impacts on ecosystems",
		"reasoning_effort": "thorough",
		"model":            "llama-3.1-sonar-large-128k-online",
	}

	b.ResetTimer()

	// Single-threaded baseline
	for i := 0; i < b.N; i++ {
		result, err := env.Server.ExecuteTool(ctx, "perplexity_research", args)
		if err != nil {
			b.Fatalf("Research tool execution failed: %v", err)
		}
		if result.IsError {
			b.Fatalf("Research tool returned error: %s", result.Content)
		}
	}
}

// BenchmarkToolListOperation benchmarks tool listing performance
func BenchmarkToolListOperation(b *testing.B) {
	helper := integration.NewBenchmarkHelper()
	env := helper.GetEnvironment()

	ctx := context.Background()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		tools, err := env.Server.ListTools(ctx)
		if err != nil {
			b.Fatalf("ListTools failed: %v", err)
		}
		if len(tools) == 0 {
			b.Fatal("No tools returned")
		}
	}
}

// BenchmarkSequentialToolExecution benchmarks executing multiple tools in sequence
// This represents a single-threaded workflow processing multiple operations
func BenchmarkSequentialToolExecution(b *testing.B) {
	helper := integration.NewBenchmarkHelper()
	env := helper.GetEnvironment()

	ctx := context.Background()

	// Define the sequential operations to benchmark
	operations := []struct {
		toolName string
		args     map[string]any
	}{
		{
			toolName: "perplexity_search",
			args: map[string]any{
				"query": "AI research trends",
			},
		},
		{
			toolName: "perplexity_chat",
			args: map[string]any{
				"messages": []map[string]string{
					{"role": "user", "content": "Summarize AI trends"},
				},
			},
		},
		{
			toolName: "perplexity_research",
			args: map[string]any{
				"topic": "Future of artificial intelligence",
			},
		},
	}

	b.ResetTimer()

	// Single-threaded baseline: process operations sequentially
	for i := 0; i < b.N; i++ {
		for _, op := range operations {
			result, err := env.Server.ExecuteTool(ctx, op.toolName, op.args)
			if err != nil {
				b.Fatalf("Tool %s execution failed: %v", op.toolName, err)
			}
			if result.IsError {
				b.Fatalf("Tool %s returned error: %s", op.toolName, result.Content)
			}
		}
	}
}

// BenchmarkInputValidation benchmarks the performance of input validation
func BenchmarkInputValidation(b *testing.B) {
	helper := integration.NewBenchmarkHelper()
	env := helper.GetEnvironment()

	// Test with a complex search request that requires validation
	request := domain.SearchRequest{
		Query:      "Complex query with multiple parameters for validation testing",
		Model:      "llama-3.1-sonar-small-128k-online",
		SearchMode: "thorough",
		MaxTokens:  2000,
		Options: map[string]string{
			"temperature":   "0.7",
			"top_p":         "0.9",
			"frequency":     "0.1",
			"presence":      "0.1",
			"focus":         "recent",
			"recency_bias":  "auto",
			"return_images": "false",
			"return_related": "true",
		},
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := env.SearchUseCase.ValidateRequest(request)
		if err != nil {
			b.Fatalf("Validation failed: %v", err)
		}
	}
}

// BenchmarkMockClientPerformance benchmarks the mock client's performance
// to establish overhead baseline for actual API performance measurement
func BenchmarkMockClientPerformance(b *testing.B) {
	helper := integration.NewBenchmarkHelper()
	env := helper.GetEnvironment()

	ctx := context.Background()
	request := domain.SearchRequest{
		Query: "Performance test query",
		Model: "llama-3.1-sonar-small-128k-online",
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		result, err := env.MockClient.Search(ctx, request)
		if err != nil {
			b.Fatalf("Mock client search failed: %v", err)
		}
		if result == nil {
			b.Fatal("Mock client returned nil result")
		}
	}
}

// BenchmarkDifferentQuerySizes benchmarks performance with varying query sizes
func BenchmarkDifferentQuerySizes(b *testing.B) {
	helper := integration.NewBenchmarkHelper()
	env := helper.GetEnvironment()

	ctx := context.Background()

	// Test with different query sizes to understand performance characteristics
	querySizes := []struct {
		name string
		size int
	}{
		{"Small", 50},
		{"Medium", 500},
		{"Large", 2000},
		{"VeryLarge", 8000},
	}

	for _, qs := range querySizes {
		// Generate query of specified size
		query := make([]byte, qs.size)
		for i := range query {
			query[i] = byte('a' + (i % 26)) // Cycle through alphabet
		}

		b.Run(qs.name, func(b *testing.B) {
			args := map[string]any{
				"query": string(query),
			}

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				result, err := env.Server.ExecuteTool(ctx, "perplexity_search", args)
				if err != nil {
					b.Fatalf("Search with %s query failed: %v", qs.name, err)
				}
				if result.IsError {
					b.Fatalf("Search with %s query returned error: %s", qs.name, result.Content)
				}
			}
		})
	}
}

// BenchmarkMemoryAllocation benchmarks memory allocation patterns
// Important for understanding GC pressure in high-throughput scenarios
func BenchmarkMemoryAllocation(b *testing.B) {
	helper := integration.NewBenchmarkHelper()
	env := helper.GetEnvironment()

	ctx := context.Background()
	args := map[string]any{
		"query": "Memory allocation test query",
	}

	// Report memory allocations
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		result, err := env.Server.ExecuteTool(ctx, "perplexity_search", args)
		if err != nil {
			b.Fatalf("Tool execution failed: %v", err)
		}
		// Don't hold reference to result to allow GC
		_ = result
	}
}

// BenchmarkErrorHandling benchmarks error handling performance
func BenchmarkErrorHandling(b *testing.B) {
	helper := integration.NewBenchmarkHelper()
	env := helper.GetEnvironment()

	ctx := context.Background()
	errorQuery := "benchmark_error_test"

	// Configure mock to return errors
	env.MockClient.SetError(errorQuery, domain.ErrAPIError)

	args := map[string]any{
		"query": errorQuery,
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		result, err := env.Server.ExecuteTool(ctx, "perplexity_search", args)
		// We expect errors in this benchmark
		if err == nil || result == nil || !result.IsError {
			b.Fatal("Expected error but none occurred")
		}
	}
}

// BenchmarkConcurrencyComparison establishes baseline for future concurrency improvements
// This benchmark provides the single-thread baseline for comparison when
// concurrency is added later (following single-thread-first policy)
func BenchmarkConcurrencyComparison(b *testing.B) {
	helper := integration.NewBenchmarkHelper()
	env := helper.GetEnvironment()

	ctx := context.Background()

	// Create a set of different queries to simulate real-world variety
	queries := []string{
		"What is artificial intelligence?",
		"Explain quantum computing",
		"Climate change effects",
		"Machine learning algorithms",
		"Renewable energy sources",
		"Space exploration missions",
		"Genetic engineering advances",
		"Cybersecurity threats",
		"Blockchain technology",
		"Neural network architectures",
	}

	b.Run("Sequential", func(b *testing.B) {
		b.ResetTimer()

		// Single-threaded baseline: process all queries sequentially
		for i := 0; i < b.N; i++ {
			for _, query := range queries {
				args := map[string]any{"query": query}
				result, err := env.Server.ExecuteTool(ctx, "perplexity_search", args)
				if err != nil {
					b.Fatalf("Sequential execution failed: %v", err)
				}
				if result.IsError {
					b.Fatalf("Sequential execution returned error: %s", result.Content)
				}
			}
		}
	})

	// Note: Concurrent version would be added here ONLY after profiling
	// shows the sequential version is a bottleneck, following single-thread-first policy
	// 
	// b.Run("Concurrent", func(b *testing.B) {
	//     // Concurrent implementation would go here
	//     // Only after sequential baseline shows performance bottleneck
	// })
}

// BenchmarkResourceCleanup benchmarks resource cleanup performance
func BenchmarkResourceCleanup(b *testing.B) {
	ctx := context.Background()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Create fresh environment for each iteration to benchmark cleanup
		helper := integration.NewBenchmarkHelper()
		env := helper.GetEnvironment()

		// Execute a tool to create some state
		args := map[string]any{"query": "cleanup test"}
		result, err := env.Server.ExecuteTool(ctx, "perplexity_search", args)
		if err != nil {
			b.Fatalf("Tool execution failed: %v", err)
		}
		_ = result

		// Environment cleanup happens when helper goes out of scope
		// This benchmarks the overhead of cleanup operations
	}
}
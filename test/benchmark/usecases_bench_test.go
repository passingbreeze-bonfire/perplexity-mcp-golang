package benchmark

import (
	"context"
	"testing"

	"github.com/yourusername/perplexity-mcp-golang/internal/core/domain"
	"github.com/yourusername/perplexity-mcp-golang/test/integration"
)

// BenchmarkSearchUseCase benchmarks search use case performance (single-threaded baseline)
func BenchmarkSearchUseCase(b *testing.B) {
	helper := integration.NewBenchmarkHelper()
	env := helper.GetEnvironment()

	ctx := context.Background()
	request := domain.SearchRequest{
		Query:      "What are the latest developments in artificial intelligence?",
		Model:      "llama-3.1-sonar-small-128k-online",
		SearchMode: "quick",
		MaxTokens:  1000,
		Options: map[string]string{
			"temperature": "0.7",
			"focus":       "recent",
		},
	}

	b.ResetTimer()

	// Single-threaded baseline - process requests sequentially
	for i := 0; i < b.N; i++ {
		result, err := env.SearchUseCase.Execute(ctx, request)
		if err != nil {
			b.Fatalf("Search use case execution failed: %v", err)
		}
		if result == nil {
			b.Fatal("Search use case returned nil result")
		}
	}
}

// BenchmarkChatUseCase benchmarks chat use case performance (single-threaded baseline)
func BenchmarkChatUseCase(b *testing.B) {
	helper := integration.NewBenchmarkHelper()
	env := helper.GetEnvironment()

	ctx := context.Background()
	request := domain.ChatRequest{
		Messages: []domain.Message{
			{Role: "system", Content: "You are a helpful AI assistant."},
			{Role: "user", Content: "Explain the concept of machine learning in simple terms."},
		},
		Model:       "llama-3.1-sonar-small-128k-chat",
		MaxTokens:   1500,
		Temperature: 0.8,
		Options: map[string]string{
			"context": "educational",
		},
	}

	b.ResetTimer()

	// Single-threaded baseline
	for i := 0; i < b.N; i++ {
		result, err := env.ChatUseCase.Execute(ctx, request)
		if err != nil {
			b.Fatalf("Chat use case execution failed: %v", err)
		}
		if result == nil {
			b.Fatal("Chat use case returned nil result")
		}
	}
}

// BenchmarkResearchUseCase benchmarks research use case performance (single-threaded baseline)
func BenchmarkResearchUseCase(b *testing.B) {
	helper := integration.NewBenchmarkHelper()
	env := helper.GetEnvironment()

	ctx := context.Background()
	request := domain.ResearchRequest{
		Topic:           "Impact of renewable energy on global economics",
		Model:           "llama-3.1-sonar-large-128k-online",
		ReasoningEffort: "thorough",
		MaxTokens:       2000,
		Options: map[string]string{
			"scope":    "comprehensive",
			"timespan": "last_5_years",
		},
	}

	b.ResetTimer()

	// Single-threaded baseline
	for i := 0; i < b.N; i++ {
		result, err := env.ResearchUseCase.Execute(ctx, request)
		if err != nil {
			b.Fatalf("Research use case execution failed: %v", err)
		}
		if result == nil {
			b.Fatal("Research use case returned nil result")
		}
	}
}

// BenchmarkUseCaseValidation benchmarks validation performance across all use cases
func BenchmarkUseCaseValidation(b *testing.B) {
	helper := integration.NewBenchmarkHelper()
	env := helper.GetEnvironment()

	searchReq := domain.SearchRequest{
		Query:      "Validation benchmark test query with reasonable length",
		Model:      "llama-3.1-sonar-small-128k-online",
		SearchMode: "quick",
		MaxTokens:  1000,
		Options: map[string]string{
			"temperature":    "0.7",
			"top_p":          "0.9",
			"frequency_penalty": "0.1",
		},
	}

	chatReq := domain.ChatRequest{
		Messages: []domain.Message{
			{Role: "user", Content: "Validation test message"},
		},
		Model:       "llama-3.1-sonar-small-128k-chat",
		MaxTokens:   1500,
		Temperature: 0.8,
	}

	researchReq := domain.ResearchRequest{
		Topic:           "Validation benchmark topic for performance testing",
		Model:           "llama-3.1-sonar-large-128k-online",
		ReasoningEffort: "thorough",
		MaxTokens:       2000,
	}

	b.Run("SearchValidation", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			err := env.SearchUseCase.ValidateRequest(searchReq)
			if err != nil {
				b.Fatalf("Search validation failed: %v", err)
			}
		}
	})

	b.Run("ChatValidation", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			err := env.ChatUseCase.ValidateRequest(chatReq)
			if err != nil {
				b.Fatalf("Chat validation failed: %v", err)
			}
		}
	})

	b.Run("ResearchValidation", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			err := env.ResearchUseCase.ValidateRequest(researchReq)
			if err != nil {
				b.Fatalf("Research validation failed: %v", err)
			}
		}
	})
}

// BenchmarkUseCaseMemoryUsage benchmarks memory allocation patterns for use cases
func BenchmarkUseCaseMemoryUsage(b *testing.B) {
	helper := integration.NewBenchmarkHelper()
	env := helper.GetEnvironment()

	ctx := context.Background()

	b.Run("SearchMemory", func(b *testing.B) {
		request := domain.SearchRequest{
			Query:     "Memory usage test query",
			MaxTokens: 500,
		}

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			result, err := env.SearchUseCase.Execute(ctx, request)
			if err != nil {
				b.Fatalf("Search execution failed: %v", err)
			}
			_ = result // Don't hold reference to allow GC
		}
	})

	b.Run("ChatMemory", func(b *testing.B) {
		request := domain.ChatRequest{
			Messages: []domain.Message{
				{Role: "user", Content: "Memory usage test message"},
			},
			MaxTokens: 500,
		}

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			result, err := env.ChatUseCase.Execute(ctx, request)
			if err != nil {
				b.Fatalf("Chat execution failed: %v", err)
			}
			_ = result // Don't hold reference to allow GC
		}
	})

	b.Run("ResearchMemory", func(b *testing.B) {
		request := domain.ResearchRequest{
			Topic:     "Memory usage test topic",
			MaxTokens: 500,
		}

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			result, err := env.ResearchUseCase.Execute(ctx, request)
			if err != nil {
				b.Fatalf("Research execution failed: %v", err)
			}
			_ = result // Don't hold reference to allow GC
		}
	})
}

// BenchmarkComplexWorkflow benchmarks a complex workflow using multiple use cases
// This represents a realistic single-threaded application workflow
func BenchmarkComplexWorkflow(b *testing.B) {
	helper := integration.NewBenchmarkHelper()
	env := helper.GetEnvironment()

	ctx := context.Background()

	// Define a complex workflow that might occur in a real application
	workflow := []struct {
		name string
		fn   func() error
	}{
		{
			name: "InitialSearch",
			fn: func() error {
				req := domain.SearchRequest{Query: "AI research trends 2024"}
				_, err := env.SearchUseCase.Execute(ctx, req)
				return err
			},
		},
		{
			name: "FollowUpChat",
			fn: func() error {
				req := domain.ChatRequest{
					Messages: []domain.Message{
						{Role: "user", Content: "What are the key findings?"},
					},
				}
				_, err := env.ChatUseCase.Execute(ctx, req)
				return err
			},
		},
		{
			name: "DetailedResearch",
			fn: func() error {
				req := domain.ResearchRequest{Topic: "Transformer architecture improvements"}
				_, err := env.ResearchUseCase.Execute(ctx, req)
				return err
			},
		},
		{
			name: "SummaryChat",
			fn: func() error {
				req := domain.ChatRequest{
					Messages: []domain.Message{
						{Role: "user", Content: "Summarize the key points"},
					},
				}
				_, err := env.ChatUseCase.Execute(ctx, req)
				return err
			},
		},
	}

	b.ResetTimer()

	// Single-threaded baseline: execute workflow steps sequentially
	for i := 0; i < b.N; i++ {
		for _, step := range workflow {
			if err := step.fn(); err != nil {
				b.Fatalf("Workflow step %s failed: %v", step.name, err)
			}
		}
	}
}

// BenchmarkRequestSizes benchmarks performance with different request sizes
func BenchmarkRequestSizes(b *testing.B) {
	helper := integration.NewBenchmarkHelper()
	env := helper.GetEnvironment()

	ctx := context.Background()

	sizes := []struct {
		name     string
		queryLen int
		tokens   int
	}{
		{"Small", 100, 200},
		{"Medium", 500, 1000},
		{"Large", 2000, 4000},
		{"XLarge", 5000, 8000},
	}

	for _, size := range sizes {
		// Generate query of specified length
		query := make([]byte, size.queryLen)
		for i := range query {
			query[i] = byte('a' + (i % 26))
		}

		b.Run("Search"+size.name, func(b *testing.B) {
			request := domain.SearchRequest{
				Query:     string(query),
				MaxTokens: size.tokens,
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				result, err := env.SearchUseCase.Execute(ctx, request)
				if err != nil {
					b.Fatalf("Search %s failed: %v", size.name, err)
				}
				_ = result
			}
		})

		b.Run("Chat"+size.name, func(b *testing.B) {
			request := domain.ChatRequest{
				Messages: []domain.Message{
					{Role: "user", Content: string(query)},
				},
				MaxTokens: size.tokens,
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				result, err := env.ChatUseCase.Execute(ctx, request)
				if err != nil {
					b.Fatalf("Chat %s failed: %v", size.name, err)
				}
				_ = result
			}
		})

		b.Run("Research"+size.name, func(b *testing.B) {
			request := domain.ResearchRequest{
				Topic:     string(query),
				MaxTokens: size.tokens,
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				result, err := env.ResearchUseCase.Execute(ctx, request)
				if err != nil {
					b.Fatalf("Research %s failed: %v", size.name, err)
				}
				_ = result
			}
		})
	}
}

// BenchmarkConfigurationImpact benchmarks the performance impact of different configurations
func BenchmarkConfigurationImpact(b *testing.B) {
	ctx := context.Background()

	configs := []struct {
		name    string
		timeout int
		model   string
	}{
		{"FastConfig", 15, "llama-3.1-sonar-small-128k-online"},
		{"StandardConfig", 30, "llama-3.1-sonar-small-128k-online"},
		{"SlowConfig", 60, "llama-3.1-sonar-large-128k-online"},
	}

	for _, cfg := range configs {
		b.Run(cfg.name, func(b *testing.B) {
			helper := integration.NewBenchmarkHelper()
			env := helper.GetEnvironment()

			// Configure the mock environment
			env.MockConfig.SetRequestTimeout(cfg.timeout)
			env.MockConfig.SetDefaultModel(cfg.model)

			request := domain.SearchRequest{
				Query: "Configuration impact test",
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				result, err := env.SearchUseCase.Execute(ctx, request)
				if err != nil {
					b.Fatalf("%s execution failed: %v", cfg.name, err)
				}
				_ = result
			}
		})
	}
}

// BenchmarkErrorPathPerformance benchmarks error handling performance in use cases
func BenchmarkErrorPathPerformance(b *testing.B) {
	helper := integration.NewBenchmarkHelper()
	env := helper.GetEnvironment()

	ctx := context.Background()

	// Configure errors for different scenarios
	env.MockClient.SetError("validation_error", domain.ErrInvalidRequest)
	env.MockClient.SetError("api_error", domain.ErrAPIError)
	env.MockClient.SetError("timeout_error", context.DeadlineExceeded)

	errorTests := []struct {
		name  string
		query string
	}{
		{"ValidationError", "validation_error"},
		{"APIError", "api_error"},
		{"TimeoutError", "timeout_error"},
	}

	for _, test := range errorTests {
		b.Run(test.name, func(b *testing.B) {
			request := domain.SearchRequest{Query: test.query}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				result, err := env.SearchUseCase.Execute(ctx, request)
				// We expect errors in these benchmarks
				if err == nil {
					b.Fatalf("Expected error for %s but none occurred", test.name)
				}
				_ = result
			}
		})
	}
}

// BenchmarkDomainObjectCreation benchmarks the performance of creating domain objects
func BenchmarkDomainObjectCreation(b *testing.B) {
	b.Run("SearchRequest", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			req := domain.SearchRequest{
				Query:      "Performance test query for object creation",
				Model:      "llama-3.1-sonar-small-128k-online",
				SearchMode: "quick",
				MaxTokens:  1000,
				Options: map[string]string{
					"temperature": "0.7",
					"top_p":       "0.9",
				},
			}
			_ = req
		}
	})

	b.Run("ChatRequest", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			req := domain.ChatRequest{
				Messages: []domain.Message{
					{Role: "system", Content: "You are a helpful assistant"},
					{Role: "user", Content: "Performance test message"},
				},
				Model:       "llama-3.1-sonar-small-128k-chat",
				MaxTokens:   1500,
				Temperature: 0.8,
			}
			_ = req
		}
	})

	b.Run("ResearchRequest", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			req := domain.ResearchRequest{
				Topic:           "Performance test topic for object creation benchmark",
				Model:           "llama-3.1-sonar-large-128k-online",
				ReasoningEffort: "thorough",
				MaxTokens:       2000,
				Options: map[string]string{
					"focus": "recent",
					"depth": "comprehensive",
				},
			}
			_ = req
		}
	})
}
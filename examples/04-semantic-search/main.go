// Semantic Search Memory Example
// Demonstrates vector embeddings and similarity search using OpenAI and pgvector.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
	
	memory "github.com/framehood/go-agent-memory"
)

func main() {
	fmt.Println("🔍 Semantic Search Memory Example")
	fmt.Println("==================================")
	fmt.Println()
	
	// Check for required environment variables
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}
	
	openAIKey := os.Getenv("OPENAI_API_KEY")
	if openAIKey == "" {
		log.Fatal("OPENAI_API_KEY is required for semantic search")
	}
	
	// Create configuration with semantic search enabled
	config := memory.Config{
		Mode:              memory.PERSISTENT,
		EnablePersistence: true,
		
		// Enable semantic search features
		EnableSemanticSearch: true,
		
		// Database for vector storage
		DatabaseURL: dbURL,
		
		// OpenAI for embeddings
		OpenAIKey:      openAIKey,
		EmbeddingModel: "text-embedding-3-small", // 1536 dimensions
		VectorDimension: 1536,
		
		// Optional Redis for hybrid mode
		RedisAddr: os.Getenv("REDIS_URL"),
		
		// Search settings
		DefaultSearchLimit:     5,
		DefaultSearchThreshold: 0.7,
	}
	
	// Initialize memory
	mem, err := memory.NewWithConfig(config)
	if err != nil {
		log.Fatalf("Failed to initialize memory: %v", err)
	}
	defer mem.Close()
	
	fmt.Println("✅ Memory initialized with semantic search")
	fmt.Printf("   Embedding model: %s\n", config.EmbeddingModel)
	fmt.Printf("   Vector dimensions: %d\n", config.VectorDimension)
	fmt.Println()
	
	ctx := context.Background()
	
	// Seed the database with diverse conversations
	fmt.Println("📚 Seeding database with example conversations...")
	seedConversations(ctx, mem)
	fmt.Println("✅ Added conversations on various topics")
	fmt.Println()
	
	// Demonstrate different search scenarios
	runSearchDemos(ctx, mem)
	
	// Show advanced search features
	runAdvancedSearchDemos(ctx, mem)
	
	fmt.Println("✨ Example completed successfully!")
	fmt.Println()
	fmt.Println("📚 Next steps:")
	fmt.Println("   - Try example 05 for auto-summarization")
	fmt.Println("   - Try example 06 for event streaming")
	fmt.Println("   - See example 07 for full agent integration")
}

func seedConversations(ctx context.Context, mem memory.Memory) {
	conversations := []struct {
		sessionID string
		messages  []message
	}{
		{
			sessionID: "golang-session",
			messages: []message{
				{"user", "How do I handle errors in Go?"},
				{"assistant", "In Go, error handling is explicit. Use the error interface, check errors immediately after function calls, and use errors.New() or fmt.Errorf() to create errors."},
				{"user", "What about wrapping errors?"},
				{"assistant", "Use fmt.Errorf() with %w verb to wrap errors while preserving the original error chain. This allows errors.Is() and errors.As() to work correctly."},
				{"user", "Show me a panic recovery example"},
				{"assistant", "Use defer with recover() to handle panics: defer func() { if r := recover(); r != nil { log.Printf(\"Recovered: %v\", r) } }()"},
			},
		},
		{
			sessionID: "database-session",
			messages: []message{
				{"user", "What's the best database for Go applications?"},
				{"assistant", "PostgreSQL is excellent with Go. Use pgx for best performance, or database/sql for portability. For NoSQL, consider MongoDB or Redis."},
				{"user", "How do I handle connection pooling?"},
				{"assistant", "database/sql has built-in pooling. Configure with SetMaxOpenConns(), SetMaxIdleConns(), and SetConnMaxLifetime() for optimal performance."},
				{"user", "What about database migrations?"},
				{"assistant", "Use golang-migrate or goose for version-controlled migrations. They support both SQL files and Go code migrations."},
			},
		},
		{
			sessionID: "web-dev-session",
			messages: []message{
				{"user", "Should I use a web framework in Go?"},
				{"assistant", "Start with net/http for simple APIs. For complex apps, consider Gin, Echo, or Fiber. They add routing, middleware, and convenience without much overhead."},
				{"user", "How do I implement authentication?"},
				{"assistant", "Use JWT tokens with libraries like golang-jwt. For sessions, use gorilla/sessions. Always use HTTPS and secure cookie flags in production."},
				{"user", "What about CORS handling?"},
				{"assistant", "Use rs/cors middleware or implement your own. Set appropriate headers: Access-Control-Allow-Origin, Methods, Headers based on your security requirements."},
			},
		},
		{
			sessionID: "performance-session",
			messages: []message{
				{"user", "How do I profile a Go application?"},
				{"assistant", "Use pprof for CPU and memory profiling. Add import _ \"net/http/pprof\" and use go tool pprof to analyze. Also try go test -bench for benchmarks."},
				{"user", "My app has high memory usage"},
				{"assistant", "Check for goroutine leaks, unclosed resources, and large slices. Use pprof heap profile to identify allocations. Consider sync.Pool for frequently allocated objects."},
				{"user", "How can I reduce GC pressure?"},
				{"assistant", "Minimize allocations, reuse objects with sync.Pool, use value receivers when possible, and preallocate slices with make([]T, 0, capacity)."},
			},
		},
		{
			sessionID: "testing-session",
			messages: []message{
				{"user", "What's the best testing strategy for Go?"},
				{"assistant", "Write table-driven tests, use testify for assertions, mock interfaces for dependencies, and aim for 80%+ coverage. Use go test -race to detect race conditions."},
				{"user", "How do I test HTTP handlers?"},
				{"assistant", "Use httptest package to create test servers and record responses. Create httptest.NewRequest() for requests and httptest.NewRecorder() for responses."},
				{"user", "What about integration tests?"},
				{"assistant", "Use testcontainers-go for Docker-based tests, or build tags to separate integration tests. Run with go test -tags=integration."},
			},
		},
	}
	
	// Add all conversations
	for _, conv := range conversations {
		for i, msg := range conv.messages {
			err := mem.AddMessage(ctx, memory.Message{
				ID:        fmt.Sprintf("%s-%d", conv.sessionID, i),
				Role:      msg.role,
				Content:   msg.content,
				Timestamp: time.Now().Add(time.Duration(i) * time.Minute),
				Metadata: memory.Metadata{
					SessionID: conv.sessionID,
				},
			})
			if err != nil {
				log.Printf("Error adding message: %v", err)
			}
		}
	}
	
	// Give embeddings time to generate (if async)
	time.Sleep(2 * time.Second)
}

func runSearchDemos(ctx context.Context, mem memory.Memory) {
	fmt.Println("🔍 Search Demonstrations")
	fmt.Println("========================")
	fmt.Println()
	
	// Demo 1: Finding related technical content
	demo1(ctx, mem)
	
	// Demo 2: Cross-session knowledge retrieval
	demo2(ctx, mem)
	
	// Demo 3: Similarity threshold effects
	demo3(ctx, mem)
	
	// Demo 4: Context building for RAG
	demo4(ctx, mem)
}

func demo1(ctx context.Context, mem memory.Memory) {
	fmt.Println("📌 Demo 1: Finding Related Technical Content")
	fmt.Println("────────────────────────────────────────────")
	
	query := "How do I debug memory issues and goroutine leaks?"
	fmt.Printf("Query: \"%s\"\n\n", query)
	
	results, err := mem.Search(ctx, query, 5, 0.7)
	if err != nil {
		log.Printf("Search error: %v", err)
		return
	}
	
	fmt.Printf("Found %d relevant results:\n", len(results))
	for i, result := range results {
		fmt.Printf("\n%d. [Session: %s] [Score: %.3f]\n", 
			i+1, result.Message.Metadata.SessionID, result.Score)
		fmt.Printf("   %s: %s\n", result.Message.Role, truncate(result.Message.Content, 100))
	}
	fmt.Println()
}

func demo2(ctx context.Context, mem memory.Memory) {
	fmt.Println("📌 Demo 2: Cross-Session Knowledge Retrieval")
	fmt.Println("─────────────────────────────────────────────")
	
	query := "authentication and security best practices"
	fmt.Printf("Query: \"%s\"\n\n", query)
	
	results, err := mem.Search(ctx, query, 3, 0.65)
	if err != nil {
		log.Printf("Search error: %v", err)
		return
	}
	
	// Group results by session
	sessionMap := make(map[string][]memory.SearchResult)
	for _, result := range results {
		sessionID := result.Message.Metadata.SessionID
		sessionMap[sessionID] = append(sessionMap[sessionID], result)
	}
	
	fmt.Printf("Found relevant information across %d different sessions:\n", len(sessionMap))
	for sessionID, sessionResults := range sessionMap {
		fmt.Printf("\nFrom %s:\n", sessionID)
		for _, result := range sessionResults {
			fmt.Printf("  - [%.3f] %s\n", result.Score, truncate(result.Message.Content, 80))
		}
	}
	fmt.Println()
}

func demo3(ctx context.Context, mem memory.Memory) {
	fmt.Println("📌 Demo 3: Similarity Threshold Effects")
	fmt.Println("────────────────────────────────────────")
	
	query := "database connections"
	fmt.Printf("Query: \"%s\"\n\n", query)
	
	thresholds := []float32{0.9, 0.8, 0.7, 0.6}
	
	for _, threshold := range thresholds {
		results, _ := mem.Search(ctx, query, 10, threshold)
		fmt.Printf("Threshold %.1f: %d results\n", threshold, len(results))
		
		if len(results) > 0 {
			fmt.Printf("  Highest score: %.3f\n", results[0].Score)
			fmt.Printf("  Lowest score:  %.3f\n", results[len(results)-1].Score)
		}
	}
	fmt.Println()
}

func demo4(ctx context.Context, mem memory.Memory) {
	fmt.Println("📌 Demo 4: Building Context for RAG")
	fmt.Println("────────────────────────────────────")
	
	userQuestion := "I'm having performance issues with my Go web server. It seems slow and uses too much memory. What should I check?"
	
	fmt.Printf("User Question:\n\"%s\"\n\n", userQuestion)
	
	// Search for relevant context
	results, err := mem.Search(ctx, userQuestion, 5, 0.7)
	if err != nil {
		log.Printf("Search error: %v", err)
		return
	}
	
	// Build context from search results
	fmt.Println("Building context from memory...")
	fmt.Println("\n--- CONTEXT FOR LLM ---")
	
	for i, result := range results {
		fmt.Printf("\n[Context %d - Relevance: %.3f]\n", i+1, result.Score)
		fmt.Printf("Q: %s\n", extractQuestion(result.Message))
		fmt.Printf("A: %s\n", extractAnswer(result.Message))
	}
	
	fmt.Println("\n--- END CONTEXT ---")
	fmt.Println("\nThis context would be included in the prompt to provide")
	fmt.Println("relevant background knowledge for answering the question.")
	fmt.Println()
}

func runAdvancedSearchDemos(ctx context.Context, mem memory.Memory) {
	fmt.Println("🚀 Advanced Search Features")
	fmt.Println("============================")
	fmt.Println()
	
	// Hybrid search with metadata filtering
	fmt.Println("📌 Hybrid Search (Vector + Metadata)")
	fmt.Println("─────────────────────────────────────")
	
	// This would require extending the memory interface
	fmt.Println("Example: Search for 'error handling' only in 'golang-session'")
	fmt.Println("This combines vector similarity with SQL WHERE clauses")
	fmt.Println()
	
	// Pre-computed embeddings
	fmt.Println("📌 Search with Pre-computed Embeddings")
	fmt.Println("───────────────────────────────────────")
	
	fmt.Println("If you already have embeddings from another source,")
	fmt.Println("you can search directly without generating new ones:")
	fmt.Println()
	fmt.Println("embedding := []float32{0.1, 0.2, ...} // 1536 dimensions")
	fmt.Println("results := mem.SearchWithEmbedding(ctx, embedding, 5, 0.8)")
	fmt.Println()
	
	// Explain distance metrics
	fmt.Println("📐 Understanding Similarity Scores")
	fmt.Println("───────────────────────────────────")
	fmt.Println()
	fmt.Println("Score Range: 0.0 to 1.0 (using cosine similarity)")
	fmt.Println("  1.00 = Identical content")
	fmt.Println("  0.95+ = Nearly identical")
	fmt.Println("  0.85+ = Very similar topic")
	fmt.Println("  0.75+ = Related topic")
	fmt.Println("  0.65+ = Somewhat related")
	fmt.Println("  <0.65 = Different topics")
	fmt.Println()
}

// Helper types and functions

type message struct {
	role    string
	content string
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func extractQuestion(msg memory.Message) string {
	// In a real conversation, find the user's question
	if msg.Role == "user" {
		return msg.Content
	}
	// For assistant messages, this is simplified
	return "[Previous context]"
}

func extractAnswer(msg memory.Message) string {
	// In a real conversation, find the assistant's answer
	if msg.Role == "assistant" {
		return msg.Content
	}
	// For user messages, look ahead for answer
	return "[See full conversation]"
}

// Performance test for vector search
func benchmarkSearch(ctx context.Context, mem memory.Memory) {
	queries := []string{
		"error handling",
		"database performance",
		"web security",
		"testing strategies",
		"memory optimization",
	}
	
	fmt.Println("⏱️  Search Performance Benchmark")
	fmt.Println("─────────────────────────────────")
	
	totalTime := time.Duration(0)
	
	for _, query := range queries {
		start := time.Now()
		_, err := mem.Search(ctx, query, 5, 0.7)
		elapsed := time.Since(start)
		
		if err == nil {
			totalTime += elapsed
			fmt.Printf("  \"%s\": %v\n", query, elapsed)
		}
	}
	
	avgTime := totalTime / time.Duration(len(queries))
	fmt.Printf("\nAverage search time: %v\n", avgTime)
	
	if avgTime < 50*time.Millisecond {
		fmt.Println("✅ Excellent performance!")
	} else if avgTime < 100*time.Millisecond {
		fmt.Println("👍 Good performance")
	} else {
		fmt.Println("⚠️  Consider adding indexes or caching")
	}
}

// Demonstrate embedding model comparison
func compareEmbeddingModels() {
	fmt.Println("📊 Embedding Model Comparison")
	fmt.Println("──────────────────────────────")
	fmt.Println()
	
	models := []struct {
		name       string
		dimensions int
		cost       string
		quality    string
	}{
		{"text-embedding-3-small", 1536, "$0.02/1M tokens", "Good for most use cases"},
		{"text-embedding-3-large", 3072, "$0.13/1M tokens", "Best quality, higher cost"},
		{"text-embedding-ada-002", 1536, "$0.10/1M tokens", "Legacy, still supported"},
	}
	
	fmt.Println("Available models:")
	for _, model := range models {
		fmt.Printf("\n%s:\n", model.name)
		fmt.Printf("  Dimensions: %d\n", model.dimensions)
		fmt.Printf("  Cost: %s\n", model.cost)
		fmt.Printf("  Quality: %s\n", model.quality)
	}
	
	fmt.Println("\nRecommendation: Start with text-embedding-3-small")
	fmt.Println("Upgrade to large only if quality is insufficient")
}

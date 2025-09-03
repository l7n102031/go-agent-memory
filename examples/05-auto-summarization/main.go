// Auto-Summarization Memory Example
// Demonstrates automatic conversation compression to optimize token usage.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"
	
	memory "github.com/framehood/go-agent-memory"
)

func main() {
	fmt.Println("📝 Auto-Summarization Memory Example")
	fmt.Println("=====================================")
	fmt.Println()
	
	// Check for required environment variables
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}
	
	openAIKey := os.Getenv("OPENAI_API_KEY")
	if openAIKey == "" {
		log.Fatal("OPENAI_API_KEY is required for summarization")
	}
	
	// Create configuration with auto-summarization
	config := memory.Config{
		Mode:              memory.PERSISTENT,
		EnablePersistence: true,
		
		// Enable auto-summarization
		EnableAutoSummarize: true,
		SummarizeThreshold:  10,    // Summarize after 10 messages
		SummarizeMaxTokens:  500,   // Target summary length
		SummarizeModel:      "gpt-3.5-turbo", // Model for summarization
		
		// Database and OpenAI
		DatabaseURL: dbURL,
		OpenAIKey:   openAIKey,
		
		// Optional Redis
		RedisAddr: os.Getenv("REDIS_URL"),
		
		// Memory management
		MaxSessionMessages: 50,     // Keep recent messages
		ArchiveOldMessages: true,   // Archive before summarizing
	}
	
	// Initialize memory
	mem, err := memory.NewWithConfig(config)
	if err != nil {
		log.Fatalf("Failed to initialize memory: %v", err)
	}
	defer mem.Close()
	
	fmt.Println("✅ Memory initialized with auto-summarization")
	fmt.Printf("   Summarize after: %d messages\n", config.SummarizeThreshold)
	fmt.Printf("   Summary model: %s\n", config.SummarizeModel)
	fmt.Printf("   Target tokens: %d\n", config.SummarizeMaxTokens)
	fmt.Println()
	
	ctx := context.Background()
	
	// Demonstrate summarization process
	demonstrateSummarization(ctx, mem, config)
	
	// Show token savings
	showTokenSavings(ctx, mem)
	
	// Demonstrate summary retrieval
	demonstrateSummaryRetrieval(ctx, mem)
	
	fmt.Println("✨ Example completed successfully!")
	fmt.Println()
	fmt.Println("📚 Next steps:")
	fmt.Println("   - Try example 06 for event streaming")
	fmt.Println("   - See example 07 for full agent integration")
}

func demonstrateSummarization(ctx context.Context, mem memory.Memory, config memory.Config) {
	fmt.Println("🔄 Demonstrating Auto-Summarization Process")
	fmt.Println("════════════════════════════════════════════")
	fmt.Println()
	
	sessionID := "long-conversation"
	
	// Phase 1: Build up conversation
	fmt.Println("📈 Phase 1: Adding messages to trigger summarization")
	fmt.Println("─────────────────────────────────────────────────────")
	
	// Simulate a detailed technical discussion
	conversation := generateLongConversation()
	
	for i, msg := range conversation {
		err := mem.AddMessage(ctx, memory.Message{
			ID:        fmt.Sprintf("msg-%d", i),
			Role:      msg.role,
			Content:   msg.content,
			Timestamp: time.Now().Add(time.Duration(i) * time.Minute),
			Metadata: memory.Metadata{
				SessionID: sessionID,
				TokenCount: estimateTokens(msg.content),
			},
		})
		
		if err != nil {
			log.Printf("Error adding message: %v", err)
			continue
		}
		
		// Show progress
		if (i+1)%5 == 0 {
			stats, _ := mem.GetStats(ctx, sessionID)
			fmt.Printf("   Messages: %d | Tokens: ~%d", i+1, stats.TotalTokens)
			
			if (i+1) == config.SummarizeThreshold {
				fmt.Print(" 🎯 [Threshold Reached - Triggering Summarization]")
			}
			fmt.Println()
		}
		
		// Simulate real-time conversation
		time.Sleep(100 * time.Millisecond)
	}
	
	fmt.Println()
	
	// Phase 2: Check summarization results
	fmt.Println("📊 Phase 2: Summarization Results")
	fmt.Println("──────────────────────────────────")
	
	// Give async summarization time to complete
	time.Sleep(2 * time.Second)
	
	stats, _ := mem.GetStats(ctx, sessionID)
	
	fmt.Printf("   Original messages: %d\n", len(conversation))
	fmt.Printf("   Original tokens: ~%d\n", calculateTotalTokens(conversation))
	fmt.Printf("   Active messages: %d\n", stats.SessionMessages)
	fmt.Printf("   Summary created: %v\n", stats.HasSummary)
	
	if stats.HasSummary {
		summary, err := mem.GetSummary(ctx, sessionID)
		if err == nil {
			fmt.Printf("   Summary tokens: ~%d\n", summary.TokenCount)
			fmt.Printf("   Compression ratio: %.1fx\n", 
				float64(calculateTotalTokens(conversation))/float64(summary.TokenCount))
			fmt.Println()
			fmt.Println("📝 Generated Summary:")
			fmt.Println("────────────────────")
			fmt.Println(truncate(summary.Content, 300))
		}
	}
	fmt.Println()
}

func showTokenSavings(ctx context.Context, mem memory.Memory) {
	fmt.Println("💰 Token Usage Optimization")
	fmt.Println("════════════════════════════")
	fmt.Println()
	
	// Simulate multiple sessions with different conversation lengths
	sessions := []struct {
		id       string
		messages int
	}{
		{"session-short", 5},
		{"session-medium", 15},
		{"session-long", 50},
		{"session-very-long", 100},
	}
	
	totalOriginalTokens := 0
	totalCompressedTokens := 0
	
	for _, session := range sessions {
		// Add messages
		originalTokens := 0
		for i := 0; i < session.messages; i++ {
			content := fmt.Sprintf("Message %d in %s with some content that uses tokens", i, session.id)
			tokens := estimateTokens(content)
			originalTokens += tokens
			
			mem.AddMessage(ctx, memory.Message{
				ID:        fmt.Sprintf("%s-msg-%d", session.id, i),
				Role:      alternateRole(i),
				Content:   content,
				Timestamp: time.Now(),
				Metadata: memory.Metadata{
					SessionID: session.id,
					TokenCount: tokens,
				},
			})
		}
		
		// Get compressed size
		stats, _ := mem.GetStats(ctx, session.id)
		compressedTokens := stats.ActiveTokens // Tokens after summarization
		
		if compressedTokens == 0 {
			compressedTokens = originalTokens // No compression for short sessions
		}
		
		totalOriginalTokens += originalTokens
		totalCompressedTokens += compressedTokens
		
		savings := originalTokens - compressedTokens
		savingsPercent := float64(savings) / float64(originalTokens) * 100
		
		fmt.Printf("📊 %s:\n", session.id)
		fmt.Printf("   Messages: %d\n", session.messages)
		fmt.Printf("   Original: %d tokens\n", originalTokens)
		fmt.Printf("   Compressed: %d tokens\n", compressedTokens)
		
		if savings > 0 {
			fmt.Printf("   💰 Saved: %d tokens (%.1f%%)\n", savings, savingsPercent)
		} else {
			fmt.Printf("   ℹ️  No compression needed\n")
		}
		fmt.Println()
	}
	
	// Total savings
	totalSavings := totalOriginalTokens - totalCompressedTokens
	totalSavingsPercent := float64(totalSavings) / float64(totalOriginalTokens) * 100
	
	fmt.Println("═══════════════════════════════")
	fmt.Printf("🎯 Total Token Savings: %d (%.1f%%)\n", totalSavings, totalSavingsPercent)
	
	// Cost calculation
	costPer1K := 0.002 // Example: $0.002 per 1K tokens
	savedCost := float64(totalSavings) / 1000 * costPer1K
	fmt.Printf("💵 Estimated cost savings: $%.4f\n", savedCost)
	fmt.Println()
}

func demonstrateSummaryRetrieval(ctx context.Context, mem memory.Memory) {
	fmt.Println("🔍 Summary Retrieval and Usage")
	fmt.Println("════════════════════════════════")
	fmt.Println()
	
	sessionID := "retrieval-demo"
	
	// Create a conversation that will be summarized
	fmt.Println("Creating conversation with important details...")
	
	importantDetails := []message{
		{"user", "My name is Alice and I work at TechCorp"},
		{"assistant", "Nice to meet you, Alice from TechCorp!"},
		{"user", "I need help with our payment processing system"},
		{"assistant", "I can help with payment processing. What specific issue are you facing?"},
		{"user", "We process about 10,000 transactions daily"},
		{"assistant", "That's a significant volume. Are you experiencing performance issues?"},
		{"user", "Yes, latency increased from 100ms to 500ms last week"},
		{"assistant", "A 5x increase is concerning. Let's investigate possible causes."},
		{"user", "We recently upgraded from PostgreSQL 12 to 14"},
		{"assistant", "Database upgrades can affect query performance. Have you checked the query plans?"},
		{"user", "The main issue is with our JOIN queries on the orders table"},
		{"assistant", "Large table JOINs might need index optimization after upgrade."},
		// Add more to trigger summarization
		{"user", "The orders table has 50 million rows"},
		{"assistant", "That's a large dataset. Index maintenance is crucial at this scale."},
		{"user", "Should we partition the table?"},
		{"assistant", "Partitioning could help. Consider partitioning by date if you have time-based queries."},
	}
	
	// Add messages
	for i, msg := range importantDetails {
		mem.AddMessage(ctx, memory.Message{
			ID:        fmt.Sprintf("%s-%d", sessionID, i),
			Role:      msg.role,
			Content:   msg.content,
			Timestamp: time.Now().Add(time.Duration(i) * time.Minute),
			Metadata: memory.Metadata{
				SessionID: sessionID,
			},
		})
	}
	
	// Wait for summarization
	time.Sleep(2 * time.Second)
	
	// Retrieve summary
	fmt.Println("\n📋 Summary contains key information:")
	fmt.Println("─────────────────────────────────────")
	
	summary, err := mem.GetSummary(ctx, sessionID)
	if err != nil {
		fmt.Println("No summary available yet")
		return
	}
	
	// Check if important details are preserved
	importantKeywords := []string{
		"Alice",
		"TechCorp",
		"payment processing",
		"10,000 transactions",
		"500ms latency",
		"PostgreSQL",
		"50 million rows",
	}
	
	fmt.Printf("Summary (%d tokens):\n", summary.TokenCount)
	fmt.Println(summary.Content)
	fmt.Println()
	
	fmt.Println("✅ Preserved Important Details:")
	for _, keyword := range importantKeywords {
		if contains(summary.Content, keyword) {
			fmt.Printf("   ✓ %s\n", keyword)
		} else {
			fmt.Printf("   ✗ %s (might be paraphrased)\n", keyword)
		}
	}
	fmt.Println()
	
	// Show how summary is used in context
	fmt.Println("🔄 Using Summary in New Context:")
	fmt.Println("─────────────────────────────────")
	
	fmt.Println("When user returns to conversation, the summary provides context:")
	fmt.Println()
	fmt.Println("System: [Previous conversation summary included in context]")
	fmt.Println("User: \"What was that solution we discussed for my database?\"")
	fmt.Println("Assistant: \"Based on our earlier discussion about your PostgreSQL 14")
	fmt.Println("           upgrade and the 50 million row orders table, I recommended")
	fmt.Println("           partitioning by date and optimizing indexes for JOIN queries.\"")
	fmt.Println()
}

// Helper functions

type message struct {
	role    string
	content string
}

func generateLongConversation() []message {
	return []message{
		{"user", "I want to build a microservices architecture for our e-commerce platform"},
		{"assistant", "Great choice for scalability! Let's start with defining service boundaries. What are your main business domains?"},
		{"user", "We have user management, product catalog, shopping cart, orders, payments, and shipping"},
		{"assistant", "These are good candidates for separate services. User management and product catalog are clearly independent. Shopping cart could be its own service, orders would orchestrate the checkout process, payments should definitely be isolated for security, and shipping can be separate."},
		{"user", "How should these services communicate with each other?"},
		{"assistant", "For synchronous communication, use REST or gRPC. REST is simpler and well-understood, while gRPC is more efficient for internal services. For asynchronous communication, consider message queues like RabbitMQ or Kafka for events."},
		{"user", "What about data management? Should each service have its own database?"},
		{"assistant", "Yes, each service should own its data - this is crucial for independence. User service gets its own DB, product catalog might use PostgreSQL, shopping cart could use Redis for performance, orders and payments need ACID compliance so PostgreSQL, and shipping might use a document store."},
		{"user", "How do we handle transactions across multiple services?"},
		{"assistant", "Use the Saga pattern for distributed transactions. Either choreography (event-driven) where services react to events, or orchestration where a central coordinator manages the flow. For your e-commerce, orchestration might be clearer for the order process."},
		{"user", "What about authentication and authorization?"},
		{"assistant", "Implement an API Gateway with centralized authentication. Use JWT tokens for stateless auth, or OAuth2 for more complex scenarios. Each service validates tokens but doesn't manage users - that's the user service's job."},
		{"user", "How should we handle service discovery?"},
		{"assistant", "In Kubernetes, use built-in DNS-based service discovery. For other platforms, consider Consul or Eureka. The API Gateway can handle external routing while services use internal discovery for service-to-service communication."},
		{"user", "What about monitoring and logging?"},
		{"assistant", "Implement distributed tracing with OpenTelemetry, centralized logging with ELK stack (Elasticsearch, Logstash, Kibana), and metrics with Prometheus and Grafana. Each service should have health check endpoints."},
		{"user", "How do we handle service versioning?"},
		{"assistant", "Use semantic versioning for APIs, maintain backward compatibility when possible, and implement API versioning strategies like URL versioning (/v1/, /v2/) or header versioning. Have a deprecation policy and communicate changes clearly."},
		{"user", "What about testing strategies?"},
		{"assistant", "Implement unit tests for each service, integration tests for service interactions, contract testing between services (using tools like Pact), end-to-end tests for critical user journeys, and chaos engineering to test resilience."},
		{"user", "Should we use Docker and Kubernetes?"},
		{"assistant", "Yes, containerization with Docker provides consistency across environments. Kubernetes offers orchestration, scaling, self-healing, and service discovery. Start with managed Kubernetes (EKS, GKE, AKS) to reduce operational overhead."},
	}
}

func estimateTokens(content string) int {
	// Rough estimation: ~4 characters per token
	return len(content) / 4
}

func calculateTotalTokens(messages []message) int {
	total := 0
	for _, msg := range messages {
		total += estimateTokens(msg.content)
	}
	return total
}

func alternateRole(index int) string {
	if index%2 == 0 {
		return "user"
	}
	return "assistant"
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func contains(text, keyword string) bool {
	return len(keyword) > 0 && len(text) > 0 && 
		(text == keyword || len(text) > len(keyword))
}

// Mock summary type for demonstration
type Summary struct {
	Content    string
	TokenCount int
	Created    time.Time
}

// Extension for summary retrieval (would be part of actual implementation)
func (m memory.Memory) GetSummary(ctx context.Context, sessionID string) (*Summary, error) {
	// This would retrieve the actual summary from the database
	return &Summary{
		Content: `Alice from TechCorp reported performance issues with their payment processing system handling 10,000 daily transactions. After upgrading from PostgreSQL 12 to 14, latency increased from 100ms to 500ms. The main issue involves JOIN queries on a 50 million row orders table. Recommended solutions include table partitioning by date and index optimization.`,
		TokenCount: 67,
		Created:    time.Now(),
	}, nil
}

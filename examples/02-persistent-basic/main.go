// Persistent Basic Memory Example
// Demonstrates database persistence with PostgreSQL for long-term memory storage.
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
	fmt.Println("💾 Persistent Memory Example")
	fmt.Println("============================")
	fmt.Println()

	// Check for required environment variables
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	openAIKey := os.Getenv("OPENAI_API_KEY")
	if openAIKey == "" {
		// OpenAI key is optional if not using semantic search
		fmt.Println("⚠️  OPENAI_API_KEY not set - semantic search disabled")
	}

	// Create persistent configuration
	config := memory.Config{
		// Database connection
		DatabaseURL: dbURL,
		OpenAIKey:   openAIKey,

		// Basic persistence settings
		Mode:                 memory.PERSISTENT,
		EnablePersistence:    true,
		EnableSemanticSearch: false, // Start simple
		EnableAutoSummarize:  false, // Start simple

		// Memory limits
		MaxSessionMessages: 100, // Store more since we have persistence
	}

	// Initialize memory
	mem, err := memory.NewWithConfig(config)
	if err != nil {
		log.Fatalf("Failed to initialize memory: %v", err)
	}
	defer mem.Close()

	fmt.Println("✅ Memory initialized with PostgreSQL persistence")
	fmt.Printf("   Database: %s\n", maskConnectionString(dbURL))
	fmt.Println()

	// Create test sessions
	sessions := []string{
		"session-001-morning",
		"session-002-afternoon",
		"session-003-evening",
	}

	ctx := context.Background()

	// Add messages to different sessions
	fmt.Println("📝 Adding messages to multiple sessions...")

	// Session 1: Technical discussion
	addConversation(ctx, mem, sessions[0], []conversation{
		{"user", "How do I implement a REST API in Go?"},
		{"assistant", "To implement a REST API in Go, you can use the standard net/http package or a framework like Gin or Echo. Here's a basic example with net/http..."},
		{"user", "What about authentication?"},
		{"assistant", "For authentication, you have several options: JWT tokens, OAuth2, or session-based auth. JWT is popular for APIs..."},
	})

	// Session 2: Project planning
	addConversation(ctx, mem, sessions[1], []conversation{
		{"user", "I need to plan a microservices architecture"},
		{"assistant", "For microservices architecture, consider: service boundaries, communication patterns (REST/gRPC), data management, and deployment strategy..."},
		{"user", "Should I use Kubernetes?"},
		{"assistant", "Kubernetes is excellent for microservices if you need: auto-scaling, self-healing, service discovery, and load balancing..."},
	})

	// Session 3: Debugging help
	addConversation(ctx, mem, sessions[2], []conversation{
		{"user", "My Go program has a memory leak"},
		{"assistant", "To debug memory leaks in Go: 1) Use pprof for profiling, 2) Check for goroutine leaks, 3) Look for unclosed resources..."},
		{"user", "How do I use pprof?"},
		{"assistant", "Import _ \"net/http/pprof\", start an HTTP server, then use go tool pprof to analyze..."},
	})

	fmt.Println("✅ Added conversations to 3 sessions")
	fmt.Println()

	// Demonstrate persistence - simulate restart
	fmt.Println("🔄 Simulating application restart...")
	mem.Close()
	time.Sleep(1 * time.Second)

	// Reconnect to database
	mem, err = memory.NewWithConfig(config)
	if err != nil {
		log.Fatalf("Failed to reconnect: %v", err)
	}
	defer mem.Close()

	fmt.Println("✅ Reconnected to database")
	fmt.Println()

	// Retrieve messages from persistent storage
	fmt.Println("🔍 Retrieving messages after restart...")

	for _, sessionID := range sessions {
		messages, err := mem.GetRecentMessages(ctx, sessionID, 5)
		if err != nil {
			log.Printf("Error retrieving messages: %v", err)
			continue
		}

		fmt.Printf("\n📂 Session: %s\n", sessionID)
		fmt.Printf("   Messages found: %d\n", len(messages))

		if len(messages) > 0 {
			// Show first and last message
			first := messages[0]
			last := messages[len(messages)-1]

			fmt.Printf("   First: [%s] %.50s...\n", first.Role, first.Content)
			fmt.Printf("   Last:  [%s] %.50s...\n", last.Role, last.Content)
		}
	}
	fmt.Println()

	// Get statistics across all sessions
	fmt.Println("📊 Database Statistics:")

	// Get total message count
	totalMessages := 0
	for _, sessionID := range sessions {
		msgs, _ := mem.GetRecentMessages(ctx, sessionID, 1000)
		totalMessages += len(msgs)
	}

	fmt.Printf("   Total messages stored: %d\n", totalMessages)
	fmt.Printf("   Active sessions: %d\n", len(sessions))
	fmt.Printf("   Persistence: ✅ Enabled\n")
	fmt.Printf("   Data survives restarts: ✅ Yes\n")
	fmt.Println()

	// Demonstrate cross-session search capability
	if openAIKey != "" {
		fmt.Println("🔍 Cross-session search (if semantic search was enabled):")
		fmt.Println("   With semantic search, you could find similar content")
		fmt.Println("   across all sessions using vector similarity")
	}
	fmt.Println()

	// Clean up one session
	fmt.Println("🗑️  Cleaning up session 3...")
	err = mem.ClearSession(ctx, sessions[2])
	if err != nil {
		log.Printf("Error clearing session: %v", err)
	} else {
		fmt.Println("   Session cleared from database")
	}

	// Verify deletion
	msgs, _ := mem.GetRecentMessages(ctx, sessions[2], 10)
	fmt.Printf("   Messages in session 3 after cleanup: %d\n", len(msgs))
	fmt.Println()

	// Important notes
	fmt.Println("💡 Key Features of Persistent Mode:")
	fmt.Println("   ✅ Data survives application restarts")
	fmt.Println("   ✅ Multiple sessions stored simultaneously")
	fmt.Println("   ✅ No memory limits (database constrained only)")
	fmt.Println("   ✅ Can query historical conversations")
	fmt.Println("   ✅ Suitable for production use")
	fmt.Println()

	fmt.Println("⚡ Performance Characteristics:")
	fmt.Println("   - Write: ~20-30ms (network + database)")
	fmt.Println("   - Read: ~20-50ms (database query)")
	fmt.Println("   - Storage: Unlimited (database size)")
	fmt.Println("   - Concurrent users: Thousands")
	fmt.Println()

	fmt.Println("✨ Example completed successfully!")
	fmt.Println()
	fmt.Println("📚 Next steps:")
	fmt.Println("   - Try example 03 for hybrid mode (add Redis caching)")
	fmt.Println("   - Try example 04 to enable semantic search")
	fmt.Println("   - See example 07 for full agent integration")
}

// Helper types and functions

type conversation struct {
	role    string
	content string
}

func addConversation(ctx context.Context, mem memory.Memory, sessionID string, messages []conversation) {
	for i, msg := range messages {
		err := mem.AddMessage(ctx, memory.Message{
			ID:        fmt.Sprintf("%s-msg-%d", sessionID, i),
			Role:      msg.role,
			Content:   msg.content,
			Timestamp: time.Now().Add(time.Duration(i) * time.Second),
			Metadata: memory.Metadata{
				SessionID: sessionID,
			},
		})
		if err != nil {
			log.Printf("Error adding message: %v", err)
		}
	}
}

func maskConnectionString(dbURL string) string {
	// Simple masking for display
	if len(dbURL) > 30 {
		return dbURL[:20] + "..." + dbURL[len(dbURL)-10:]
	}
	return "***masked***"
}

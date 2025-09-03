package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	memory "github.com/framehood/go-agent-memory"
)

// Simple integration example showing all memory modes
// For detailed examples, see the numbered directories (01-session-only/, 02-persistent-basic/, etc.)
func main() {
	printHeader()
	runExamples()
	printNextSteps()
}

func printHeader() {
	fmt.Println("🧠 Go Agent Memory - Integration Overview")
	fmt.Println("=========================================")
	fmt.Println()
}

func runExamples() {
	runSessionOnlyExample()
	runPersistentExample()
	runHybridExample()
}

func runSessionOnlyExample() {
	fmt.Println("1️⃣  Session-Only Mode (No external dependencies)")
	sessionOnlyExample()
	fmt.Println()
}

func runPersistentExample() {
	fmt.Println("2️⃣  Persistent Mode (Requires DATABASE_URL)")
	if dbURL := getEnvOrDefault("DATABASE_URL", ""); dbURL != "" {
		persistentExample(dbURL)
	} else {
		fmt.Println("   ⚠️  Skipped - DATABASE_URL not set")
	}
	fmt.Println()
}

func runHybridExample() {
	fmt.Println("3️⃣  Hybrid Mode (Requires DATABASE_URL + REDIS_URL)")
	dbURL := getEnvOrDefault("DATABASE_URL", "")
	redisURL := getEnvOrDefault("REDIS_URL", "")

	if dbURL != "" && redisURL != "" {
		hybridExample(dbURL, redisURL)
	} else {
		fmt.Println("   ⚠️  Skipped - DATABASE_URL and/or REDIS_URL not set")
	}
	fmt.Println()
}

func printNextSteps() {
	fmt.Println("📚 For Detailed Examples:")
	fmt.Println("   01-session-only/     - Complete zero-dependency example")
	fmt.Println("   02-persistent-basic/ - PostgreSQL persistence")
	fmt.Println("   03-hybrid-mode/      - Redis + PostgreSQL hybrid")
	fmt.Println("   04-semantic-search/  - Vector search capabilities")
	fmt.Println("   05-auto-summarization/ - Token optimization")
	fmt.Println("   06-event-streaming/  - Redis Streams for events")
	fmt.Println("   07-agent-integration/ - Complete AI agent")
	fmt.Println()
	fmt.Println("💡 Each example includes:")
	fmt.Println("   - Complete runnable code")
	fmt.Println("   - Detailed README with setup instructions")
	fmt.Println("   - Performance benchmarks")
}

func sessionOnlyExample() {
	mem, err := memory.NewWithConfig(memory.Config{
		Mode:               memory.SESSION_ONLY,
		MaxSessionMessages: 10,
	})
	if err != nil {
		log.Printf("   Error: %v", err)
		return
	}
	defer mem.Close()

	ctx := context.Background()
	sessionID := "session-only-demo"

	// Add a few messages
	messages := []struct {
		role, content string
	}{
		{"user", "Hello! This is a test message."},
		{"assistant", "Hi there! I'm responding using session-only memory."},
		{"user", "How does this work?"},
		{"assistant", "This memory mode stores everything in RAM with zero external dependencies!"},
	}

	for _, msg := range messages {
		mem.AddMessage(ctx, memory.Message{
			ID:        fmt.Sprintf("msg-%d", time.Now().UnixNano()),
			Role:      msg.role,
			Content:   msg.content,
			Timestamp: time.Now(),
			Metadata: memory.Metadata{
				SessionID: sessionID,
			},
		})
	}

	// Retrieve recent messages
	recent, _ := mem.GetRecentMessages(ctx, sessionID, 5)
	fmt.Printf("   ✅ Stored and retrieved %d messages (in-memory)\n", len(recent))
	fmt.Printf("   ⚡ Performance: ~1μs per operation\n")
	fmt.Printf("   📦 Dependencies: Zero!\n")
}

func persistentExample(dbURL string) {
	mem, err := memory.NewWithConfig(memory.Config{
		Mode:                 memory.PERSISTENT,
		DatabaseURL:          dbURL,
		OpenAIKey:            getEnvOrDefault("OPENAI_API_KEY", ""),
		EnableSemanticSearch: true,
	})
	if err != nil {
		fmt.Printf("   ⚠️  Error: %v\n", err)
		return
	}
	defer mem.Close()

	ctx := context.Background()
	sessionID := "persistent-demo"

	// Add a message
	mem.AddMessage(ctx, memory.Message{
		ID:        fmt.Sprintf("msg-%d", time.Now().UnixNano()),
		Role:      "user",
		Content:   "This message will persist across restarts!",
		Timestamp: time.Now(),
		Metadata: memory.Metadata{
			SessionID: sessionID,
		},
	})

	// Get stats
	stats, _ := mem.GetStats(ctx, sessionID)
	fmt.Printf("   ✅ Connected to PostgreSQL\n")
	fmt.Printf("   💾 Messages in session: %d\n", stats.SessionMessages)
	fmt.Printf("   🔍 Semantic search: %v\n", getEnvOrDefault("OPENAI_API_KEY", "") != "")
}

func hybridExample(dbURL, redisURL string) {
	mem, err := memory.NewWithConfig(memory.Config{
		Mode:                 memory.HYBRID,
		DatabaseURL:          dbURL,
		RedisAddr:            redisURL,
		OpenAIKey:            getEnvOrDefault("OPENAI_API_KEY", ""),
		EnableSemanticSearch: true,
		EnableAutoSummarize:  true,
		MaxSessionMessages:   20,
		SessionTTL:           time.Hour,
	})
	if err != nil {
		fmt.Printf("   ⚠️  Error: %v\n", err)
		return
	}
	defer mem.Close()

	ctx := context.Background()
	sessionID := "hybrid-demo"

	// Add a message
	mem.AddMessage(ctx, memory.Message{
		ID:        fmt.Sprintf("msg-%d", time.Now().UnixNano()),
		Role:      "user",
		Content:   "This message uses the best of both worlds: Redis speed + PostgreSQL persistence!",
		Timestamp: time.Now(),
		Metadata: memory.Metadata{
			SessionID: sessionID,
		},
	})

	fmt.Printf("   ✅ Connected to Redis + PostgreSQL\n")
	fmt.Printf("   ⚡ Cache performance: ~2-5ms\n")
	fmt.Printf("   💾 Persistent storage: ✅\n")
	fmt.Printf("   🔍 Semantic search: %v\n", getEnvOrDefault("OPENAI_API_KEY", "") != "")
	fmt.Printf("   🎯 Auto-summarization: ✅\n")
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

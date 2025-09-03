package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	memory "github.com/framehood/go-agent-memory"
)

// Example showing how to integrate memory with an AI agent
func main() {
	// Initialize memory from environment variables
	mem := initializeMemory()
	if mem == nil {
		fmt.Println("Running without memory support")
	} else {
		defer mem.Close()
		fmt.Println("Memory initialized successfully!")
	}

	// Simulate a conversation
	sessionID := "demo-session-001"
	userID := "demo-user"

	// Example 1: Add messages to memory
	if mem != nil {
		// User message
		err := mem.AddMessage(context.Background(), memory.Message{
			ID:      "msg-001",
			Role:    "user",
			Content: "Can you help me plan a trip to Tokyo?",
			Metadata: memory.Metadata{
				SessionID:  sessionID,
				UserID:     userID,
				TokenCount: 10,
			},
			Timestamp: time.Now(),
		})
		if err != nil {
			log.Printf("Error adding message: %v", err)
		}

		// Assistant response
		err = mem.AddMessage(context.Background(), memory.Message{
			ID:      "msg-002",
			Role:    "assistant",
			Content: "I'd be happy to help you plan a trip to Tokyo! Tokyo is an amazing destination. When are you planning to visit, and what are your main interests?",
			Metadata: memory.Metadata{
				SessionID:   sessionID,
				UserID:      userID,
				TokenCount:  25,
				Model:       "gpt-4",
				Temperature: 0.7,
			},
			Timestamp: time.Now(),
		})
		if err != nil {
			log.Printf("Error adding message: %v", err)
		}
	}

	// Example 2: Retrieve recent messages
	if mem != nil {
		fmt.Println("\n📚 Recent Messages:")
		messages, err := mem.GetRecentMessages(context.Background(), sessionID, 10)
		if err != nil {
			log.Printf("Error getting messages: %v", err)
		} else {
			for _, msg := range messages {
				fmt.Printf("  [%s] %s: %s\n", msg.Timestamp.Format("15:04:05"), msg.Role, msg.Content)
			}
		}
	}

	// Example 3: Semantic search
	if mem != nil {
		fmt.Println("\n🔍 Semantic Search for 'travel Japan':")
		results, err := mem.Search(context.Background(), "travel Japan", 5, 0.7)
		if err != nil {
			log.Printf("Error searching: %v", err)
		} else {
			for i, result := range results {
				fmt.Printf("  %d. [Score: %.2f] %s\n", i+1, result.Score, result.Message.Content[:min(100, len(result.Message.Content))])
			}
		}
	}

	// Example 4: Get statistics
	if mem != nil {
		fmt.Println("\n📊 Memory Statistics:")
		stats, err := mem.GetStats(context.Background(), sessionID)
		if err != nil {
			log.Printf("Error getting stats: %v", err)
		} else {
			fmt.Printf("  Total Messages: %d\n", stats.TotalMessages)
			fmt.Printf("  Session Messages: %d\n", stats.SessionMessages)
			fmt.Printf("  Total Tokens: %d\n", stats.TotalTokens)
		}
	}

	// Example 5: Generate summary (for long conversations)
	if mem != nil {
		fmt.Println("\n📝 Generating Summary:")
		summary, err := mem.Summarize(context.Background(), sessionID, 1000)
		if err != nil {
			log.Printf("Error generating summary: %v", err)
		} else if summary != "" {
			fmt.Printf("  Summary: %s\n", summary)
		}
	}
}

// initializeMemory creates a memory instance if environment variables are set
func initializeMemory() memory.Memory {
	// Check for required environment variables
	dbURL := os.Getenv("DATABASE_URL")
	openAIKey := os.Getenv("OPENAI_API_KEY")

	if dbURL == "" || openAIKey == "" {
		fmt.Println("⚠️  Memory disabled: Set DATABASE_URL and OPENAI_API_KEY to enable")
		return nil
	}

	// Create configuration
	config := memory.Config{
		DatabaseURL:    dbURL,
		OpenAIKey:      openAIKey,
		EmbeddingModel: "text-embedding-3-small",

		// Optional: Add Redis for faster session access
		RedisAddr:     os.Getenv("REDIS_URL"),
		RedisPassword: os.Getenv("REDIS_PASSWORD"),

		// Memory settings
		MaxSessionMessages: 50,
		SessionTTL:         24 * time.Hour,
		AutoSummarize:      true,
		VectorDimension:    1536,
	}

	// Initialize memory
	mem, err := memory.New(config)
	if err != nil {
		log.Printf("Failed to initialize memory: %v", err)
		return nil
	}

	return mem
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

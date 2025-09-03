// Session-Only Memory Example
// This example demonstrates the simplest memory configuration with no external dependencies.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	memory "github.com/framehood/go-agent-memory"
)

func main() {
	fmt.Println("🧠 Session-Only Memory Example")
	fmt.Println("==============================")
	fmt.Println()

	// Create session-only configuration
	// No database or Redis needed!
	config := memory.Config{
		// Core settings
		Mode:               memory.SESSION_ONLY,
		MaxSessionMessages: 20, // Keep only last 20 messages per session

		// All these are disabled for session-only mode
		EnablePersistence:    false,
		EnableSemanticSearch: false,
		EnableAutoSummarize:  false,

		// No external connections needed!
		// DatabaseURL: "", // Not needed
		// RedisAddr:   "", // Not needed
		// OpenAIKey:   "", // Not needed
	}

	// Initialize memory
	mem, err := memory.NewWithConfig(config)
	if err != nil {
		log.Fatalf("Failed to initialize memory: %v", err)
	}
	defer func() {
		if err := mem.Close(); err != nil {
			log.Printf("Error closing memory: %v", err)
		}
	}()

	fmt.Println("✅ Memory initialized (session-only mode)")
	fmt.Println("   - No database connection")
	fmt.Println("   - No Redis connection")
	fmt.Println("   - Pure in-memory storage")
	fmt.Println()

	// Create a session
	sessionID := "demo-session-001"
	ctx := context.Background()

	// Simulate a conversation
	fmt.Println("📝 Adding messages to session...")
	messages := []struct {
		role    string
		content string
	}{
		{"user", "Hello! Can you help me with Go programming?"},
		{"assistant", "Of course! I'd be happy to help with Go programming. What would you like to know?"},
		{"user", "How do I create a slice in Go?"},
		{"assistant", "You can create a slice in Go in several ways:\n1. Using make: `slice := make([]int, 5)`\n2. Using literal: `slice := []int{1, 2, 3}`\n3. From array: `slice := array[1:4]`"},
		{"user", "What's the difference between slice and array?"},
		{"assistant", "Great question! Key differences:\n- Arrays have fixed size, slices are dynamic\n- Arrays are values, slices are references\n- Slices have capacity and length\n- Slices are more commonly used in Go"},
		{"user", "Can you show me an example?"},
		{"assistant", "```go\n// Array (fixed size)\nvar arr [3]int = [3]int{1, 2, 3}\n\n// Slice (dynamic)\nvar slice []int = []int{1, 2, 3}\nslice = append(slice, 4) // Can grow\n```"},
	}

	// Add messages to memory
	for i, msg := range messages {
		if err := mem.AddMessage(ctx, memory.Message{
			ID:        fmt.Sprintf("msg-%d", i+1),
			Role:      msg.role,
			Content:   msg.content,
			Timestamp: time.Now().Add(time.Duration(i) * time.Second),
			Metadata: memory.Metadata{
				SessionID: sessionID,
			},
		}); err != nil {
			log.Printf("Error adding message: %v", err)
		}

		// Show progress
		fmt.Printf("   Added message %d: [%s] %.50s...\n", i+1, msg.role, msg.content)
		time.Sleep(100 * time.Millisecond) // Simulate typing delay
	}

	fmt.Println()
	fmt.Println("✅ Added", len(messages), "messages to session")
	fmt.Println()

	// Retrieve recent messages
	fmt.Println("🔍 Retrieving recent messages...")
	recent, err := mem.GetRecentMessages(ctx, sessionID, 5)
	if err != nil {
		log.Printf("Error retrieving messages: %v", err)
	}

	fmt.Printf("   Found %d recent messages (requested last 5):\n", len(recent))
	for i, msg := range recent {
		preview := msg.Content
		if len(preview) > 60 {
			preview = preview[:57] + "..."
		}
		fmt.Printf("   %d. [%s] %s\n", i+1, msg.Role, preview)
	}
	fmt.Println()

	// Get session statistics
	fmt.Println("📊 Session Statistics:")
	stats, err := mem.GetStats(ctx, sessionID)
	if err != nil {
		log.Printf("Error getting stats: %v", err)
	} else {
		fmt.Printf("   - Total messages: %d\n", stats.SessionMessages)
		fmt.Printf("   - Session started: %s\n", stats.OldestMessage.Format("15:04:05"))
		fmt.Printf("   - Last activity: %s\n", stats.LatestMessage.Format("15:04:05"))
		fmt.Printf("   - Memory usage: ~%d KB\n", stats.StorageSize/1024)
	}
	fmt.Println()

	// Demonstrate memory limits
	fmt.Println("🔄 Testing memory limits (max 20 messages)...")

	// Add more messages to exceed limit
	for i := len(messages); i < 25; i++ {
		mem.AddMessage(ctx, memory.Message{
			ID:        fmt.Sprintf("msg-%d", i+1),
			Role:      "user",
			Content:   fmt.Sprintf("Additional message %d", i+1),
			Timestamp: time.Now(),
			Metadata: memory.Metadata{
				SessionID: sessionID,
			},
		})
	}

	// Check that old messages were removed
	allMessages, _ := mem.GetRecentMessages(ctx, sessionID, 100)
	fmt.Printf("   Added 25 total messages, but only %d kept (due to limit)\n", len(allMessages))
	fmt.Printf("   Oldest message is now: msg-%d\n", 25-len(allMessages)+1)
	fmt.Println()

	// Demonstrate multiple sessions
	fmt.Println("👥 Testing multiple sessions...")

	session2ID := "demo-session-002"
	mem.AddMessage(ctx, memory.Message{
		ID:        "session2-msg1",
		Role:      "user",
		Content:   "This is a different session",
		Timestamp: time.Now(),
		Metadata: memory.Metadata{
			SessionID: session2ID,
		},
	})

	session1Messages, _ := mem.GetRecentMessages(ctx, sessionID, 100)
	session2Messages, _ := mem.GetRecentMessages(ctx, session2ID, 100)

	fmt.Printf("   Session 1 messages: %d\n", len(session1Messages))
	fmt.Printf("   Session 2 messages: %d\n", len(session2Messages))
	fmt.Println("   ✅ Sessions are properly isolated")
	fmt.Println()

	// Clear a session
	fmt.Println("🗑️  Clearing session 2...")
	err = mem.ClearSession(ctx, session2ID)
	if err != nil {
		log.Printf("Error clearing session: %v", err)
	}

	session2Messages, _ = mem.GetRecentMessages(ctx, session2ID, 100)
	fmt.Printf("   Session 2 messages after clear: %d\n", len(session2Messages))
	fmt.Println()

	// Important notes
	fmt.Println("⚠️  Important Notes:")
	fmt.Println("   - All data is in-memory only")
	fmt.Println("   - Data will be lost when program exits")
	fmt.Println("   - No persistence between runs")
	fmt.Println("   - Perfect for development and testing!")
	fmt.Println()

	fmt.Println("✨ Example completed successfully!")
	fmt.Println()
	fmt.Println("📚 Next steps:")
	fmt.Println("   - Try example 02 for persistent storage")
	fmt.Println("   - Try example 03 for hybrid mode with caching")
	fmt.Println("   - See example 07 for full agent integration")
}



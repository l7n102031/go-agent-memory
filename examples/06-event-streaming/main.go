// Event Streaming Memory Example
// Demonstrates Redis Streams for event sourcing and audit trails.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
	
	"github.com/redis/go-redis/v9"
	memory "github.com/framehood/go-agent-memory"
)

// Event types
const (
	EventMessageAdded    = "message.added"
	EventSessionStarted  = "session.started"
	EventSessionEnded    = "session.ended"
	EventSummaryCreated  = "summary.created"
	EventSearchPerformed = "search.performed"
	EventMemoryCleared   = "memory.cleared"
)

// Event represents a memory system event
type Event struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	SessionID string                 `json:"session_id"`
	UserID    string                 `json:"user_id,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// EventStreamMemory wraps memory with event streaming
type EventStreamMemory struct {
	memory.Memory
	redis      *redis.Client
	streamName string
}

func main() {
	fmt.Println("📡 Event Streaming Memory Example")
	fmt.Println("==================================")
	fmt.Println()
	
	// Initialize Redis for event streaming
	redisAddr := os.Getenv("REDIS_URL")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	
	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})
	
	// Test Redis connection
	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	
	fmt.Println("✅ Connected to Redis for event streaming")
	fmt.Printf("   Stream server: %s\n", redisAddr)
	fmt.Println()
	
	// Create event-enabled memory
	esm, err := createEventStreamMemory(redisClient)
	if err != nil {
		log.Fatalf("Failed to create memory: %v", err)
	}
	defer esm.Close()
	
	// Demonstrate event streaming
	demonstrateEventStreaming(ctx, esm, redisClient)
	
	// Show event replay capability
	demonstrateEventReplay(ctx, redisClient)
	
	// Demonstrate consumer groups
	demonstrateConsumerGroups(ctx, redisClient)
	
	// Show event analytics
	showEventAnalytics(ctx, redisClient)
	
	fmt.Println("✨ Example completed successfully!")
	fmt.Println()
	fmt.Println("📚 Next steps:")
	fmt.Println("   - See example 07 for complete agent integration")
}

func createEventStreamMemory(redisClient *redis.Client) (*EventStreamMemory, error) {
	// Initialize base memory (could be any type)
	config := memory.Config{
		Mode:               memory.SESSION_ONLY, // Simple for demo
		MaxSessionMessages: 50,
	}
	
	baseMem, err := memory.NewWithConfig(config)
	if err != nil {
		return nil, err
	}
	
	return &EventStreamMemory{
		Memory:     baseMem,
		redis:      redisClient,
		streamName: "memory:events",
	}, nil
}

func demonstrateEventStreaming(ctx context.Context, esm *EventStreamMemory, redisClient *redis.Client) {
	fmt.Println("🎬 Demonstrating Event Streaming")
	fmt.Println("═════════════════════════════════")
	fmt.Println()
	
	sessionID := "event-demo-session"
	userID := "user-123"
	
	// Start event consumer in background
	go consumeEvents(ctx, redisClient, esm.streamName)
	
	// Generate various events
	fmt.Println("📝 Generating events...")
	fmt.Println()
	
	// Session started
	publishEvent(ctx, esm, Event{
		Type:      EventSessionStarted,
		SessionID: sessionID,
		UserID:    userID,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"ip":         "192.168.1.1",
			"user_agent": "Mozilla/5.0",
		},
	})
	
	// Messages added
	messages := []struct {
		role    string
		content string
	}{
		{"user", "Hello, I need help with Redis Streams"},
		{"assistant", "I'll help you understand Redis Streams for event sourcing."},
		{"user", "How do I implement consumer groups?"},
		{"assistant", "Consumer groups allow multiple consumers to process events..."},
	}
	
	for i, msg := range messages {
		// Add to memory
		esm.AddMessage(ctx, memory.Message{
			ID:        fmt.Sprintf("msg-%d", i),
			Role:      msg.role,
			Content:   msg.content,
			Timestamp: time.Now(),
			Metadata: memory.Metadata{
				SessionID: sessionID,
				UserID:    userID,
			},
		})
		
		// Publish event
		publishEvent(ctx, esm, Event{
			Type:      EventMessageAdded,
			SessionID: sessionID,
			UserID:    userID,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"message_id": fmt.Sprintf("msg-%d", i),
				"role":       msg.role,
				"length":     len(msg.content),
			},
		})
		
		time.Sleep(100 * time.Millisecond)
	}
	
	// Search performed
	publishEvent(ctx, esm, Event{
		Type:      EventSearchPerformed,
		SessionID: sessionID,
		UserID:    userID,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"query":         "consumer groups",
			"results_count": 3,
			"threshold":     0.75,
		},
	})
	
	// Summary created
	publishEvent(ctx, esm, Event{
		Type:      EventSummaryCreated,
		SessionID: sessionID,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"original_messages": 50,
			"summary_tokens":    200,
			"compression_ratio": 5.5,
		},
	})
	
	// Session ended
	publishEvent(ctx, esm, Event{
		Type:      EventSessionEnded,
		SessionID: sessionID,
		UserID:    userID,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"duration_minutes": 15,
			"total_messages":   4,
		},
	})
	
	// Give consumer time to process
	time.Sleep(500 * time.Millisecond)
	fmt.Println()
}

func publishEvent(ctx context.Context, esm *EventStreamMemory, event Event) {
	event.ID = fmt.Sprintf("evt-%d", time.Now().UnixNano())
	
	// Convert event to map
	eventData := map[string]interface{}{
		"id":         event.ID,
		"type":       event.Type,
		"session_id": event.SessionID,
		"user_id":    event.UserID,
		"timestamp":  event.Timestamp.Unix(),
	}
	
	// Add custom data
	for k, v := range event.Data {
		eventData[k] = v
	}
	
	// Publish to stream
	_, err := esm.redis.XAdd(ctx, &redis.XAddArgs{
		Stream: esm.streamName,
		Values: eventData,
	}).Result()
	
	if err != nil {
		log.Printf("Failed to publish event: %v", err)
	}
}

func consumeEvents(ctx context.Context, redisClient *redis.Client, streamName string) {
	fmt.Println("👂 Event Consumer Started")
	fmt.Println("─────────────────────────")
	
	lastID := "0"
	
	for {
		// Read from stream
		results, err := redisClient.XRead(ctx, &redis.XReadArgs{
			Streams: []string{streamName, lastID},
			Count:   10,
			Block:   1 * time.Second,
		}).Result()
		
		if err != nil {
			if err != redis.Nil {
				log.Printf("Error reading stream: %v", err)
			}
			continue
		}
		
		// Process events
		for _, result := range results {
			for _, message := range result.Messages {
				processEvent(message.ID, message.Values)
				lastID = message.ID
			}
		}
	}
}

func processEvent(id string, data map[string]interface{}) {
	eventType, _ := data["type"].(string)
	sessionID, _ := data["session_id"].(string)
	
	// Format event for display
	switch eventType {
	case EventSessionStarted:
		fmt.Printf("🚀 [%s] Session started: %s\n", id, sessionID)
		
	case EventMessageAdded:
		role, _ := data["role"].(string)
		messageID, _ := data["message_id"].(string)
		fmt.Printf("💬 [%s] Message added: %s (%s)\n", id, messageID, role)
		
	case EventSearchPerformed:
		query, _ := data["query"].(string)
		count := data["results_count"]
		fmt.Printf("🔍 [%s] Search performed: \"%s\" (%v results)\n", id, query, count)
		
	case EventSummaryCreated:
		ratio := data["compression_ratio"]
		fmt.Printf("📝 [%s] Summary created: %.1fx compression\n", id, ratio)
		
	case EventSessionEnded:
		duration := data["duration_minutes"]
		fmt.Printf("🏁 [%s] Session ended: %v minutes\n", id, duration)
		
	default:
		fmt.Printf("📌 [%s] Event: %s\n", id, eventType)
	}
}

func demonstrateEventReplay(ctx context.Context, redisClient *redis.Client) {
	fmt.Println("\n🔄 Event Replay Capability")
	fmt.Println("═══════════════════════════")
	fmt.Println()
	
	streamName := "memory:events"
	
	// Get all events from beginning
	fmt.Println("Replaying all events from stream...")
	
	events, err := redisClient.XRange(ctx, streamName, "-", "+").Result()
	if err != nil {
		log.Printf("Error reading events: %v", err)
		return
	}
	
	fmt.Printf("Found %d events in stream\n\n", len(events))
	
	// Group events by session
	sessionEvents := make(map[string][]redis.XMessage)
	for _, event := range events {
		if sessionID, ok := event.Values["session_id"].(string); ok {
			sessionEvents[sessionID] = append(sessionEvents[sessionID], event)
		}
	}
	
	// Show session timeline
	for sessionID, events := range sessionEvents {
		fmt.Printf("📅 Session: %s\n", sessionID)
		fmt.Printf("   Events: %d\n", len(events))
		
		if len(events) > 0 {
			first := events[0]
			last := events[len(events)-1]
			
			firstType, _ := first.Values["type"].(string)
			lastType, _ := last.Values["type"].(string)
			
			fmt.Printf("   First: %s (%s)\n", firstType, first.ID)
			fmt.Printf("   Last:  %s (%s)\n", lastType, last.ID)
		}
		fmt.Println()
	}
	
	// Demonstrate time-based replay
	fmt.Println("⏰ Time-based Event Replay")
	fmt.Println("──────────────────────────")
	
	// Get events from last minute
	oneMinuteAgo := time.Now().Add(-1 * time.Minute).UnixMilli()
	recentEvents, _ := redisClient.XRange(ctx, streamName, 
		fmt.Sprintf("%d", oneMinuteAgo), "+").Result()
	
	fmt.Printf("Events in last minute: %d\n", len(recentEvents))
	fmt.Println()
}

func demonstrateConsumerGroups(ctx context.Context, redisClient *redis.Client) {
	fmt.Println("👥 Consumer Groups Demo")
	fmt.Println("════════════════════════")
	fmt.Println()
	
	streamName := "memory:events"
	groupName := "analytics-group"
	
	// Create consumer group
	err := redisClient.XGroupCreateMkStream(ctx, streamName, groupName, "0").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		log.Printf("Error creating group: %v", err)
	}
	
	fmt.Printf("✅ Consumer group '%s' ready\n", groupName)
	
	// Simulate multiple consumers
	consumers := []string{"analytics-1", "analytics-2", "audit-1"}
	
	for _, consumer := range consumers {
		// Read pending messages
		messages, err := redisClient.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    groupName,
			Consumer: consumer,
			Streams:  []string{streamName, ">"},
			Count:    2,
			NoAck:    false,
		}).Result()
		
		if err != nil && err != redis.Nil {
			log.Printf("Error reading for %s: %v", consumer, err)
			continue
		}
		
		messageCount := 0
		for _, result := range messages {
			messageCount += len(result.Messages)
		}
		
		fmt.Printf("   Consumer %s: processed %d messages\n", consumer, messageCount)
	}
	
	// Show pending messages
	pending, _ := redisClient.XPending(ctx, streamName, groupName).Result()
	fmt.Printf("\n📊 Group Statistics:\n")
	fmt.Printf("   Pending messages: %d\n", pending.Count)
	if pending.Count > 0 {
		fmt.Printf("   Oldest pending: %s\n", pending.Lower)
		fmt.Printf("   Newest pending: %s\n", pending.Higher)
	}
	fmt.Println()
}

func showEventAnalytics(ctx context.Context, redisClient *redis.Client) {
	fmt.Println("📊 Event Analytics")
	fmt.Println("═══════════════════")
	fmt.Println()
	
	streamName := "memory:events"
	
	// Get stream info
	info, err := redisClient.XInfoStream(ctx, streamName).Result()
	if err != nil {
		log.Printf("Error getting stream info: %v", err)
		return
	}
	
	fmt.Printf("Stream Statistics:\n")
	fmt.Printf("   Total events: %d\n", info.Length)
	fmt.Printf("   First entry: %s\n", info.FirstEntry.ID)
	fmt.Printf("   Last entry: %s\n", info.LastEntry.ID)
	fmt.Println()
	
	// Analyze event types
	events, _ := redisClient.XRange(ctx, streamName, "-", "+").Result()
	
	eventCounts := make(map[string]int)
	sessionCounts := make(map[string]int)
	
	for _, event := range events {
		if eventType, ok := event.Values["type"].(string); ok {
			eventCounts[eventType]++
		}
		if sessionID, ok := event.Values["session_id"].(string); ok {
			sessionCounts[sessionID]++
		}
	}
	
	fmt.Println("Event Type Distribution:")
	for eventType, count := range eventCounts {
		percentage := float64(count) / float64(len(events)) * 100
		fmt.Printf("   %s: %d (%.1f%%)\n", eventType, count, percentage)
	}
	fmt.Println()
	
	fmt.Printf("Unique sessions: %d\n", len(sessionCounts))
	fmt.Printf("Average events per session: %.1f\n", 
		float64(len(events))/float64(len(sessionCounts)))
	fmt.Println()
	
	// Memory usage
	memUsage, _ := redisClient.MemoryUsage(ctx, streamName).Result()
	fmt.Printf("Stream memory usage: %d bytes (%.2f KB)\n", 
		memUsage, float64(memUsage)/1024)
	fmt.Println()
}

package memory_test

import (
	"context"
	"fmt"
	"testing"
	"time"
	
	memory "github.com/framehood/go-agent-memory"
)

// MockMemory implements Memory interface for testing
type MockMemory struct {
	messages []memory.Message
}

func NewMockMemory() memory.Memory {
	return &MockMemory{
		messages: make([]memory.Message, 0),
	}
}

func (m *MockMemory) AddMessage(ctx context.Context, msg memory.Message) error {
	m.messages = append(m.messages, msg)
	return nil
}

func (m *MockMemory) GetRecentMessages(ctx context.Context, sessionID string, limit int) ([]memory.Message, error) {
	var result []memory.Message
	for _, msg := range m.messages {
		if msg.Metadata.SessionID == sessionID {
			result = append(result, msg)
			if len(result) >= limit {
				break
			}
		}
	}
	return result, nil
}

func (m *MockMemory) ClearSession(ctx context.Context, sessionID string) error {
	var filtered []memory.Message
	for _, msg := range m.messages {
		if msg.Metadata.SessionID != sessionID {
			filtered = append(filtered, msg)
		}
	}
	m.messages = filtered
	return nil
}

func (m *MockMemory) Store(ctx context.Context, msg memory.Message) error {
	return m.AddMessage(ctx, msg)
}

func (m *MockMemory) Search(ctx context.Context, query string, limit int, threshold float32) ([]memory.SearchResult, error) {
	// Simple mock search - just return first messages
	var results []memory.SearchResult
	for i, msg := range m.messages {
		if i >= limit {
			break
		}
		results = append(results, memory.SearchResult{
			Message:  msg,
			Score:    0.9,
			Distance: 0.1,
		})
	}
	return results, nil
}

func (m *MockMemory) SearchWithEmbedding(ctx context.Context, embedding []float32, limit int, threshold float32) ([]memory.SearchResult, error) {
	return m.Search(ctx, "", limit, threshold)
}

func (m *MockMemory) Summarize(ctx context.Context, sessionID string, maxTokens int) (string, error) {
	messages, _ := m.GetRecentMessages(ctx, sessionID, 100) // Get up to 100 messages
	if len(messages) == 0 {
		return "", nil
	}
	return "Mock summary of conversation", nil
}

func (m *MockMemory) GetStats(ctx context.Context, sessionID string) (*memory.Stats, error) {
	count := 0
	for _, msg := range m.messages {
		if msg.Metadata.SessionID == sessionID {
			count++
		}
	}
	
	return &memory.Stats{
		SessionID:       sessionID,
		TotalMessages:   len(m.messages),
		SessionMessages: count,
		OldestMessage:   time.Now().Add(-24 * time.Hour),
		LatestMessage:   time.Now(),
	}, nil
}

func (m *MockMemory) Close() error {
	return nil
}

// Tests

func TestMockMemory(t *testing.T) {
	mem := NewMockMemory()
	ctx := context.Background()
	sessionID := "test-session"
	
	// Test adding a message
	msg := memory.Message{
		ID:      "msg-1",
		Role:    "user",
		Content: "Hello, world!",
		Metadata: memory.Metadata{
			SessionID: sessionID,
		},
		Timestamp: time.Now(),
	}
	
	err := mem.AddMessage(ctx, msg)
	if err != nil {
		t.Fatalf("Failed to add message: %v", err)
	}
	
	// Test retrieving messages
	messages, err := mem.GetRecentMessages(ctx, sessionID, 10)
	if err != nil {
		t.Fatalf("Failed to get messages: %v", err)
	}
	
	if len(messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(messages))
	}
	
	if messages[0].Content != "Hello, world!" {
		t.Errorf("Expected content 'Hello, world!', got %s", messages[0].Content)
	}
	
	// Test stats
	stats, err := mem.GetStats(ctx, sessionID)
	if err != nil {
		t.Fatalf("Failed to get stats: %v", err)
	}
	
	if stats.TotalMessages != 1 {
		t.Errorf("Expected 1 total message, got %d", stats.TotalMessages)
	}
	
	// Test search
	results, err := mem.Search(ctx, "hello", 5, 0.7)
	if err != nil {
		t.Fatalf("Failed to search: %v", err)
	}
	
	if len(results) != 1 {
		t.Errorf("Expected 1 search result, got %d", len(results))
	}
	
	// Test clear session
	err = mem.ClearSession(ctx, sessionID)
	if err != nil {
		t.Fatalf("Failed to clear session: %v", err)
	}
	
	messages, _ = mem.GetRecentMessages(ctx, sessionID, 10)
	if len(messages) != 0 {
		t.Errorf("Expected 0 messages after clear, got %d", len(messages))
	}
}

func TestMessageMetadata(t *testing.T) {
	metadata := memory.Metadata{
		SessionID:   "session-123",
		UserID:      "user-456",
		TokenCount:  100,
		Model:       "gpt-4",
		Temperature: 0.7,
		Extra: map[string]interface{}{
			"custom_field": "value",
		},
	}
	
	if metadata.SessionID != "session-123" {
		t.Errorf("Expected session ID 'session-123', got %s", metadata.SessionID)
	}
	
	if metadata.Extra["custom_field"] != "value" {
		t.Errorf("Expected custom field 'value', got %v", metadata.Extra["custom_field"])
	}
}

func BenchmarkAddMessage(b *testing.B) {
	mem := NewMockMemory()
	ctx := context.Background()
	
	msg := memory.Message{
		ID:      "msg-bench",
		Role:    "user",
		Content: "Benchmark message",
		Metadata: memory.Metadata{
			SessionID: "bench-session",
		},
		Timestamp: time.Now(),
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mem.AddMessage(ctx, msg)
	}
}

func BenchmarkGetRecentMessages(b *testing.B) {
	mem := NewMockMemory()
	ctx := context.Background()
	sessionID := "bench-session"
	
	// Add some messages
	for i := 0; i < 100; i++ {
		mem.AddMessage(ctx, memory.Message{
			ID:      fmt.Sprintf("msg-%d", i),
			Role:    "user",
			Content: "Test message",
			Metadata: memory.Metadata{
				SessionID: sessionID,
			},
		})
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mem.GetRecentMessages(ctx, sessionID, 10)
	}
}
package memory

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"
)

// SessionOnlyMemory implements in-memory session storage with no persistence
type SessionOnlyMemory struct {
	config   Config
	sessions map[string][]Message
	mutex    sync.RWMutex
	stats    map[string]*Stats
}

// NewSessionOnlyMemory creates a new session-only memory instance
func NewSessionOnlyMemory(cfg Config) (Memory, error) {
	return &SessionOnlyMemory{
		config:   cfg,
		sessions: make(map[string][]Message),
		stats:    make(map[string]*Stats),
	}, nil
}

// AddMessage adds a message to the session
func (sm *SessionOnlyMemory) AddMessage(ctx context.Context, msg Message) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	
	sessionID := msg.Metadata.SessionID
	if sessionID == "" {
		return fmt.Errorf("session ID is required")
	}
	
	// Add message to session
	sm.sessions[sessionID] = append(sm.sessions[sessionID], msg)
	
	// Enforce message limit
	if len(sm.sessions[sessionID]) > sm.config.MaxSessionMessages {
		// Keep only the most recent messages
		start := len(sm.sessions[sessionID]) - sm.config.MaxSessionMessages
		sm.sessions[sessionID] = sm.sessions[sessionID][start:]
	}
	
	// Update stats
	sm.updateStats(sessionID)
	
	return nil
}

// GetRecentMessages retrieves recent messages from the session
func (sm *SessionOnlyMemory) GetRecentMessages(ctx context.Context, sessionID string, limit int) ([]Message, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	
	messages, exists := sm.sessions[sessionID]
	if !exists {
		return []Message{}, nil
	}
	
	// Return the most recent messages up to limit
	start := 0
	if len(messages) > limit {
		start = len(messages) - limit
	}
	
	result := make([]Message, len(messages)-start)
	copy(result, messages[start:])
	
	return result, nil
}

// ClearSession removes all messages from a session
func (sm *SessionOnlyMemory) ClearSession(ctx context.Context, sessionID string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	
	delete(sm.sessions, sessionID)
	delete(sm.stats, sessionID)
	
	return nil
}

// Store is a no-op for session-only memory (same as AddMessage)
func (sm *SessionOnlyMemory) Store(ctx context.Context, msg Message) error {
	return sm.AddMessage(ctx, msg)
}

// Search performs basic text search (no semantic search in session-only mode)
func (sm *SessionOnlyMemory) Search(ctx context.Context, query string, limit int, threshold float32) ([]SearchResult, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	
	var results []SearchResult
	
	// Simple text matching across all sessions
	for _, messages := range sm.sessions {
		for _, msg := range messages {
			if contains(msg.Content, query) {
				// Simple relevance scoring based on exact matches
				score := calculateSimpleScore(msg.Content, query)
				if score >= threshold {
					results = append(results, SearchResult{
						Message: msg,
						Score:   score,
					})
				}
			}
		}
	}
	
	// Sort by score descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})
	
	// Limit results
	if len(results) > limit {
		results = results[:limit]
	}
	
	return results, nil
}

// SearchWithEmbedding is not supported in session-only mode
func (sm *SessionOnlyMemory) SearchWithEmbedding(ctx context.Context, embedding []float32, limit int, threshold float32) ([]SearchResult, error) {
	return nil, fmt.Errorf("semantic search not supported in session-only mode")
}

// Summarize creates a simple summary (no LLM in session-only mode)
func (sm *SessionOnlyMemory) Summarize(ctx context.Context, sessionID string, maxTokens int) (string, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	
	messages, exists := sm.sessions[sessionID]
	if !exists || len(messages) == 0 {
		return "", fmt.Errorf("no messages found for session %s", sessionID)
	}
	
	// Simple summary: count messages by role
	userMessages := 0
	assistantMessages := 0
	
	for _, msg := range messages {
		switch msg.Role {
		case "user":
			userMessages++
		case "assistant":
			assistantMessages++
		}
	}
	
	summary := fmt.Sprintf("Session contains %d messages: %d from user, %d from assistant", 
		len(messages), userMessages, assistantMessages)
	
	return summary, nil
}

// GetSummary returns a basic summary structure
func (sm *SessionOnlyMemory) GetSummary(ctx context.Context, sessionID string) (*Summary, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	
	messages, exists := sm.sessions[sessionID]
	if !exists || len(messages) == 0 {
		return nil, fmt.Errorf("no messages found for session %s", sessionID)
	}
	
	content, _ := sm.Summarize(ctx, sessionID, 0)
	
	var startTime, endTime time.Time
	if len(messages) > 0 {
		startTime = messages[0].Timestamp
		endTime = messages[len(messages)-1].Timestamp
	}
	
	return &Summary{
		SessionID:    sessionID,
		Content:      content,
		TokenCount:   len(content) / 4, // Rough estimation
		MessageCount: len(messages),
		StartTime:    startTime,
		EndTime:      endTime,
		Created:      time.Now(),
	}, nil
}

// GetStats returns session statistics
func (sm *SessionOnlyMemory) GetStats(ctx context.Context, sessionID string) (*Stats, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	
	if stats, exists := sm.stats[sessionID]; exists {
		return stats, nil
	}
	
	// Generate stats if not cached
	sm.updateStats(sessionID)
	return sm.stats[sessionID], nil
}

// Close cleans up resources (no-op for session-only)
func (sm *SessionOnlyMemory) Close() error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	
	// Clear all data
	sm.sessions = make(map[string][]Message)
	sm.stats = make(map[string]*Stats)
	
	return nil
}

// Helper methods

func (sm *SessionOnlyMemory) updateStats(sessionID string) {
	messages, exists := sm.sessions[sessionID]
	if !exists {
		return
	}
	
	stats := &Stats{
		SessionID:       sessionID,
		TotalMessages:   len(messages),
		SessionMessages: len(messages),
		HasSummary:      false, // No real summarization in session-only
	}
	
	// Calculate total tokens (rough estimation)
	totalTokens := 0
	for _, msg := range messages {
		totalTokens += len(msg.Content) / 4 // 4 chars per token approximation
	}
	stats.TotalTokens = totalTokens
	stats.ActiveTokens = totalTokens
	
	// Set time bounds
	if len(messages) > 0 {
		oldest := messages[0].Timestamp
		latest := messages[len(messages)-1].Timestamp
		stats.OldestMessage = &oldest
		stats.LatestMessage = &latest
	}
	
	// Calculate storage size (rough estimation)
	storageSize := int64(0)
	for _, msg := range messages {
		storageSize += int64(len(msg.Content) + len(msg.ID) + len(msg.Role))
	}
	stats.StorageSize = storageSize
	
	sm.stats[sessionID] = stats
}

func contains(text, query string) bool {
	return len(text) > 0 && len(query) > 0 && 
		(text == query || len(text) > len(query))
}

func calculateSimpleScore(content, query string) float32 {
	// Simple scoring: exact match = 1.0, contains = 0.8
	if content == query {
		return 1.0
	}
	if contains(content, query) {
		return 0.8
	}
	return 0.0
}

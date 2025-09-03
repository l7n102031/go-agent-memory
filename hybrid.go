package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// HybridMemory combines Redis for fast session memory and Supabase for semantic search
type HybridMemory struct {
	supabase           *SupabaseMemory
	redis              *redis.Client
	maxSessionMessages int
	sessionTTL         time.Duration
}

// NewHybridMemory creates a memory system with both Redis and Supabase
func NewHybridMemory(cfg Config) (Memory, error) {
	// Initialize Supabase memory
	supabaseMem, err := NewSupabaseMemory(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Supabase memory: %w", err)
	}

	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	// Test Redis connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	// Set defaults
	maxMessages := cfg.MaxSessionMessages
	if maxMessages == 0 {
		maxMessages = 50
	}

	sessionTTL := cfg.SessionTTL
	if sessionTTL == 0 {
		sessionTTL = 24 * time.Hour
	}

	return &HybridMemory{
		supabase:           supabaseMem.(*SupabaseMemory),
		redis:              redisClient,
		maxSessionMessages: maxMessages,
		sessionTTL:         sessionTTL,
	}, nil
}

// AddMessage adds a message to both Redis (for fast access) and Supabase (for persistence)
func (hm *HybridMemory) AddMessage(ctx context.Context, msg Message) error {
	// Add to Supabase for persistence and semantic search
	if err := hm.supabase.AddMessage(ctx, msg); err != nil {
		// Log but don't fail if Supabase write fails
		fmt.Printf("Warning: failed to persist message to Supabase: %v\n", err)
	}

	// Add to Redis for fast session access
	sessionKey := fmt.Sprintf("session:%s:messages", msg.Metadata.SessionID)

	// Serialize message
	msgJSON, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to serialize message: %w", err)
	}

	// Add to Redis list (LPUSH for newest first)
	if err := hm.redis.LPush(ctx, sessionKey, msgJSON).Err(); err != nil {
		return fmt.Errorf("failed to add message to Redis: %w", err)
	}

	// Trim list to max size
	if err := hm.redis.LTrim(ctx, sessionKey, 0, int64(hm.maxSessionMessages-1)).Err(); err != nil {
		fmt.Printf("Warning: failed to trim message list: %v\n", err)
	}

	// Set TTL on the key
	if err := hm.redis.Expire(ctx, sessionKey, hm.sessionTTL).Err(); err != nil {
		fmt.Printf("Warning: failed to set TTL: %v\n", err)
	}

	// Update session metadata
	metaKey := fmt.Sprintf("session:%s:meta", msg.Metadata.SessionID)
	hm.redis.HSet(ctx, metaKey,
		"last_message_time", msg.Timestamp.Unix(),
		"last_message_role", msg.Role,
		"user_id", msg.Metadata.UserID,
		"model", msg.Metadata.Model,
	)
	hm.redis.HIncrBy(ctx, metaKey, "message_count", 1)
	hm.redis.HIncrBy(ctx, metaKey, "total_tokens", int64(msg.Metadata.TokenCount))
	hm.redis.Expire(ctx, metaKey, hm.sessionTTL)

	return nil
}

// GetRecentMessages retrieves recent messages from Redis first, falls back to Supabase
func (hm *HybridMemory) GetRecentMessages(ctx context.Context, sessionID string, limit int) ([]Message, error) {
	if limit <= 0 {
		limit = hm.maxSessionMessages
	}

	sessionKey := fmt.Sprintf("session:%s:messages", sessionID)

	// Try to get from Redis first
	results, err := hm.redis.LRange(ctx, sessionKey, 0, int64(limit-1)).Result()
	if err == nil && len(results) > 0 {
		messages := make([]Message, 0, len(results))
		for i := len(results) - 1; i >= 0; i-- { // Reverse for chronological order
			var msg Message
			if err := json.Unmarshal([]byte(results[i]), &msg); err != nil {
				fmt.Printf("Warning: failed to deserialize message: %v\n", err)
				continue
			}
			messages = append(messages, msg)
		}

		if len(messages) > 0 {
			return messages, nil
		}
	}

	// Fallback to Supabase if Redis doesn't have the data
	messages, err := hm.supabase.GetRecentMessages(ctx, sessionID, limit)
	if err != nil {
		return nil, err
	}

	// Optionally repopulate Redis cache
	go hm.repopulateRedisCache(context.Background(), sessionID, messages)

	return messages, nil
}

// repopulateRedisCache asynchronously repopulates Redis cache from Supabase
func (hm *HybridMemory) repopulateRedisCache(ctx context.Context, sessionID string, messages []Message) {
	sessionKey := fmt.Sprintf("session:%s:messages", sessionID)

	// Clear existing cache
	hm.redis.Del(ctx, sessionKey)

	// Add messages in reverse order (newest first in Redis)
	for i := len(messages) - 1; i >= 0; i-- {
		msgJSON, err := json.Marshal(messages[i])
		if err != nil {
			continue
		}
		hm.redis.LPush(ctx, sessionKey, msgJSON)
	}

	// Set TTL
	hm.redis.Expire(ctx, sessionKey, hm.sessionTTL)
}

// ClearSession clears messages from both Redis and Supabase
func (hm *HybridMemory) ClearSession(ctx context.Context, sessionID string) error {
	// Clear from Redis
	sessionKey := fmt.Sprintf("session:%s:messages", sessionID)
	metaKey := fmt.Sprintf("session:%s:meta", sessionID)

	hm.redis.Del(ctx, sessionKey, metaKey)

	// Clear from Supabase
	return hm.supabase.ClearSession(ctx, sessionID)
}

// Store saves a message for long-term memory (delegates to Supabase)
func (hm *HybridMemory) Store(ctx context.Context, msg Message) error {
	return hm.supabase.Store(ctx, msg)
}

// Search performs semantic search (delegates to Supabase)
func (hm *HybridMemory) Search(ctx context.Context, query string, limit int, threshold float32) ([]SearchResult, error) {
	return hm.supabase.Search(ctx, query, limit, threshold)
}

// SearchWithEmbedding searches using a pre-computed embedding (delegates to Supabase)
func (hm *HybridMemory) SearchWithEmbedding(ctx context.Context, embedding []float32, limit int, threshold float32) ([]SearchResult, error) {
	return hm.supabase.SearchWithEmbedding(ctx, embedding, limit, threshold)
}

// Summarize creates a summary of session messages
func (hm *HybridMemory) Summarize(ctx context.Context, sessionID string, maxTokens int) (string, error) {
	// Check if we have a recent summary in Redis
	summaryKey := fmt.Sprintf("session:%s:summary", sessionID)
	summary, err := hm.redis.Get(ctx, summaryKey).Result()
	if err == nil && summary != "" {
		return summary, nil
	}

	// Generate new summary
	summary, err = hm.supabase.Summarize(ctx, sessionID, maxTokens)
	if err != nil {
		return "", err
	}

	// Cache summary in Redis for 1 hour
	hm.redis.Set(ctx, summaryKey, summary, time.Hour)

	return summary, nil
}

// GetSummary retrieves a summary for the session (delegates to Supabase)
func (hm *HybridMemory) GetSummary(ctx context.Context, sessionID string) (*Summary, error) {
	return hm.supabase.GetSummary(ctx, sessionID)
}

// GetStats returns statistics about memory usage
func (hm *HybridMemory) GetStats(ctx context.Context, sessionID string) (*Stats, error) {
	// Get stats from Supabase
	stats, err := hm.supabase.GetStats(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	// Enhance with Redis session data if available
	if sessionID != "" {
		metaKey := fmt.Sprintf("session:%s:meta", sessionID)
		meta := hm.redis.HGetAll(ctx, metaKey).Val()

		if msgCount, ok := meta["message_count"]; ok {
			// Parse and update if Redis has more recent data
			var count int
			fmt.Sscanf(msgCount, "%d", &count)
			if count > stats.SessionMessages {
				stats.SessionMessages = count
			}
		}

		if tokens, ok := meta["total_tokens"]; ok {
			var tokenCount int
			fmt.Sscanf(tokens, "%d", &tokenCount)
			stats.TotalTokens = tokenCount
		}
	}

	return stats, nil
}

// Close closes all connections
func (hm *HybridMemory) Close() error {
	if err := hm.redis.Close(); err != nil {
		fmt.Printf("Warning: failed to close Redis: %v\n", err)
	}
	return hm.supabase.Close()
}

// ClearCache clears Redis cache for a specific session
func (hm *HybridMemory) ClearCache(ctx context.Context, sessionID string) error {
	sessionKey := fmt.Sprintf("session:%s:messages", sessionID)
	metaKey := fmt.Sprintf("session:%s:meta", sessionID)
	summaryKey := fmt.Sprintf("session:%s:summary", sessionID)

	return hm.redis.Del(ctx, sessionKey, metaKey, summaryKey).Err()
}

// CacheStats represents cache statistics
type CacheStats struct {
	Hits         int64
	Misses       int64
	SessionCount int64
	MemoryUsage  int64
}

// GetCacheStats returns Redis cache statistics
func (hm *HybridMemory) GetCacheStats(ctx context.Context) (*CacheStats, error) {
	// Parse Redis INFO stats (simplified)
	stats := &CacheStats{
		Hits:   0, // Would parse from INFO stats
		Misses: 0, // Would parse from INFO stats
	}

	// Count active sessions in cache
	keys := hm.redis.Keys(ctx, "session:*:messages").Val()
	stats.SessionCount = int64(len(keys))

	// Estimate memory usage (simplified)
	stats.MemoryUsage = stats.SessionCount * 1024 // Rough estimate

	return stats, nil
}

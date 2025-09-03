// Package memory provides a modular conversation memory system for AI agents.
// It supports both short-term session memory (via Redis) and long-term semantic memory (via Supabase pgvector).
package memory

import (
	"context"
	"time"
)

// Message represents a single conversation message
type Message struct {
	ID        string    `json:"id"`
	Role      string    `json:"role"`      // "user", "assistant", "system"
	Content   string    `json:"content"`
	Metadata  Metadata  `json:"metadata"`
	Timestamp time.Time `json:"timestamp"`
	Embedding []float32 `json:"embedding,omitempty"`
}

// Metadata contains additional message information
type Metadata struct {
	SessionID   string                 `json:"session_id"`
	UserID      string                 `json:"user_id,omitempty"`
	TokenCount  int                    `json:"token_count,omitempty"`
	Model       string                 `json:"model,omitempty"`
	Temperature float64                `json:"temperature,omitempty"`
	Extra       map[string]interface{} `json:"extra,omitempty"`
}

// SearchResult represents a semantic search result
type SearchResult struct {
	Message    Message `json:"message"`
	Score      float32 `json:"score"`
	Distance   float32 `json:"distance"`
}

// Memory interface defines the contract for memory implementations
type Memory interface {
	// Session Memory (fast, recent messages)
	AddMessage(ctx context.Context, msg Message) error
	GetRecentMessages(ctx context.Context, sessionID string, limit int) ([]Message, error)
	ClearSession(ctx context.Context, sessionID string) error
	
	// Semantic Memory (long-term, searchable)
	Store(ctx context.Context, msg Message) error
	Search(ctx context.Context, query string, limit int, threshold float32) ([]SearchResult, error)
	SearchWithEmbedding(ctx context.Context, embedding []float32, limit int, threshold float32) ([]SearchResult, error)
	
	// Management
	Summarize(ctx context.Context, sessionID string, maxTokens int) (string, error)
	GetStats(ctx context.Context, sessionID string) (*Stats, error)
	Close() error
}

// Stats provides memory usage statistics
type Stats struct {
	SessionID           string    `json:"session_id"`
	TotalMessages       int       `json:"total_messages"`
	SessionMessages     int       `json:"session_messages"`
	TotalTokens         int       `json:"total_tokens"`
	OldestMessage       time.Time `json:"oldest_message"`
	LatestMessage       time.Time `json:"latest_message"`
	UniqueUsers         int       `json:"unique_users,omitempty"`
	StorageSize         int64     `json:"storage_size,omitempty"`
}

// Config holds configuration for memory initialization
type Config struct {
	// Supabase Configuration (required for semantic memory)
	SupabaseURL    string `json:"supabase_url"`
	SupabaseKey    string `json:"supabase_key"`
	DatabaseURL    string `json:"database_url"` // Direct PostgreSQL connection
	
	// Redis Configuration (optional for fast session cache)
	RedisAddr      string `json:"redis_addr,omitempty"`     // e.g., "localhost:6379"
	RedisPassword  string `json:"redis_password,omitempty"`
	RedisDB        int    `json:"redis_db,omitempty"`
	
	// OpenAI Configuration (for embeddings)
	OpenAIKey      string `json:"openai_key"`
	EmbeddingModel string `json:"embedding_model,omitempty"` // default: "text-embedding-3-small"
	
	// Memory Settings
	MaxSessionMessages int           `json:"max_session_messages,omitempty"` // default: 50
	SessionTTL        time.Duration `json:"session_ttl,omitempty"`          // default: 24h
	AutoSummarize     bool          `json:"auto_summarize,omitempty"`       // auto-summarize old messages
	VectorDimension   int           `json:"vector_dimension,omitempty"`     // default: 1536
}

// New creates a new memory instance based on configuration
func New(cfg Config) (Memory, error) {
	// If Redis is configured, use hybrid memory (Redis + Supabase)
	if cfg.RedisAddr != "" {
		return NewHybridMemory(cfg)
	}
	
	// Otherwise, use Supabase-only memory
	return NewSupabaseMemory(cfg)
}

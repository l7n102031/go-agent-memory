// Package memory provides a modular conversation memory system for AI agents.
// It supports both short-term session memory (via Redis) and long-term semantic memory (via Supabase pgvector).
package memory

import (
	"context"
	"time"
)

// MemoryMode represents the type of memory implementation
type MemoryMode string

// Memory mode constants
const (
	SESSION_ONLY MemoryMode = "session_only" // In-memory only, no persistence
	PERSISTENT   MemoryMode = "persistent"   // PostgreSQL only
	HYBRID       MemoryMode = "hybrid"       // Redis + PostgreSQL
)

// Message represents a single conversation message
type Message struct {
	ID        string    `json:"id"`
	Role      string    `json:"role"` // "user", "assistant", "system"
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
	Message  Message `json:"message"`
	Score    float32 `json:"score"`
	Distance float32 `json:"distance"`
}

// Summary represents a conversation summary
type Summary struct {
	SessionID    string    `json:"session_id"`
	Content      string    `json:"content"`
	TokenCount   int       `json:"token_count"`
	MessageCount int       `json:"message_count"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	Created      time.Time `json:"created"`
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
	GetSummary(ctx context.Context, sessionID string) (*Summary, error)
	GetStats(ctx context.Context, sessionID string) (*Stats, error)
	Close() error
}

// Stats provides memory usage statistics
type Stats struct {
	SessionID       string     `json:"session_id"`
	TotalMessages   int        `json:"total_messages"`
	SessionMessages int        `json:"session_messages"`
	TotalTokens     int        `json:"total_tokens"`
	ActiveTokens    int        `json:"active_tokens,omitempty"` // Tokens after summarization
	OldestMessage   *time.Time `json:"oldest_message,omitempty"`
	LatestMessage   *time.Time `json:"latest_message,omitempty"`
	UniqueUsers     int        `json:"unique_users,omitempty"`
	StorageSize     int64      `json:"storage_size,omitempty"`
	HasSummary      bool       `json:"has_summary,omitempty"` // Whether session has summary
}

// Config holds configuration for memory initialization
type Config struct {
	// Mode Selection
	Mode MemoryMode `json:"mode,omitempty"` // SESSION_ONLY, PERSISTENT, HYBRID

	// Feature Flags
	EnablePersistence    bool `json:"enable_persistence,omitempty"`
	EnableSemanticSearch bool `json:"enable_semantic_search,omitempty"`
	EnableAutoSummarize  bool `json:"enable_auto_summarize,omitempty"`

	// Database Configuration
	SupabaseURL string `json:"supabase_url,omitempty"`
	SupabaseKey string `json:"supabase_key,omitempty"`
	DatabaseURL string `json:"database_url,omitempty"` // Direct PostgreSQL connection

	// Redis Configuration (optional for fast session cache)
	RedisAddr     string `json:"redis_addr,omitempty"` // e.g., "localhost:6379"
	RedisPassword string `json:"redis_password,omitempty"`
	RedisDB       int    `json:"redis_db,omitempty"`

	// OpenAI Configuration (for embeddings)
	OpenAIKey      string `json:"openai_key,omitempty"`
	EmbeddingModel string `json:"embedding_model,omitempty"` // default: "text-embedding-3-small"

	// Memory Settings
	MaxSessionMessages int           `json:"max_session_messages,omitempty"` // default: 50
	SessionTTL         time.Duration `json:"session_ttl,omitempty"`          // default: 24h
	AutoSummarize      bool          `json:"auto_summarize,omitempty"`       // auto-summarize old messages
	VectorDimension    int           `json:"vector_dimension,omitempty"`     // default: 1536

	// Summarization Settings
	SummarizeThreshold int    `json:"summarize_threshold,omitempty"`  // Messages before summarization
	SummarizeMaxTokens int    `json:"summarize_max_tokens,omitempty"` // Target summary length
	SummarizeModel     string `json:"summarize_model,omitempty"`      // Model for summarization
	ArchiveOldMessages bool   `json:"archive_old_messages,omitempty"` // Keep originals

	// Search Settings
	DefaultSearchLimit     int     `json:"default_search_limit,omitempty"`
	DefaultSearchThreshold float32 `json:"default_search_threshold,omitempty"`
}

// New creates a new memory instance based on configuration (legacy)
func New(cfg Config) (Memory, error) {
	// Determine mode if not set
	if cfg.Mode == "" {
		if cfg.RedisAddr != "" {
			cfg.Mode = HYBRID
		} else if cfg.DatabaseURL != "" {
			cfg.Mode = PERSISTENT
		} else {
			cfg.Mode = SESSION_ONLY
		}
	}

	return NewWithConfig(cfg)
}

// NewWithConfig creates a new memory instance with explicit configuration
func NewWithConfig(cfg Config) (Memory, error) {
	// Set defaults
	if cfg.MaxSessionMessages == 0 {
		cfg.MaxSessionMessages = 50
	}
	if cfg.SessionTTL == 0 {
		cfg.SessionTTL = 24 * time.Hour
	}
	if cfg.EmbeddingModel == "" {
		cfg.EmbeddingModel = "text-embedding-3-small"
	}
	if cfg.VectorDimension == 0 {
		cfg.VectorDimension = 1536
	}
	if cfg.SummarizeThreshold == 0 {
		cfg.SummarizeThreshold = 30
	}
	if cfg.SummarizeMaxTokens == 0 {
		cfg.SummarizeMaxTokens = 500
	}
	if cfg.SummarizeModel == "" {
		cfg.SummarizeModel = "gpt-3.5-turbo"
	}
	if cfg.DefaultSearchLimit == 0 {
		cfg.DefaultSearchLimit = 5
	}
	if cfg.DefaultSearchThreshold == 0 {
		cfg.DefaultSearchThreshold = 0.7
	}

	// Create memory based on mode
	switch cfg.Mode {
	case SESSION_ONLY:
		return NewSessionOnlyMemory(cfg)
	case PERSISTENT:
		return NewSupabaseMemory(cfg)
	case HYBRID:
		return NewHybridMemory(cfg)
	default:
		// Auto-detect based on configuration
		if cfg.RedisAddr != "" && cfg.DatabaseURL != "" {
			return NewHybridMemory(cfg)
		} else if cfg.DatabaseURL != "" {
			return NewSupabaseMemory(cfg)
		} else {
			return NewSessionOnlyMemory(cfg)
		}
	}
}

package memory

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
	"github.com/sashabaranov/go-openai"
)

// SupabaseMemory implements Memory using Supabase PostgreSQL with pgvector
type SupabaseMemory struct {
	db         *pgxpool.Pool
	openai     *openai.Client
	config     Config
	embModel   string
	dimension  int
}

// NewSupabaseMemory creates a new Supabase-based memory instance
func NewSupabaseMemory(cfg Config) (Memory, error) {
	// Connect to PostgreSQL
	config, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}
	
	db, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	
	// Create OpenAI client for embeddings
	openaiClient := openai.NewClient(cfg.OpenAIKey)
	
	// Set defaults
	if cfg.EmbeddingModel == "" {
		cfg.EmbeddingModel = "text-embedding-3-small"
	}
	if cfg.VectorDimension == 0 {
		cfg.VectorDimension = 1536
	}
	
	sm := &SupabaseMemory{
		db:         db,
		openai:     openaiClient,
		config:     cfg,
		embModel:   cfg.EmbeddingModel,
		dimension:  cfg.VectorDimension,
	}
	
	// Initialize database schema
	if err := sm.initSchema(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}
	
	return sm, nil
}

// initSchema creates the necessary tables and indexes
func (sm *SupabaseMemory) initSchema(ctx context.Context) error {
	schema := fmt.Sprintf(`
		-- Enable pgvector extension
		CREATE EXTENSION IF NOT EXISTS vector;
		
		-- Create messages table
		CREATE TABLE IF NOT EXISTS agent_messages (
			id SERIAL PRIMARY KEY,
			message_id TEXT UNIQUE NOT NULL,
			session_id TEXT NOT NULL,
			user_id TEXT,
			role TEXT NOT NULL,
			content TEXT NOT NULL,
			metadata JSONB,
			embedding vector(%d),
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW()
		);
		
		-- Create indexes
		CREATE INDEX IF NOT EXISTS idx_messages_session ON agent_messages(session_id);
		CREATE INDEX IF NOT EXISTS idx_messages_user ON agent_messages(user_id);
		CREATE INDEX IF NOT EXISTS idx_messages_created ON agent_messages(created_at DESC);
		
		-- Create HNSW index for fast similarity search
		CREATE INDEX IF NOT EXISTS idx_messages_embedding ON agent_messages 
		USING hnsw (embedding vector_cosine_ops)
		WITH (m = 16, ef_construction = 64);
		
		-- Create summaries table
		CREATE TABLE IF NOT EXISTS agent_summaries (
			id SERIAL PRIMARY KEY,
			session_id TEXT NOT NULL,
			summary TEXT NOT NULL,
			message_count INT,
			token_count INT,
			start_time TIMESTAMPTZ,
			end_time TIMESTAMPTZ,
			created_at TIMESTAMPTZ DEFAULT NOW()
		);
		
		CREATE INDEX IF NOT EXISTS idx_summaries_session ON agent_summaries(session_id);
	`, sm.dimension)
	
	_, err := sm.db.Exec(ctx, schema)
	return err
}

// AddMessage adds a message to session memory
func (sm *SupabaseMemory) AddMessage(ctx context.Context, msg Message) error {
	// Generate embedding if not provided
	if len(msg.Embedding) == 0 && msg.Content != "" {
		embedding, err := sm.generateEmbedding(ctx, msg.Content)
		if err != nil {
			// Continue without embedding on error
			fmt.Printf("Warning: failed to generate embedding: %v\n", err)
		} else {
			msg.Embedding = embedding
		}
	}
	
	metadataJSON, err := json.Marshal(msg.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}
	
	query := `
		INSERT INTO agent_messages (
			message_id, session_id, user_id, role, content, 
			metadata, embedding, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (message_id) DO UPDATE SET
			content = EXCLUDED.content,
			metadata = EXCLUDED.metadata,
			embedding = EXCLUDED.embedding,
			updated_at = NOW()
	`
	
	var embedding interface{}
	if len(msg.Embedding) > 0 {
		embedding = pgvector.NewVector(msg.Embedding)
	}
	
	_, err = sm.db.Exec(ctx, query,
		msg.ID,
		msg.Metadata.SessionID,
		msg.Metadata.UserID,
		msg.Role,
		msg.Content,
		metadataJSON,
		embedding,
		msg.Timestamp,
	)
	
	return err
}

// GetRecentMessages retrieves recent messages for a session
func (sm *SupabaseMemory) GetRecentMessages(ctx context.Context, sessionID string, limit int) ([]Message, error) {
	if limit <= 0 {
		limit = 50
	}
	
	query := `
		SELECT message_id, session_id, user_id, role, content, 
		       metadata, created_at
		FROM agent_messages
		WHERE session_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`
	
	rows, err := sm.db.Query(ctx, query, sessionID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var messages []Message
	for rows.Next() {
		var msg Message
		var metadataJSON []byte
		var userID *string
		
		err := rows.Scan(
			&msg.ID,
			&msg.Metadata.SessionID,
			&userID,
			&msg.Role,
			&msg.Content,
			&metadataJSON,
			&msg.Timestamp,
		)
		if err != nil {
			return nil, err
		}
		
		if userID != nil {
			msg.Metadata.UserID = *userID
		}
		
		if len(metadataJSON) > 0 {
			json.Unmarshal(metadataJSON, &msg.Metadata)
		}
		
		messages = append(messages, msg)
	}
	
	// Reverse to get chronological order
	for i := len(messages)/2 - 1; i >= 0; i-- {
		opp := len(messages) - 1 - i
		messages[i], messages[opp] = messages[opp], messages[i]
	}
	
	return messages, nil
}

// Store saves a message for long-term semantic memory
func (sm *SupabaseMemory) Store(ctx context.Context, msg Message) error {
	// Same as AddMessage for Supabase implementation
	return sm.AddMessage(ctx, msg)
}

// Search performs semantic search on messages
func (sm *SupabaseMemory) Search(ctx context.Context, query string, limit int, threshold float32) ([]SearchResult, error) {
	// Generate embedding for query
	embedding, err := sm.generateEmbedding(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}
	
	return sm.SearchWithEmbedding(ctx, embedding, limit, threshold)
}

// SearchWithEmbedding searches using a pre-computed embedding
func (sm *SupabaseMemory) SearchWithEmbedding(ctx context.Context, embedding []float32, limit int, threshold float32) ([]SearchResult, error) {
	if limit <= 0 {
		limit = 10
	}
	if threshold <= 0 {
		threshold = 0.7
	}
	
	query := `
		SELECT 
			message_id, session_id, user_id, role, content, 
			metadata, created_at,
			1 - (embedding <=> $1::vector) as score,
			embedding <=> $1::vector as distance
		FROM agent_messages
		WHERE embedding IS NOT NULL
		  AND 1 - (embedding <=> $1::vector) > $2
		ORDER BY embedding <=> $1::vector
		LIMIT $3
	`
	
	rows, err := sm.db.Query(ctx, query, pgvector.NewVector(embedding), threshold, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var results []SearchResult
	for rows.Next() {
		var result SearchResult
		var metadataJSON []byte
		var userID *string
		
		err := rows.Scan(
			&result.Message.ID,
			&result.Message.Metadata.SessionID,
			&userID,
			&result.Message.Role,
			&result.Message.Content,
			&metadataJSON,
			&result.Message.Timestamp,
			&result.Score,
			&result.Distance,
		)
		if err != nil {
			return nil, err
		}
		
		if userID != nil {
			result.Message.Metadata.UserID = *userID
		}
		
		if len(metadataJSON) > 0 {
			json.Unmarshal(metadataJSON, &result.Message.Metadata)
		}
		
		results = append(results, result)
	}
	
	return results, nil
}

// ClearSession removes all messages for a session
func (sm *SupabaseMemory) ClearSession(ctx context.Context, sessionID string) error {
	_, err := sm.db.Exec(ctx, "DELETE FROM agent_messages WHERE session_id = $1", sessionID)
	return err
}

// Summarize creates a summary of a session's messages
func (sm *SupabaseMemory) Summarize(ctx context.Context, sessionID string, maxTokens int) (string, error) {
	// Get all messages for the session
	messages, err := sm.GetRecentMessages(ctx, sessionID, 1000) // Get up to 1000 messages
	if err != nil {
		return "", err
	}
	
	if len(messages) == 0 {
		return "", nil
	}
	
	// Build conversation text
	var conversation string
	tokenCount := 0
	for _, msg := range messages {
		msgText := fmt.Sprintf("%s: %s\n", msg.Role, msg.Content)
		msgTokens := len(msgText) / 4 // Rough token estimation
		
		if maxTokens > 0 && tokenCount+msgTokens > maxTokens {
			break
		}
		
		conversation += msgText
		tokenCount += msgTokens
	}
	
	// Use OpenAI to generate summary
	resp, err := sm.openai.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: "gpt-4o-mini",
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "Summarize the following conversation concisely, preserving key information and context:",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: conversation,
			},
		},
		Temperature: 0.3,
		MaxTokens:   500,
	})
	
	if err != nil {
		return "", fmt.Errorf("failed to generate summary: %w", err)
	}
	
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no summary generated")
	}
	
	summary := resp.Choices[0].Message.Content
	
	// Store the summary
	_, err = sm.db.Exec(ctx, `
		INSERT INTO agent_summaries (session_id, summary, message_count, token_count, start_time, end_time)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, sessionID, summary, len(messages), tokenCount, 
	   messages[0].Timestamp, messages[len(messages)-1].Timestamp)
	
	return summary, err
}

// GetStats returns statistics about memory usage
func (sm *SupabaseMemory) GetStats(ctx context.Context, sessionID string) (*Stats, error) {
	stats := &Stats{SessionID: sessionID}
	
	// Get message counts
	err := sm.db.QueryRow(ctx, `
		SELECT 
			COUNT(*) as total,
			COUNT(DISTINCT session_id) as sessions,
			COUNT(DISTINCT user_id) as users,
			MIN(created_at) as oldest,
			MAX(created_at) as latest,
			SUM(LENGTH(content)) as size
		FROM agent_messages
		WHERE ($1 = '' OR session_id = $1)
	`, sessionID).Scan(
		&stats.TotalMessages,
		&stats.SessionMessages,
		&stats.UniqueUsers,
		&stats.OldestMessage,
		&stats.LatestMessage,
		&stats.StorageSize,
	)
	
	return stats, err
}

// generateEmbedding creates an embedding for text using OpenAI
func (sm *SupabaseMemory) generateEmbedding(ctx context.Context, text string) ([]float32, error) {
	// Map string model names to OpenAI constants
	var model openai.EmbeddingModel
	switch sm.embModel {
	case "text-embedding-3-large":
		model = openai.LargeEmbedding3
	case "text-embedding-3-small":
		model = openai.SmallEmbedding3  
	case "text-embedding-ada-002":
		model = openai.AdaEmbeddingV2
	default:
		model = openai.SmallEmbedding3 // Default to small
	}
	
	resp, err := sm.openai.CreateEmbeddings(ctx, openai.EmbeddingRequest{
		Model: model,
		Input: []string{text},
	})
	
	if err != nil {
		return nil, err
	}
	
	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("no embedding generated")
	}
	
	// Convert float64 to float32
	embedding := make([]float32, len(resp.Data[0].Embedding))
	for i, v := range resp.Data[0].Embedding {
		embedding[i] = float32(v)
	}
	
	return embedding, nil
}

// Close closes database connections
func (sm *SupabaseMemory) Close() error {
	sm.db.Close()
	return nil
}

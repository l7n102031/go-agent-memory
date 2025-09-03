// Complete Agent Integration Example
// Production-ready AI agent with configurable memory system.
package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	memory "github.com/framehood/go-agent-memory"
	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/option"
)

// Agent configuration
const (
	MODEL         = "gpt-4"
	TEMPERATURE   = 0.7
	MAX_TOKENS    = 2000
	SYSTEM_PROMPT = "You are a helpful AI assistant with memory of our conversation."
)

// Agent represents our AI agent with optional memory
type Agent struct {
	client openai.Client
	memory memory.Memory // Can be nil
	config AgentConfig
}

type AgentConfig struct {
	Model        string
	Temperature  float64
	MaxTokens    int
	SystemPrompt string

	// Memory configuration
	MemoryEnabled bool
	MemoryMode    string // "none", "session_only", "persistent", "hybrid"
}

func main() {
	fmt.Println("🤖 AI Agent with Memory Integration")
	fmt.Println("====================================")
	fmt.Println()

	// Initialize agent
	agent, err := initializeAgent()
	if err != nil {
		log.Fatalf("Failed to initialize agent: %v", err)
	}
	defer agent.Cleanup()

	// Show configuration
	agent.PrintConfig()

	// Start interactive session
	agent.InteractiveSession()
}

func initializeAgent() (*Agent, error) {
	// Get OpenAI key
	openAIKey := os.Getenv("OPENAI_API_KEY")
	if openAIKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable is required")
	}

	// Create OpenAI client
	client := openai.NewClient(
		option.WithAPIKey(openAIKey),
	)

	// Determine memory mode from environment
	memoryMode := os.Getenv("MEMORY_MODE")
	if memoryMode == "" {
		memoryMode = "none"
	}

	// Create agent configuration
	config := AgentConfig{
		Model:         MODEL,
		Temperature:   TEMPERATURE,
		MaxTokens:     MAX_TOKENS,
		SystemPrompt:  SYSTEM_PROMPT,
		MemoryEnabled: memoryMode != "none",
		MemoryMode:    memoryMode,
	}

	// Initialize agent
	agent := &Agent{
		client: client,
		config: config,
	}

	// Initialize memory if enabled
	if config.MemoryEnabled {
		mem, err := initializeMemory(memoryMode)
		if err != nil {
			fmt.Printf("⚠️  Memory initialization failed: %v\n", err)
			fmt.Println("   Continuing without memory...")
			agent.config.MemoryEnabled = false
		} else {
			agent.memory = mem
			fmt.Println("✅ Memory system initialized")
		}
	} else {
		fmt.Println("ℹ️  Memory disabled (set MEMORY_MODE to enable)")
	}

	return agent, nil
}

func initializeMemory(mode string) (memory.Memory, error) {
	var config memory.Config

	switch mode {
	case "session_only":
		config = memory.Config{
			Mode:               memory.SESSION_ONLY,
			MaxSessionMessages: 50,
		}

	case "persistent":
		dbURL := os.Getenv("DATABASE_URL")
		if dbURL == "" {
			return nil, fmt.Errorf("DATABASE_URL required for persistent mode")
		}

		config = memory.Config{
			Mode:                 memory.PERSISTENT,
			DatabaseURL:          dbURL,
			OpenAIKey:            os.Getenv("OPENAI_API_KEY"),
			EnableSemanticSearch: true,
			EnableAutoSummarize:  true,
			SummarizeThreshold:   30,
		}

	case "hybrid":
		dbURL := os.Getenv("DATABASE_URL")
		redisURL := os.Getenv("REDIS_URL")

		if dbURL == "" || redisURL == "" {
			return nil, fmt.Errorf("DATABASE_URL and REDIS_URL required for hybrid mode")
		}

		config = memory.Config{
			Mode:                 memory.HYBRID,
			DatabaseURL:          dbURL,
			RedisAddr:            redisURL,
			RedisPassword:        os.Getenv("REDIS_PASSWORD"),
			OpenAIKey:            os.Getenv("OPENAI_API_KEY"),
			EnableSemanticSearch: true,
			EnableAutoSummarize:  true,
			MaxSessionMessages:   30,
			SessionTTL:           2 * time.Hour,
		}

	default:
		return nil, fmt.Errorf("unknown memory mode: %s", mode)
	}

	return memory.NewWithConfig(config)
}

func (a *Agent) PrintConfig() {
	fmt.Println("📋 Agent Configuration:")
	fmt.Println("───────────────────────")
	fmt.Printf("   Model: %s\n", a.config.Model)
	fmt.Printf("   Temperature: %.1f\n", a.config.Temperature)
	fmt.Printf("   Max Tokens: %d\n", a.config.MaxTokens)
	fmt.Printf("   Memory: %s\n", a.config.MemoryMode)
	fmt.Println()
}

func (a *Agent) InteractiveSession() {
	sessionID := fmt.Sprintf("session-%d", time.Now().Unix())
	ctx := context.Background()

	fmt.Println("💬 Interactive Chat Session")
	fmt.Println("──────────────────────────")
	fmt.Println("Commands:")
	fmt.Println("  /memory   - Show memory stats")
	fmt.Println("  /search   - Search conversation history")
	fmt.Println("  /clear    - Clear session memory")
	fmt.Println("  /help     - Show this help")
	fmt.Println("  /exit     - Exit the agent")
	fmt.Println()
	fmt.Println("Start chatting! Type your message and press Enter.")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("You: ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		// Handle commands
		if strings.HasPrefix(input, "/") {
			if !a.handleCommand(ctx, sessionID, input) {
				break
			}
			continue
		}

		// Process chat message
		response, err := a.Chat(ctx, sessionID, input)
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			continue
		}

		fmt.Printf("\n🤖 Assistant: %s\n\n", response)
	}

	fmt.Println("\n👋 Goodbye!")
}

func (a *Agent) handleCommand(ctx context.Context, sessionID string, command string) bool {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return true
	}

	switch parts[0] {
	case "/exit", "/quit":
		return false

	case "/help":
		fmt.Println("\n📚 Available Commands:")
		fmt.Println("  /memory        - Show memory statistics")
		fmt.Println("  /search <query>- Search conversation history")
		fmt.Println("  /clear         - Clear current session")
		fmt.Println("  /help          - Show this help")
		fmt.Println("  /exit          - Exit the agent")

	case "/memory":
		if a.memory == nil {
			fmt.Println("ℹ️  Memory is not enabled")
		} else {
			a.showMemoryStats(ctx, sessionID)
		}

	case "/search":
		if a.memory == nil {
			fmt.Println("ℹ️  Memory is not enabled")
		} else if len(parts) < 2 {
			fmt.Println("Usage: /search <query>")
		} else {
			query := strings.Join(parts[1:], " ")
			a.searchMemory(ctx, query)
		}

	case "/clear":
		if a.memory == nil {
			fmt.Println("ℹ️  Memory is not enabled")
		} else {
			err := a.memory.ClearSession(ctx, sessionID)
			if err != nil {
				fmt.Printf("❌ Error clearing session: %v\n", err)
			} else {
				fmt.Println("✅ Session cleared")
			}
		}

	default:
		fmt.Printf("Unknown command: %s (type /help for commands)\n", parts[0])
	}

	return true
}

func (a *Agent) Chat(ctx context.Context, sessionID string, userMessage string) (string, error) {
	// Store user message in memory
	if a.memory != nil {
		a.memory.AddMessage(ctx, memory.Message{
			ID:        fmt.Sprintf("msg-%d", time.Now().UnixNano()),
			Role:      "user",
			Content:   userMessage,
			Timestamp: time.Now(),
			Metadata: memory.Metadata{
				SessionID: sessionID,
			},
		})
	}

	// Build conversation context
	messages := a.buildContext(ctx, sessionID, userMessage)

	// Call OpenAI API
	completion, err := a.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model:       openai.ChatModel(a.config.Model),
		Messages:    messages,
		Temperature: openai.Float(a.config.Temperature),
		MaxTokens:   openai.Int(int64(a.config.MaxTokens)),
	})

	if err != nil {
		return "", fmt.Errorf("OpenAI API error: %w", err)
	}

	// Get response
	if len(completion.Choices) == 0 {
		return "", fmt.Errorf("no response from model")
	}

	response := completion.Choices[0].Message.Content

	// Store assistant response in memory
	if a.memory != nil {
		a.memory.AddMessage(ctx, memory.Message{
			ID:        fmt.Sprintf("msg-%d", time.Now().UnixNano()),
			Role:      "assistant",
			Content:   response,
			Timestamp: time.Now(),
			Metadata: memory.Metadata{
				SessionID: sessionID,
				Model:     a.config.Model,
			},
		})
	}

	return response, nil
}

func (a *Agent) buildContext(ctx context.Context, sessionID string, currentMessage string) []openai.ChatCompletionMessageParamUnion {
	var messages []openai.ChatCompletionMessageParamUnion

	// Add system prompt
	messages = append(messages, openai.SystemMessage(a.config.SystemPrompt))

	// Add conversation history from memory
	if a.memory != nil {
		// Get recent messages
		recent, err := a.memory.GetRecentMessages(ctx, sessionID, 10)
		if err == nil {
			for _, msg := range recent {
				switch msg.Role {
				case "user":
					messages = append(messages, openai.UserMessage(msg.Content))
				case "assistant":
					messages = append(messages, openai.AssistantMessage(msg.Content))
				}
			}
		}

		// Search for relevant context (if semantic search is enabled)
		if a.config.MemoryMode == "persistent" || a.config.MemoryMode == "hybrid" {
			results, err := a.memory.Search(ctx, currentMessage, 3, 0.75)
			if err == nil && len(results) > 0 {
				// Add relevant context as system message
				contextMsg := "Relevant context from previous conversations:\n"
				for _, result := range results {
					contextMsg += fmt.Sprintf("- %s\n", result.Message.Content)
				}
				messages = append(messages, openai.SystemMessage(contextMsg))
			}
		}
	}

	// Add current message
	messages = append(messages, openai.UserMessage(currentMessage))

	return messages
}

func (a *Agent) showMemoryStats(ctx context.Context, sessionID string) {
	stats, err := a.memory.GetStats(ctx, sessionID)
	if err != nil {
		fmt.Printf("❌ Error getting stats: %v\n", err)
		return
	}

	fmt.Println("\n📊 Memory Statistics:")
	fmt.Println("─────────────────────")
	fmt.Printf("   Session ID: %s\n", sessionID)
	fmt.Printf("   Total messages: %d\n", stats.SessionMessages)
	fmt.Printf("   Total tokens: ~%d\n", stats.TotalTokens)
	fmt.Printf("   Memory mode: %s\n", a.config.MemoryMode)

	if stats.OldestMessage != nil {
		fmt.Printf("   First message: %s\n", stats.OldestMessage.Format("15:04:05"))
	}
	if stats.LatestMessage != nil {
		fmt.Printf("   Last message: %s\n", stats.LatestMessage.Format("15:04:05"))
	}

	fmt.Println()
}

func (a *Agent) searchMemory(ctx context.Context, query string) {
	fmt.Printf("\n🔍 Searching for: \"%s\"\n", query)

	results, err := a.memory.Search(ctx, query, 5, 0.7)
	if err != nil {
		fmt.Printf("❌ Search error: %v\n", err)
		return
	}

	if len(results) == 0 {
		fmt.Println("No matching results found.")
	} else {
		fmt.Printf("\nFound %d results:\n", len(results))
		fmt.Println("─────────────────")

		for i, result := range results {
			fmt.Printf("\n%d. [Score: %.2f] [%s]\n", i+1, result.Score, result.Message.Role)

			// Truncate long messages
			content := result.Message.Content
			if len(content) > 150 {
				content = content[:147] + "..."
			}
			fmt.Printf("   %s\n", content)
		}
	}

	fmt.Println()
}

func (a *Agent) Cleanup() {
	if a.memory != nil {
		a.memory.Close()
	}
}

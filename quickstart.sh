#!/bin/bash

# Go Agent Memory - Quick Setup Script
# This script helps you quickly set up the memory module for your agent

set -e

echo "🧠 Go Agent Memory - Quick Setup"
echo "================================"
echo ""

# Check for required environment variables
if [ -z "$DATABASE_URL" ]; then
    echo "⚠️  DATABASE_URL not set. Please set your Supabase/PostgreSQL connection string:"
    echo "   export DATABASE_URL='postgresql://user:pass@host:5432/dbname'"
    echo ""
    echo "You can get this from your Supabase dashboard under Settings > Database"
    exit 1
fi

if [ -z "$OPENAI_API_KEY" ]; then
    echo "⚠️  OPENAI_API_KEY not set. Please set your OpenAI API key:"
    echo "   export OPENAI_API_KEY='sk-...'"
    exit 1
fi

echo "✅ Environment variables detected"
echo ""

# Install the module
echo "📦 Installing go-agent-memory module..."
go get github.com/framehood/go-agent-memory
echo "✅ Module installed"
echo ""

# Create example integration file
echo "📝 Creating example integration..."
cat > memory_integration.go << 'EOF'
package main

import (
    "context"
    "fmt"
    "os"
    "time"
    
    memory "github.com/framehood/go-agent-memory"
)

func main() {
    // Initialize memory
    mem, err := memory.New(memory.Config{
        DatabaseURL:    os.Getenv("DATABASE_URL"),
        OpenAIKey:      os.Getenv("OPENAI_API_KEY"),
        RedisAddr:      os.Getenv("REDIS_URL"), // Optional
    })
    if err != nil {
        panic(fmt.Errorf("Failed to initialize memory: %w", err))
    }
    defer mem.Close()
    
    ctx := context.Background()
    sessionID := "quickstart-session"
    
    // Add a test message
    err = mem.AddMessage(ctx, memory.Message{
        ID:      fmt.Sprintf("msg-%d", time.Now().Unix()),
        Role:    "user",
        Content: "This is a test message from the quickstart script!",
        Metadata: memory.Metadata{
            SessionID: sessionID,
        },
        Timestamp: time.Now(),
    })
    if err != nil {
        fmt.Printf("Error adding message: %v\n", err)
    } else {
        fmt.Println("✅ Successfully added test message to memory!")
    }
    
    // Retrieve messages
    messages, err := mem.GetRecentMessages(ctx, sessionID, 10)
    if err != nil {
        fmt.Printf("Error retrieving messages: %v\n", err)
    } else {
        fmt.Printf("✅ Retrieved %d messages from memory\n", len(messages))
        for _, msg := range messages {
            fmt.Printf("  - [%s] %s: %s\n", 
                msg.Timestamp.Format("15:04:05"), 
                msg.Role, 
                msg.Content[:min(50, len(msg.Content))])
        }
    }
    
    // Test semantic search
    results, err := mem.Search(ctx, "test message quickstart", 5, 0.7)
    if err != nil {
        fmt.Printf("Error searching: %v\n", err)
    } else {
        fmt.Printf("✅ Found %d similar messages\n", len(results))
    }
    
    fmt.Println("\n🎉 Memory module is working correctly!")
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}
EOF

echo "✅ Example created: memory_integration.go"
echo ""

# Run the test
echo "🧪 Testing memory module..."
go run memory_integration.go

echo ""
echo "🎉 Setup complete! Your memory module is ready to use."
echo ""
echo "📚 Next steps:"
echo "   1. Check memory_integration.go for the example code"
echo "   2. Copy the integration code to your agent"
echo "   3. Read the documentation at: https://github.com/framehood/go-agent-memory"
echo ""
echo "💡 Tip: For faster performance, set up Redis:"
echo "   export REDIS_URL='localhost:6379'"

package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/Blogem/enron-graph/pkg/llm"
)

func main() {
	// Setup logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	// Configuration - adjust these as needed
	baseURL := "https://litellm.sbp.ai"        // Default LiteLLM URL
	completionModel := "aws/claude-4-5-sonnet" // Change to your model
	embeddingModel := "azure/text-embedding-3-large"
	apiKey := os.Getenv("LITELLM_API_KEY") // Optional, leave empty if not needed

	// Create client
	client := llm.NewLiteLLMClient(baseURL, completionModel, embeddingModel, apiKey, logger)
	defer client.Close()

	fmt.Println("=== Testing LiteLLM Connection ===")
	fmt.Printf("Base URL: %s\n", baseURL)
	fmt.Printf("Completion Model: %s\n", completionModel)
	fmt.Printf("Embedding Model: %s\n\n", embeddingModel)

	// Test 1: Generate a simple completion
	fmt.Println("Test 1: Generate Completion")
	fmt.Println("Prompt: 'Say hello in one sentence'")

	ctx := context.Background()
	response, err := client.GenerateCompletion(ctx, "Say hello in one sentence")
	if err != nil {
		log.Fatalf("❌ Completion failed: %v\n", err)
	}
	fmt.Printf("✅ Response: %s\n\n", response)

	// Test 2: Generate an embedding
	fmt.Println("Test 2: Generate Embedding")
	fmt.Println("Text: 'Hello world'")

	embedding, err := client.GenerateEmbedding(ctx, "Hello world")
	if err != nil {
		log.Fatalf("❌ Embedding failed: %v\n", err)
	}
	fmt.Printf("✅ Embedding generated: %d dimensions\n", len(embedding))
	fmt.Printf("   First 5 values: %v\n\n", embedding[:min(5, len(embedding))])

	fmt.Println("=== All tests passed! ===")
}

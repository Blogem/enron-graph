package llm

import (
	"context"
)

// Client defines the interface for LLM operations
type Client interface {
	// GenerateCompletion generates a text completion from a prompt
	GenerateCompletion(ctx context.Context, prompt string) (string, error)

	// GenerateEmbedding generates a vector embedding for the given text
	GenerateEmbedding(ctx context.Context, text string) ([]float32, error)

	// GenerateEmbeddings generates embeddings for multiple texts in batch
	GenerateEmbeddings(ctx context.Context, texts []string) ([][]float32, error)

	// Close releases any resources held by the client
	Close() error
}

// CompletionOptions provides options for text generation
type CompletionOptions struct {
	Temperature float64
	MaxTokens   int
	TopP        float64
}

// DefaultCompletionOptions returns sensible defaults
func DefaultCompletionOptions() *CompletionOptions {
	return &CompletionOptions{
		Temperature: 0.7,
		MaxTokens:   2048,
		TopP:        0.9,
	}
}

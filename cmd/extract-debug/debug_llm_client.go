package main

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/Blogem/enron-graph/pkg/llm"
)

// DebugLLMClient wraps an LLM client and logs prompts and responses
type DebugLLMClient struct {
	base    llm.Client
	logger  *slog.Logger
	verbose bool
}

// NewDebugLLMClient creates a new debug LLM client wrapper
func NewDebugLLMClient(base llm.Client, logger *slog.Logger, verbose bool) *DebugLLMClient {
	return &DebugLLMClient{
		base:    base,
		logger:  logger,
		verbose: verbose,
	}
}

// GenerateCompletion wraps the base client's GenerateCompletion with debug logging
func (d *DebugLLMClient) GenerateCompletion(ctx context.Context, prompt string) (string, error) {
	// Log the prompt
	fmt.Println(strings.Repeat("-", 80))
	fmt.Println("LLM PROMPT:")
	fmt.Println(strings.Repeat("-", 80))
	if d.verbose {
		// Show full prompt in verbose mode
		fmt.Println(prompt)
	} else {
		// Show truncated prompt
		fmt.Println(truncate(prompt, 500))
	}
	fmt.Println(strings.Repeat("-", 80))

	// Call the base client
	response, err := d.base.GenerateCompletion(ctx, prompt)

	// Log the response
	fmt.Println("\nLLM RESPONSE:")
	fmt.Println(strings.Repeat("-", 80))
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
	} else {
		if d.verbose {
			// Show full response in verbose mode
			fmt.Println(response)
		} else {
			// Show truncated response
			fmt.Println(truncate(response, 500))
		}
	}
	fmt.Println(strings.Repeat("-", 80))

	return response, err
}

// GenerateEmbedding wraps the base client's GenerateEmbedding (no logging needed for embeddings)
func (d *DebugLLMClient) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	d.logger.Debug("Generating embedding", "text_length", len(text))
	return d.base.GenerateEmbedding(ctx, text)
}

// GenerateEmbeddings wraps the base client's GenerateEmbeddings (no logging needed for embeddings)
func (d *DebugLLMClient) GenerateEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	d.logger.Debug("Generating embeddings", "count", len(texts))
	return d.base.GenerateEmbeddings(ctx, texts)
}

// Close wraps the base client's Close method
func (d *DebugLLMClient) Close() error {
	return d.base.Close()
}

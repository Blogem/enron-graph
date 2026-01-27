package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/Blogem/enron-graph/internal/chat"
	"github.com/Blogem/enron-graph/pkg/llm"
	"github.com/Blogem/enron-graph/pkg/utils"
)

// llmAdapter adapts pkg/llm.Client to internal/chat.LLMClient interface
type llmAdapter struct {
	client llm.Client
}

// newLLMAdapter creates a new adapter wrapping a pkg/llm.Client
func newLLMAdapter(client llm.Client) chat.LLMClient {
	return &llmAdapter{client: client}
}

// GenerateCompletion implements chat.LLMClient
func (a *llmAdapter) GenerateCompletion(ctx context.Context, prompt string) (string, error) {
	return a.client.GenerateCompletion(ctx, prompt)
}

// GenerateEmbedding implements chat.LLMClient
func (a *llmAdapter) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	return a.client.GenerateEmbedding(ctx, text)
}

// newProductionLLMClient creates an LLM client based on config
func newProductionLLMClient(cfg *utils.Config) chat.LLMClient {
	// Create logger
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	var client llm.Client

	// Choose client based on LLM_PROVIDER setting
	if cfg.LLMProvider == "litellm" {
		client = llm.NewLiteLLMClient(
			cfg.LiteLLMURL,
			cfg.CompletionModel,
			cfg.EmbeddingModel,
			cfg.LiteLLMAPIKey,
			logger,
		)
	} else {
		// Default to Ollama
		client = llm.NewOllamaClient(
			cfg.OllamaURL,
			cfg.CompletionModel,
			cfg.EmbeddingModel,
			logger,
		)
	}

	// Wrap in adapter
	return newLLMAdapter(client)
}

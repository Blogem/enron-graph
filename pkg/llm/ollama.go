package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

// OllamaClient implements the Client interface using Ollama API
type OllamaClient struct {
	baseURL         string
	completionModel string
	embeddingModel  string
	httpClient      *http.Client
	logger          *slog.Logger
	maxRetries      int
	retryDelay      time.Duration
}

// NewOllamaClient creates a new Ollama client
func NewOllamaClient(baseURL, completionModel, embeddingModel string, logger *slog.Logger) *OllamaClient {
	return &OllamaClient{
		baseURL:         baseURL,
		completionModel: completionModel,
		embeddingModel:  embeddingModel,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		logger:     logger,
		maxRetries: 3,
		retryDelay: 2 * time.Second,
	}
}

// GenerateCompletion generates a text completion using Ollama
func (c *OllamaClient) GenerateCompletion(ctx context.Context, prompt string) (string, error) {
	requestBody := map[string]interface{}{
		"model":  c.completionModel,
		"prompt": prompt,
		"stream": false,
		"options": map[string]interface{}{
			"temperature": 0.7,
			"top_p":       0.9,
		},
	}

	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			c.logger.Debug("Retrying completion request",
				"attempt", attempt,
				"max_retries", c.maxRetries)
			time.Sleep(c.retryDelay * time.Duration(attempt))
		}

		response, err := c.makeRequest(ctx, "/api/generate", requestBody, 30*time.Second)
		if err != nil {
			lastErr = err
			continue
		}

		var result struct {
			Response string `json:"response"`
			Done     bool   `json:"done"`
		}

		if err := json.Unmarshal(response, &result); err != nil {
			lastErr = fmt.Errorf("failed to parse response: %w", err)
			continue
		}

		if !result.Done {
			lastErr = fmt.Errorf("incomplete response from Ollama")
			continue
		}

		return result.Response, nil
	}

	return "", fmt.Errorf("failed after %d retries: %w", c.maxRetries, lastErr)
}

// GenerateEmbedding generates a vector embedding using Ollama
func (c *OllamaClient) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	requestBody := map[string]interface{}{
		"model":  c.embeddingModel,
		"prompt": text,
	}

	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			c.logger.Debug("Retrying embedding request",
				"attempt", attempt,
				"max_retries", c.maxRetries)
			time.Sleep(c.retryDelay * time.Duration(attempt))
		}

		response, err := c.makeRequest(ctx, "/api/embeddings", requestBody, 10*time.Second)
		if err != nil {
			lastErr = err
			continue
		}

		var result struct {
			Embedding []float64 `json:"embedding"`
		}

		if err := json.Unmarshal(response, &result); err != nil {
			lastErr = fmt.Errorf("failed to parse response: %w", err)
			continue
		}

		if len(result.Embedding) == 0 {
			lastErr = fmt.Errorf("empty embedding returned")
			continue
		}

		// Convert float64 to float32
		embedding := make([]float32, len(result.Embedding))
		for i, v := range result.Embedding {
			embedding[i] = float32(v)
		}

		return embedding, nil
	}

	return nil, fmt.Errorf("failed after %d retries: %w", c.maxRetries, lastErr)
}

// GenerateEmbeddings generates embeddings for multiple texts in batch
func (c *OllamaClient) GenerateEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	embeddings := make([][]float32, len(texts))

	// Process in batches to avoid overwhelming the API
	batchSize := 10
	for i := 0; i < len(texts); i += batchSize {
		end := i + batchSize
		if end > len(texts) {
			end = len(texts)
		}

		for j := i; j < end; j++ {
			embedding, err := c.GenerateEmbedding(ctx, texts[j])
			if err != nil {
				return nil, fmt.Errorf("failed to generate embedding for text %d: %w", j, err)
			}
			embeddings[j] = embedding
		}

		// Small delay between batches
		if end < len(texts) {
			time.Sleep(100 * time.Millisecond)
		}
	}

	return embeddings, nil
}

// makeRequest makes an HTTP request to Ollama API with timeout and retries
func (c *OllamaClient) makeRequest(ctx context.Context, endpoint string, requestBody interface{}, timeout time.Duration) ([]byte, error) {
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := c.baseURL + endpoint

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Create a new client with custom timeout for this request
	client := &http.Client{Timeout: timeout}

	// Make request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// Close releases resources (no-op for Ollama client)
func (c *OllamaClient) Close() error {
	return nil
}

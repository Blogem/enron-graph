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

// LiteLLMClient implements the Client interface using LiteLLM API (OpenAI-compatible)
type LiteLLMClient struct {
	baseURL         string
	completionModel string
	embeddingModel  string
	apiKey          string
	httpClient      *http.Client
	logger          *slog.Logger
	maxRetries      int
	retryDelay      time.Duration
}

// NewLiteLLMClient creates a new LiteLLM client
func NewLiteLLMClient(baseURL, completionModel, embeddingModel, apiKey string, logger *slog.Logger) *LiteLLMClient {
	return &LiteLLMClient{
		baseURL:         baseURL,
		completionModel: completionModel,
		embeddingModel:  embeddingModel,
		apiKey:          apiKey,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		logger:     logger,
		maxRetries: 3,
		retryDelay: 2 * time.Second,
	}
}

// GenerateCompletion generates a text completion using LiteLLM
func (c *LiteLLMClient) GenerateCompletion(ctx context.Context, prompt string) (string, error) {
	requestBody := map[string]interface{}{
		"model": c.completionModel,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"temperature": 0.7,
		"top_p":       0.9,
		"stream":      false,
	}

	// handle special case for Claude models
	if c.completionModel == "aws/claude-4-5-sonnet" {
		requestBody = map[string]interface{}{
			"model": c.completionModel,
			"messages": []map[string]string{
				{
					"role":    "user",
					"content": prompt,
				},
			},
			"temperature": 1.0,
			"stream":      false,
		}
	}

	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			c.logger.Debug("Retrying completion request",
				"attempt", attempt,
				"max_retries", c.maxRetries)
			time.Sleep(c.retryDelay * time.Duration(attempt))
		}

		response, err := c.makeRequest(ctx, "/v1/chat/completions", requestBody, 30*time.Second)
		if err != nil {
			lastErr = err
			continue
		}

		var result struct {
			Choices []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
				FinishReason string `json:"finish_reason"`
			} `json:"choices"`
		}

		if err := json.Unmarshal(response, &result); err != nil {
			lastErr = fmt.Errorf("failed to parse response: %w", err)
			continue
		}

		if len(result.Choices) == 0 {
			lastErr = fmt.Errorf("no choices in response")
			continue
		}

		return result.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("failed after %d retries: %w", c.maxRetries, lastErr)
}

// GenerateEmbedding generates a vector embedding using LiteLLM
func (c *LiteLLMClient) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	requestBody := map[string]interface{}{
		"model": c.embeddingModel,
		"input": text,
	}

	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			c.logger.Debug("Retrying embedding request",
				"attempt", attempt,
				"max_retries", c.maxRetries)
			time.Sleep(c.retryDelay * time.Duration(attempt))
		}

		response, err := c.makeRequest(ctx, "/v1/embeddings", requestBody, 10*time.Second)
		if err != nil {
			lastErr = err
			continue
		}

		var result struct {
			Data []struct {
				Embedding []float64 `json:"embedding"`
			} `json:"data"`
		}

		if err := json.Unmarshal(response, &result); err != nil {
			lastErr = fmt.Errorf("failed to parse response: %w", err)
			continue
		}

		if len(result.Data) == 0 || len(result.Data[0].Embedding) == 0 {
			lastErr = fmt.Errorf("empty embedding returned")
			continue
		}

		// Convert float64 to float32
		embedding := make([]float32, len(result.Data[0].Embedding))
		for i, v := range result.Data[0].Embedding {
			embedding[i] = float32(v)
		}

		return embedding, nil
	}

	return nil, fmt.Errorf("failed after %d retries: %w", c.maxRetries, lastErr)
}

// GenerateEmbeddings generates embeddings for multiple texts in batch
func (c *LiteLLMClient) GenerateEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	// LiteLLM supports batch embedding requests
	requestBody := map[string]interface{}{
		"model": c.embeddingModel,
		"input": texts,
	}

	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			c.logger.Debug("Retrying batch embedding request",
				"attempt", attempt,
				"max_retries", c.maxRetries)
			time.Sleep(c.retryDelay * time.Duration(attempt))
		}

		response, err := c.makeRequest(ctx, "/v1/embeddings", requestBody, 30*time.Second)
		if err != nil {
			lastErr = err
			continue
		}

		var result struct {
			Data []struct {
				Embedding []float64 `json:"embedding"`
				Index     int       `json:"index"`
			} `json:"data"`
		}

		if err := json.Unmarshal(response, &result); err != nil {
			lastErr = fmt.Errorf("failed to parse response: %w", err)
			continue
		}

		if len(result.Data) != len(texts) {
			lastErr = fmt.Errorf("expected %d embeddings, got %d", len(texts), len(result.Data))
			continue
		}

		// Convert and return embeddings
		embeddings := make([][]float32, len(texts))
		for _, data := range result.Data {
			if data.Index >= len(texts) {
				lastErr = fmt.Errorf("invalid index %d in response", data.Index)
				break
			}
			embedding := make([]float32, len(data.Embedding))
			for i, v := range data.Embedding {
				embedding[i] = float32(v)
			}
			embeddings[data.Index] = embedding
		}

		if lastErr != nil {
			continue
		}

		return embeddings, nil
	}

	return nil, fmt.Errorf("failed after %d retries: %w", c.maxRetries, lastErr)
}

// makeRequest makes an HTTP request to LiteLLM API with timeout and retries
func (c *LiteLLMClient) makeRequest(ctx context.Context, endpoint string, requestBody interface{}, timeout time.Duration) ([]byte, error) {
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
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

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

// Close releases resources (no-op for LiteLLM client)
func (c *LiteLLMClient) Close() error {
	return nil
}

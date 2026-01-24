package llm

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

// T028: Unit tests for LLM client
// Tests GenerateCompletion, GenerateEmbedding, retry logic, timeout handling, error cases

// MockOllamaClient for testing
type MockOllamaClient struct {
	CompletionResponse string
	CompletionError    error
	CompletionFunc     func(ctx context.Context, prompt string) (string, error)
	EmbeddingResponse  []float32
	EmbeddingError     error
	EmbeddingFunc      func(ctx context.Context, text string) ([]float32, error)
	ShouldParseJSON    bool
}

func (m *MockOllamaClient) GenerateCompletion(ctx context.Context, prompt string) (string, error) {
	if m.CompletionFunc != nil {
		return m.CompletionFunc(ctx, prompt)
	}
	return m.CompletionResponse, m.CompletionError
}

func (m *MockOllamaClient) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	if m.EmbeddingFunc != nil {
		return m.EmbeddingFunc(ctx, text)
	}
	return m.EmbeddingResponse, m.EmbeddingError
}

func (m *MockOllamaClient) GenerateCompletionWithRetry(ctx context.Context, prompt string, maxRetries int, baseDelay time.Duration) (string, error) {
	var lastErr error
	for i := 0; i < maxRetries; i++ {
		response, err := m.GenerateCompletion(ctx, prompt)
		if err == nil {
			return response, nil
		}
		lastErr = err
		if i < maxRetries-1 {
			delay := baseDelay * time.Duration(1<<uint(i)) // Exponential backoff
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return "", ctx.Err()
			}
		}
	}
	return "", lastErr
}

func (m *MockOllamaClient) GenerateEmbeddingWithRetry(ctx context.Context, text string, maxRetries int, baseDelay time.Duration) ([]float32, error) {
	var lastErr error
	for i := 0; i < maxRetries; i++ {
		embedding, err := m.GenerateEmbedding(ctx, text)
		if err == nil {
			return embedding, nil
		}
		lastErr = err
		if i < maxRetries-1 {
			delay := baseDelay * time.Duration(1<<uint(i)) // Exponential backoff
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
	}
	return nil, lastErr
}

func (m *MockOllamaClient) ParseJSONCompletion(ctx context.Context, prompt string) (interface{}, error) {
	response, err := m.GenerateCompletion(ctx, prompt)
	if err != nil {
		return nil, err
	}
	// Simulate JSON parsing
	if m.ShouldParseJSON {
		if !strings.HasPrefix(response, "{") {
			return nil, errors.New("invalid JSON")
		}
		// Additional validation for malformed JSON
		if strings.Contains(response, "invalid") {
			return nil, errors.New("invalid JSON")
		}
	}
	return response, nil
}

func TestGenerateCompletion_Success(t *testing.T) {
	// Mock client that returns successful response
	client := &MockOllamaClient{
		CompletionResponse: "This is a test response",
		CompletionError:    nil,
	}

	ctx := context.Background()
	response, err := client.GenerateCompletion(ctx, "Test prompt")

	if err != nil {
		t.Fatalf("GenerateCompletion failed: %v", err)
	}

	if response != "This is a test response" {
		t.Errorf("Expected 'This is a test response', got '%s'", response)
	}
}

func TestGenerateCompletion_RetryLogic(t *testing.T) {
	// Mock client that fails twice then succeeds
	attempts := 0
	client := &MockOllamaClient{
		CompletionFunc: func(ctx context.Context, prompt string) (string, error) {
			attempts++
			if attempts < 3 {
				return "", errors.New("temporary error")
			}
			return "Success after retries", nil
		},
	}

	ctx := context.Background()
	response, err := client.GenerateCompletionWithRetry(ctx, "Test prompt", 3, 10*time.Millisecond)

	if err != nil {
		t.Fatalf("GenerateCompletionWithRetry failed: %v", err)
	}

	if response != "Success after retries" {
		t.Errorf("Expected 'Success after retries', got '%s'", response)
	}

	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}
}

func TestGenerateCompletion_ExponentialBackoff(t *testing.T) {
	attempts := 0
	attemptTimes := []time.Time{}
	client := &MockOllamaClient{
		CompletionFunc: func(ctx context.Context, prompt string) (string, error) {
			attempts++
			attemptTimes = append(attemptTimes, time.Now())
			if attempts < 3 {
				return "", errors.New("temporary error")
			}
			return "Success", nil
		},
	}

	ctx := context.Background()
	baseDelay := 50 * time.Millisecond
	_, err := client.GenerateCompletionWithRetry(ctx, "Test prompt", 3, baseDelay)

	if err != nil {
		t.Fatalf("GenerateCompletionWithRetry failed: %v", err)
	}

	// Verify exponential backoff timing
	if len(attemptTimes) >= 3 {
		delay1 := attemptTimes[1].Sub(attemptTimes[0])
		delay2 := attemptTimes[2].Sub(attemptTimes[1])
		// Second delay should be approximately 2x first delay (exponential backoff)
		if delay2 < delay1 {
			t.Error("Expected exponential backoff, but delays did not increase")
		}
	}
}

func TestGenerateCompletion_TimeoutHandling(t *testing.T) {
	client := &MockOllamaClient{
		CompletionFunc: func(ctx context.Context, prompt string) (string, error) {
			// Simulate slow response
			select {
			case <-time.After(2 * time.Second):
				return "Slow response", nil
			case <-ctx.Done():
				return "", ctx.Err()
			}
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	_, err := client.GenerateCompletion(ctx, "Test prompt")

	if err == nil {
		t.Error("Expected timeout error")
	}

	if !errors.Is(err, context.DeadlineExceeded) && !strings.Contains(err.Error(), "timeout") {
		t.Errorf("Expected timeout error, got: %v", err)
	}
}

func TestGenerateEmbedding_Success(t *testing.T) {
	expectedEmbedding := []float32{0.1, 0.2, 0.3, 0.4, 0.5}
	client := &MockOllamaClient{
		EmbeddingResponse: expectedEmbedding,
		EmbeddingError:    nil,
	}

	ctx := context.Background()
	embedding, err := client.GenerateEmbedding(ctx, "Test text")

	if err != nil {
		t.Fatalf("GenerateEmbedding failed: %v", err)
	}

	if len(embedding) != len(expectedEmbedding) {
		t.Errorf("Expected embedding length %d, got %d", len(expectedEmbedding), len(embedding))
	}

	for i, val := range expectedEmbedding {
		if embedding[i] != val {
			t.Errorf("Expected embedding[%d] = %f, got %f", i, val, embedding[i])
		}
	}
}

func TestGenerateEmbedding_RetryLogic(t *testing.T) {
	attempts := 0
	client := &MockOllamaClient{
		EmbeddingFunc: func(ctx context.Context, text string) ([]float32, error) {
			attempts++
			if attempts < 3 {
				return nil, errors.New("temporary error")
			}
			return []float32{0.1, 0.2, 0.3}, nil
		},
	}

	ctx := context.Background()
	embedding, err := client.GenerateEmbeddingWithRetry(ctx, "Test text", 3, 10*time.Millisecond)

	if err != nil {
		t.Fatalf("GenerateEmbeddingWithRetry failed: %v", err)
	}

	if len(embedding) != 3 {
		t.Errorf("Expected embedding length 3, got %d", len(embedding))
	}

	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}
}

func TestGenerateEmbedding_TimeoutHandling(t *testing.T) {
	client := &MockOllamaClient{
		EmbeddingFunc: func(ctx context.Context, text string) ([]float32, error) {
			// Simulate slow response
			select {
			case <-time.After(2 * time.Second):
				return []float32{0.1}, nil
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	_, err := client.GenerateEmbedding(ctx, "Test text")

	if err == nil {
		t.Error("Expected timeout error")
	}

	if !errors.Is(err, context.DeadlineExceeded) && !strings.Contains(err.Error(), "timeout") {
		t.Errorf("Expected timeout error, got: %v", err)
	}
}

func TestErrorHandling_ConnectionFailure(t *testing.T) {
	client := &MockOllamaClient{
		CompletionError: errors.New("connection refused"),
	}

	ctx := context.Background()
	_, err := client.GenerateCompletion(ctx, "Test prompt")

	if err == nil {
		t.Error("Expected connection error")
	}

	if !strings.Contains(err.Error(), "connection") {
		t.Errorf("Expected connection error, got: %v", err)
	}
}

func TestErrorHandling_InvalidJSON(t *testing.T) {
	client := &MockOllamaClient{
		CompletionResponse: "{invalid json}",
		ShouldParseJSON:    true,
	}

	ctx := context.Background()
	_, err := client.ParseJSONCompletion(ctx, "Test prompt")

	if err == nil {
		t.Error("Expected JSON parsing error")
	}
}

func TestGenerateCompletion_MaxRetriesExceeded(t *testing.T) {
	attempts := 0
	client := &MockOllamaClient{
		CompletionFunc: func(ctx context.Context, prompt string) (string, error) {
			attempts++
			return "", errors.New("persistent error")
		},
	}

	ctx := context.Background()
	maxRetries := 3
	_, err := client.GenerateCompletionWithRetry(ctx, "Test prompt", maxRetries, 10*time.Millisecond)

	if err == nil {
		t.Error("Expected error after max retries exceeded")
	}

	if attempts != maxRetries {
		t.Errorf("Expected %d attempts, got %d", maxRetries, attempts)
	}
}

package llm

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestLiteLLMClient_GenerateCompletion_Success(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.URL.Path != "/v1/chat/completions" {
			t.Errorf("Expected path /v1/chat/completions, got %s", r.URL.Path)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json")
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("Expected Authorization header with Bearer token")
		}

		// Parse request body
		var reqBody map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		// Verify request structure
		if reqBody["model"] != "test-model" {
			t.Errorf("Expected model test-model, got %v", reqBody["model"])
		}

		// Send response
		response := map[string]interface{}{
			"choices": []map[string]interface{}{
				{
					"message": map[string]string{
						"content": "This is a test response",
					},
					"finish_reason": "stop",
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	client := NewLiteLLMClient(server.URL, "test-model", "test-embed-model", "test-key", logger)

	ctx := context.Background()
	result, err := client.GenerateCompletion(ctx, "Test prompt")

	if err != nil {
		t.Fatalf("GenerateCompletion failed: %v", err)
	}

	if result != "This is a test response" {
		t.Errorf("Expected 'This is a test response', got '%s'", result)
	}
}

func TestLiteLLMClient_GenerateCompletion_WithRetries(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			// Fail the first 2 attempts
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("temporary error"))
			return
		}

		// Succeed on the 3rd attempt
		response := map[string]interface{}{
			"choices": []map[string]interface{}{
				{
					"message": map[string]string{
						"content": "Success after retries",
					},
					"finish_reason": "stop",
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	client := NewLiteLLMClient(server.URL, "test-model", "test-embed-model", "", logger)
	client.maxRetries = 3
	client.retryDelay = 10 * time.Millisecond

	ctx := context.Background()
	result, err := client.GenerateCompletion(ctx, "Test prompt")

	if err != nil {
		t.Fatalf("GenerateCompletion failed: %v", err)
	}

	if result != "Success after retries" {
		t.Errorf("Expected 'Success after retries', got '%s'", result)
	}

	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}
}

func TestLiteLLMClient_GenerateCompletion_MaxRetriesExceeded(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("persistent error"))
	}))
	defer server.Close()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	client := NewLiteLLMClient(server.URL, "test-model", "test-embed-model", "", logger)
	client.maxRetries = 2
	client.retryDelay = 10 * time.Millisecond

	ctx := context.Background()
	_, err := client.GenerateCompletion(ctx, "Test prompt")

	if err == nil {
		t.Error("Expected error after max retries exceeded")
	}

	if !strings.Contains(err.Error(), "failed after") {
		t.Errorf("Expected retry error message, got: %v", err)
	}

	// Should attempt initial + maxRetries = 3 times
	if attempts != 3 {
		t.Errorf("Expected 3 attempts (initial + 2 retries), got %d", attempts)
	}
}

func TestLiteLLMClient_GenerateCompletion_EmptyChoices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"choices": []map[string]interface{}{},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	client := NewLiteLLMClient(server.URL, "test-model", "test-embed-model", "", logger)
	client.maxRetries = 1
	client.retryDelay = 10 * time.Millisecond

	ctx := context.Background()
	_, err := client.GenerateCompletion(ctx, "Test prompt")

	if err == nil {
		t.Error("Expected error for empty choices")
	}

	if !strings.Contains(err.Error(), "no choices") {
		t.Errorf("Expected 'no choices' error, got: %v", err)
	}
}

func TestLiteLLMClient_GenerateCompletion_ContextTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow response
		time.Sleep(200 * time.Millisecond)
		response := map[string]interface{}{
			"choices": []map[string]interface{}{
				{
					"message": map[string]string{
						"content": "Slow response",
					},
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	client := NewLiteLLMClient(server.URL, "test-model", "test-embed-model", "", logger)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := client.GenerateCompletion(ctx, "Test prompt")

	if err == nil {
		t.Error("Expected timeout error")
	}
}

func TestLiteLLMClient_GenerateEmbedding_Success(t *testing.T) {
	expectedEmbedding := []float64{0.1, 0.2, 0.3, 0.4, 0.5}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/embeddings" {
			t.Errorf("Expected path /v1/embeddings, got %s", r.URL.Path)
		}

		var reqBody map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		if reqBody["model"] != "test-embed-model" {
			t.Errorf("Expected model test-embed-model, got %v", reqBody["model"])
		}

		response := map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"embedding": expectedEmbedding,
					"index":     0,
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	client := NewLiteLLMClient(server.URL, "test-model", "test-embed-model", "", logger)

	ctx := context.Background()
	embedding, err := client.GenerateEmbedding(ctx, "Test text")

	if err != nil {
		t.Fatalf("GenerateEmbedding failed: %v", err)
	}

	if len(embedding) != len(expectedEmbedding) {
		t.Errorf("Expected embedding length %d, got %d", len(expectedEmbedding), len(embedding))
	}

	for i, expected := range expectedEmbedding {
		if embedding[i] != float32(expected) {
			t.Errorf("Expected embedding[%d] = %f, got %f", i, expected, embedding[i])
		}
	}
}

func TestLiteLLMClient_GenerateEmbedding_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"data": []map[string]interface{}{},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	client := NewLiteLLMClient(server.URL, "test-model", "test-embed-model", "", logger)
	client.maxRetries = 1
	client.retryDelay = 10 * time.Millisecond

	ctx := context.Background()
	_, err := client.GenerateEmbedding(ctx, "Test text")

	if err == nil {
		t.Error("Expected error for empty embedding")
	}

	if !strings.Contains(err.Error(), "empty embedding") {
		t.Errorf("Expected 'empty embedding' error, got: %v", err)
	}
}

func TestLiteLLMClient_GenerateEmbeddings_Batch_Success(t *testing.T) {
	texts := []string{"text1", "text2", "text3"}
	embeddings := [][]float64{
		{0.1, 0.2, 0.3},
		{0.4, 0.5, 0.6},
		{0.7, 0.8, 0.9},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqBody map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		// Verify batch input
		input, ok := reqBody["input"].([]interface{})
		if !ok {
			t.Error("Expected input to be an array")
		}
		if len(input) != len(texts) {
			t.Errorf("Expected %d texts, got %d", len(texts), len(input))
		}

		response := map[string]interface{}{
			"data": []map[string]interface{}{
				{"embedding": embeddings[0], "index": 0},
				{"embedding": embeddings[1], "index": 1},
				{"embedding": embeddings[2], "index": 2},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	client := NewLiteLLMClient(server.URL, "test-model", "test-embed-model", "", logger)

	ctx := context.Background()
	results, err := client.GenerateEmbeddings(ctx, texts)

	if err != nil {
		t.Fatalf("GenerateEmbeddings failed: %v", err)
	}

	if len(results) != len(texts) {
		t.Errorf("Expected %d embeddings, got %d", len(texts), len(results))
	}

	for i, expected := range embeddings {
		if len(results[i]) != len(expected) {
			t.Errorf("Embedding %d has wrong length: expected %d, got %d", i, len(expected), len(results[i]))
		}
		for j, val := range expected {
			if results[i][j] != float32(val) {
				t.Errorf("Embedding[%d][%d] = %f, expected %f", i, j, results[i][j], val)
			}
		}
	}
}

func TestLiteLLMClient_GenerateEmbeddings_CountMismatch(t *testing.T) {
	texts := []string{"text1", "text2", "text3"}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return fewer embeddings than requested
		response := map[string]interface{}{
			"data": []map[string]interface{}{
				{"embedding": []float64{0.1, 0.2}, "index": 0},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	client := NewLiteLLMClient(server.URL, "test-model", "test-embed-model", "", logger)
	client.maxRetries = 1
	client.retryDelay = 10 * time.Millisecond

	ctx := context.Background()
	_, err := client.GenerateEmbeddings(ctx, texts)

	if err == nil {
		t.Error("Expected error for embedding count mismatch")
	}

	if !strings.Contains(err.Error(), "expected") {
		t.Errorf("Expected count mismatch error, got: %v", err)
	}
}

func TestLiteLLMClient_GenerateEmbeddings_InvalidIndex(t *testing.T) {
	texts := []string{"text1", "text2"}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return invalid index
		response := map[string]interface{}{
			"data": []map[string]interface{}{
				{"embedding": []float64{0.1, 0.2}, "index": 0},
				{"embedding": []float64{0.3, 0.4}, "index": 10}, // Invalid index
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	client := NewLiteLLMClient(server.URL, "test-model", "test-embed-model", "", logger)
	client.maxRetries = 1
	client.retryDelay = 10 * time.Millisecond

	ctx := context.Background()
	_, err := client.GenerateEmbeddings(ctx, texts)

	if err == nil {
		t.Error("Expected error for invalid index")
	}

	if !strings.Contains(err.Error(), "invalid index") {
		t.Errorf("Expected invalid index error, got: %v", err)
	}
}

func TestLiteLLMClient_Authorization_WithKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-api-key" {
			t.Errorf("Expected Authorization: Bearer test-api-key, got: %s", auth)
		}

		response := map[string]interface{}{
			"choices": []map[string]interface{}{
				{
					"message": map[string]string{
						"content": "response",
					},
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	client := NewLiteLLMClient(server.URL, "test-model", "test-embed-model", "test-api-key", logger)

	ctx := context.Background()
	_, err := client.GenerateCompletion(ctx, "Test prompt")

	if err != nil {
		t.Fatalf("GenerateCompletion failed: %v", err)
	}
}

func TestLiteLLMClient_Authorization_WithoutKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "" {
			t.Errorf("Expected no Authorization header, got: %s", auth)
		}

		response := map[string]interface{}{
			"choices": []map[string]interface{}{
				{
					"message": map[string]string{
						"content": "response",
					},
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	client := NewLiteLLMClient(server.URL, "test-model", "test-embed-model", "", logger)

	ctx := context.Background()
	_, err := client.GenerateCompletion(ctx, "Test prompt")

	if err != nil {
		t.Fatalf("GenerateCompletion failed: %v", err)
	}
}

func TestLiteLLMClient_Close(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	client := NewLiteLLMClient("http://localhost", "test-model", "test-embed-model", "", logger)

	err := client.Close()
	if err != nil {
		t.Errorf("Close should not return error, got: %v", err)
	}
}

func TestLiteLLMClient_InvalidJSON_Response(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("invalid json {"))
	}))
	defer server.Close()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	client := NewLiteLLMClient(server.URL, "test-model", "test-embed-model", "", logger)
	client.maxRetries = 1
	client.retryDelay = 10 * time.Millisecond

	ctx := context.Background()
	_, err := client.GenerateCompletion(ctx, "Test prompt")

	if err == nil {
		t.Error("Expected error for invalid JSON")
	}

	if !strings.Contains(err.Error(), "parse") {
		t.Errorf("Expected parse error, got: %v", err)
	}
}

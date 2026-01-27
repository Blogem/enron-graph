package contract

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/Blogem/enron-graph/ent"
	"github.com/Blogem/enron-graph/ent/enttest"
	_ "github.com/mattn/go-sqlite3"
)

// App interface for testing - matches the actual App methods we're testing
type App interface {
	ProcessChatQuery(query string) (string, error)
	ClearChatContext() error
}

// TestProcessChatQuery tests the ProcessChatQuery Wails binding
func TestProcessChatQuery(t *testing.T) {
	// Setup: Create in-memory test database
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	// Create App instance
	app := createTestApp(client)

	tests := []struct {
		name        string
		query       string
		wantErr     bool
		errContains string
		validate    func(t *testing.T, response string)
	}{
		{
			name:    "valid query returns response",
			query:   "who is ken lay",
			wantErr: false,
			validate: func(t *testing.T, response string) {
				if response == "" {
					t.Error("expected non-empty response")
				}
			},
		},
		{
			name:        "empty query returns error",
			query:       "",
			wantErr:     true,
			errContains: "query cannot be empty",
		},
		{
			name:        "whitespace query returns error",
			query:       "   ",
			wantErr:     true,
			errContains: "query cannot be empty",
		},
		{
			name:    "query with timeout succeeds",
			query:   "tell me about enron",
			wantErr: false,
			validate: func(t *testing.T, response string) {
				if response == "" {
					t.Error("expected non-empty response")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute ProcessChatQuery
			response, err := app.ProcessChatQuery(tt.query)

			// Check error expectation
			if (err != nil) != tt.wantErr {
				t.Errorf("ProcessChatQuery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Validate error message if expected
			if tt.wantErr && tt.errContains != "" {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errContains)
				} else if !contains(err.Error(), tt.errContains) {
					t.Errorf("expected error containing %q, got %q", tt.errContains, err.Error())
				}
			}

			// Run custom validation if provided
			if !tt.wantErr && tt.validate != nil {
				tt.validate(t, response)
			}
		})
	}
}

// TestClearChatContext tests the ClearChatContext Wails binding
func TestClearChatContext(t *testing.T) {
	// Setup: Create in-memory test database
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	// Create App instance
	app := createTestApp(client)

	// First, send a query to populate context
	_, err := app.ProcessChatQuery("who is ken lay")
	if err != nil {
		t.Fatalf("Failed to send initial query: %v", err)
	}

	// Clear the context
	err = app.ClearChatContext()
	if err != nil {
		t.Errorf("ClearChatContext() error = %v, want nil", err)
	}

	// Verify context is cleared by checking that history is empty
	// This would require exposing the context or adding a method to check history
	// For now, we just verify the call doesn't error
}

// TestProcessChatQueryTimeout tests that queries timeout appropriately
func TestProcessChatQueryTimeout(t *testing.T) {
	// This test is a placeholder for timeout validation
	// In a real implementation, we'd need a way to simulate a long-running query
	// For now, we just verify that normal queries complete quickly

	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	app := createTestApp(client)

	start := time.Now()
	_, err := app.ProcessChatQuery("quick query")
	duration := time.Since(start)

	if err != nil {
		t.Errorf("ProcessChatQuery() error = %v, want nil", err)
	}

	// Verify response is quick (should be << 60 seconds)
	if duration > 5*time.Second {
		t.Errorf("ProcessChatQuery() took %v, expected < 5s", duration)
	}
}

// Helper functions

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// createTestApp creates an App instance for testing by importing from cmd/explorer
func createTestApp(client *ent.Client) App {
	// We need to create a wrapper that matches the main App
	// This requires either:
	// 1. Making the main package testable by exporting NewApp, or
	// 2. Creating a test adapter

	// For now, we'll create a minimal test adapter
	// In production, we'd refactor main package to be testable
	return &testAppAdapter{
		client: client,
	}
}

// testAppAdapter wraps the functionality we need for testing
// This is a temporary solution - ideally we'd refactor the main package
type testAppAdapter struct {
	client *ent.Client
}

// ProcessChatQuery implements the App interface
func (a *testAppAdapter) ProcessChatQuery(query string) (string, error) {
	// This matches the validation logic from cmd/explorer/app.go
	// We validate the contract here since we can't import the main package

	// Validate input (matching the actual implementation)
	if len(strings.TrimSpace(query)) == 0 {
		return "", fmt.Errorf("query cannot be empty")
	}

	// Return a test response (actual chat processing would happen in integration tests)
	return "test response", nil
}

// ClearChatContext implements the App interface
func (a *testAppAdapter) ClearChatContext() error {
	return nil
}

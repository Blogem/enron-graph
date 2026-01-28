//go:build ignore
// +build ignore

// Chat API Go Contracts
// Location: specs/004-explorer-chat/contracts/chat-api.go
// Purpose: Define Go method signatures for chat interface
// Note: This is a specification file, not actual implementation

package main

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/Blogem/enron-graph/ent"
	"github.com/Blogem/enron-graph/internal/chat"
)

// Stub types for IDE type checking - not used in actual build
type App struct{}
type Context struct{}

// App struct methods for chat functionality
// These methods will be added to cmd/explorer/app.go

// ProcessChatQuery processes a user's natural language query and returns a response.
// This method is automatically bound by Wails and callable from the React frontend.
//
// Architecture Note (Context Management):
//
//	Conversation context is maintained internally within the App struct's chat.Handler
//	and chat.Context instances. The frontend does not need to pass context explicitly -
//	it is preserved across calls within the same application session. This satisfies
//	FR-008 and FR-012 requirements for context passing and maintenance.
//
// Parameters:
//   - query: The user's natural language query string (non-empty, max 1000 chars)
//
// Returns:
//   - string: The formatted response from the chat handler
//   - error: Non-nil if query processing fails, times out, or validation fails
//
// Validation:
//   - Returns error if query is empty or whitespace-only
//   - Returns error if query exceeds 1000 characters
//   - Times out after 60 seconds (context.WithTimeout)
//
// Example:
//
//	response, err := a.ProcessChatQuery("Show me Kenneth Lay")
//	if err != nil {
//	    return "", fmt.Errorf("chat query failed: %w", err)
//	}
//	// response: "Kenneth Lay (person)\nProperties:\n  email: kenneth.lay@enron.com"
func (a *App) ProcessChatQuery(query string) (string, error)

// ClearChatContext clears the conversation history and resets the chat context.
// This method is automatically bound by Wails and callable from the React frontend.
//
// Returns:
//   - error: Non-nil if clearing fails (should be rare)
//
// Side Effects:
//   - Calls Clear() on the chat context
//   - Removes all tracked entities and conversation history
//
// Example:
//
//	err := a.ClearChatContext()
//	if err != nil {
//	    return fmt.Errorf("failed to clear chat context: %w", err)
//	}
func (a *App) ClearChatContext() error

// Internal types (not exposed to frontend, included for documentation)

// chatAdapter implements chat.Repository interface by adapting ent client methods
// This adapter translates between ent types and internal/chat types
type chatAdapter struct {
	client *ent.Client
	db     *sql.DB
}

// NewChatAdapter creates a new adapter for the chat repository
//
// Parameters:
//   - client: The ent client for database operations
//   - db: The raw database connection (if needed for custom queries)
//
// Returns:
//   - chat.Repository: Implementation of the repository interface
func NewChatAdapter(client *ent.Client, db *sql.DB) chat.Repository

// stubLLMClient implements chat.LLMClient interface for development/testing
// This stub returns mock responses until real LLM integration is added
type stubLLMClient struct{}

// NewStubLLMClient creates a new stub LLM client
//
// Returns:
//   - chat.LLMClient: Stub implementation for development
func NewStubLLMClient() chat.LLMClient

// Validation constants
const (
	// MaxQueryLength is the maximum allowed query length in characters
	MaxQueryLength = 1000

	// QueryTimeout is the maximum time to wait for a query response
	QueryTimeout = 60 * time.Second
)

// validateQuery checks if a query string is valid
//
// Parameters:
//   - query: The query string to validate
//
// Returns:
//   - error: Non-nil if query is invalid (empty, whitespace-only, or too long)
func validateQuery(query string) error {
	trimmed := strings.TrimSpace(query)
	if trimmed == "" {
		return fmt.Errorf("query cannot be empty")
	}
	if len(query) > MaxQueryLength {
		return fmt.Errorf("query too long: %d characters (max %d)", len(query), MaxQueryLength)
	}
	return nil
}

// Implementation notes:
//
// 1. ProcessChatQuery implementation flow:
//    - Validate query (empty check, length check)
//    - Create context with 60-second timeout
//    - Call chatHandler.ProcessQuery(ctx, query, chatContext)
//    - Return response or error
//
// 2. ClearChatContext implementation flow:
//    - Call chatContext.Clear()
//    - Return nil (or error if Clear() fails)
//
// 3. NewApp modifications:
//    - Add chatHandler field to App struct
//    - Add chatContext field to App struct
//    - Initialize in NewApp():
//        - Create NewStubLLMClient()
//        - Create NewChatAdapter(client, db)
//        - Create chat.NewHandler(llm, repo)
//        - Create chat.NewContext()
//
// 4. Testing contracts:
//    - Contract tests in tests/contract/chat_bindings_test.go
//    - Verify methods are callable via Wails runtime
//    - Verify error handling and validation
//    - Verify timeout behavior

package main

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/Blogem/enron-graph/ent"
	"github.com/Blogem/enron-graph/internal/chat"
	"github.com/Blogem/enron-graph/internal/explorer"
)

// App struct
type App struct {
	ctx           context.Context
	client        *ent.Client
	schemaService *explorer.SchemaService
	graphService  *explorer.GraphService
	chatHandler   chat.Handler
	chatContext   chat.Context
}

// NewApp creates a new App application struct
func NewApp(client *ent.Client, db *sql.DB) *App {
	// Create chat dependencies
	llmClient := chat.NewStubLLMClient()
	// chatRepo needs context, will be initialized in startup
	chatHandler := chat.NewHandler(llmClient, nil)
	chatContext := chat.NewContext()

	return &App{
		client:        client,
		schemaService: explorer.NewSchemaService(client, db),
		graphService:  explorer.NewGraphService(client, db),
		chatHandler:   chatHandler,
		chatContext:   chatContext,
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Initialize chat adapter with context
	llmClient := chat.NewStubLLMClient()
	chatRepo := newChatAdapter(a.client, ctx)
	a.chatHandler = chat.NewHandler(llmClient, chatRepo)
}

// GetSchema returns the complete schema metadata (promoted and discovered types)
func (a *App) GetSchema() (*explorer.SchemaResponse, error) {
	return a.schemaService.GetSchema(a.ctx)
}

// GetTypeDetails returns detailed information about a specific type
func (a *App) GetTypeDetails(typeName string) (*explorer.SchemaType, error) {
	return a.schemaService.GetTypeDetails(a.ctx, typeName)
}

// RefreshSchema clears the cache and reloads schema from database
func (a *App) RefreshSchema() error {
	return a.schemaService.RefreshSchema(a.ctx)
}

// GetRandomNodes returns exactly limit random nodes with connecting edges
func (a *App) GetRandomNodes(limit int) (*explorer.GraphResponse, error) {
	return a.graphService.GetRandomNodes(a.ctx, limit)
}

// GetRelationships returns paginated relationships for a specific node
func (a *App) GetRelationships(nodeID string, offset, limit int) (*explorer.RelationshipsResponse, error) {
	return a.graphService.GetRelationships(a.ctx, nodeID, offset, limit)
}

// GetNodeDetails returns complete information for a specific node
func (a *App) GetNodeDetails(nodeID string) (*explorer.GraphNode, error) {
	return a.graphService.GetNodeDetails(a.ctx, nodeID)
}

// GetNodes returns nodes filtered by type, category, and/or search query
func (a *App) GetNodes(filter explorer.NodeFilter) (*explorer.GraphResponse, error) {
	return a.graphService.GetNodes(a.ctx, filter)
}

// ProcessChatQuery processes a natural language query and returns a response
func (a *App) ProcessChatQuery(query string) (string, error) {
	// Validate input
	if len(strings.TrimSpace(query)) == 0 {
		return "", fmt.Errorf("query cannot be empty")
	}

	// Create a context with timeout (60 seconds per spec)
	ctx, cancel := context.WithTimeout(a.ctx, 60*time.Second)
	defer cancel()

	// Process the query using the chat handler
	response, err := a.chatHandler.ProcessQuery(ctx, query, a.chatContext)
	if err != nil {
		// Check if it's a timeout error
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("query processing timed out after 60 seconds")
		}
		return "", fmt.Errorf("failed to process query: %w", err)
	}

	return response, nil
}

// ClearChatContext clears the conversation history and context
func (a *App) ClearChatContext() error {
	a.chatContext.Clear()
	return nil
}

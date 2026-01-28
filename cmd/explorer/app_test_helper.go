package main

import (
	"database/sql"

	"github.com/Blogem/enron-graph/ent"
	"github.com/Blogem/enron-graph/internal/chat"
	"github.com/Blogem/enron-graph/internal/explorer"
	"github.com/Blogem/enron-graph/pkg/llm"
	"github.com/Blogem/enron-graph/pkg/utils"
)

// NewAppWithStub creates a new App with a stub LLM client for testing
func NewAppWithStub(client *ent.Client, db *sql.DB, cfg *utils.Config, llmClient llm.Client) *App {
	// Create chat dependencies with stub
	llmClientChat := chat.NewStubLLMClient()
	chatHandler := chat.NewHandler(llmClientChat, nil)
	chatContext := chat.NewContext()

	return &App{
		client:        client,
		config:        cfg,
		schemaService: explorer.NewSchemaService(client, db),
		graphService:  explorer.NewGraphService(client, db, llmClient),
		chatHandler:   chatHandler,
		chatContext:   chatContext,
	}
}

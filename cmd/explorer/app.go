package main

import (
	"context"
	"database/sql"

	"github.com/Blogem/enron-graph/ent"
	"github.com/Blogem/enron-graph/internal/explorer"
)

// App struct
type App struct {
	ctx           context.Context
	client        *ent.Client
	schemaService *explorer.SchemaService
	graphService  *explorer.GraphService
}

// NewApp creates a new App application struct
func NewApp(client *ent.Client, db *sql.DB) *App {
	return &App{
		client:        client,
		schemaService: explorer.NewSchemaService(client, db),
		graphService:  explorer.NewGraphService(client, db),
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
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

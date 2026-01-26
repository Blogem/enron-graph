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
}

// NewApp creates a new App application struct
func NewApp(client *ent.Client, db *sql.DB) *App {
	return &App{
		client:        client,
		schemaService: explorer.NewSchemaService(client, db),
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

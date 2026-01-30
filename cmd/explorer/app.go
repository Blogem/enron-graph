package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Blogem/enron-graph/ent"
	"github.com/Blogem/enron-graph/internal/analyst"
	"github.com/Blogem/enron-graph/internal/chat"
	"github.com/Blogem/enron-graph/internal/explorer"
	"github.com/Blogem/enron-graph/internal/promoter"
	"github.com/Blogem/enron-graph/pkg/llm"
	"github.com/Blogem/enron-graph/pkg/utils"
)

// App struct
type App struct {
	ctx           context.Context
	client        *ent.Client
	db            *sql.DB
	config        *utils.Config
	schemaService *explorer.SchemaService
	graphService  *explorer.GraphService
	chatHandler   chat.Handler
	chatContext   chat.Context
}

// AnalysisRequest contains parameters for entity analysis
type AnalysisRequest struct {
	MinOccurrences int     `json:"minOccurrences"`
	MinConsistency float64 `json:"minConsistency"`
	TopN           int     `json:"topN"`
}

// AnalysisResponse contains the results of entity analysis
type AnalysisResponse struct {
	Candidates []TypeCandidate `json:"candidates"`
	TotalTypes int             `json:"totalTypes"`
}

// TypeCandidate represents a candidate type for promotion
type TypeCandidate struct {
	Rank        int     `json:"rank"`
	TypeName    string  `json:"typeName"`
	Frequency   int     `json:"frequency"`
	Density     float64 `json:"density"`
	Consistency float64 `json:"consistency"`
	Score       float64 `json:"score"`
}

// PromotionRequest contains the type name to promote
type PromotionRequest struct {
	TypeName string `json:"typeName"`
}

// PromotionResponse contains the results of entity promotion
type PromotionResponse struct {
	Success          bool           `json:"success"`
	SchemaFilePath   string         `json:"schemaFilePath"`
	EntitiesMigrated int            `json:"entitiesMigrated"`
	ValidationErrors int            `json:"validationErrors"`
	Error            string         `json:"error,omitempty"`
	Properties       []PropertyInfo `json:"properties"`
}

// PropertyInfo describes a property in the promoted schema
type PropertyInfo struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
}

// NewApp creates a new App application struct
func NewApp(client *ent.Client, db *sql.DB, cfg *utils.Config, llmClient llm.Client) *App {
	// Create chat dependencies
	// TODO: avoid creating multiple LLM clients
	llmClientPrd := newProductionLLMClient(cfg)
	// chatRepo needs context, will be initialized in startup
	chatHandler := chat.NewHandler(llmClientPrd, nil)
	chatContext := chat.NewContext()

	return &App{
		client:        client,
		db:            db,
		config:        cfg,
		schemaService: explorer.NewSchemaService(client, db),
		graphService:  explorer.NewGraphService(client, db, llmClient),
		chatHandler:   chatHandler,
		chatContext:   chatContext,
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Initialize chat adapter with context
	llmClient := newProductionLLMClient(a.config)
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

// AnalyzeEntities analyzes discovered entities and returns ranked candidates for type promotion
func (a *App) AnalyzeEntities(req AnalysisRequest) (*AnalysisResponse, error) {
	// Validate parameters
	if req.MinOccurrences < 1 {
		return nil, fmt.Errorf("Minimum occurrences must be at least 1")
	}
	if req.MinConsistency < 0.0 || req.MinConsistency > 1.0 {
		return nil, fmt.Errorf("Consistency must be between 0.0 and 1.0")
	}
	if req.TopN < 1 {
		return nil, fmt.Errorf("Top N must be at least 1")
	}

	// Call analyst package to analyze and rank candidates
	candidates, err := analyst.AnalyzeAndRankCandidates(
		a.ctx,
		a.client,
		req.MinOccurrences,
		req.MinConsistency,
		req.TopN,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze entities: %w", err)
	}

	// Transform analyst.TypeCandidate to our TypeCandidate format
	responseCandidates := make([]TypeCandidate, len(candidates))
	for i, candidate := range candidates {
		responseCandidates[i] = TypeCandidate{
			Rank:        i + 1,
			TypeName:    candidate.Type,
			Frequency:   candidate.Frequency,
			Density:     candidate.Density,
			Consistency: candidate.Consistency,
			Score:       candidate.Score,
		}
	}

	return &AnalysisResponse{
		Candidates: responseCandidates,
		TotalTypes: len(responseCandidates),
	}, nil
}

// PromoteEntity promotes a discovered entity type to a formal Ent schema
func (a *App) PromoteEntity(req PromotionRequest) (*PromotionResponse, error) {
	// Validate type name
	if strings.TrimSpace(req.TypeName) == "" {
		return &PromotionResponse{Success: false, Error: "Type name cannot be empty"},
			fmt.Errorf("Type name cannot be empty")
	}

	// Calculate project root
	projectRoot, err := a.calculateProjectRoot()
	if err != nil {
		return &PromotionResponse{Success: false, Error: err.Error()},
			fmt.Errorf("failed to calculate project root: %w", err)
	}

	// Generate schema for the type using analyst package
	schema, err := analyst.GenerateSchemaForType(a.ctx, a.client, req.TypeName)
	if err != nil {
		return &PromotionResponse{Success: false, Error: err.Error()},
			fmt.Errorf("failed to generate schema: %w", err)
	}

	// Convert analyst.SchemaDefinition to promoter.SchemaDefinition
	promoterSchema := convertSchemaDefinition(schema)

	// Create promoter request
	promoterReq := promoter.PromotionRequest{
		TypeName:         req.TypeName,
		SchemaDefinition: promoterSchema,
		OutputDir:        filepath.Join(projectRoot, "ent", "schema"),
		ProjectRoot:      projectRoot,
	}

	// Execute promotion workflow
	promo := promoter.NewPromoter(a.client)
	promo.SetDB(a.db)
	result, err := promo.PromoteType(a.ctx, promoterReq)

	// Build response
	response := &PromotionResponse{
		Success:          result.Success,
		SchemaFilePath:   result.SchemaFilePath,
		EntitiesMigrated: result.EntitiesMigrated,
		ValidationErrors: result.ValidationErrors,
		Properties:       convertPropertiesToPropertyInfo(schema),
	}

	if err != nil {
		response.Error = err.Error()
		return response, err
	}

	return response, nil
}

// calculateProjectRoot calculates the project root directory
func (a *App) calculateProjectRoot() (string, error) {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Navigate up from current directory to find project root (where go.mod exists)
	// The explorer binary might be running from different locations
	current := cwd
	for {
		// Check if go.mod exists in current directory
		if _, err := os.Stat(filepath.Join(current, "go.mod")); err == nil {
			return current, nil
		}

		// Go up one level
		parent := filepath.Dir(current)
		if parent == current {
			// Reached root without finding go.mod
			return "", fmt.Errorf("could not find project root (go.mod not found)")
		}
		current = parent
	}
}

// convertSchemaDefinition converts analyst.SchemaDefinition to promoter.SchemaDefinition
func convertSchemaDefinition(schema *analyst.SchemaDefinition) promoter.SchemaDefinition {
	properties := make(map[string]promoter.PropertyDefinition)

	for propName, propDef := range schema.Properties {
		// Convert validation rules
		var validationRules []promoter.ValidationRule
		for _, rule := range propDef.ValidationRules {
			validationRules = append(validationRules, promoter.ValidationRule{
				Type:  rule.Type,
				Value: rule.Value,
			})
		}

		properties[propName] = promoter.PropertyDefinition{
			Type:            propDef.Type,
			Required:        propDef.Required,
			ValidationRules: validationRules,
		}
	}

	return promoter.SchemaDefinition{
		Type:       schema.Type,
		Properties: properties,
	}
}

// convertPropertiesToPropertyInfo converts schema properties to PropertyInfo for response
func convertPropertiesToPropertyInfo(schema *analyst.SchemaDefinition) []PropertyInfo {
	var properties []PropertyInfo

	for propName, propDef := range schema.Properties {
		properties = append(properties, PropertyInfo{
			Name:     propName,
			Type:     propDef.Type,
			Required: propDef.Required,
		})
	}

	return properties
}

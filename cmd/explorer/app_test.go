package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/Blogem/enron-graph/ent"
	"github.com/Blogem/enron-graph/ent/enttest"
	"github.com/Blogem/enron-graph/pkg/utils"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestApp(t *testing.T) (*App, *ent.Client, *sql.DB) {
	// Create test database
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")

	// Open raw SQL connection
	db, err := sql.Open("sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	require.NoError(t, err)

	// Create test config
	cfg := &utils.Config{
		DatabaseURL: "file:ent?mode=memory&cache=shared&_fk=1",
	}

	// Create app
	app := &App{
		client: client,
		db:     db,
		config: cfg,
		ctx:    context.Background(),
	}

	return app, client, db
}

// Test 2.1: AnalyzeEntities with valid parameters
func TestAnalyzeEntities_ValidParameters(t *testing.T) {
	app, client, db := setupTestApp(t)
	defer client.Close()
	defer db.Close()

	req := AnalysisRequest{
		MinOccurrences: 5,
		MinConsistency: 0.4,
		TopN:           10,
	}

	resp, err := app.AnalyzeEntities(req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.GreaterOrEqual(t, resp.TotalTypes, 0)
	assert.NotNil(t, resp.Candidates)
}

// Test 2.2: AnalyzeEntities parameter validation
func TestAnalyzeEntities_ParameterValidation(t *testing.T) {
	app, client, db := setupTestApp(t)
	defer client.Close()
	defer db.Close()

	tests := []struct {
		name        string
		req         AnalysisRequest
		expectedErr string
	}{
		{
			name: "negative occurrences",
			req: AnalysisRequest{
				MinOccurrences: -1,
				MinConsistency: 0.4,
				TopN:           10,
			},
			expectedErr: "Minimum occurrences must be at least 1",
		},
		{
			name: "zero occurrences",
			req: AnalysisRequest{
				MinOccurrences: 0,
				MinConsistency: 0.4,
				TopN:           10,
			},
			expectedErr: "Minimum occurrences must be at least 1",
		},
		{
			name: "consistency below range",
			req: AnalysisRequest{
				MinOccurrences: 5,
				MinConsistency: -0.1,
				TopN:           10,
			},
			expectedErr: "Consistency must be between 0.0 and 1.0",
		},
		{
			name: "consistency above range",
			req: AnalysisRequest{
				MinOccurrences: 5,
				MinConsistency: 1.1,
				TopN:           10,
			},
			expectedErr: "Consistency must be between 0.0 and 1.0",
		},
		{
			name: "negative topN",
			req: AnalysisRequest{
				MinOccurrences: 5,
				MinConsistency: 0.4,
				TopN:           -1,
			},
			expectedErr: "Top N must be at least 1",
		},
		{
			name: "zero topN",
			req: AnalysisRequest{
				MinOccurrences: 5,
				MinConsistency: 0.4,
				TopN:           0,
			},
			expectedErr: "Top N must be at least 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := app.AnalyzeEntities(tt.req)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

// Test 2.3: AnalyzeEntities with database error
func TestAnalyzeEntities_DatabaseError(t *testing.T) {
	app, client, db := setupTestApp(t)
	client.Close() // Close client to simulate database error
	defer db.Close()

	req := AnalysisRequest{
		MinOccurrences: 5,
		MinConsistency: 0.4,
		TopN:           10,
	}

	_, err := app.AnalyzeEntities(req)
	require.Error(t, err)
}

// Test 2.4: AnalyzeEntities with empty results
func TestAnalyzeEntities_EmptyResults(t *testing.T) {
	app, client, db := setupTestApp(t)
	defer client.Close()
	defer db.Close()

	req := AnalysisRequest{
		MinOccurrences: 10000, // Very high threshold to ensure no results
		MinConsistency: 0.99,
		TopN:           10,
	}

	resp, err := app.AnalyzeEntities(req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 0, resp.TotalTypes)
	assert.Len(t, resp.Candidates, 0)
}
// Test 3.1: PromoteEntity with valid type name
func TestPromoteEntity_ValidTypeName(t *testing.T) {
	t.Skip("Requires actual database with discovered entities and project setup")
	app, client, db := setupTestApp(t)
	defer client.Close()
	defer db.Close()
	
	// Would need to set up discovered entities first
	req := PromotionRequest{
		TypeName: "TestType",
	}
	
	resp, err := app.PromoteEntity(req)
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.NotEmpty(t, resp.SchemaFilePath)
	assert.Greater(t, resp.EntitiesMigrated, 0)
}

// Test 3.2: PromoteEntity with invalid/non-existent type name
func TestPromoteEntity_InvalidTypeName(t *testing.T) {
	app, client, db := setupTestApp(t)
	defer client.Close()
	defer db.Close()
	
	req := PromotionRequest{
		TypeName: "NonExistentType",
	}
	
	resp, err := app.PromoteEntity(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no entities found for type")
	assert.NotNil(t, resp)
	assert.False(t, resp.Success)
}

// Test 3.3: PromoteEntity schema generation error
func TestPromoteEntity_SchemaGenerationError(t *testing.T) {
	app, client, db := setupTestApp(t)
	client.Close() // Close to cause error
	defer db.Close()
	
	req := PromotionRequest{
		TypeName: "TestType",
	}
	
	resp, err := app.PromoteEntity(req)
	assert.Error(t, err)
	assert.NotNil(t, resp)
	assert.False(t, resp.Success)
}

// Test 3.4: PromoteEntity file write error
func TestPromoteEntity_FileWriteError(t *testing.T) {
	t.Skip("Requires mocking file system or setting up permissions")
}

// Test 3.5: PromoteEntity database migration error
func TestPromoteEntity_DatabaseMigrationError(t *testing.T) {
	t.Skip("Requires mocking migration execution")
}

// Test 3.6: Project root calculation
func TestPromoteEntity_ProjectRootCalculation(t *testing.T) {
	_, client, db := setupTestApp(t)
	defer client.Close()
	defer db.Close()
	
	// Test that we can calculate project root
	projectRoot, err := calculateProjectRoot()
	assert.NoError(t, err)
	assert.NotEmpty(t, projectRoot)
	assert.Contains(t, projectRoot, "enron-graph")
}

func calculateProjectRoot() (string, error) {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	
	// Navigate up from cmd/explorer to project root
	// cwd should be .../enron-graph/cmd/explorer
	// We need to go up 2 levels
	projectRoot := filepath.Dir(filepath.Dir(cwd))
	
	// Verify we're at the project root by checking for go.mod
	if _, err := os.Stat(filepath.Join(projectRoot, "go.mod")); err != nil {
		return "", fmt.Errorf("could not find project root (go.mod not found)")
	}
	
	return projectRoot, nil
}
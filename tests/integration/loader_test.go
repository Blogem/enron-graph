package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/Blogem/enron-graph/ent"
	"github.com/Blogem/enron-graph/internal/graph"
	"github.com/Blogem/enron-graph/internal/loader"
	"github.com/Blogem/enron-graph/pkg/utils"

	_ "github.com/lib/pq"
)

// TestLoaderIntegration tests the email loading functionality (T045)
func TestLoaderIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup: Connect to test database
	ctx := context.Background()
	client, cleanup := setupTestDB(t)
	defer cleanup()

	// Create repository
	repo := graph.NewRepository(client)
	logger := utils.NewLogger()

	// Load test CSV
	testCSV := filepath.Join("..", "fixtures", "sample_emails.csv")
	records, errors, err := loader.ParseCSV(testCSV)
	if err != nil {
		t.Fatalf("Failed to parse CSV: %v", err)
	}

	// Process emails
	processor := loader.NewProcessor(repo, logger, 5)
	if err := processor.ProcessBatch(ctx, records, errors); err != nil {
		t.Fatalf("ProcessBatch failed: %v", err)
	}

	// Verify: Check emails were inserted
	stats := processor.GetStats()
	if stats.Processed == 0 {
		t.Error("Expected some emails to be processed")
	}

	t.Logf("Processed %d emails", stats.Processed)

	// Verify: Query emails from database
	emails, err := client.Email.Query().All(ctx)
	if err != nil {
		t.Fatalf("Failed to query emails: %v", err)
	}

	if len(emails) == 0 {
		t.Fatal("No emails found in database")
	}

	t.Logf("Successfully loaded %d emails", len(emails))

	// Verify: Check metadata correctness
	for _, email := range emails {
		if email.MessageID == "" {
			t.Errorf("Email %d has empty MessageID", email.ID)
		}
		if email.From == "" {
			t.Errorf("Email %d has empty From field", email.ID)
		}
		if email.Date.IsZero() {
			t.Errorf("Email %d has zero Date", email.ID)
		}
	}

	// Verify: No duplicate message-ids
	messageIDs := make(map[string]bool)
	for _, email := range emails {
		if messageIDs[email.MessageID] {
			t.Errorf("Duplicate MessageID found: %s", email.MessageID)
		}
		messageIDs[email.MessageID] = true
	}
}

// setupTestDB creates a test database connection and returns cleanup function
func setupTestDB(t *testing.T) (*ent.Client, func()) {
	// Use test database URL from environment or default
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://enron:enron123@localhost:5433/enron_graph?sslmode=disable"
	}

	client, err := ent.Open("postgres", dbURL)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v\nMake sure PostgreSQL is running on port 5433", err)
	}

	// Run migrations
	ctx := context.Background()
	if err := client.Schema.Create(ctx); err != nil {
		client.Close()
		t.Fatalf("Failed to create schema: %v", err)
	}

	cleanup := func() {
		// Clean up test data
		ctx := context.Background()
		client.Email.Delete().ExecX(ctx)
		client.DiscoveredEntity.Delete().ExecX(ctx)
		client.Relationship.Delete().ExecX(ctx)
		client.Close()
	}

	return client, cleanup
}

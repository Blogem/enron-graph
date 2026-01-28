package integration

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/Blogem/enron-graph/internal/graph"
	"github.com/Blogem/enron-graph/internal/loader"
	"github.com/Blogem/enron-graph/pkg/utils"
)

// TestLoaderIntegration tests the email loading functionality (T045)
func TestLoaderIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup: Connect to test database
	ctx := context.Background()
	client := SetupTestDB(t)

	// Create repository
	logger := utils.NewLogger()
	repo := graph.NewRepository(client, logger)

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

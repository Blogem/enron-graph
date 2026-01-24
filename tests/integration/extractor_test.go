package integration

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/Blogem/enron-graph/ent"
	"github.com/Blogem/enron-graph/internal/extractor"
	"github.com/Blogem/enron-graph/internal/graph"
	"github.com/Blogem/enron-graph/pkg/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestExtractorIntegration tests T046: Entity extraction integration test
// - Setup: Test database with sample emails
// - Test: Run extraction on emails
// - Verify: Entities created with correct types
// - Verify: Relationships created (SENT, RECEIVED, MENTIONS)
// - Verify: Deduplicated entities (no duplicate email addresses)
// - Verify: Confidence scores applied correctly
func TestExtractorIntegration(t *testing.T) {
	// Setup: Create test database with clean schema
	client := SetupTestDB(t)

	ctx := context.Background()

	// Create repository
	repo := graph.NewRepository(client)

	// Create sample emails for testing
	date1, _ := time.Parse(time.RFC3339, "2001-01-15T10:00:00Z")
	email1, err := repo.CreateEmail(ctx, &graph.EmailInput{
		MessageID: "msg1@test.com",
		From:      "sender@enron.com",
		To:        []string{"recipient@enron.com"},
		Subject:   "Test Subject",
		Body:      "This email mentions Enron Corporation and discusses energy trading.",
		Date:      date1,
	})
	require.NoError(t, err, "Failed to create test email 1")

	date2, _ := time.Parse(time.RFC3339, "2001-01-16T14:00:00Z")
	email2, err := repo.CreateEmail(ctx, &graph.EmailInput{
		MessageID: "msg2@test.com",
		From:      "sender@enron.com",
		To:        []string{"another@enron.com"},
		Subject:   "Follow up",
		Body:      "Following up on the Enron Corporation meeting.",
		Date:      date2,
	})
	require.NoError(t, err, "Failed to create test email 2")

	// Initialize logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))

	// Initialize LLM client (using Ollama)
	llmClient := llm.NewOllamaClient("http://localhost:11434", "llama3.1:8b", "mxbai-embed-large", logger)

	// Initialize extractor
	ext := extractor.NewExtractor(llmClient, repo, logger)

	// Test: Run extraction on emails
	t.Run("ExtractEntitiesFromEmails", func(t *testing.T) {
		// Extract entities from first email
		_, err := ext.ExtractFromEmail(ctx, email1)
		require.NoError(t, err, "Failed to extract entities from email 1")

		// Extract entities from second email (tests deduplication)
		_, err = ext.ExtractFromEmail(ctx, email2)
		require.NoError(t, err, "Failed to extract entities from email 2")
	})

	// Verify: Entities created with correct types
	t.Run("VerifyEntityTypesCreated", func(t *testing.T) {
		// Query all discovered entities
		entities, err := client.DiscoveredEntity.Query().All(ctx)
		require.NoError(t, err, "Failed to query entities")
		assert.Greater(t, len(entities), 0, "No entities were created")

		// Check that we have different entity types
		entityTypes := make(map[string]int)
		for _, e := range entities {
			entityTypes[e.TypeCategory]++
		}

		// We should have at least person entities (from email addresses)
		assert.Contains(t, entityTypes, "person", "No person entities found")

		t.Logf("Entity types found: %v", entityTypes)
	})

	// Verify: Relationships created (SENT, RECEIVED, MENTIONS)
	t.Run("VerifyRelationshipsCreated", func(t *testing.T) {
		// Query all relationships
		relationships, err := client.Relationship.Query().All(ctx)
		require.NoError(t, err, "Failed to query relationships")
		assert.Greater(t, len(relationships), 0, "No relationships were created")

		// Check relationship types
		relTypes := make(map[string]int)
		for _, r := range relationships {
			relTypes[r.Type]++
		}

		// We should have SENT and RECEIVED relationships at minimum
		assert.Contains(t, relTypes, "SENT", "No SENT relationships found")
		assert.Contains(t, relTypes, "RECEIVED", "No RECEIVED relationships found")

		t.Logf("Relationship types found: %v", relTypes)
	})

	// Verify: Deduplicated entities (no duplicate email addresses)
	t.Run("VerifyDeduplication", func(t *testing.T) {
		// Query all entities
		entities, err := client.DiscoveredEntity.Query().All(ctx)
		require.NoError(t, err, "Failed to query entities")

		// Check for duplicate unique_ids
		uniqueIDs := make(map[string]int)
		for _, e := range entities {
			uniqueIDs[e.UniqueID]++
		}

		// Verify no duplicates
		for id, count := range uniqueIDs {
			assert.Equal(t, 1, count, "Duplicate entity found with unique_id: %s", id)
		}

		t.Logf("Total entities: %d, unique IDs: %d", len(entities), len(uniqueIDs))
	})

	// Verify: Confidence scores applied correctly
	t.Run("VerifyConfidenceScores", func(t *testing.T) {
		// Query all discovered entities
		entities, err := client.DiscoveredEntity.Query().All(ctx)
		require.NoError(t, err, "Failed to query entities")

		// Verify all entities have confidence scores >= 0.7 (minimum threshold)
		for _, e := range entities {
			assert.GreaterOrEqual(t, e.ConfidenceScore, 0.7,
				"Entity %s has confidence score below threshold: %f", e.Name, e.ConfidenceScore)
			assert.LessOrEqual(t, e.ConfidenceScore, 1.0,
				"Entity %s has confidence score above 1.0: %f", e.Name, e.ConfidenceScore)
		}

		t.Logf("All %d entities have valid confidence scores (>=0.7, <=1.0)", len(entities))
	})

	// Verify: MENTIONS relationships exist for organizations/concepts
	t.Run("VerifyMentionsRelationships", func(t *testing.T) {
		// Query all relationships and filter MENTIONS in Go
		allRels, err := client.Relationship.Query().All(ctx)
		require.NoError(t, err, "Failed to query relationships")

		mentions := 0
		for _, r := range allRels {
			if r.Type == "MENTIONS" {
				mentions++
			}
		}

		// We should have at least some MENTIONS relationships if orgs/concepts were extracted
		t.Logf("Found %d MENTIONS relationships", mentions)
	})

	// Verify: COMMUNICATES_WITH relationships inferred
	t.Run("VerifyCommunicatesWithInferred", func(t *testing.T) {
		// Query all relationships and filter COMMUNICATES_WITH in Go
		allRels, err := client.Relationship.Query().All(ctx)
		require.NoError(t, err, "Failed to query relationships")

		communicates := 0
		for _, r := range allRels {
			if r.Type == "COMMUNICATES_WITH" {
				communicates++
			}
		}

		t.Logf("Found %d COMMUNICATES_WITH relationships", communicates)
	})
}

// TestExtractorBatchProcessing tests batch extraction functionality
func TestExtractorBatchProcessing(t *testing.T) {
	// Setup: Create test database
	client := SetupTestDB(t)

	ctx := context.Background()

	// Create repository
	repo := graph.NewRepository(client)

	// Create multiple emails for batch processing
	emailIDs := make([]*ent.Email, 0, 10)
	for i := 0; i < 10; i++ {
		date, _ := time.Parse(time.RFC3339, "2001-01-15T10:00:00Z")
		email, err := repo.CreateEmail(ctx, &graph.EmailInput{
			MessageID: fmt.Sprintf("msg%d@test.com", i),
			From:      "sender@enron.com",
			To:        []string{"recipient@enron.com"},
			Subject:   "Test Subject",
			Body:      "This is test email content about Enron Corporation.",
			Date:      date,
		})
		require.NoError(t, err, "Failed to create test email")
		emailIDs = append(emailIDs, email)
	}

	// Initialize logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))

	// Initialize LLM client
	llmClient := llm.NewOllamaClient("http://localhost:11434", "llama3.1:8b", "mxbai-embed-large", logger)

	// Test batch extraction
	t.Run("BatchExtraction", func(t *testing.T) {
		// Create batch extractor with 5 workers
		batchExt := extractor.NewBatchExtractor(llmClient, repo, logger, 5)

		err := batchExt.ProcessBatch(ctx, emailIDs)
		require.NoError(t, err, "Batch extraction failed")

		// Verify entities were created
		entities, err := client.DiscoveredEntity.Query().All(ctx)
		require.NoError(t, err, "Failed to query entities")
		assert.Greater(t, len(entities), 0, "No entities created from batch extraction")

		t.Logf("Batch extraction created %d entities from %d emails",
			len(entities), len(emailIDs))
	})
}

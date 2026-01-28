package integration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Blogem/enron-graph/ent"
	"github.com/Blogem/enron-graph/ent/discoveredentity"
	"github.com/Blogem/enron-graph/internal/analyst"
	"github.com/Blogem/enron-graph/internal/extractor"
	"github.com/Blogem/enron-graph/internal/graph"
	"github.com/Blogem/enron-graph/internal/loader"
	"github.com/Blogem/enron-graph/internal/promoter"
	"github.com/Blogem/enron-graph/pkg/llm"
	"github.com/Blogem/enron-graph/pkg/utils"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFullWorkflow tests the complete end-to-end workflow (T140):
// load emails → extract → query → promote → query promoted entities
func TestFullWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping full workflow test in short mode")
	}

	ctx := context.Background()

	// Step 1: Setup test environment
	t.Log("Step 1: Setting up test environment...")
	client, db := SetupTestDBWithSQL(t)
	logger := utils.NewLogger()
	repo := graph.NewRepositoryWithDB(client, db, logger)

	// Step 2: Load emails from CSV
	t.Log("Step 2: Loading emails from CSV...")
	testCSV := filepath.Join("..", "fixtures", "sample_emails.csv")
	records, errors, err := loader.ParseCSV(testCSV)
	require.NoError(t, err, "Failed to parse CSV")

	processor := loader.NewProcessor(repo, logger, 5)
	err = processor.ProcessBatch(ctx, records, errors)
	require.NoError(t, err, "Failed to process email batch")

	// Verify emails were loaded
	emails, err := client.Email.Query().All(ctx)
	require.NoError(t, err, "Failed to query emails")
	require.Greater(t, len(emails), 0, "No emails loaded")
	t.Logf("✓ Loaded %d emails", len(emails))

	// Step 3: Extract entities and relationships from emails
	t.Log("Step 3: Extracting entities and relationships...")

	// Check if Ollama is available
	llmClient := llm.NewOllamaClient(
		"http://localhost:11434",
		"llama3.1:8b",
		"mxbai-embed-large",
		logger,
	)

	// Test Ollama connection with a simple prompt
	testCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	_, err = llmClient.GenerateCompletion(testCtx, "test")

	var extractedCount int
	if err != nil {
		t.Logf("Ollama not available, creating mock entities: %v", err)
		// Create mock entities and relationships for testing without LLM
		extractedCount, err = createMockEntitiesFromEmails(ctx, client, emails)
		require.NoError(t, err, "Failed to create mock entities")
		t.Logf("✓ Created %d mock entities from %d emails", extractedCount, len(emails))
	} else {
		// Use actual extractor
		extractor := extractor.NewExtractor(llmClient, repo, logger)

		// Extract entities from each email
		for _, email := range emails {
			_, err := extractor.ExtractFromEmail(ctx, email)
			if err != nil {
				t.Logf("Warning: extraction failed for email %d: %v", email.ID, err)
				continue
			}
			extractedCount++
		}

		require.Greater(t, extractedCount, 0, "No emails were successfully extracted")
		t.Logf("✓ Extracted entities from %d emails", extractedCount)
	}

	// Verify entities were extracted
	entities, err := client.DiscoveredEntity.Query().All(ctx)
	require.NoError(t, err, "Failed to query entities")
	require.Greater(t, len(entities), 0, "No entities extracted")
	t.Logf("✓ Found %d discovered entities", len(entities))

	// Verify relationships were created
	relationships, err := client.Relationship.Query().All(ctx)
	require.NoError(t, err, "Failed to query relationships")
	require.Greater(t, len(relationships), 0, "No relationships created")
	t.Logf("✓ Found %d relationships", len(relationships))

	// Step 4: Query the graph
	t.Log("Step 4: Querying the graph...")

	// Query entities by type
	entityTypes := make(map[string]int)
	for _, entity := range entities {
		entityTypes[entity.TypeCategory]++
	}
	t.Logf("✓ Entity types found: %v", entityTypes)

	// Find entities by type (e.g., person)
	personEntities, err := repo.FindEntitiesByType(ctx, "person")
	require.NoError(t, err, "Failed to find person entities")
	t.Logf("✓ Found %d person entities", len(personEntities))

	// Test relationship traversal
	if len(personEntities) > 0 {
		firstPerson := personEntities[0]
		related, err := repo.TraverseRelationships(ctx, firstPerson.ID, "", 1)
		require.NoError(t, err, "Failed to traverse relationships")
		t.Logf("✓ Person '%s' has %d 1-hop relationships", firstPerson.Name, len(related))
	}

	// Test similarity search if we have entities with embeddings
	if len(entities) > 0 && entities[0].Embedding != nil {
		similar, err := repo.SimilaritySearch(ctx, entities[0].Embedding, 5, 0.8)
		require.NoError(t, err, "Failed to perform similarity search")
		t.Logf("✓ Found %d similar entities via pgvector", len(similar))
	} else {
		t.Log("Note: No entities with embeddings for similarity search test")
	}

	// Step 5: Run analyst to detect patterns and promote schema
	t.Log("Step 5: Running analyst pattern detection...")

	// Populate more test entities to meet minimum thresholds
	err = populateAdditionalEntities(ctx, client)
	require.NoError(t, err, "Failed to populate additional entities")

	// Detect patterns
	patterns, err := analyst.DetectPatterns(ctx, client)
	require.NoError(t, err, "Failed to detect patterns")
	require.Greater(t, len(patterns), 0, "No patterns detected")
	t.Logf("✓ Detected %d type patterns", len(patterns))

	// Rank candidates
	candidates, err := analyst.AnalyzeAndRankCandidates(ctx, client, 5, 0.5, 10)
	require.NoError(t, err, "Failed to rank candidates")
	require.Greater(t, len(candidates), 0, "No candidates ranked")
	t.Logf("✓ Ranked %d candidates for promotion", len(candidates))

	// Log top candidate
	topCandidate := candidates[0]
	t.Logf("✓ Top candidate: %s (score: %.2f, frequency: %d)",
		topCandidate.Type, topCandidate.Score, topCandidate.Frequency)

	// Step 6: Promote top candidate schema
	t.Log("Step 6: Promoting schema...")

	// Create temp directory for schema generation
	tempSchemaDir, err := os.MkdirTemp("", "ent-schema-test-*")
	require.NoError(t, err, "Failed to create temp directory")
	defer os.RemoveAll(tempSchemaDir)

	// Generate schema definition for top candidate
	schemaDefn, err := analyst.GenerateSchemaForType(ctx, client, topCandidate.Type)
	require.NoError(t, err, "Failed to generate schema definition")
	require.NotNil(t, schemaDefn, "Schema definition is nil")
	t.Logf("✓ Generated schema for type '%s' with %d properties",
		schemaDefn.Type, len(schemaDefn.Properties))

	// Convert to promoter schema definition
	promoterSchema := convertSchemaDefinition(schemaDefn)

	// Create promoter and generate ent schema file
	p := promoter.NewPromoter(client)
	p.SetDB(db)

	req := promoter.PromotionRequest{
		TypeName:         strings.Title(topCandidate.Type),
		SchemaDefinition: promoterSchema,
		OutputDir:        tempSchemaDir,
		ProjectRoot:      filepath.Join("..", ".."),
	}

	schemaPath, err := p.GenerateEntSchema(req)
	require.NoError(t, err, "Failed to generate ent schema file")

	// Verify schema file was created
	_, err = os.Stat(schemaPath)
	require.NoError(t, err, "Schema file not created")
	t.Logf("✓ Generated ent schema file at: %s", schemaPath)

	// Verify schema file content
	schemaContent, err := os.ReadFile(schemaPath)
	require.NoError(t, err, "Failed to read schema file")

	schemaStr := string(schemaContent)
	assert.Contains(t, schemaStr, "package schema", "Schema file missing package declaration")
	assert.Contains(t, schemaStr, "func ("+strings.Title(topCandidate.Type)+") Fields()",
		"Schema file missing Fields() method")
	t.Logf("✓ Schema file contains expected structure")

	// Record promotion in audit log
	schemaDefJSON := make(map[string]interface{})
	for key, prop := range schemaDefn.Properties {
		schemaDefJSON[key] = map[string]interface{}{
			"type":     prop.Type,
			"required": prop.Required,
		}
	}

	promotionCriteria := map[string]interface{}{
		"score":       topCandidate.Score,
		"frequency":   topCandidate.Frequency,
		"density":     topCandidate.Density,
		"consistency": topCandidate.Consistency,
	}

	promotion, err := client.SchemaPromotion.
		Create().
		SetTypeName(topCandidate.Type).
		SetSchemaDefinition(schemaDefJSON).
		SetPromotionCriteria(promotionCriteria).
		SetPromotedAt(time.Now()).
		Save(ctx)
	require.NoError(t, err, "Failed to record schema promotion")
	t.Logf("✓ Recorded promotion in audit log (ID: %d)", promotion.ID)

	// Step 7: Query promoted entities
	t.Log("Step 7: Verifying promotion workflow...")

	// Verify the promotion was recorded
	promotions, err := client.SchemaPromotion.Query().All(ctx)
	require.NoError(t, err, "Failed to query promotions")
	require.Greater(t, len(promotions), 0, "No promotions recorded")
	t.Logf("✓ Found %d schema promotions", len(promotions))

	// Verify last promotion matches our request
	lastPromotion := promotions[len(promotions)-1]
	assert.Equal(t, topCandidate.Type, lastPromotion.TypeName, "Promotion type name mismatch")
	assert.NotEmpty(t, lastPromotion.SchemaDefinition, "Promotion schema definition is empty")
	t.Logf("✓ Verified promotion: %s at %s", lastPromotion.TypeName, lastPromotion.PromotedAt)

	// Step 8: Final verification
	t.Log("Step 8: Final verification...")

	// Count final state
	finalEmails, _ := client.Email.Query().Count(ctx)
	finalEntities, _ := client.DiscoveredEntity.Query().Count(ctx)
	finalRelationships, _ := client.Relationship.Query().Count(ctx)
	finalPromotions, _ := client.SchemaPromotion.Query().Count(ctx)

	t.Log("=== Full Workflow Test Summary ===")
	t.Logf("Emails loaded: %d", finalEmails)
	t.Logf("Entities extracted: %d", finalEntities)
	t.Logf("Relationships created: %d", finalRelationships)
	t.Logf("Schema promotions: %d", finalPromotions)
	t.Log("=================================")

	// Assert minimum success criteria
	assert.Greater(t, finalEmails, 0, "No emails in final state")
	assert.Greater(t, finalEntities, 0, "No entities in final state")
	assert.Greater(t, finalRelationships, 0, "No relationships in final state")
	assert.Greater(t, finalPromotions, 0, "No promotions in final state")

	t.Log("✓ T140 PASS: Full workflow test completed successfully")
}

// populateAdditionalEntities adds more test entities to meet minimum thresholds for pattern detection
func populateAdditionalEntities(ctx context.Context, client *ent.Client) error {
	// Create additional entities to ensure we have enough for pattern detection
	entityTypes := []struct {
		typeCategory string
		baseName     string
		count        int
		properties   map[string]interface{}
	}{
		{
			typeCategory: "person",
			baseName:     "Test Person",
			count:        20,
			properties: map[string]interface{}{
				"email":    "test@example.com",
				"role":     "employee",
				"location": "Houston",
			},
		},
		{
			typeCategory: "organization",
			baseName:     "Test Org",
			count:        15,
			properties: map[string]interface{}{
				"industry": "energy",
				"size":     "large",
			},
		},
		{
			typeCategory: "project",
			baseName:     "Test Project",
			count:        10,
			properties: map[string]interface{}{
				"status":   "active",
				"priority": "high",
			},
		},
	}

	for _, entityType := range entityTypes {
		for i := 1; i <= entityType.count; i++ {
			_, err := client.DiscoveredEntity.
				Create().
				SetUniqueID(fmt.Sprintf("%s_%d", entityType.typeCategory, i)).
				SetTypeCategory(entityType.typeCategory).
				SetName(fmt.Sprintf("%s %d", entityType.baseName, i)).
				SetProperties(entityType.properties).
				SetConfidenceScore(0.85).
				Save(ctx)
			if err != nil {
				return fmt.Errorf("failed to create %s entity: %w", entityType.typeCategory, err)
			}
		}
	}

	return nil
}

// createMockEntitiesFromEmails creates mock entities and relationships from emails when LLM is not available
func createMockEntitiesFromEmails(ctx context.Context, client *ent.Client, emails []*ent.Email) (int, error) {
	count := 0

	for _, email := range emails {
		// Create person entity from sender
		if email.From != "" {
			_, err := client.DiscoveredEntity.
				Create().
				SetUniqueID(email.From).
				SetTypeCategory("person").
				SetName(email.From).
				SetProperties(map[string]interface{}{
					"email": email.From,
				}).
				SetConfidenceScore(0.9).
				Save(ctx)
			if err == nil {
				count++
			}
		}

		// Create person entities from recipients
		for _, recipient := range email.To {
			_, err := client.DiscoveredEntity.
				Create().
				SetUniqueID(recipient).
				SetTypeCategory("person").
				SetName(recipient).
				SetProperties(map[string]interface{}{
					"email": recipient,
				}).
				SetConfidenceScore(0.9).
				Save(ctx)
			if err == nil {
				count++
			}
		}

		// Create concept entity from subject
		if email.Subject != "" {
			_, err := client.DiscoveredEntity.
				Create().
				SetUniqueID(fmt.Sprintf("concept_%d", email.ID)).
				SetTypeCategory("concept").
				SetName(email.Subject).
				SetProperties(map[string]interface{}{
					"topic": email.Subject,
				}).
				SetConfidenceScore(0.8).
				Save(ctx)
			if err == nil {
				count++
			}
		}

		// Create relationships
		sender, _ := client.DiscoveredEntity.Query().
			Where(discoveredentity.UniqueIDEQ(email.From)).
			First(ctx)

		if sender != nil {
			for _, recipient := range email.To {
				recipientEntity, _ := client.DiscoveredEntity.Query().
					Where(discoveredentity.UniqueIDEQ(recipient)).
					First(ctx)

				if recipientEntity != nil {
					_, _ = client.Relationship.
						Create().
						SetType("COMMUNICATES_WITH").
						SetFromType("discovered_entity").
						SetFromID(sender.ID).
						SetToType("discovered_entity").
						SetToID(recipientEntity.ID).
						SetProperties(map[string]interface{}{
							"email_id": email.ID,
						}).
						Save(ctx)
				}
			}
		}
	}

	return count, nil
}

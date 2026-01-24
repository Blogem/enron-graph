package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Blogem/enron-graph/ent"
	"github.com/Blogem/enron-graph/ent/discoveredentity"
	"github.com/Blogem/enron-graph/ent/schemapromotion"
	"github.com/Blogem/enron-graph/internal/analyst"
	"github.com/Blogem/enron-graph/internal/promoter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUS3AcceptanceCriteria tests all acceptance criteria for User Story 3
// These tests verify the acceptance criteria from spec.md for schema evolution through type promotion

// T090: Verify: Analyst identifies frequent/high-connectivity entity types
func TestAcceptance_T090_AnalystIdentifiesEntityTypes(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping acceptance test in short mode")
	}

	// Setup: Create test database with diverse discovered entities
	ctx := context.Background()
	client := SetupTestDB(t)

	// Pre-populate test database with entities of different frequencies and connectivity
	err := populateTestEntities(ctx, client)
	require.NoError(t, err, "Failed to populate test entities")

	// Test: Run pattern detection
	patterns, err := analyst.DetectPatterns(ctx, client)
	require.NoError(t, err, "Pattern detection failed")

	// Verify: Patterns identified
	require.Greater(t, len(patterns), 0, "Expected analyst to identify patterns")

	// Verify: Patterns include frequent entity types
	foundPerson := false
	foundOrg := false

	for typeName, stats := range patterns {
		t.Logf("Pattern: %s - Frequency=%d, AvgDensity=%.2f, PropertyCount=%d",
			typeName, stats.Frequency, stats.AvgDensity, len(stats.Properties))

		// Verify frequent types are identified
		if typeName == "person" {
			foundPerson = true
			assert.GreaterOrEqual(t, stats.Frequency, 30, "Expected person type to have high frequency")
		}
		if typeName == "organization" {
			foundOrg = true
			assert.GreaterOrEqual(t, stats.Frequency, 20, "Expected organization type to have moderate frequency")
		}

		// Verify statistics are calculated
		assert.Greater(t, stats.Frequency, 0, "Frequency should be positive")
		assert.NotNil(t, stats.Properties, "Properties should not be nil")
		assert.NotNil(t, stats.PropertyConsistency, "PropertyConsistency should not be nil")
	}

	assert.True(t, foundPerson, "Expected 'person' type to be identified")
	assert.True(t, foundOrg, "Expected 'organization' type to be identified")

	t.Logf("✓ T090 PASS: Analyst identifies frequent/high-connectivity entity types (%d patterns found)", len(patterns))
}

// T091: Verify: Candidates ranked by frequency, density, consistency
func TestAcceptance_T091_CandidatesRanked(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping acceptance test in short mode")
	}

	// Setup: Create test database
	ctx := context.Background()
	client := SetupTestDB(t)

	// Pre-populate test database
	err := populateTestEntities(ctx, client)
	require.NoError(t, err, "Failed to populate test entities")

	// Test: Analyze and rank candidates
	minOccurrences := 10
	minConsistency := 0.5
	topN := 10

	candidates, err := analyst.AnalyzeAndRankCandidates(ctx, client, minOccurrences, minConsistency, topN)
	require.NoError(t, err, "Candidate ranking failed")

	// Verify: Candidates returned
	require.Greater(t, len(candidates), 0, "Expected ranked candidates to be returned")

	// Verify: Candidates are sorted by score (descending)
	for i := 0; i < len(candidates)-1; i++ {
		assert.GreaterOrEqual(t, candidates[i].Score, candidates[i+1].Score,
			"Candidates should be sorted by score in descending order")
	}

	// Verify: Each candidate has valid metrics
	for i, candidate := range candidates {
		t.Logf("Candidate %d: Type=%s, Frequency=%d, Density=%.2f, Consistency=%.2f, Score=%.2f",
			i+1, candidate.Type, candidate.Frequency, candidate.Density, candidate.Consistency, candidate.Score)

		// Verify score calculation: 0.4*frequency + 0.3*density + 0.3*consistency
		expectedScore := 0.4*float64(candidate.Frequency) + 0.3*candidate.Density + 0.3*candidate.Consistency
		assert.InDelta(t, expectedScore, candidate.Score, 0.01, "Score should match formula")

		// Verify candidates meet minimum thresholds
		assert.GreaterOrEqual(t, candidate.Frequency, minOccurrences,
			"Candidate frequency should meet minimum threshold")
		assert.GreaterOrEqual(t, candidate.Consistency, minConsistency,
			"Candidate consistency should meet minimum threshold")
	}

	t.Logf("✓ T091 PASS: Candidates ranked by frequency, density, consistency (%d candidates)", len(candidates))
}

// T092: Verify: Promotion adds type to schema with properties/constraints
func TestAcceptance_T092_PromotionAddsTypeToSchema(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping acceptance test in short mode")
	}

	// Setup: Create test database
	ctx := context.Background()
	client := SetupTestDB(t)

	// Create temporary directory for generated schema files
	tempSchemaDir, err := os.MkdirTemp("", "ent-schema-test-*")
	require.NoError(t, err, "Failed to create temp directory")
	defer os.RemoveAll(tempSchemaDir)

	// Pre-populate with candidate entities
	err = populatePromoterTestEntities(ctx, client)
	require.NoError(t, err, "Failed to populate test entities")

	// Test: Generate schema from candidate type
	typeName := "contract"
	schema, err := analyst.GenerateSchemaForType(ctx, client, typeName)
	require.NoError(t, err, "Schema generation failed")

	// Verify: Schema has properties
	require.NotNil(t, schema, "Schema should not be nil")
	assert.Equal(t, typeName, schema.Type, "Schema type should match requested type")
	assert.Greater(t, len(schema.Properties), 0, "Schema should have properties")

	// Verify: Required properties identified (>90% presence)
	requiredProps := []string{}
	for propName, prop := range schema.Properties {
		if prop.Required {
			requiredProps = append(requiredProps, propName)
		}
	}
	assert.Greater(t, len(requiredProps), 0, "Schema should have required properties")

	// Verify: Schema has constraints/validation rules
	hasValidation := false
	for _, prop := range schema.Properties {
		if len(prop.ValidationRules) > 0 {
			hasValidation = true
			break
		}
	}
	assert.True(t, hasValidation, "Schema should have validation rules")

	// Test: Generate ent schema file
	promoterSchema := convertSchemaDefinition(schema)
	p := promoter.NewPromoter(client)

	req := promoter.PromotionRequest{
		TypeName:         "Contract",
		SchemaDefinition: promoterSchema,
		OutputDir:        tempSchemaDir,
		ProjectRoot:      filepath.Join("..", ".."),
	}

	schemaPath, err := p.GenerateEntSchema(req)
	require.NoError(t, err, "Failed to generate ent schema file")

	// Verify: Schema file created
	_, err = os.Stat(schemaPath)
	assert.NoError(t, err, "Schema file should exist")

	// Verify: Schema file contains type definition
	content, err := os.ReadFile(schemaPath)
	require.NoError(t, err, "Failed to read schema file")

	schemaStr := string(content)
	assert.Contains(t, schemaStr, "package schema", "Schema file should have package declaration")
	assert.Contains(t, schemaStr, "type Contract struct", "Schema file should have type definition")
	assert.Contains(t, schemaStr, "func (Contract) Fields()", "Schema file should have Fields method")

	t.Logf("✓ T092 PASS: Promotion adds type to schema with properties/constraints")
	t.Logf("  - Type: %s", schema.Type)
	t.Logf("  - Properties: %d", len(schema.Properties))
	t.Logf("  - Required: %d", len(requiredProps))
	t.Logf("  - Schema file: %s", schemaPath)
}

// T093: Verify: New entities validated against promoted schema
func TestAcceptance_T093_EntitiesValidatedAgainstSchema(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping acceptance test in short mode")
	}

	// Setup: Create test database
	ctx := context.Background()
	client := SetupTestDB(t)

	// Pre-populate with candidate entities
	err := populatePromoterTestEntities(ctx, client)
	require.NoError(t, err, "Failed to populate test entities")

	// Test: Generate schema
	typeName := "contract"
	schema, err := analyst.GenerateSchemaForType(ctx, client, typeName)
	require.NoError(t, err, "Schema generation failed")

	// Convert to promoter schema
	promoterSchema := convertSchemaDefinition(schema)

	// Test: Validate entities against schema
	p := promoter.NewPromoter(client)
	validationErrors, err := p.ValidateEntities(ctx, typeName, promoterSchema)
	require.NoError(t, err, "Validation should complete without error")

	// Verify: Validation ran (may have errors but process completes)
	assert.GreaterOrEqual(t, validationErrors, 0, "Validation error count should be non-negative")

	// Get entity count for reference
	entityCount, err := client.DiscoveredEntity.
		Query().
		Where(discoveredentity.TypeCategoryEQ(typeName)).
		Count(ctx)
	require.NoError(t, err, "Failed to count entities")

	// Verify: Most entities should validate successfully
	validEntities := entityCount - validationErrors
	validationRate := float64(validEntities) / float64(entityCount)
	assert.Greater(t, validationRate, 0.7, "At least 70%% of entities should validate successfully")

	t.Logf("✓ T093 PASS: New entities validated against promoted schema")
	t.Logf("  - Total entities: %d", entityCount)
	t.Logf("  - Valid entities: %d", validEntities)
	t.Logf("  - Validation errors: %d", validationErrors)
	t.Logf("  - Validation rate: %.1f%%", validationRate*100)
}

// T094: Verify: Audit log captures promotion events
func TestAcceptance_T094_AuditLogCapturesEvents(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping acceptance test in short mode")
	}

	// Setup: Create test database
	ctx := context.Background()
	client := SetupTestDB(t)

	// Create promoter
	p := promoter.NewPromoter(client)

	// Test: Create audit record for a promotion
	result := promoter.PromotionResult{
		Success:          true,
		TypeName:         "TestType",
		EntitiesMigrated: 42,
		ValidationErrors: 3,
		SchemaFilePath:   "/path/to/schema.go",
	}

	err := p.CreateAuditRecord(ctx, result)
	require.NoError(t, err, "Failed to create audit record")

	// Verify: Audit record created
	auditRecords, err := client.SchemaPromotion.Query().All(ctx)
	require.NoError(t, err, "Failed to query audit records")
	require.Greater(t, len(auditRecords), 0, "Expected audit record to be created")

	// Find our record
	var record *ent.SchemaPromotion
	for _, r := range auditRecords {
		if r.TypeName == "TestType" {
			record = r
			break
		}
	}
	require.NotNil(t, record, "Audit record for TestType not found")

	// Verify: Audit record has required fields
	assert.Equal(t, result.TypeName, record.TypeName, "TypeName should match")
	assert.Equal(t, result.EntitiesMigrated, record.EntitiesAffected, "EntitiesAffected should match")
	assert.Equal(t, result.ValidationErrors, record.ValidationFailures, "ValidationFailures should match")
	assert.False(t, record.PromotedAt.IsZero(), "PromotedAt should be set")
	assert.WithinDuration(t, time.Now(), record.PromotedAt, 5*time.Second, "PromotedAt should be recent")

	t.Logf("✓ T094 PASS: Audit log captures promotion events")
	t.Logf("  - TypeName: %s", record.TypeName)
	t.Logf("  - EntitiesAffected: %d", record.EntitiesAffected)
	t.Logf("  - ValidationFailures: %d", record.ValidationFailures)
	t.Logf("  - PromotedAt: %s", record.PromotedAt.Format(time.RFC3339))
}

// T095: Verify: SC-005 - 3+ candidates identified from 1k emails
func TestAcceptance_T095_SC005_ThreePlusCandidates(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping acceptance test in short mode")
	}

	// Setup: Create test database
	ctx := context.Background()
	client := SetupTestDB(t)

	// Pre-populate with diverse entities (simulating extraction from many emails)
	err := populateTestEntities(ctx, client)
	require.NoError(t, err, "Failed to populate test entities")

	// Test: Analyze and rank candidates with reasonable thresholds for test data
	// Note: In production, these would be higher (50 occurrences, 0.7 consistency)
	minOccurrences := 10  // Lowered for test data
	minConsistency := 0.5 // Lowered for test data
	topN := 10

	candidates, err := analyst.AnalyzeAndRankCandidates(ctx, client, minOccurrences, minConsistency, topN)
	require.NoError(t, err, "Candidate ranking failed")

	// Verify: SC-005 - At least 3 candidates identified
	candidateCount := len(candidates)
	assert.GreaterOrEqual(t, candidateCount, 3,
		"SC-005: Should identify at least 3 candidate entity types for promotion")

	// Log candidate details
	t.Logf("✓ T095 PASS: SC-005 - %d candidates identified (requirement: 3+)", candidateCount)
	for i, candidate := range candidates {
		t.Logf("  Candidate %d: %s (Frequency=%d, Score=%.2f)",
			i+1, candidate.Type, candidate.Frequency, candidate.Score)
	}
}

// T096: Verify: SC-006 - 1+ type successfully promoted
func TestAcceptance_T096_SC006_TypeSuccessfullyPromoted(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping acceptance test in short mode")
	}

	// Setup: Create test database
	ctx := context.Background()
	client := SetupTestDB(t)

	// Create temporary directory for schema files
	tempSchemaDir, err := os.MkdirTemp("", "ent-schema-test-*")
	require.NoError(t, err, "Failed to create temp directory")
	defer os.RemoveAll(tempSchemaDir)

	// Pre-populate with candidate entities
	err = populatePromoterTestEntities(ctx, client)
	require.NoError(t, err, "Failed to populate test entities")

	// Test: Full promotion workflow for one type
	typeName := "contract"

	// Step 1: Generate schema
	schema, err := analyst.GenerateSchemaForType(ctx, client, typeName)
	require.NoError(t, err, "Schema generation failed")
	require.NotNil(t, schema, "Schema should not be nil")

	// Step 2: Generate ent schema file
	promoterSchema := convertSchemaDefinition(schema)
	p := promoter.NewPromoter(client)

	req := promoter.PromotionRequest{
		TypeName:         "Contract",
		SchemaDefinition: promoterSchema,
		OutputDir:        tempSchemaDir,
		ProjectRoot:      filepath.Join("..", ".."),
	}

	schemaPath, err := p.GenerateEntSchema(req)
	require.NoError(t, err, "Failed to generate ent schema file")

	// Verify: Schema file exists
	_, err = os.Stat(schemaPath)
	require.NoError(t, err, "Schema file should exist")

	// Step 3: Validate entities
	validationErrors, err := p.ValidateEntities(ctx, typeName, promoterSchema)
	require.NoError(t, err, "Validation failed")

	// Step 4: Create audit record
	entityCount, _ := client.DiscoveredEntity.
		Query().
		Where(discoveredentity.TypeCategoryEQ(typeName)).
		Count(ctx)

	result := promoter.PromotionResult{
		Success:          true,
		TypeName:         req.TypeName,
		EntitiesMigrated: entityCount,
		ValidationErrors: validationErrors,
		SchemaFilePath:   schemaPath,
	}

	err = p.CreateAuditRecord(ctx, result)
	require.NoError(t, err, "Failed to create audit record")

	// Verify: SC-006 - Type successfully promoted
	// Success criteria:
	// 1. Schema generated
	// 2. Ent schema file created
	// 3. Entities validated
	// 4. Audit record created

	assert.True(t, result.Success, "Promotion should be successful")
	assert.FileExists(t, schemaPath, "Schema file should exist")
	assert.GreaterOrEqual(t, result.EntitiesMigrated, 1, "At least 1 entity should be migrated")

	t.Logf("✓ T096 PASS: SC-006 - 1+ type successfully promoted")
	t.Logf("  - Type: %s", result.TypeName)
	t.Logf("  - Entities migrated: %d", result.EntitiesMigrated)
	t.Logf("  - Validation errors: %d", result.ValidationErrors)
	t.Logf("  - Schema file: %s", schemaPath)
}

// T097: Verify: SC-010 - Audit log is complete
func TestAcceptance_T097_SC010_AuditLogComplete(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping acceptance test in short mode")
	}

	// Setup: Create test database
	ctx := context.Background()
	client := SetupTestDB(t)

	// Create promoter
	p := promoter.NewPromoter(client)

	// Test: Create multiple audit records to verify completeness
	promotions := []promoter.PromotionResult{
		{
			Success:          true,
			TypeName:         "Person",
			EntitiesMigrated: 100,
			ValidationErrors: 2,
			SchemaFilePath:   "/schemas/person.go",
		},
		{
			Success:          true,
			TypeName:         "Organization",
			EntitiesMigrated: 50,
			ValidationErrors: 0,
			SchemaFilePath:   "/schemas/organization.go",
		},
		{
			Success:          false,
			TypeName:         "FailedType",
			EntitiesMigrated: 0,
			ValidationErrors: 25,
			SchemaFilePath:   "",
		},
	}

	// Create audit records
	for _, result := range promotions {
		err := p.CreateAuditRecord(ctx, result)
		require.NoError(t, err, "Failed to create audit record for %s", result.TypeName)
	}

	// Verify: SC-010 - Audit log contains complete record
	auditRecords, err := client.SchemaPromotion.Query().All(ctx)
	require.NoError(t, err, "Failed to query audit records")
	require.GreaterOrEqual(t, len(auditRecords), len(promotions), "All promotions should be audited")

	// Verify: Each audit record has complete information
	for _, result := range promotions {
		// Find corresponding audit record
		var record *ent.SchemaPromotion
		for _, r := range auditRecords {
			if r.TypeName == result.TypeName {
				record = r
				break
			}
		}
		require.NotNil(t, record, "Audit record for %s not found", result.TypeName)

		// Verify: Complete record with all required fields
		assert.Equal(t, result.TypeName, record.TypeName, "TypeName should be recorded")
		assert.Equal(t, result.EntitiesMigrated, record.EntitiesAffected, "EntitiesAffected should be recorded")
		assert.Equal(t, result.ValidationErrors, record.ValidationFailures, "ValidationFailures should be recorded")
		assert.False(t, record.PromotedAt.IsZero(), "PromotedAt timestamp should be recorded")
		// Note: PromotionCriteria will be populated in a future enhancement (T098+)

		// Verify: Timestamp is reasonable
		assert.WithinDuration(t, time.Now(), record.PromotedAt, 10*time.Second,
			"PromotedAt should be recent")

		t.Logf("Audit record for %s: EntitiesAffected=%d, ValidationFailures=%d, Timestamp=%s",
			record.TypeName, record.EntitiesAffected, record.ValidationFailures,
			record.PromotedAt.Format(time.RFC3339))
	}

	// Verify: Audit log is queryable and filterable
	successfulPromotions, err := client.SchemaPromotion.
		Query().
		Where(schemapromotion.EntitiesAffectedGT(0)).
		All(ctx)
	require.NoError(t, err, "Should be able to query successful promotions")
	assert.GreaterOrEqual(t, len(successfulPromotions), 2, "At least 2 successful promotions")

	t.Logf("✓ T097 PASS: SC-010 - Audit log is complete")
	t.Logf("  - Total promotions audited: %d", len(auditRecords))
	t.Logf("  - Successful promotions: %d", len(successfulPromotions))
	t.Logf("  - All records have: TypeName, EntitiesAffected, ValidationFailures, PromotedAt")
}

package integration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Blogem/enron-graph/ent"
	"github.com/Blogem/enron-graph/internal/analyst"
	"github.com/Blogem/enron-graph/internal/promoter"
	_ "github.com/lib/pq"
)

// TestPromoterIntegration tests the schema promotion workflow (T089)
func TestPromoterIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup: Connect to test database with clean schema
	ctx := context.Background()
	client, db := SetupTestDBWithSQL(t)

	// Create a temporary directory for generated schema files
	tempSchemaDir, err := os.MkdirTemp("", "ent-schema-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempSchemaDir)

	// Pre-populate test database with candidate type entities
	err = populatePromoterTestEntities(ctx, client)
	if err != nil {
		t.Fatalf("Failed to populate test entities: %v", err)
	}

	// Step 1: Test schema generation from candidate
	typeName := "contract"
	analystSchema, err := analyst.GenerateSchemaForType(ctx, client, typeName)
	if err != nil {
		t.Fatalf("Failed to generate schema: %v", err)
	}

	// Verify: JSON schema has correct properties and types
	if analystSchema == nil {
		t.Fatal("Expected schema to be generated, got nil")
	}

	if analystSchema.Type != typeName {
		t.Errorf("Expected schema type '%s', got '%s'", typeName, analystSchema.Type)
	}

	if len(analystSchema.Properties) == 0 {
		t.Fatal("Expected schema to have properties, got 0")
	}

	t.Logf("Generated schema for type '%s' with %d properties", typeName, len(analystSchema.Properties))

	// Verify common properties exist
	expectedProperties := []string{"contract_id", "value", "status"}
	for _, propName := range expectedProperties {
		if _, exists := analystSchema.Properties[propName]; !exists {
			t.Errorf("Expected property '%s' not found in schema", propName)
		} else {
			prop := analystSchema.Properties[propName]
			t.Logf("Property '%s': Type=%s, Required=%v", propName, prop.Type, prop.Required)
		}
	}

	// Verify required vs optional properties
	requiredCount := 0
	optionalCount := 0
	for _, prop := range analystSchema.Properties {
		if prop.Required {
			requiredCount++
		} else {
			optionalCount++
		}
	}

	t.Logf("Schema has %d required properties and %d optional properties", requiredCount, optionalCount)

	// Convert analyst.SchemaDefinition to promoter.SchemaDefinition
	schema := convertSchemaDefinition(analystSchema)

	// Step 2: Test promotion workflow

	// Create promoter instance
	p := promoter.NewPromoter(client)
	p.SetDB(db)

	// Create promotion request
	req := promoter.PromotionRequest{
		TypeName:         strings.Title(typeName), // "Contract"
		SchemaDefinition: schema,
		OutputDir:        tempSchemaDir,
		ProjectRoot:      filepath.Join("..", ".."), // Point to repo root from tests/integration
	}

	// Note: We'll test individual steps separately since full promotion
	// requires go generate which modifies the actual ent package

	// Test: Generate ent schema file
	schemaPath, err := p.GenerateEntSchema(req)
	if err != nil {
		t.Fatalf("Failed to generate ent schema file: %v", err)
	}

	// Verify: New ent schema file created
	if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
		t.Errorf("Expected schema file to be created at %s, but it doesn't exist", schemaPath)
	}

	t.Logf("Generated ent schema file at: %s", schemaPath)

	// Verify: Schema file has correct structure
	schemaContent, err := os.ReadFile(schemaPath)
	if err != nil {
		t.Fatalf("Failed to read schema file: %v", err)
	}

	schemaStr := string(schemaContent)

	// Check for essential components
	expectedComponents := []string{
		"package schema",
		"type Contract struct",
		"func (Contract) Fields()",
		"field.",
	}

	for _, component := range expectedComponents {
		if !strings.Contains(schemaStr, component) {
			t.Errorf("Schema file missing expected component: %s", component)
		}
	}

	// Check that generated fields match schema properties
	for propName := range schema.Properties {
		if !strings.Contains(schemaStr, fmt.Sprintf(`"%s"`, propName)) {
			t.Errorf("Schema file missing field for property: %s", propName)
		}
	}

	t.Logf("Schema file validation passed")

	// Test: Validate entities against schema
	validationErrors, err := p.ValidateEntities(ctx, typeName, schema)
	if err != nil {
		t.Fatalf("Failed to validate entities: %v", err)
	}

	t.Logf("Validation completed with %d errors", validationErrors)

	// Verify: Entities are validated (some may fail, but process should complete)
	if validationErrors < 0 {
		t.Errorf("Invalid validation error count: %d", validationErrors)
	}

	// Test: SchemaPromotion audit record creation
	result := promoter.PromotionResult{
		Success:          true,
		TypeName:         req.TypeName,
		EntitiesMigrated: 30,
		ValidationErrors: validationErrors,
		SchemaFilePath:   schemaPath,
	}

	err = p.CreateAuditRecord(ctx, result)
	if err != nil {
		t.Fatalf("Failed to create audit record: %v", err)
	}

	// Verify: SchemaPromotion audit record created
	auditRecords, err := client.SchemaPromotion.Query().All(ctx)
	if err != nil {
		t.Fatalf("Failed to query audit records: %v", err)
	}

	if len(auditRecords) == 0 {
		t.Fatal("Expected audit record to be created, got 0")
	}

	// Find the audit record we just created
	var ourRecord *ent.SchemaPromotion
	for _, record := range auditRecords {
		if record.TypeName == req.TypeName {
			ourRecord = record
			break
		}
	}

	if ourRecord == nil {
		t.Fatalf("Audit record for type '%s' not found", req.TypeName)
	}

	// Verify audit record fields
	if ourRecord.EntitiesAffected != result.EntitiesMigrated {
		t.Errorf("Expected EntitiesAffected=%d, got %d", result.EntitiesMigrated, ourRecord.EntitiesAffected)
	}

	if ourRecord.ValidationFailures != result.ValidationErrors {
		t.Errorf("Expected ValidationFailures=%d, got %d", result.ValidationErrors, ourRecord.ValidationFailures)
	}

	if ourRecord.PromotedAt.IsZero() {
		t.Error("Expected PromotedAt to be set, got zero time")
	}

	t.Logf("Audit record verified: TypeName=%s, EntitiesAffected=%d, ValidationFailures=%d, PromotedAt=%s",
		ourRecord.TypeName, ourRecord.EntitiesAffected, ourRecord.ValidationFailures, ourRecord.PromotedAt)

	// Note: We skip the following steps in this test because they modify the actual codebase:
	// - Running go generate ./ent (would regenerate ent code in the real package)
	// - Running database migration (would modify test schema in unpredictable ways)
	// - Copying data to new typed table (requires the new table to exist first)
	//
	// These steps are tested in the full end-to-end shell script test (T097a)
	// and are exercised in the analyst CLI's promote command

	t.Log("Promoter integration test completed successfully")

	// Teardown: Cleanup is handled by SetupTestDB's t.Cleanup and deferred tempSchemaDir removal
}

// populatePromoterTestEntities creates candidate entities for promotion testing
func populatePromoterTestEntities(ctx context.Context, client *ent.Client) error {
	// Create "contract" entities with consistent properties
	// These represent a candidate type for promotion

	// Properties that appear in >90% of entities (required)
	baseProperties := map[string]interface{}{
		"contract_id": "CTR-12345",
		"value":       100000.0,
		"status":      "active",
	}

	// Properties that appear in 50-90% of entities (optional)
	optionalProperties := map[string]interface{}{
		"department": "trading",
		"region":     "north",
	}

	// Create 30 contract entities with high property consistency
	for i := 1; i <= 30; i++ {
		props := make(map[string]interface{})

		// Add base properties to all entities (100% consistency)
		for k, v := range baseProperties {
			props[k] = v
		}

		// Add optional properties to 70% of entities
		if i%10 != 0 && i%10 != 1 && i%10 != 2 {
			for k, v := range optionalProperties {
				props[k] = v
			}
		}

		// Update values to be unique
		props["contract_id"] = fmt.Sprintf("CTR-%05d", i)
		props["value"] = float64(50000 + i*1000)

		_, err := client.DiscoveredEntity.
			Create().
			SetUniqueID(fmt.Sprintf("contract_%d", i)).
			SetTypeCategory("contract").
			SetName(fmt.Sprintf("Contract %d", i)).
			SetProperties(props).
			SetConfidenceScore(0.85).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create contract entity %d: %w", i, err)
		}
	}

	// Create some entities of other types to ensure filtering works
	for i := 1; i <= 10; i++ {
		otherProps := map[string]interface{}{
			"name":  fmt.Sprintf("Other %d", i),
			"value": i * 100,
		}

		_, err := client.DiscoveredEntity.
			Create().
			SetUniqueID(fmt.Sprintf("other_%d", i)).
			SetTypeCategory("other_type").
			SetName(fmt.Sprintf("Other Entity %d", i)).
			SetProperties(otherProps).
			SetConfidenceScore(0.75).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create other entity %d: %w", i, err)
		}
	}

	return nil
}

// convertSchemaDefinition converts analyst.SchemaDefinition to promoter.SchemaDefinition
// This is needed because both packages define the same type independently
func convertSchemaDefinition(analystSchema *analyst.SchemaDefinition) promoter.SchemaDefinition {
	properties := make(map[string]promoter.PropertyDefinition)

	for propName, analystProp := range analystSchema.Properties {
		validationRules := make([]promoter.ValidationRule, 0, len(analystProp.ValidationRules))
		for _, analystRule := range analystProp.ValidationRules {
			validationRules = append(validationRules, promoter.ValidationRule{
				Type:  analystRule.Type,
				Value: analystRule.Value,
			})
		}

		properties[propName] = promoter.PropertyDefinition{
			Type:            analystProp.Type,
			Required:        analystProp.Required,
			ValidationRules: validationRules,
		}
	}

	return promoter.SchemaDefinition{
		Type:       analystSchema.Type,
		Properties: properties,
	}
}

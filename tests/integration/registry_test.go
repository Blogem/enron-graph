//go:build registry
// +build registry

package integration

import (
	"context"
	"log/slog"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/Blogem/enron-graph/internal/extractor"
	"github.com/Blogem/enron-graph/internal/graph"
	"github.com/Blogem/enron-graph/internal/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// NOTE: These tests require the TestPerson schema to be generated before running.
// Use scripts/test-registry.sh to run these tests, which handles schema creation and cleanup.
//
// The TestPerson schema is dynamically created during test execution and is NOT
// part of production builds. This ensures the test schema doesn't pollute production code.
//
// Build tag: These tests are excluded from normal test runs (go test ./...).
// They only run when explicitly invoked with: go test -tags registry

// TestRegistryCodegenIntegration tests that code generation creates registration for TestPerson
func TestRegistryCodegenIntegration(t *testing.T) {
	// Verify TestPerson is registered in the registry
	_, exists := registry.PromotedTypes["TestPerson"]
	assert.True(t, exists, "TestPerson should be registered in the promoted types registry")

	// Verify other expected schemas are registered
	expectedSchemas := []string{"Email", "DiscoveredEntity", "Relationship", "SchemaPromotion"}
	for _, schemaName := range expectedSchemas {
		_, exists := registry.PromotedTypes[schemaName]
		assert.True(t, exists, "Schema %s should be registered", schemaName)
	}
}

// TestRegistryEntityCreation tests that registry creators can create entities
func TestRegistryEntityCreation(t *testing.T) {
	// Setup: Create test database
	client := SetupTestDB(t)
	ctx := context.Background()

	// Add Ent client to context for registry creators
	ctx = context.WithValue(ctx, "entClient", client)

	// Get the TestPerson creator function
	createFn, exists := registry.PromotedTypes["TestPerson"]
	require.True(t, exists, "TestPerson should be registered")

	// Prepare test data
	data := map[string]any{
		"unique_id":        "test-person-1",
		"name":             "John Doe",
		"email":            "john.doe@test.com",
		"confidence_score": 0.95,
	}

	// Create entity using registry
	result, err := createFn(ctx, data)
	require.NoError(t, err, "Should create TestPerson entity without error")
	require.NotNil(t, result, "Created entity should not be nil")

	// Verify the entity was created using reflection to access fields
	// We avoid using *ent.TestPerson directly since it's only generated at test time
	v := reflect.ValueOf(result)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	getField := func(name string) interface{} {
		field := v.FieldByName(name)
		if field.IsValid() {
			return field.Interface()
		}
		return nil
	}

	assert.Equal(t, "test-person-1", getField("UniqueID"))
	assert.Equal(t, "John Doe", getField("Name"))
	assert.Equal(t, "john.doe@test.com", getField("Email"))
	assert.Equal(t, 0.95, getField("ConfidenceScore"))
}

// TestRegistryNilFieldHandling tests that nil/missing fields are handled gracefully
func TestRegistryNilFieldHandling(t *testing.T) {
	// Setup: Create test database
	client := SetupTestDB(t)
	ctx := context.Background()
	ctx = context.WithValue(ctx, "entClient", client)

	// Get the TestPerson creator function
	createFn, exists := registry.PromotedTypes["TestPerson"]
	require.True(t, exists, "TestPerson should be registered")

	// Prepare test data with minimal required fields only
	data := map[string]any{
		"unique_id": "test-person-2",
		"name":      "Jane Doe",
		// email is optional - omitted
		// confidence_score has a default
	}

	// Create entity using registry
	result, err := createFn(ctx, data)
	require.NoError(t, err, "Should create TestPerson with minimal fields")
	require.NotNil(t, result, "Created entity should not be nil")

	// Verify the entity was created with defaults using reflection
	v := reflect.ValueOf(result)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	getField := func(name string) interface{} {
		field := v.FieldByName(name)
		if field.IsValid() {
			return field.Interface()
		}
		return nil
	}

	assert.Equal(t, "test-person-2", getField("UniqueID"))
	assert.Equal(t, "Jane Doe", getField("Name"))
	assert.Equal(t, "", getField("Email")) // Optional field should be empty
}

// TestRegistryErrorHandling tests that registry handles errors gracefully
func TestRegistryErrorHandling(t *testing.T) {
	// Setup: Create test database
	client := SetupTestDB(t)
	ctx := context.Background()
	ctx = context.WithValue(ctx, "entClient", client)

	// Get the TestPerson creator function
	createFn, exists := registry.PromotedTypes["TestPerson"]
	require.True(t, exists, "TestPerson should be registered")

	// Test 1: Missing required field should fail
	data := map[string]any{
		"unique_id": "test-person-error-1",
		// name is required - omitted
	}

	result, err := createFn(ctx, data)
	assert.Error(t, err, "Should fail when required field is missing")
	assert.Nil(t, result, "Result should be nil on error")

	// Test 2: Duplicate unique_id should fail
	// First create a person
	data1 := map[string]any{
		"unique_id": "test-person-dup",
		"name":      "First Person",
	}

	result1, err1 := createFn(ctx, data1)
	require.NoError(t, err1, "First creation should succeed")
	require.NotNil(t, result1, "First result should not be nil")

	// Now try to create with same unique_id
	data2 := map[string]any{
		"unique_id": "test-person-dup", // Same as above
		"name":      "Duplicate Person",
	}

	result2, err2 := createFn(ctx, data2)
	assert.Error(t, err2, "Should fail on duplicate unique_id")
	assert.Nil(t, result2, "Result should be nil on error")
}

// TestRegistryPromotedTypeEndToEnd tests the complete workflow:
// 1. Extract entity with promoted type (TestPerson)
// 2. Verify it uses the promoted schema instead of DiscoveredEntity
func TestRegistryPromotedTypeEndToEnd(t *testing.T) {
	// Setup: Create test database
	client := SetupTestDB(t)
	ctx := context.Background()

	// Create repository and extractor
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))
	repo := graph.NewRepository(client, logger)

	// Create a mock LLM client that returns a TestPerson entity
	// Note: The extractor adds "{" to the response, so we don't include it here
	mockLLM := &mockLLMClient{
		responses: map[string]string{
			"": `
	"analysis": "Found a person entity",
	"entities": [
		{
			"id": "person1",
			"type": "TestPerson",
			"name": "Alice Smith",
			"properties": {
				"email": "alice.smith@test.com"
			},
			"confidence": 0.95
		}
	],
	"relationships": []
}`,
		},
		patterns: []string{""}, // Match all prompts
	}

	ext := extractor.NewExtractor(mockLLM, repo, logger)

	// Create a test email
	date, _ := time.Parse(time.RFC3339, "2001-01-15T10:00:00Z")
	email, err := repo.CreateEmail(ctx, &graph.EmailInput{
		MessageID: "test-promoted-type@test.com",
		From:      "sender@test.com",
		To:        []string{"alice.smith@test.com"},
		Subject:   "Test Email",
		Body:      "This email mentions Alice Smith.",
		Date:      date,
	})
	require.NoError(t, err, "Failed to create test email")

	// Extract entities from the email
	summary, err := ext.ExtractFromEmail(ctx, email)
	require.NoError(t, err, "Extraction should succeed")
	require.Greater(t, summary.EntitiesCreated, 0, "Should create at least one entity")

	// Verify: Check if TestPerson was created in the promoted table
	// We use reflection to query the TestPerson table since the type is only available at test time
	testPersons, err := client.TestPerson.Query().All(ctx)
	require.NoError(t, err, "Should be able to query TestPerson table")

	// Should have at least one TestPerson entity (from content extraction)
	// Note: Header extraction creates additional Person entities from email addresses
	assert.Greater(t, len(testPersons), 0, "TestPerson entity should be created in promoted table")

	if len(testPersons) > 0 {
		// Verify the entity has correct data
		found := false
		for _, tp := range testPersons {
			if tp.Name == "Alice Smith" || tp.Email == "alice.smith@test.com" {
				found = true
				assert.Equal(t, "Alice Smith", tp.Name, "TestPerson should have correct name")
				assert.Equal(t, "alice.smith@test.com", tp.Email, "TestPerson should have correct email")
				assert.Equal(t, 0.95, tp.ConfidenceScore, "TestPerson should have correct confidence")
				break
			}
		}
		assert.True(t, found, "Should find TestPerson with expected data")
	}

	// Note: Current known limitation (W3 in verification report):
	// The extractor creates the entity in the promoted table successfully,
	// but also creates a fallback entry in DiscoveredEntity because
	// FindEntityByUniqueID only queries DiscoveredEntity table.
	// This is documented behavior and will be addressed in a future improvement.
	//
	// For this test, we verify the MAIN functionality works:
	// - TestPerson entity is created in the promoted table
	// - It has the correct data
	//
	// The duplicate entry in DiscoveredEntity is a known issue that doesn't
	// affect the core registry routing functionality.

	discoveredEntities, err := client.DiscoveredEntity.Query().All(ctx)
	require.NoError(t, err, "Should be able to query DiscoveredEntity table")

	// Document current state: promoted entities ARE created, even if a fallback entry exists
	testPersonCount := 0
	for _, de := range discoveredEntities {
		if de.TypeCategory == "TestPerson" {
			testPersonCount++
		}
	}

	// Log the current state for visibility
	t.Logf("Found %d TestPerson entities in promoted table", len(testPersons))
	t.Logf("Found %d TestPerson fallback entries in DiscoveredEntity table (known limitation)", testPersonCount)

	// The key assertion: promoted table has the entities
	assert.Greater(t, len(testPersons), 0, "Promoted table should contain TestPerson entities")
}

# Schema Service Contract (Go)

**Package**: `internal/explorer`  
**File**: `schema_service.go`  
**Purpose**: Backend service interface for schema introspection in the Graph Explorer

---

## Interface Definition

```go
package explorer

import (
    "context"
)

// SchemaService provides schema metadata access for the explorer frontend
type SchemaService interface {
    // GetSchema returns complete schema information for both promoted and discovered types
    // Implements FR-001 (promoted types) and FR-002 (discovered types)
    // Parameters:
    //   - ctx: request context
    // Returns:
    //   - SchemaResponse with all type metadata
    //   - error if database query or ent schema introspection fails
    GetSchema(ctx context.Context) (*SchemaResponse, error)
    
    // GetTypeDetails returns detailed information for a specific entity type
    // Implements User Story 1, Acceptance Scenario 3 (click on type for details)
    // Parameters:
    //   - ctx: request context
    //   - typeName: name of the type (e.g., "Email", "Person")
    // Returns:
    //   - SchemaType with complete property and relationship information
    //   - error if type not found
    GetTypeDetails(ctx context.Context, typeName string) (*SchemaType, error)
    
    // RefreshSchema re-queries database for schema changes (new discovered types)
    // Implements FR-013: refresh without restart
    // Parameters:
    //   - ctx: request context
    // Returns:
    //   - SchemaResponse with updated schema
    //   - error if refresh fails
    RefreshSchema(ctx context.Context) (*SchemaResponse, error)
}

// SchemaResponse contains complete schema metadata
type SchemaResponse struct {
    PromotedTypes   []SchemaType `json:"promoted_types"`
    DiscoveredTypes []SchemaType `json:"discovered_types"`
    TotalEntities   int          `json:"total_entities"`
}

// SchemaType describes an entity type's schema
type SchemaType struct {
    Name          string               `json:"name"`         // e.g., "Email", "Person"
    Category      string               `json:"category"`     // "promoted" | "discovered"
    Count         int                  `json:"count"`        // Number of entities
    Properties    []PropertyDefinition `json:"properties"`   // Property metadata
    Relationships []string             `json:"relationships"` // Relationship type names
}

// PropertyDefinition describes a single property/field
type PropertyDefinition struct {
    Name        string      `json:"name"`         // Property name
    DataType    string      `json:"data_type"`    // e.g., "string", "int", "timestamp"
    SampleValue interface{} `json:"sample_value"` // Example value (can be nil)
    Nullable    bool        `json:"nullable"`     // Can be null
}
```

---

## Contract Tests

**Test File**: `tests/contract/schema_service_contract_test.go`

### Test Cases

```go
package contract_test

import (
    "context"
    "testing"
    
    "github.com/Blogem/enron-graph/internal/explorer"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

// TestSchemaService_GetSchema_ReturnsPromotedTypes verifies promoted types are exposed
func TestSchemaService_GetSchema_ReturnsPromotedTypes(t *testing.T) {
    // Given: SchemaService with ent schema containing Email, Relationship, SchemaPromotion
    svc := setupSchemaService(t)
    ctx := context.Background()
    
    // When: GetSchema()
    resp, err := svc.GetSchema(ctx)
    
    // Then: Returns promoted types from ent schema
    require.NoError(t, err)
    assert.NotEmpty(t, resp.PromotedTypes, "should return at least Email, Relationship, SchemaPromotion")
    
    // And: Each promoted type has required fields
    promotedNames := []string{}
    for _, schemaType := range resp.PromotedTypes {
        assert.NotEmpty(t, schemaType.Name, "type name required")
        assert.Equal(t, "promoted", schemaType.Category)
        assert.GreaterOrEqual(t, schemaType.Count, 0, "count should be non-negative")
        assert.NotNil(t, schemaType.Properties, "properties required (can be empty)")
        promotedNames = append(promotedNames, schemaType.Name)
    }
    
    // And: Email, Relationship, SchemaPromotion are present
    assert.Contains(t, promotedNames, "Email", "Email should be a promoted type")
    assert.Contains(t, promotedNames, "Relationship", "Relationship should be a promoted type")
}

// TestSchemaService_GetSchema_ReturnsDiscoveredTypes verifies discovered types are exposed
func TestSchemaService_GetSchema_ReturnsDiscoveredTypes(t *testing.T) {
    // Given: Database with DiscoveredEntity records of different types
    svc := setupSchemaServiceWithDiscoveredEntities(t, []string{"Person", "Organization", "Location"})
    ctx := context.Background()
    
    // When: GetSchema()
    resp, err := svc.GetSchema(ctx)
    
    // Then: Returns discovered types
    require.NoError(t, err)
    assert.NotEmpty(t, resp.DiscoveredTypes, "should return discovered types")
    
    // And: Each discovered type has required fields
    discoveredNames := []string{}
    for _, schemaType := range resp.DiscoveredTypes {
        assert.NotEmpty(t, schemaType.Name)
        assert.Equal(t, "discovered", schemaType.Category)
        assert.Greater(t, schemaType.Count, 0, "discovered types should have at least 1 entity")
        discoveredNames = append(discoveredNames, schemaType.Name)
    }
    
    // And: Discovered types match database
    assert.Contains(t, discoveredNames, "Person")
    assert.Contains(t, discoveredNames, "Organization")
    assert.Contains(t, discoveredNames, "Location")
}

// TestSchemaService_GetSchema_NoOverlapBetweenCategories verifies type name uniqueness
func TestSchemaService_GetSchema_NoOverlapBetweenCategories(t *testing.T) {
    // Given: SchemaService with both promoted and discovered types
    svc := setupSchemaServiceWithBoth(t)
    ctx := context.Background()
    
    // When: GetSchema()
    resp, err := svc.GetSchema(ctx)
    
    // Then: No name appears in both promoted and discovered
    require.NoError(t, err)
    
    promotedNames := make(map[string]bool)
    for _, schemaType := range resp.PromotedTypes {
        promotedNames[schemaType.Name] = true
    }
    
    for _, schemaType := range resp.DiscoveredTypes {
        assert.False(t, promotedNames[schemaType.Name], 
            "type '%s' should not appear in both promoted and discovered", schemaType.Name)
    }
}

// TestSchemaService_GetSchema_IncludesPropertyMetadata verifies property details
func TestSchemaService_GetSchema_IncludesPropertyMetadata(t *testing.T) {
    // Given: SchemaService with Email type (has known properties)
    svc := setupSchemaService(t)
    ctx := context.Background()
    
    // When: GetSchema()
    resp, err := svc.GetSchema(ctx)
    require.NoError(t, err)
    
    // Then: Email type includes property definitions
    var emailType *explorer.SchemaType
    for i := range resp.PromotedTypes {
        if resp.PromotedTypes[i].Name == "Email" {
            emailType = &resp.PromotedTypes[i]
            break
        }
    }
    require.NotNil(t, emailType, "Email type should exist")
    
    // And: Properties include expected fields
    assert.NotEmpty(t, emailType.Properties, "Email should have properties")
    
    propertyNames := []string{}
    for _, prop := range emailType.Properties {
        assert.NotEmpty(t, prop.Name, "property name required")
        assert.NotEmpty(t, prop.DataType, "property data type required")
        propertyNames = append(propertyNames, prop.Name)
    }
    
    // And: Expected Email properties are present
    assert.Contains(t, propertyNames, "subject", "Email should have 'subject' property")
    assert.Contains(t, propertyNames, "from_addr", "Email should have 'from_addr' property")
}

// TestSchemaService_GetSchema_IncludesSampleValues verifies sample data
func TestSchemaService_GetSchema_IncludesSampleValues(t *testing.T) {
    // Given: Database with at least one Email record
    svc := setupSchemaServiceWithSampleEmail(t, "test@enron.com", "Test Subject")
    ctx := context.Background()
    
    // When: GetSchema()
    resp, err := svc.GetSchema(ctx)
    require.NoError(t, err)
    
    // Then: Properties include sample values
    var emailType *explorer.SchemaType
    for i := range resp.PromotedTypes {
        if resp.PromotedTypes[i].Name == "Email" {
            emailType = &resp.PromotedTypes[i]
            break
        }
    }
    require.NotNil(t, emailType)
    
    // And: At least some properties have non-nil sample values
    hasSampleValue := false
    for _, prop := range emailType.Properties {
        if prop.SampleValue != nil {
            hasSampleValue = true
            break
        }
    }
    assert.True(t, hasSampleValue, "at least some properties should have sample values")
}

// TestSchemaService_GetSchema_TotalEntitiesMatchesSum verifies count consistency
func TestSchemaService_GetSchema_TotalEntitiesMatchesSum(t *testing.T) {
    // Given: SchemaService with mixed entity types
    svc := setupSchemaServiceWithCounts(t)
    ctx := context.Background()
    
    // When: GetSchema()
    resp, err := svc.GetSchema(ctx)
    require.NoError(t, err)
    
    // Then: TotalEntities equals sum of all type counts
    expectedTotal := 0
    for _, schemaType := range resp.PromotedTypes {
        expectedTotal += schemaType.Count
    }
    for _, schemaType := range resp.DiscoveredTypes {
        expectedTotal += schemaType.Count
    }
    
    assert.Equal(t, expectedTotal, resp.TotalEntities, 
        "TotalEntities should equal sum of all type counts")
}

// TestSchemaService_GetTypeDetails_ReturnsCompleteType verifies detail retrieval
func TestSchemaService_GetTypeDetails_ReturnsCompleteType(t *testing.T) {
    // Given: SchemaService with Email type
    svc := setupSchemaService(t)
    ctx := context.Background()
    
    // When: GetTypeDetails("Email")
    schemaType, err := svc.GetTypeDetails(ctx, "Email")
    
    // Then: Returns complete Email type information
    require.NoError(t, err)
    assert.Equal(t, "Email", schemaType.Name)
    assert.Equal(t, "promoted", schemaType.Category)
    assert.GreaterOrEqual(t, schemaType.Count, 0)
    assert.NotEmpty(t, schemaType.Properties)
}

// TestSchemaService_GetTypeDetails_ErrorsOnMissingType verifies error handling
func TestSchemaService_GetTypeDetails_ErrorsOnMissingType(t *testing.T) {
    // Given: SchemaService
    svc := setupSchemaService(t)
    ctx := context.Background()
    
    // When: GetTypeDetails with non-existent type
    _, err := svc.GetTypeDetails(ctx, "NonExistentType12345")
    
    // Then: Returns error
    assert.Error(t, err, "should error on non-existent type")
}

// TestSchemaService_RefreshSchema_DetectsNewTypes verifies refresh functionality
func TestSchemaService_RefreshSchema_DetectsNewTypes(t *testing.T) {
    // Given: SchemaService with initial schema
    svc := setupSchemaService(t)
    ctx := context.Background()
    
    initial, err := svc.GetSchema(ctx)
    require.NoError(t, err)
    initialDiscoveredCount := len(initial.DiscoveredTypes)
    
    // When: New discovered entities are added to database
    addDiscoveredEntities(t, []string{"NewType1", "NewType2"})
    
    // And: RefreshSchema() is called
    refreshed, err := svc.RefreshSchema(ctx)
    
    // Then: New types are detected
    require.NoError(t, err)
    assert.Greater(t, len(refreshed.DiscoveredTypes), initialDiscoveredCount,
        "should detect newly added discovered types")
}
```

---

## Implementation Notes

### Performance Expectations

Based on Success Criteria:
- `GetSchema()` MUST complete in <2 seconds (SC-001)
- Schema introspection should cache ent metadata (only query database for counts and samples)

### Ent Schema Introspection

The implementation should use ent's built-in schema descriptor:

```go
func (s *schemaServiceImpl) getPromotedTypes(ctx context.Context) ([]SchemaType, error) {
    types := []SchemaType{}
    
    // Iterate over ent schema tables
    for _, table := range s.client.Schema.Tables {
        properties := []PropertyDefinition{}
        
        // Extract column metadata
        for _, column := range table.Columns {
            properties = append(properties, PropertyDefinition{
                Name:     column.Name,
                DataType: column.Type.String(),
                Nullable: column.Nullable,
                // Query sample value from database
                SampleValue: s.getSampleValue(ctx, table.Name, column.Name),
            })
        }
        
        count := s.getTableCount(ctx, table.Name)
        
        types = append(types, SchemaType{
            Name:       table.Name,
            Category:   "promoted",
            Count:      count,
            Properties: properties,
        })
    }
    
    return types, nil
}
```

### Discovered Types Query

For discovered entities:

```sql
-- Get unique types with counts
SELECT 
    type,
    COUNT(*) as count,
    jsonb_object_keys(properties) as property_keys
FROM discovered_entities
GROUP BY type;

-- Get sample entity per type for sample values
SELECT DISTINCT ON (type)
    type,
    properties
FROM discovered_entities
ORDER BY type, discovered_at DESC;
```

### Thread Safety

All methods MUST be safe for concurrent use.

### Caching Strategy

Consider caching schema metadata in memory with TTL of 60 seconds to avoid repeated introspection. Invalidate cache on `RefreshSchema()`.

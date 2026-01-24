package analyst

import (
	"testing"
)

// T075: Unit tests for pattern detection
// Tests frequency calculation, relationship density, property consistency, grouping by type_category

func TestCalculateFrequency(t *testing.T) {
	tests := []struct {
		name           string
		entities       []TypeGroup
		expectedCounts map[string]int
	}{
		{
			name: "single type with multiple entities",
			entities: []TypeGroup{
				{Type: "person", Count: 150},
			},
			expectedCounts: map[string]int{
				"person": 150,
			},
		},
		{
			name: "multiple types with different frequencies",
			entities: []TypeGroup{
				{Type: "person", Count: 200},
				{Type: "organization", Count: 75},
				{Type: "concept", Count: 50},
			},
			expectedCounts: map[string]int{
				"person":       200,
				"organization": 75,
				"concept":      50,
			},
		},
		{
			name:           "empty entities list",
			entities:       []TypeGroup{},
			expectedCounts: map[string]int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateFrequency(tt.entities)

			if len(result) != len(tt.expectedCounts) {
				t.Errorf("Expected %d type groups, got %d", len(tt.expectedCounts), len(result))
			}

			for _, group := range result {
				expected, ok := tt.expectedCounts[group.Type]
				if !ok {
					t.Errorf("Unexpected type '%s' in result", group.Type)
					continue
				}
				if group.Count != expected {
					t.Errorf("Type '%s': expected count %d, got %d", group.Type, expected, group.Count)
				}
			}
		})
	}
}

func TestCalculateRelationshipDensity(t *testing.T) {
	tests := []struct {
		name            string
		entities        []EntityWithRelationships
		expectedDensity map[string]float64
	}{
		{
			name: "entities with relationships",
			entities: []EntityWithRelationships{
				{Type: "person", ID: 1, RelationshipCount: 10},
				{Type: "person", ID: 2, RelationshipCount: 20},
				{Type: "person", ID: 3, RelationshipCount: 30},
			},
			expectedDensity: map[string]float64{
				"person": 20.0, // average: (10 + 20 + 30) / 3 = 20
			},
		},
		{
			name: "multiple types with different densities",
			entities: []EntityWithRelationships{
				{Type: "person", ID: 1, RelationshipCount: 10},
				{Type: "person", ID: 2, RelationshipCount: 20},
				{Type: "organization", ID: 3, RelationshipCount: 50},
				{Type: "organization", ID: 4, RelationshipCount: 30},
			},
			expectedDensity: map[string]float64{
				"person":       15.0, // (10 + 20) / 2 = 15
				"organization": 40.0, // (50 + 30) / 2 = 40
			},
		},
		{
			name: "single entity with no relationships",
			entities: []EntityWithRelationships{
				{Type: "concept", ID: 1, RelationshipCount: 0},
			},
			expectedDensity: map[string]float64{
				"concept": 0.0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateRelationshipDensity(tt.entities)

			if len(result) != len(tt.expectedDensity) {
				t.Errorf("Expected %d type groups, got %d", len(tt.expectedDensity), len(result))
			}

			for entityType, density := range result {
				expected, ok := tt.expectedDensity[entityType]
				if !ok {
					t.Errorf("Unexpected type '%s' in result", entityType)
					continue
				}
				// Use small epsilon for float comparison
				if abs(density-expected) > 0.001 {
					t.Errorf("Type '%s': expected density %.2f, got %.2f", entityType, expected, density)
				}
			}
		})
	}
}

func TestCalculatePropertyConsistency(t *testing.T) {
	tests := []struct {
		name                  string
		entities              []EntityWithProperties
		expectedConsistencies map[string]map[string]float64 // type -> property -> consistency
	}{
		{
			name: "person entities with email property",
			entities: []EntityWithProperties{
				{Type: "person", Properties: map[string]interface{}{"email": "alice@enron.com", "name": "Alice"}},
				{Type: "person", Properties: map[string]interface{}{"email": "bob@enron.com", "name": "Bob"}},
				{Type: "person", Properties: map[string]interface{}{"name": "Charlie"}}, // missing email
			},
			expectedConsistencies: map[string]map[string]float64{
				"person": {
					"email": 0.6667, // 2 out of 3 have email
					"name":  1.0,    // 3 out of 3 have name
				},
			},
		},
		{
			name: "organization entities with varying properties",
			entities: []EntityWithProperties{
				{Type: "organization", Properties: map[string]interface{}{"name": "Enron", "industry": "energy"}},
				{Type: "organization", Properties: map[string]interface{}{"name": "Dynegy", "industry": "energy"}},
				{Type: "organization", Properties: map[string]interface{}{"name": "ExxonMobil"}},
			},
			expectedConsistencies: map[string]map[string]float64{
				"organization": {
					"name":     1.0,    // 3 out of 3
					"industry": 0.6667, // 2 out of 3
				},
			},
		},
		{
			name: "mixed types",
			entities: []EntityWithProperties{
				{Type: "person", Properties: map[string]interface{}{"email": "test@test.com"}},
				{Type: "person", Properties: map[string]interface{}{"email": "test2@test.com"}},
				{Type: "concept", Properties: map[string]interface{}{"description": "test"}},
			},
			expectedConsistencies: map[string]map[string]float64{
				"person": {
					"email": 1.0, // 2 out of 2 persons
				},
				"concept": {
					"description": 1.0, // 1 out of 1 concept
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculatePropertyConsistency(tt.entities)

			if len(result) != len(tt.expectedConsistencies) {
				t.Errorf("Expected %d types, got %d", len(tt.expectedConsistencies), len(result))
			}

			for entityType, propConsistencies := range result {
				expected, ok := tt.expectedConsistencies[entityType]
				if !ok {
					t.Errorf("Unexpected type '%s' in result", entityType)
					continue
				}

				if len(propConsistencies) != len(expected) {
					t.Errorf("Type '%s': expected %d properties, got %d", entityType, len(expected), len(propConsistencies))
				}

				for prop, consistency := range propConsistencies {
					expectedConsistency, ok := expected[prop]
					if !ok {
						t.Errorf("Type '%s': unexpected property '%s'", entityType, prop)
						continue
					}
					// Use small epsilon for float comparison
					if abs(consistency-expectedConsistency) > 0.01 {
						t.Errorf("Type '%s', property '%s': expected consistency %.4f, got %.4f", entityType, prop, expectedConsistency, consistency)
					}
				}
			}
		})
	}
}

func TestGroupByTypeCategory(t *testing.T) {
	tests := []struct {
		name          string
		entities      []Entity
		expectedTypes []string
	}{
		{
			name: "group entities by type",
			entities: []Entity{
				{ID: 1, Type: "person"},
				{ID: 2, Type: "person"},
				{ID: 3, Type: "organization"},
				{ID: 4, Type: "concept"},
				{ID: 5, Type: "person"},
			},
			expectedTypes: []string{"person", "organization", "concept"},
		},
		{
			name: "single type",
			entities: []Entity{
				{ID: 1, Type: "person"},
				{ID: 2, Type: "person"},
			},
			expectedTypes: []string{"person"},
		},
		{
			name:          "empty entities",
			entities:      []Entity{},
			expectedTypes: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GroupByTypeCategory(tt.entities)

			if len(result) != len(tt.expectedTypes) {
				t.Errorf("Expected %d type groups, got %d", len(tt.expectedTypes), len(result))
			}

			foundTypes := make(map[string]bool)
			for _, group := range result {
				foundTypes[group.Type] = true
			}

			for _, expectedType := range tt.expectedTypes {
				if !foundTypes[expectedType] {
					t.Errorf("Expected type '%s' not found in result", expectedType)
				}
			}
		})
	}
}

// Helper function for absolute value
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

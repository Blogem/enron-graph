package analyst

import (
	"context"

	"github.com/Blogem/enron-graph/ent"
)

// T080: Pattern detection implementation
// Query discovered entities grouped by type_category, calculate frequency, density, and consistency

// TypeGroup represents a group of entities by type with their count
type TypeGroup struct {
	Type  string
	Count int
}

// EntityWithRelationships represents an entity with its relationship count
type EntityWithRelationships struct {
	ID                int
	Type              string
	RelationshipCount int
}

// EntityWithProperties represents an entity with its properties
type EntityWithProperties struct {
	Type       string
	Properties map[string]interface{}
}

// Entity represents a basic entity
type Entity struct {
	ID   int
	Type string
}

// CalculateFrequency counts the number of entities per type
func CalculateFrequency(entities []TypeGroup) []TypeGroup {
	return entities
}

// CalculateRelationshipDensity calculates average relationships per entity type
func CalculateRelationshipDensity(entities []EntityWithRelationships) map[string]float64 {
	// Group relationship counts by type
	typeGroups := make(map[string][]int)
	for _, e := range entities {
		typeGroups[e.Type] = append(typeGroups[e.Type], e.RelationshipCount)
	}

	// Calculate average for each type
	result := make(map[string]float64)
	for entityType, counts := range typeGroups {
		sum := 0
		for _, count := range counts {
			sum += count
		}
		if len(counts) > 0 {
			result[entityType] = float64(sum) / float64(len(counts))
		}
	}

	return result
}

// CalculatePropertyConsistency calculates the percentage of entities with each property per type
func CalculatePropertyConsistency(entities []EntityWithProperties) map[string]map[string]float64 {
	// Count entities per type
	typeCounts := make(map[string]int)
	// Count properties per type
	propertyPresence := make(map[string]map[string]int)

	for _, e := range entities {
		typeCounts[e.Type]++

		if propertyPresence[e.Type] == nil {
			propertyPresence[e.Type] = make(map[string]int)
		}

		for prop := range e.Properties {
			propertyPresence[e.Type][prop]++
		}
	}

	// Calculate consistency percentages
	result := make(map[string]map[string]float64)
	for entityType, props := range propertyPresence {
		if result[entityType] == nil {
			result[entityType] = make(map[string]float64)
		}

		totalEntities := typeCounts[entityType]
		for prop, count := range props {
			if totalEntities > 0 {
				result[entityType][prop] = float64(count) / float64(totalEntities)
			}
		}
	}

	return result
}

// GroupByTypeCategory groups entities by their type
func GroupByTypeCategory(entities []Entity) []TypeGroup {
	typeMap := make(map[string]int)
	for _, e := range entities {
		typeMap[e.Type]++
	}

	result := make([]TypeGroup, 0, len(typeMap))
	for entityType, count := range typeMap {
		result = append(result, TypeGroup{
			Type:  entityType,
			Count: count,
		})
	}

	return result
}

// PatternStats holds statistics for a type pattern
type PatternStats struct {
	Type                string
	Frequency           int
	TotalRelationships  int
	AvgDensity          float64
	Properties          map[string]int
	PropertyConsistency map[string]float64
}

// DetectPatterns analyzes discovered entities and returns statistics
func DetectPatterns(ctx context.Context, client *ent.Client) (map[string]*PatternStats, error) {
	// Query all discovered entities
	entities, err := client.DiscoveredEntity.
		Query().
		All(ctx)
	if err != nil {
		return nil, err
	}

	// Group by type
	typeGroups := make(map[string]*PatternStats)
	entityTypeMap := make(map[int]string) // Track entity ID to type mapping

	for _, entity := range entities {
		if typeGroups[entity.TypeCategory] == nil {
			typeGroups[entity.TypeCategory] = &PatternStats{
				Type:       entity.TypeCategory,
				Frequency:  0,
				Properties: make(map[string]int),
			}
		}

		stats := typeGroups[entity.TypeCategory]
		entityTypeMap[entity.ID] = entity.TypeCategory

		// Count frequency
		stats.Frequency++

		// Parse and count properties
		for prop := range entity.Properties {
			stats.Properties[prop]++
		}
	}

	// Query all relationships involving discovered entities
	relationships, err := client.Relationship.
		Query().
		All(ctx)
	if err != nil {
		return nil, err
	}

	// Count relationships per entity type
	for _, rel := range relationships {
		// Count relationships where discovered_entity is the source
		if rel.FromType == "discovered_entity" {
			if entityType, ok := entityTypeMap[rel.FromID]; ok {
				if typeGroups[entityType] != nil {
					typeGroups[entityType].TotalRelationships++
				}
			}
		}
		// Count relationships where discovered_entity is the target
		if rel.ToType == "discovered_entity" {
			if entityType, ok := entityTypeMap[rel.ToID]; ok {
				if typeGroups[entityType] != nil {
					typeGroups[entityType].TotalRelationships++
				}
			}
		}
	}

	// Calculate averages and consistency
	for _, stats := range typeGroups {
		if stats.Frequency > 0 {
			stats.AvgDensity = float64(stats.TotalRelationships) / float64(stats.Frequency)

			// Calculate property consistency
			stats.PropertyConsistency = make(map[string]float64)
			for prop, count := range stats.Properties {
				stats.PropertyConsistency[prop] = float64(count) / float64(stats.Frequency)
			}
		}
	}

	return typeGroups, nil
}

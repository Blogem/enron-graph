package analyst

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/Blogem/enron-graph/ent"
	"github.com/Blogem/enron-graph/ent/discoveredentity"
)

// T083: Schema generator implementation
// Infer required/optional properties, data types, validation rules

// EntitySample represents an entity with its properties for analysis
type EntitySample struct {
	Properties map[string]interface{}
}

// ValidationRule represents a validation constraint
type ValidationRule struct {
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}

// PropertyDefinition defines a property in the schema
type PropertyDefinition struct {
	Type            string           `json:"type"`
	Required        bool             `json:"required"`
	ValidationRules []ValidationRule `json:"validation_rules,omitempty"`
}

// SchemaDefinition represents a complete schema for a type
type SchemaDefinition struct {
	Type       string                        `json:"type"`
	Properties map[string]PropertyDefinition `json:"properties"`
}

// InferRequiredProperties identifies properties that appear in >threshold% of entities
func InferRequiredProperties(entities []EntitySample, threshold float64) []string {
	if len(entities) == 0 {
		return []string{}
	}

	// Count property occurrences
	propertyCounts := make(map[string]int)
	for _, entity := range entities {
		for prop := range entity.Properties {
			propertyCounts[prop]++
		}
	}

	// Calculate percentages and filter
	required := []string{}
	total := float64(len(entities))
	for prop, count := range propertyCounts {
		if float64(count)/total >= threshold {
			required = append(required, prop)
		}
	}

	return required
}

// InferOptionalProperties identifies properties in the minThreshold-maxThreshold range
func InferOptionalProperties(entities []EntitySample, minThreshold, maxThreshold float64) []string {
	if len(entities) == 0 {
		return []string{}
	}

	// Count property occurrences
	propertyCounts := make(map[string]int)
	for _, entity := range entities {
		for prop := range entity.Properties {
			propertyCounts[prop]++
		}
	}

	// Calculate percentages and filter
	optional := []string{}
	total := float64(len(entities))
	for prop, count := range propertyCounts {
		percentage := float64(count) / total
		if percentage >= minThreshold && percentage < maxThreshold {
			optional = append(optional, prop)
		}
	}

	return optional
}

// InferDataType determines the data type from sample values
func InferDataType(samples []interface{}) string {
	if len(samples) == 0 {
		return "string"
	}

	// Check types of samples
	hasInt := false
	hasFloat := false
	hasString := false
	hasBool := false

	for _, sample := range samples {
		switch sample.(type) {
		case int, int32, int64:
			hasInt = true
		case float32, float64:
			hasFloat = true
		case string:
			hasString = true
		case bool:
			hasBool = true
		}
	}

	// Determine predominant type
	if hasBool && !hasInt && !hasFloat && !hasString {
		return "boolean"
	}
	if hasFloat {
		return "number"
	}
	if hasInt && !hasFloat {
		return "integer"
	}

	return "string"
}

// GenerateValidationRules creates validation rules based on property name and samples
func GenerateValidationRules(property, dataType string, samples []interface{}) []ValidationRule {
	rules := []ValidationRule{}

	// Email validation
	if strings.Contains(strings.ToLower(property), "email") {
		rules = append(rules, ValidationRule{
			Type:  "format",
			Value: "email",
		})
	}

	// String length validation
	if dataType == "string" {
		rules = append(rules, ValidationRule{
			Type:  "minLength",
			Value: 1,
		})

		// Check if it's not an email (emails can be longer)
		if !strings.Contains(strings.ToLower(property), "email") {
			rules = append(rules, ValidationRule{
				Type:  "maxLength",
				Value: 100,
			})
		}
	}

	// Integer minimum validation
	if dataType == "integer" {
		rules = append(rules, ValidationRule{
			Type:  "minimum",
			Value: 0,
		})
	}

	return rules
}

// GenerateJSONSchema creates a complete JSON schema from entity samples
func GenerateJSONSchema(typeName string, entities []EntitySample) SchemaDefinition {
	schema := SchemaDefinition{
		Type:       typeName,
		Properties: make(map[string]PropertyDefinition),
	}

	if len(entities) == 0 {
		return schema
	}

	// Infer required and optional properties
	required := InferRequiredProperties(entities, 0.90)
	optional := InferOptionalProperties(entities, 0.30, 0.90)

	requiredMap := make(map[string]bool)
	for _, prop := range required {
		requiredMap[prop] = true
	}

	// Collect all properties
	allProps := make(map[string]bool)
	for _, prop := range required {
		allProps[prop] = true
	}
	for _, prop := range optional {
		allProps[prop] = true
	}

	// Collect samples for each property
	propertySamples := make(map[string][]interface{})
	for _, entity := range entities {
		for prop, value := range entity.Properties {
			if allProps[prop] {
				propertySamples[prop] = append(propertySamples[prop], value)
			}
		}
	}

	// Generate property definitions
	for prop := range allProps {
		samples := propertySamples[prop]
		dataType := InferDataType(samples)

		propDef := PropertyDefinition{
			Type:            dataType,
			Required:        requiredMap[prop],
			ValidationRules: GenerateValidationRules(prop, dataType, samples),
		}

		schema.Properties[prop] = propDef
	}

	return schema
}

// GenerateSchemaForType generates a schema definition for a specific entity type
func GenerateSchemaForType(ctx context.Context, client *ent.Client, typeName string) (*SchemaDefinition, error) {
	// Query entities of the specified type
	entities, err := client.DiscoveredEntity.
		Query().
		Where(discoveredentity.TypeCategory(typeName)).
		Limit(1000). // Sample limit to avoid memory issues
		All(ctx)
	if err != nil {
		return nil, err
	}

	if len(entities) == 0 {
		return nil, fmt.Errorf("no entities found for type: %s", typeName)
	}

	// Convert to EntitySample
	samples := make([]EntitySample, 0, len(entities))
	for _, entity := range entities {
		samples = append(samples, EntitySample{
			Properties: entity.Properties,
		})
	}

	schema := GenerateJSONSchema(typeName, samples)
	return &schema, nil
}

// Email regex pattern
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

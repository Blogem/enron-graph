// Package registry maintains a global mapping of promoted entity type names
// to their corresponding Ent creation functions.
//
// The registry enables runtime routing of discovered entities to their promoted
// schemas when available. When a type is promoted (via the promoter workflow),
// code generation automatically registers the new schema, making it available
// to the extractor for subsequent entity creation.
//
// Usage:
//   - Registration: Typically done automatically via generated code in init() functions
//   - Lookup: Check if a type is promoted before creating entities
//
// Example:
//
//	if createFn, exists := registry.PromotedTypes["Person"]; exists {
//	    entity, err := createFn(ctx, properties)
//	    // Use promoted schema
//	} else {
//	    // Fallback to generic DiscoveredEntity
//	}
package registry

import "context"

// EntityCreator is a function that creates an entity of a promoted type.
// It accepts a context (which should contain the Ent client) and a map of
// property values extracted from the LLM, and returns the created entity.
//
// The function should:
//   - Extract the Ent client from context
//   - Map properties from the data map to the appropriate Ent builder setters
//   - Handle missing/nil properties gracefully
//   - Return the created entity or an error
type EntityCreator func(ctx context.Context, data map[string]any) (any, error)

// PromotedTypes maps entity type names to their creation functions.
// This map is populated automatically during initialization via generated code
// in ent/registry_generated.go.
//
// Keys are Ent schema names (e.g., "Person", "Organization").
// Values are EntityCreator functions that know how to create entities of that type.
var PromotedTypes = make(map[string]EntityCreator)

// EntityFinder is a function that finds an entity of a promoted type by unique_id.
// It accepts a context (which should contain the Ent client) and a unique_id string,
// and returns the found entity.
//
// The function should:
//   - Extract the Ent client from context
//   - Query the promoted table by unique_id field
//   - Return the found entity or an error (including sql.ErrNoRows for not found)
type EntityFinder func(ctx context.Context, uniqueID string) (any, error)

// PromotedFinders maps entity type names to their finder functions.
// This map is populated automatically during initialization via generated code
// in ent/registry.go, in parallel to PromotedTypes.
//
// Keys are Ent schema names (e.g., "Person", "Organization").
// Values are EntityFinder functions that know how to find entities of that type.
var PromotedFinders = make(map[string]EntityFinder)

// Register adds a new promoted type to the global registry.
// This function is typically called from generated code during package initialization.
//
// Parameters:
//   - typeName: The name of the Ent schema (e.g., "Person")
//   - fn: The EntityCreator function that creates entities of this type
func Register(typeName string, fn EntityCreator) {
	PromotedTypes[typeName] = fn
}

// RegisterFinder adds a new entity finder to the global registry.
// This function is typically called from generated code during package initialization.
//
// Parameters:
//   - typeName: The name of the Ent schema (e.g., "Person")
//   - fn: The EntityFinder function that finds entities of this type
func RegisterFinder(typeName string, fn EntityFinder) {
	PromotedFinders[typeName] = fn
}

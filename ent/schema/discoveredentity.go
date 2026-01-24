package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// DiscoveredEntity holds the schema definition for the DiscoveredEntity entity.
type DiscoveredEntity struct {
	ent.Schema
}

// Fields of the DiscoveredEntity.
func (DiscoveredEntity) Fields() []ent.Field {
	return []ent.Field{
		field.String("unique_id").
			Unique().
			NotEmpty().
			Comment("Unique identifier (email, domain, etc.)"),
		field.String("type_category").
			NotEmpty().
			Comment("Entity type: person, organization, concept, etc."),
		field.String("name").
			NotEmpty().
			Comment("Entity name or label"),
		field.JSON("properties", map[string]interface{}{}).
			Optional().
			SchemaType(map[string]string{
				dialect.Postgres: "jsonb",
			}).
			Comment("Flexible properties stored as JSONB"),
		field.JSON("embedding", []float32{}).
			Optional().
			Comment("Vector embedding (stored as JSON, used with pgvector)"),
		field.Float("confidence_score").
			Default(0.0).
			Min(0.0).
			Max(1.0).
			Comment("Confidence score for entity extraction (0-1)"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

// Edges of the DiscoveredEntity.
func (DiscoveredEntity) Edges() []ent.Edge {
	return nil
}

// Indexes of the DiscoveredEntity.
func (DiscoveredEntity) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("type_category"),
		index.Fields("name"),
		index.Fields("confidence_score"),
	}
}

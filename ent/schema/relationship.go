package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Relationship holds the schema definition for the Relationship entity.
type Relationship struct {
	ent.Schema
}

// Fields of the Relationship.
func (Relationship) Fields() []ent.Field {
	return []ent.Field{
		field.String("type").
			NotEmpty().
			Comment("Relationship type: SENT, RECEIVED, MENTIONS, COMMUNICATES_WITH"),
		field.String("from_type").
			NotEmpty().
			Comment("Source entity type: email, discovered_entity, person, etc."),
		field.Int("from_id").
			Positive().
			Comment("Source entity ID"),
		field.String("to_type").
			NotEmpty().
			Comment("Target entity type"),
		field.Int("to_id").
			Positive().
			Comment("Target entity ID"),
		field.Time("timestamp").
			Default(time.Now).
			Comment("When the relationship occurred or was created"),
		field.Float("confidence_score").
			Default(1.0).
			Min(0.0).
			Max(1.0).
			Comment("Confidence score for relationship (0-1)"),
		field.JSON("properties", map[string]interface{}{}).
			Optional().
			SchemaType(map[string]string{
				dialect.Postgres: "jsonb",
			}).
			Comment("Additional metadata as JSONB"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

// Edges of the Relationship.
func (Relationship) Edges() []ent.Edge {
	return nil
}

// Indexes of the Relationship.
func (Relationship) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("type"),
		index.Fields("from_type", "from_id"),
		index.Fields("to_type", "to_id"),
		index.Fields("timestamp"),
	}
}

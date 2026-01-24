package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// SchemaPromotion holds the schema definition for the SchemaPromotion entity.
type SchemaPromotion struct {
	ent.Schema
}

// Fields of the SchemaPromotion.
func (SchemaPromotion) Fields() []ent.Field {
	return []ent.Field{
		field.String("type_name").
			NotEmpty().
			Comment("Name of the promoted type (e.g., Person, Organization)"),
		field.Time("promoted_at").
			Default(time.Now).
			Comment("When the type was promoted to core schema"),
		field.JSON("promotion_criteria", map[string]interface{}{}).
			Optional().
			SchemaType(map[string]string{
				dialect.Postgres: "jsonb",
			}).
			Comment("Criteria used for promotion: frequency, density, consistency"),
		field.Int("entities_affected").
			Default(0).
			NonNegative().
			Comment("Number of entities migrated to new type"),
		field.Int("validation_failures").
			Default(0).
			NonNegative().
			Comment("Number of entities that failed validation"),
		field.JSON("schema_definition", map[string]interface{}{}).
			Optional().
			SchemaType(map[string]string{
				dialect.Postgres: "jsonb",
			}).
			Comment("Generated schema definition rules"),
	}
}

// Edges of the SchemaPromotion.
func (SchemaPromotion) Edges() []ent.Edge {
	return nil
}

// Indexes of the SchemaPromotion.
func (SchemaPromotion) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("type_name"),
		index.Fields("promoted_at"),
	}
}

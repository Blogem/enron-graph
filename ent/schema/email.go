package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Email holds the schema definition for the Email entity.
type Email struct {
	ent.Schema
}

// Fields of the Email.
func (Email) Fields() []ent.Field {
	return []ent.Field{
		field.String("message_id").
			Unique().
			NotEmpty().
			Comment("Unique message identifier from email headers"),
		field.String("from").
			NotEmpty().
			Comment("Sender email address"),
		field.JSON("to", []string{}).
			Optional().
			Comment("Recipient email addresses"),
		field.JSON("cc", []string{}).
			Optional().
			Comment("CC email addresses"),
		field.JSON("bcc", []string{}).
			Optional().
			Comment("BCC email addresses"),
		field.String("subject").
			Default("").
			Comment("Email subject line"),
		field.Time("date").
			Default(time.Now).
			Comment("Email send date"),
		field.Text("body").
			Default("").
			Comment("Email body content"),
		field.String("file_path").
			Optional().
			Comment("Original file path in dataset"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

// Edges of the Email.
func (Email) Edges() []ent.Edge {
	return nil
}

// Indexes of the Email.
func (Email) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("date"),
		index.Fields("from"),
	}
}

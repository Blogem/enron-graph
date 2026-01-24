package graph

import (
	"context"

	"github.com/Blogem/enron-graph/ent"
	"github.com/Blogem/enron-graph/ent/migrate"
)

// CreateIndexes creates database indexes for optimized queries
func CreateIndexes(ctx context.Context, client *ent.Client) error {
	// Use ent's migration system to create indexes
	// Indexes are typically defined in the schema files using ent's
	// Index() method, but we can also create them manually

	// Get the migration driver
	err := client.Schema.Create(
		ctx,
		migrate.WithDropIndex(true),
		migrate.WithDropColumn(true),
	)
	if err != nil {
		return err
	}

	return nil
}

// Note: Indexes should be defined in ent schema files using:
//
// func (DiscoveredEntity) Indexes() []ent.Index {
//     return []ent.Index{
//         index.Fields("name"),
//         index.Fields("type_category"),
//         index.Fields("unique_id").Unique(),
//         index.Fields("type_category", "name"),
//         index.Fields("confidence_score"),
//     }
// }
//
// This file is provided for manual index management if needed.

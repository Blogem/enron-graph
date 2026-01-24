package main

import (
	"context"
	"fmt"
	"log"

	_ "github.com/lib/pq"
	"github.com/Blogem/enron-graph/ent"
	"github.com/Blogem/enron-graph/pkg/utils"
)

func main() {
	cfg, err := utils.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	client, err := ent.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	ctx := context.Background()

	// Count entities
	entityCount, _ := client.DiscoveredEntity.Query().Count(ctx)
	fmt.Printf("Total Entities: %d\n", entityCount)

	// List entities
	entities, _ := client.DiscoveredEntity.Query().All(ctx)
	for _, e := range entities {
		fmt.Printf("  - %s (%s) [confidence: %.2f]\n", e.Name, e.TypeCategory, e.ConfidenceScore)
	}

	// Count relationships
	relCount, _ := client.Relationship.Query().Count(ctx)
	fmt.Printf("\nTotal Relationships: %d\n", relCount)

	// List relationships
	rels, _ := client.Relationship.Query().All(ctx)
	for _, r := range rels {
		fmt.Printf("  - [%s] (from_id=%d, to_id=%d, confidence=%.2f)\n", 
			r.Type, r.FromID, r.ToID, r.ConfidenceScore)
	}
}
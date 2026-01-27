package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/Blogem/enron-graph/ent"
	_ "github.com/lib/pq"
)

func main() {
	// Get node count from args (default 1000)
	nodeCount := 1000
	if len(os.Args) > 1 {
		fmt.Sscanf(os.Args[1], "%d", &nodeCount)
	}

	// Connect to database
	connStr := "host=localhost port=5432 user=enron password=enron123 dbname=enron_graph sslmode=disable"
	if env := os.Getenv("DATABASE_URL"); env != "" {
		connStr = env
	}

	client, err := ent.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	log.Printf("🌱 Seeding %d performance test nodes...", nodeCount)

	// Create nodes with varying degrees of connectivity
	entityIDs := make([]int, nodeCount)

	for i := 0; i < nodeCount; i++ {
		entity, err := client.DiscoveredEntity.Create().
			SetUniqueID(fmt.Sprintf("perf-node-%d", i)).
			SetTypeCategory(getTypeForIndex(i)).
			SetName(fmt.Sprintf("Node %d", i)).
			SetConfidenceScore(0.80 + float64(i%20)/100.0).
			SetProperties(map[string]interface{}{
				"index":       i,
				"group":       i % 10,
				"description": fmt.Sprintf("Performance test node %d", i),
			}).
			Save(ctx)
		if err != nil {
			log.Fatalf("Failed to create node %d: %v", i, err)
		}
		entityIDs[i] = entity.ID

		if (i+1)%100 == 0 {
			log.Printf("  ✓ Created %d nodes...", i+1)
		}
	}

	// Create relationships with varying connectivity patterns
	edgeCount := 0
	for i := 0; i < nodeCount; i++ {
		connectionCount := getConnectionCountForIndex(i, nodeCount)

		for j := 0; j < connectionCount; j++ {
			targetIdx := (i + j + 1) % nodeCount
			if targetIdx != i {
				_, err := client.Relationship.Create().
					SetFromType("discovered_entity").
					SetFromID(entityIDs[i]).
					SetToType("discovered_entity").
					SetToID(entityIDs[targetIdx]).
					SetType(getRelationshipType(i, j)).
					SetConfidenceScore(0.75 + float64(j%10)/40.0).
					Save(ctx)
				if err == nil {
					edgeCount++
				}
			}
		}

		if (i+1)%100 == 0 {
			log.Printf("  ✓ Created relationships for %d nodes (%d edges so far)...", i+1, edgeCount)
		}
	}

	log.Printf("\n✅ Successfully seeded %d nodes and %d edges!", nodeCount, edgeCount)
	log.Println("\n🚀 Next steps:")
	log.Println("   1. cd cmd/explorer")
	log.Println("   2. wails dev")
	log.Println("   3. The graph will auto-load 100 nodes on startup")
	log.Println("   4. Try panning, zooming, clicking nodes, and expanding relationships!")
}

func getTypeForIndex(i int) string {
	types := []string{"person", "organization", "project", "location", "event"}
	return types[i%len(types)]
}

func getConnectionCountForIndex(i, total int) int {
	if i%10 == 0 {
		return min(20, total/10) // Hub nodes with many connections
	} else if i%5 == 0 {
		return min(10, total/20) // Medium connectivity
	}
	return min(3, total/50) // Most nodes have few connections
}

func getRelationshipType(i, j int) string {
	types := []string{"connected_to", "related_to", "works_with", "part_of", "manages"}
	return types[(i+j)%len(types)]
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

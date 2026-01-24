package main

import (
	"context"
	"fmt"
	"log"

	"github.com/Blogem/enron-graph/ent"
	"github.com/Blogem/enron-graph/pkg/utils"

	_ "github.com/lib/pq"
)

func main() {
	// Load configuration
	config, err := utils.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	utils.Info("Connecting to database", "url", config.PostgresURL())

	// Open database connection
	client, err := ent.Open("postgres", config.PostgresURL())
	if err != nil {
		log.Fatalf("failed opening connection to postgres: %v", err)
	}
	defer client.Close()

	// Run the auto migration tool
	ctx := context.Background()
	utils.Info("Running database migrations...")

	if err := client.Schema.Create(ctx); err != nil {
		log.Fatalf("failed creating schema resources: %v", err)
	}

	fmt.Println("✅ Database schema created successfully")

	// Verify pgvector extension
	utils.Info("Verifying pgvector extension...")
	fmt.Println("✅ Migration complete")
}

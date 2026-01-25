package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/Blogem/enron-graph/ent"
	"github.com/Blogem/enron-graph/internal/graph"
	"github.com/Blogem/enron-graph/internal/tui"
	"github.com/Blogem/enron-graph/pkg/llm"
	"github.com/Blogem/enron-graph/pkg/utils"
	tea "github.com/charmbracelet/bubbletea"
	_ "github.com/lib/pq"
)

func main() {
	// Initialize logger
	logger := utils.NewLogger()
	logger.Info("Starting TUI application")

	// Load configuration
	cfg, err := utils.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to database
	client, err := ent.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer client.Close()

	// Initialize repository
	repo := graph.NewRepository(client)

	// Initialize LLM client (optional - chat will still work without it)
	llmClient := llm.NewOllamaClient(
		cfg.OllamaURL,
		"llama3.1:8b",
		"mxbai-embed-large",
		logger,
	)
	logger.Info("LLM client initialized", "ollama_url", cfg.OllamaURL)

	// Create TUI model with repository
	model := tui.NewModel(repo)

	// Set LLM client for chat functionality
	model.SetLLMClient(llmClient)

	// Load initial data
	ctx := context.Background()
	entities, err := loadEntities(ctx, repo)
	if err != nil {
		logger.Error("Failed to load entities", "error", err)
	} else {
		model.LoadEntities(entities)
	}

	// Start Bubble Tea program
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running TUI: %v\n", err)
		os.Exit(1)
	}
}

// loadEntities loads initial entities from the database
func loadEntities(ctx context.Context, repo graph.Repository) ([]tui.Entity, error) {
	// Get discovered entities
	discoveredEntities, err := repo.FindEntitiesByType(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("failed to query entities: %w", err)
	}

	// Limit to 1000 entities for performance
	if len(discoveredEntities) > 1000 {
		discoveredEntities = discoveredEntities[:1000]
	}

	// Convert to TUI entities
	entities := make([]tui.Entity, len(discoveredEntities))
	for i, de := range discoveredEntities {
		entities[i] = tui.Entity{
			ID:         de.ID,
			Type:       de.TypeCategory,
			Name:       de.Name,
			Confidence: de.ConfidenceScore,
		}
	}

	return entities, nil
}

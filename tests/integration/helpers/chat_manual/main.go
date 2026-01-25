package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/Blogem/enron-graph/ent"
	"github.com/Blogem/enron-graph/internal/chat"
	"github.com/Blogem/enron-graph/internal/graph"
	"github.com/Blogem/enron-graph/pkg/llm"
	"github.com/Blogem/enron-graph/pkg/utils"
	_ "github.com/lib/pq"
)

type chatRepoAdapter struct {
	repo graph.Repository
	ctx  context.Context
}

func newChatRepo(repo graph.Repository) chat.Repository {
	return &chatRepoAdapter{repo: repo, ctx: context.Background()}
}

func (a *chatRepoAdapter) FindEntityByName(name string) (*chat.Entity, error) {
	entities, err := a.repo.FindEntitiesByType(a.ctx, "")
	if err != nil {
		return nil, err
	}
	nameLower := strings.ToLower(name)
	for _, e := range entities {
		if strings.Contains(strings.ToLower(e.Name), nameLower) {
			return &chat.Entity{ID: e.ID, Name: e.Name, Type: e.TypeCategory, Properties: e.Properties}, nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func (a *chatRepoAdapter) TraverseRelationships(entityID int, relType string) ([]*chat.Entity, error) {
	entities, err := a.repo.TraverseRelationships(a.ctx, entityID, relType, 1)
	if err != nil {
		return nil, err
	}
	result := make([]*chat.Entity, len(entities))
	for i, e := range entities {
		result[i] = &chat.Entity{ID: e.ID, Name: e.Name, Type: e.TypeCategory, Properties: e.Properties}
	}
	return result, nil
}

func (a *chatRepoAdapter) FindShortestPath(sourceID, targetID int) ([]*chat.PathNode, error) {
	rels, err := a.repo.FindShortestPath(a.ctx, sourceID, targetID)
	if err != nil {
		return nil, err
	}
	if len(rels) == 0 {
		return nil, nil
	}

	path := []*chat.PathNode{}
	src, _ := a.repo.FindEntityByID(a.ctx, sourceID)
	path = append(path, &chat.PathNode{
		Entity: &chat.Entity{ID: src.ID, Name: src.Name, Type: src.TypeCategory, Properties: src.Properties},
	})

	for _, rel := range rels {
		tgt, _ := a.repo.FindEntityByID(a.ctx, rel.ToID)
		path = append(path, &chat.PathNode{
			Entity:       &chat.Entity{ID: tgt.ID, Name: tgt.Name, Type: tgt.TypeCategory, Properties: tgt.Properties},
			Relationship: rel.Type,
		})
	}
	return path, nil
}

func (a *chatRepoAdapter) SimilaritySearch(embedding []float32, limit int) ([]*chat.Entity, error) {
	entities, err := a.repo.SimilaritySearch(a.ctx, embedding, limit, 0.7)
	if err != nil {
		return nil, err
	}
	result := make([]*chat.Entity, len(entities))
	for i, e := range entities {
		result[i] = &chat.Entity{ID: e.ID, Name: e.Name, Type: e.TypeCategory, Properties: e.Properties}
	}
	return result, nil
}

func (a *chatRepoAdapter) CountRelationships(entityID int, relType string) (int, error) {
	rels, err := a.repo.FindRelationshipsByEntity(a.ctx, "discovered_entity", entityID)
	if err != nil {
		return 0, err
	}
	if relType == "" {
		return len(rels), nil
	}
	count := 0
	for _, rel := range rels {
		if rel.Type == relType {
			count++
		}
	}
	return count, nil
}

func main() {
	cfg, _ := utils.LoadConfig()
	client, err := ent.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	repo := graph.NewRepository(client)
	chatRepo := newChatRepo(repo)
	logger := utils.NewLogger()
	llmClient := llm.NewOllamaClient(cfg.OllamaURL, "llama3.1:8b", "mxbai-embed-large", logger)

	handler := chat.NewHandler(llmClient, chatRepo)
	chatContext := chat.NewContext()
	ctx := context.Background()

	queries := []string{
		"Who is Jeff Skilling?",
		"Who did Jeff Skilling email?",
		"How are Jeff Skilling and Kenneth Lay connected?",
	}

	for i, query := range queries {
		fmt.Print("\n" + strings.Repeat("=", 60) + "\n")
		fmt.Printf("[Query %d] %s\n", i+1, query)
		fmt.Print(strings.Repeat("-", 60) + "\n")

		// Test entity lookup directly first
		if i == 0 {
			fmt.Println("\n[DEBUG] Testing direct entity lookup:")

			// List all entities in database
			allEntities, _ := chatRepo.(*chatRepoAdapter).repo.FindEntitiesByType(chatRepo.(*chatRepoAdapter).ctx, "")
			fmt.Printf("  Found %d entities in database:\n", len(allEntities))
			for j, e := range allEntities {
				if j < 5 {
					fmt.Printf("    %d. %s (ID: %d, Type: %s)\n", j+1, e.Name, e.ID, e.TypeCategory)
				}
			}

			entity, err := chatRepo.FindEntityByName("Jeff Skilling")
			if err != nil {
				fmt.Printf("  ❌ Direct lookup failed: %v\n", err)
			} else {
				fmt.Printf("  ✓ Found entity: %s (ID: %d, Type: %s)\n", entity.Name, entity.ID, entity.Type)
			}
			fmt.Println()
		}

		response, err := handler.ProcessQuery(ctx, query, chatContext)
		if err != nil {
			fmt.Printf("❌ Error processing query: %v\n", err)
			// Don't exit, continue to see all responses
			continue
		}
		fmt.Printf("\n[LLM Response]\n%s\n", response)
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("✓ All chat queries processed")
}

package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
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

	logger := utils.NewLogger()
	repo := graph.NewRepository(client, logger)
	chatRepo := newChatRepo(repo)
	llmClient := llm.NewOllamaClient(cfg.OllamaURL, "llama3.1:8b", "mxbai-embed-large", logger)

	handler := chat.NewHandler(llmClient, chatRepo)
	chatContext := chat.NewContext()
	ctx := context.Background()

	// Interactive mode if no arguments
	if len(os.Args) == 1 {
		fmt.Println("╔════════════════════════════════════════════════════════════╗")
		fmt.Println("║              Interactive Chat Mode                         ║")
		fmt.Println("║  Type your questions or 'quit' to exit                     ║")
		fmt.Println("╚════════════════════════════════════════════════════════════╝")
		fmt.Println("")

		scanner := bufio.NewScanner(os.Stdin)
		for {
			fmt.Print("\n🤖 You: ")
			if !scanner.Scan() {
				break
			}
			query := strings.TrimSpace(scanner.Text())
			if query == "" {
				continue
			}
			if query == "quit" || query == "exit" {
				fmt.Println("\nGoodbye!")
				break
			}

			fmt.Print("💭 Thinking")
			response, err := handler.ProcessQuery(ctx, query, chatContext)
			fmt.Print("\r                    \r") // Clear "Thinking..."

			if err != nil {
				fmt.Printf("❌ Error: %v\n", err)
				continue
			}
			fmt.Printf("🔍 Assistant:\n%s\n", response)
		}
	} else {
		// Batch mode - process command line queries
		for _, query := range os.Args[1:] {
			fmt.Printf("\n📝 Query: %s\n", query)
			fmt.Print("💭 Processing...")

			response, err := handler.ProcessQuery(ctx, query, chatContext)
			fmt.Print("\r                    \r")

			if err != nil {
				fmt.Printf("❌ Error: %v\n", err)
				continue
			}
			fmt.Printf("🔍 Response:\n%s\n", response)
			fmt.Println(strings.Repeat("-", 60))
		}
	}
}

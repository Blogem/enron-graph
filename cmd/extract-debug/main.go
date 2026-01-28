package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/Blogem/enron-graph/ent"
	"github.com/Blogem/enron-graph/internal/extractor"
	"github.com/Blogem/enron-graph/internal/graph"
	"github.com/Blogem/enron-graph/pkg/llm"
	"github.com/Blogem/enron-graph/pkg/utils"

	_ "github.com/lib/pq"
)

func main() {
	// Command line flags
	emailID := flag.Int("email-id", 0, "Extract from email with specific ID")
	messageID := flag.String("message-id", "", "Extract from email with specific Message-ID")
	limit := flag.Int("limit", 1, "Number of random emails to extract from (ignored if email-id or message-id specified)")
	verbose := flag.Bool("verbose", false, "Show detailed extraction results (entities and relationships)")
	dbURL := flag.String("db", "", "Database connection URL (optional, uses config if not provided)")

	flag.Parse()

	// Initialize logger
	utils.SetLogLevel(slog.LevelDebug)
	logger := utils.NewLogger()
	logger.Info("Starting extraction debug utility",
		"email_id", *emailID,
		"message_id", *messageID,
		"limit", *limit,
		"verbose", *verbose)

	// Load configuration
	config, err := utils.LoadConfig()
	if err != nil {
		logger.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Use provided DB URL or from config
	connStr := *dbURL
	if connStr == "" {
		connStr = config.DatabaseURL
	}

	// Connect to database
	client, err := ent.Open("postgres", connStr)
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer client.Close()

	// Also open a direct SQL connection for raw queries
	sqlDB, err := sql.Open("postgres", connStr)
	if err != nil {
		logger.Error("Failed to open SQL database connection", slog.Any("error", err))
		os.Exit(1)
	}
	defer sqlDB.Close()

	// Create repository (wrapped to prevent writes)
	baseRepo := graph.NewRepositoryWithDB(client, sqlDB, logger)
	repo := NewReadOnlyRepository(baseRepo, logger)

	// Initialize LLM client based on provider
	var llmClient llm.Client
	switch config.LLMProvider {
	case "litellm":
		logger.Info("Using LiteLLM provider",
			"url", config.LiteLLMURL,
			"completion_model", config.CompletionModel,
			"embedding_model", config.EmbeddingModel)
		llmClient = llm.NewLiteLLMClient(
			config.LiteLLMURL,
			config.CompletionModel,
			config.EmbeddingModel,
			config.LiteLLMAPIKey,
			logger,
		)
	default:
		// Default to Ollama
		ollamaURL := config.OllamaURL
		if ollamaURL == "" {
			ollamaURL = "http://localhost:11434"
		}
		logger.Info("Using Ollama provider",
			"url", ollamaURL,
			"completion_model", config.CompletionModel,
			"embedding_model", config.EmbeddingModel)
		llmClient = llm.NewOllamaClient(
			ollamaURL,
			config.CompletionModel,
			config.EmbeddingModel,
			logger,
		)
	}

	// Wrap LLM client with debug logging
	debugLLMClient := NewDebugLLMClient(llmClient, logger, *verbose)

	// Create extractor
	extr := extractor.NewExtractor(debugLLMClient, repo, logger)

	// Query emails based on flags
	ctx := context.Background()
	var emails []*ent.Email

	if *emailID > 0 {
		email, err := client.Email.Get(ctx, *emailID)
		if err != nil {
			logger.Error("Failed to find email by ID", "id", *emailID, "error", err)
			os.Exit(1)
		}
		emails = []*ent.Email{email}
	} else if *messageID != "" {
		email, err := baseRepo.FindEmailByMessageID(ctx, *messageID)
		if err != nil {
			logger.Error("Failed to find email by message ID", "message_id", *messageID, "error", err)
			os.Exit(1)
		}
		emails = []*ent.Email{email}
	} else {
		// Get random sample
		// First get the total count
		count, err := client.Email.Query().Count(ctx)
		if err != nil {
			logger.Error("Failed to count emails", "error", err)
			os.Exit(1)
		}

		if count == 0 {
			logger.Info("No emails found in database")
			return
		}

		// Generate random offset
		var offset int
		if count > *limit {
			offset = rand.Intn(count - *limit)
		}

		emails, err = client.Email.Query().
			Offset(offset).
			Limit(*limit).
			All(ctx)
		if err != nil {
			logger.Error("Failed to query emails", "error", err)
			os.Exit(1)
		}
	}

	if len(emails) == 0 {
		logger.Info("No emails found")
		return
	}

	// Process each email
	for i, email := range emails {
		fmt.Println(strings.Repeat("=", 80))
		fmt.Printf("Email %d of %d\n", i+1, len(emails))
		fmt.Println(strings.Repeat("=", 80))
		fmt.Printf("ID: %d\n", email.ID)
		fmt.Printf("Message-ID: %s\n", email.MessageID)
		fmt.Printf("From: %s\n", email.From)
		fmt.Printf("To: %s\n", strings.Join(email.To, ", "))
		if len(email.Cc) > 0 {
			fmt.Printf("CC: %s\n", strings.Join(email.Cc, ", "))
		}
		fmt.Printf("Subject: %s\n", email.Subject)
		fmt.Printf("Date: %s\n", email.Date.Format(time.RFC3339))
		fmt.Printf("\nBody (first 200 chars):\n%s\n", truncate(email.Body, 200))
		fmt.Println(strings.Repeat("-", 80))

		// Extract entities
		fmt.Println("\nRunning extraction...")
		startTime := time.Now()

		summary, err := extr.ExtractFromEmail(ctx, email)
		duration := time.Since(startTime)

		if err != nil {
			logger.Error("Extraction failed", "error", err)
			fmt.Printf("\nERROR: %v\n", err)
		} else {
			fmt.Printf("\nExtraction Summary:\n")
			fmt.Printf("  Entities Created: %d\n", summary.EntitiesCreated)
			fmt.Printf("  Relationships Created: %d\n", summary.RelationshipsCreated)
			fmt.Printf("  Duration: %v\n", duration.Round(time.Millisecond))
		}

		// Show detailed results if verbose
		if *verbose {
			// The read-only repo will have captured what would have been created
			readOnlyRepo := repo
			fmt.Println("\nDetailed Results:")

			if len(readOnlyRepo.capturedEntities) > 0 {
				// Group entities by source
				var headerEntities, contentEntities []*ent.DiscoveredEntity
				for _, entity := range readOnlyRepo.capturedEntities {
					if source, ok := entity.Properties["source"].(string); ok && source == "header" {
						headerEntities = append(headerEntities, entity)
					} else {
						contentEntities = append(contentEntities, entity)
					}
				}

				fmt.Printf("\nEntities (%d total: %d from headers, %d from content):\n",
					len(readOnlyRepo.capturedEntities), len(headerEntities), len(contentEntities))

				if len(headerEntities) > 0 {
					fmt.Println("\n  FROM HEADERS:")
					for _, entity := range headerEntities {
						fmt.Printf("    - [%s] %s (confidence: %.2f)\n",
							entity.TypeCategory,
							entity.Name,
							entity.ConfidenceScore)
					}
				}

				if len(contentEntities) > 0 {
					fmt.Println("\n  FROM CONTENT:")
					for _, entity := range contentEntities {
						fmt.Printf("    - [%s] %s (confidence: %.2f)\n",
							entity.TypeCategory,
							entity.Name,
							entity.ConfidenceScore)
						if len(entity.Properties) > 1 { // More than just 'source'
							fmt.Printf("      Properties: %v\n", entity.Properties)
						}
					}
				}
			}

			if len(readOnlyRepo.capturedRelationships) > 0 {
				fmt.Printf("\nRelationships (%d):\n", len(readOnlyRepo.capturedRelationships))
				for _, rel := range readOnlyRepo.capturedRelationships {
					fmt.Printf("  - [%s] %d -[%s]-> [%s] %d (confidence: %.2f)\n",
						rel.FromType,
						rel.FromID,
						rel.Type,
						rel.ToType,
						rel.ToID,
						rel.ConfidenceScore)
					if len(rel.Properties) > 0 {
						fmt.Printf("    Properties: %v\n", rel.Properties)
					}
				}
			}

			// Reset for next email
			readOnlyRepo.capturedEntities = nil
			readOnlyRepo.capturedRelationships = nil
		}

		fmt.Println()
	}

	logger.Info("Debug session complete", "emails_processed", len(emails))
}

// truncate truncates a string to the specified length, adding "..." if truncated
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

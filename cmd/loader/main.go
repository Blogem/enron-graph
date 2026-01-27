package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Blogem/enron-graph/ent"
	"github.com/Blogem/enron-graph/internal/extractor"
	"github.com/Blogem/enron-graph/internal/graph"
	"github.com/Blogem/enron-graph/internal/loader"
	"github.com/Blogem/enron-graph/pkg/llm"
	"github.com/Blogem/enron-graph/pkg/utils"

	_ "github.com/lib/pq"
)

func main() {
	// Command line flags
	csvPath := flag.String("csv-path", "", "Path to Enron emails CSV file (required)")
	workers := flag.Int("workers", 50, "Number of concurrent workers (10-100)")
	extract := flag.Bool("extract", false, "Enable entity extraction (requires LLM)")
	dbURL := flag.String("db", "", "Database connection URL (optional, uses config if not provided)")

	flag.Parse()

	// Validate flags
	if *csvPath == "" {
		log.Fatal("--csv-path is required")
	}

	if *workers < 1 || *workers > 100 {
		log.Fatal("--workers must be between 1 and 100")
	}

	// Initialize logger
	logger := utils.NewLogger()
	logger.Info("Starting email loader",
		"csv_path", *csvPath,
		"workers", *workers,
		"extract", *extract)

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

	// Create repository
	repo := graph.NewRepository(client)

	// Parse CSV
	logger.Info("Parsing CSV file", "path", *csvPath)
	records, errors, err := loader.ParseCSV(*csvPath)
	if err != nil {
		logger.Error("Failed to open CSV file", "error", err)
		os.Exit(1)
	}

	// Create processor
	processor := loader.NewProcessor(repo, logger, *workers)

	// Process emails
	ctx := context.Background()
	startTime := time.Now()

	logger.Info("Processing emails...")
	if err := processor.ProcessBatch(ctx, records, errors); err != nil {
		logger.Error("Processing failed", "error", err)
		os.Exit(1)
	}

	// Get final stats
	stats := processor.GetStats()
	duration := time.Since(startTime)

	// Report summary
	logger.Info("Processing complete",
		"total_processed", stats.Processed,
		"failures", stats.Failures,
		"duplicates_skipped", stats.Skipped,
		"duration", duration.Round(time.Second),
		"rate", fmt.Sprintf("%.1f emails/sec", float64(stats.Processed)/duration.Seconds()))

	// Exit with error if failure rate too high
	total := stats.Processed + stats.Failures + stats.Skipped
	if total > 0 {
		failureRate := float64(stats.Failures) / float64(total)
		if failureRate > 0.02 {
			logger.Error("Failure rate exceeds threshold",
				"failure_rate", fmt.Sprintf("%.2f%%", failureRate*100),
				"threshold", "2%")
			os.Exit(1)
		}
	}

	logger.Info("Email loading successful")

	// Run entity extraction if enabled
	if *extract {
		logger.Info("Starting entity extraction...")

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

		// Query emails that were just loaded
		logger.Info("Querying emails for extraction...")

		emails, err := client.Email.Query().
			Limit(int(stats.Processed)).
			All(ctx)
		if err != nil {
			logger.Error("Failed to query emails", "error", err)
			os.Exit(1)
		}

		logger.Info("Retrieved emails from database", "count", len(emails))

		// Run batch extraction
		extractionStart := time.Now()
		batchExtractor := extractor.NewBatchExtractor(llmClient, repo, logger, *workers)

		if err := batchExtractor.ProcessBatch(ctx, emails); err != nil {
			logger.Error("Extraction failed", "error", err)
			os.Exit(1)
		}

		extractionStats := batchExtractor.GetStats()
		extractionDuration := time.Since(extractionStart)

		// Report extraction summary
		logger.Info("Extraction complete",
			"emails_processed", extractionStats.EmailsProcessed,
			"entities_created", extractionStats.EntitiesCreated,
			"relationships_created", extractionStats.RelationshipsCreated,
			"failures", extractionStats.Failures,
			"duration", extractionDuration.Round(time.Second))
	}
}

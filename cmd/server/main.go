package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Blogem/enron-graph/ent"
	"github.com/Blogem/enron-graph/internal/api"
	"github.com/Blogem/enron-graph/internal/graph"
	"github.com/Blogem/enron-graph/pkg/llm"
	"github.com/Blogem/enron-graph/pkg/utils"
	"github.com/go-chi/chi/v5"

	_ "github.com/lib/pq"
)

func main() {
	// Parse command-line flags
	var (
		port     = flag.Int("port", 8080, "HTTP server port")
		dbURL    = flag.String("db", "", "PostgreSQL connection URL (default: from config)")
		logLevel = flag.String("log-level", "info", "Log level (debug, info, warn, error)")
	)
	flag.Parse()

	// Setup structured logging
	level := slog.LevelInfo
	switch *logLevel {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))
	slog.SetDefault(logger)

	// Load configuration
	cfg, err := utils.LoadConfig()
	if err != nil {
		logger.Error("Failed to load configuration", slog.Any("error", err))
		os.Exit(1)
	}

	// Override database URL if provided via flag
	if *dbURL != "" {
		cfg.DatabaseURL = *dbURL
	}

	logger.Info("Starting Enron Graph API Server",
		slog.Int("port", *port),
		slog.String("db", cfg.DatabaseURL),
	)

	// Initialize database connection (ent client)
	entClient, err := ent.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		logger.Error("Failed to connect to database", slog.Any("error", err))
		os.Exit(1)
	}
	defer entClient.Close()

	// Also open a direct SQL connection for raw queries (like pgvector search)
	sqlDB, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		logger.Error("Failed to open SQL database connection", slog.Any("error", err))
		os.Exit(1)
	}
	defer sqlDB.Close()

	// Create repository with both ent client and SQL connection
	repo := graph.NewRepositoryWithDB(entClient, sqlDB)

	logger.Info("Connected to database")

	// Initialize LLM client based on provider (optional - for semantic search)
	var llmClient llm.Client
	switch cfg.LLMProvider {
	case "litellm":
		logger.Info("Using LiteLLM provider",
			slog.String("url", cfg.LiteLLMURL),
			slog.String("completion_model", cfg.CompletionModel),
			slog.String("embedding_model", cfg.EmbeddingModel))
		llmClient = llm.NewLiteLLMClient(
			cfg.LiteLLMURL,
			cfg.CompletionModel,
			cfg.EmbeddingModel,
			cfg.LiteLLMAPIKey,
			logger,
		)
	default:
		// Default to Ollama
		ollamaURL := cfg.OllamaURL
		if ollamaURL == "" {
			ollamaURL = "http://localhost:11434"
		}
		logger.Info("Using Ollama provider",
			slog.String("url", ollamaURL),
			slog.String("completion_model", cfg.CompletionModel),
			slog.String("embedding_model", cfg.EmbeddingModel))
		llmClient = llm.NewOllamaClient(
			ollamaURL,
			cfg.CompletionModel,
			cfg.EmbeddingModel,
			logger,
		)
	}

	// Create API handler
	handler := api.NewHandlerWithLLM(repo, llmClient)

	// Setup Chi router
	r := chi.NewRouter()

	// Apply middleware
	r.Use(api.RecoveryMiddleware(logger))
	r.Use(api.LoggingMiddleware(logger))
	r.Use(api.CORSMiddleware())

	// Define routes
	r.Route("/api/v1", func(r chi.Router) {
		// Entity endpoints
		r.Get("/entities/{id}", handler.GetEntity)
		r.Get("/entities", handler.SearchEntities)
		r.Get("/entities/{id}/relationships", handler.GetEntityRelationships)
		r.Get("/entities/{id}/neighbors", handler.GetEntityNeighbors)

		// Graph operations
		r.Post("/entities/path", handler.FindPath)
		r.Post("/entities/search", handler.SemanticSearch)
	})

	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	})

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", *port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Server listening", slog.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed", slog.Any("error", err))
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", slog.Any("error", err))
		os.Exit(1)
	}

	logger.Info("Server stopped gracefully")
}

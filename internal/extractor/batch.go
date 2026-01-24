package extractor

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Blogem/enron-graph/ent"
	"github.com/Blogem/enron-graph/internal/graph"
	"github.com/Blogem/enron-graph/pkg/llm"
)

// BatchExtractor handles concurrent entity extraction from multiple emails
type BatchExtractor struct {
	extractor *Extractor
	logger    *slog.Logger
	workers   int
	stats     *BatchStats
}

// BatchStats tracks extraction statistics
type BatchStats struct {
	EmailsProcessed      int64
	EntitiesCreated      int64
	RelationshipsCreated int64
	Failures             int64
	StartTime            time.Time
}

// NewBatchExtractor creates a new batch extractor
func NewBatchExtractor(llmClient llm.Client, repo graph.Repository, logger *slog.Logger, workers int) *BatchExtractor {
	return &BatchExtractor{
		extractor: NewExtractor(llmClient, repo, logger),
		logger:    logger,
		workers:   workers,
		stats: &BatchStats{
			StartTime: time.Now(),
		},
	}
}

// ProcessBatch processes multiple emails concurrently
func (b *BatchExtractor) ProcessBatch(ctx context.Context, emails []*ent.Email) error {
	var wg sync.WaitGroup
	workers := make(chan struct{}, b.workers)

	for _, email := range emails {
		select {
		case <-ctx.Done():
			b.logger.Info("Batch extraction cancelled", "reason", ctx.Err())
			wg.Wait()
			return ctx.Err()
		default:
		}

		wg.Add(1)
		workers <- struct{}{} // Acquire worker slot

		go func(e *ent.Email) {
			defer wg.Done()
			defer func() { <-workers }() // Release worker slot

			if err := b.processEmail(ctx, e); err != nil {
				atomic.AddInt64(&b.stats.Failures, 1)
				b.logger.Error("Failed to extract from email",
					"message_id", e.MessageID,
					"error", err)
			} else {
				atomic.AddInt64(&b.stats.EmailsProcessed, 1)
			}

			// Log progress every 50 emails
			if b.stats.EmailsProcessed%50 == 0 {
				b.logProgress()
			}
		}(email)
	}

	// Wait for all workers to finish
	wg.Wait()

	// Final progress report
	b.logProgress()

	return nil
}

// processEmail processes a single email
func (b *BatchExtractor) processEmail(ctx context.Context, email *ent.Email) error {
	summary, err := b.extractor.ExtractFromEmail(ctx, email)
	if err != nil {
		return err
	}

	atomic.AddInt64(&b.stats.EntitiesCreated, int64(summary.EntitiesCreated))
	atomic.AddInt64(&b.stats.RelationshipsCreated, int64(summary.RelationshipsCreated))

	return nil
}

// logProgress logs the current extraction progress
func (b *BatchExtractor) logProgress() {
	elapsed := time.Since(b.stats.StartTime)
	processed := atomic.LoadInt64(&b.stats.EmailsProcessed)
	entities := atomic.LoadInt64(&b.stats.EntitiesCreated)
	relationships := atomic.LoadInt64(&b.stats.RelationshipsCreated)
	failures := atomic.LoadInt64(&b.stats.Failures)

	rate := float64(processed) / elapsed.Seconds()

	b.logger.Info("Extraction progress",
		"emails_processed", processed,
		"entities_created", entities,
		"relationships_created", relationships,
		"failures", failures,
		"rate", rate,
		"elapsed", elapsed.Round(time.Second))
}

// GetStats returns the current extraction statistics
func (b *BatchExtractor) GetStats() BatchStats {
	return BatchStats{
		EmailsProcessed:      atomic.LoadInt64(&b.stats.EmailsProcessed),
		EntitiesCreated:      atomic.LoadInt64(&b.stats.EntitiesCreated),
		RelationshipsCreated: atomic.LoadInt64(&b.stats.RelationshipsCreated),
		Failures:             atomic.LoadInt64(&b.stats.Failures),
		StartTime:            b.stats.StartTime,
	}
}

// ProcessEmailsWithExtraction is a convenience function that processes emails
// and extracts entities in one operation
func ProcessEmailsWithExtraction(ctx context.Context, emails []*ent.Email, llmClient llm.Client, repo graph.Repository, logger *slog.Logger, workers int) error {
	batchExtractor := NewBatchExtractor(llmClient, repo, logger, workers)
	return batchExtractor.ProcessBatch(ctx, emails)
}

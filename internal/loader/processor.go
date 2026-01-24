package loader

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Blogem/enron-graph/internal/graph"
)

// ProcessorStats tracks processing statistics
type ProcessorStats struct {
	Processed int64
	Failures  int64
	Skipped   int64 // Duplicates
	StartTime time.Time
}

// Processor handles batch email processing
type Processor struct {
	repo    graph.Repository
	logger  *slog.Logger
	workers int
	stats   *ProcessorStats
}

// NewProcessor creates a new email processor
func NewProcessor(repo graph.Repository, logger *slog.Logger, workers int) *Processor {
	return &Processor{
		repo:    repo,
		logger:  logger,
		workers: workers,
		stats: &ProcessorStats{
			StartTime: time.Now(),
		},
	}
}

// ProcessBatch processes a stream of email records concurrently
func (p *Processor) ProcessBatch(ctx context.Context, records <-chan EmailRecord, errors <-chan error) error {
	var wg sync.WaitGroup

	// Worker pool
	workers := make(chan struct{}, p.workers)

	// Error accumulator
	var lastError error
	var errorMu sync.Mutex

	// Process errors from CSV parser
	go func() {
		for err := range errors {
			p.logger.Warn("CSV parsing error", "error", err)
		}
	}()

	// Process records
	for record := range records {
		select {
		case <-ctx.Done():
			p.logger.Info("Processing cancelled", "reason", ctx.Err())
			wg.Wait()
			return ctx.Err()
		default:
		}

		wg.Add(1)
		workers <- struct{}{} // Acquire worker slot

		go func(rec EmailRecord) {
			defer wg.Done()
			defer func() { <-workers }() // Release worker slot

			if err := p.processEmail(ctx, rec); err != nil {
				atomic.AddInt64(&p.stats.Failures, 1)
				errorMu.Lock()
				lastError = err
				errorMu.Unlock()
				p.logger.Error("Failed to process email",
					"file", rec.File,
					"error", err)
			} else {
				atomic.AddInt64(&p.stats.Processed, 1)
			}

			// Log progress every 100 emails
			if p.stats.Processed%100 == 0 {
				p.logProgress()
			}
		}(record)
	}

	// Wait for all workers to finish
	wg.Wait()

	// Final progress report
	p.logProgress()

	// Check failure rate
	failureRate := float64(p.stats.Failures) / float64(p.stats.Processed+p.stats.Failures+p.stats.Skipped)
	if failureRate > 0.02 { // 2% threshold
		return fmt.Errorf("failure rate %.2f%% exceeds 2%% threshold", failureRate*100)
	}

	return lastError
}

// processEmail processes a single email record
func (p *Processor) processEmail(ctx context.Context, record EmailRecord) error {
	// Parse email headers
	metadata, err := ParseEmailHeaders(record.Message)
	if err != nil {
		return fmt.Errorf("failed to parse email headers: %w", err)
	}

	// Check for duplicate by message-id
	if metadata.MessageID != "" {
		existing, err := p.repo.FindEmailByMessageID(ctx, metadata.MessageID)
		if err == nil && existing != nil {
			// Duplicate found, skip
			atomic.AddInt64(&p.stats.Skipped, 1)
			p.logger.Debug("Skipping duplicate email", "message_id", metadata.MessageID)
			return nil
		}
	}

	// Create email entity
	emailInput := &graph.EmailInput{
		MessageID: metadata.MessageID,
		From:      metadata.From,
		To:        metadata.To,
		CC:        metadata.CC,
		BCC:       metadata.BCC,
		Subject:   metadata.Subject,
		Date:      metadata.Date,
		Body:      metadata.Body,
		FilePath:  record.File,
	}

	_, err = p.repo.CreateEmail(ctx, emailInput)
	if err != nil {
		return fmt.Errorf("failed to create email entity: %w", err)
	}

	return nil
}

// logProgress logs the current processing progress
func (p *Processor) logProgress() {
	elapsed := time.Since(p.stats.StartTime)
	total := p.stats.Processed + p.stats.Failures + p.stats.Skipped
	rate := float64(total) / elapsed.Seconds()

	p.logger.Info("Processing progress",
		"processed", p.stats.Processed,
		"failures", p.stats.Failures,
		"skipped", p.stats.Skipped,
		"total", total,
		"rate", fmt.Sprintf("%.1f emails/sec", rate),
		"elapsed", elapsed.Round(time.Second))
}

// GetStats returns the current processing statistics
func (p *Processor) GetStats() ProcessorStats {
	return ProcessorStats{
		Processed: atomic.LoadInt64(&p.stats.Processed),
		Failures:  atomic.LoadInt64(&p.stats.Failures),
		Skipped:   atomic.LoadInt64(&p.stats.Skipped),
		StartTime: p.stats.StartTime,
	}
}

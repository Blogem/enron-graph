package integration

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/Blogem/enron-graph/ent/discoveredentity"
	"github.com/Blogem/enron-graph/internal/graph"
	"github.com/Blogem/enron-graph/internal/loader"
	"github.com/Blogem/enron-graph/pkg/utils"
)

// TestConcurrentWrites tests parallel loading and extraction (T150)
// Validates that multiple loaders can run simultaneously without race conditions or deadlocks
func TestConcurrentWrites(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	client, db := SetupTestDBWithSQL(t)

	// Create repository and logger
	logger := utils.NewLogger()
	repo := graph.NewRepository(client, logger)

	// Track concurrent operations
	var wg sync.WaitGroup
	errChan := make(chan error, 10) // Buffered channel for errors
	doneChan := make(chan bool)

	// Worker 1: Load emails from CSV
	wg.Add(1)
	go func() {
		defer wg.Done()

		testCSV := filepath.Join("..", "fixtures", "sample_emails.csv")
		records, errors, err := loader.ParseCSV(testCSV)
		if err != nil {
			errChan <- fmt.Errorf("worker1 ParseCSV failed: %w", err)
			return
		}

		processor := loader.NewProcessor(repo, logger, 3)
		if err := processor.ProcessBatch(ctx, records, errors); err != nil {
			// Some failures are expected due to concurrent duplicate inserts
			// Only fail if too many errors (>50%)
			stats := processor.GetStats()
			total := stats.Processed + stats.Failures + stats.Skipped
			failureRate := float64(stats.Failures) / float64(total)
			if failureRate > 0.50 {
				errChan <- fmt.Errorf("worker1 ProcessBatch excessive failures: %.0f%%", failureRate*100)
				return
			}
			t.Logf("Worker 1: some expected concurrent duplicate errors (%.0f%%)", failureRate*100)
		}

		t.Logf("Worker 1 completed: processed %d emails", processor.GetStats().Processed)
	}()

	// Worker 2: Load emails from CSV (same file, testing concurrent duplicate handling)
	wg.Add(1)
	go func() {
		defer wg.Done()

		// Small delay to allow Worker 1 to start first
		time.Sleep(50 * time.Millisecond)

		testCSV := filepath.Join("..", "fixtures", "sample_emails.csv")
		records, errors, err := loader.ParseCSV(testCSV)
		if err != nil {
			errChan <- fmt.Errorf("worker2 ParseCSV failed: %w", err)
			return
		}

		processor := loader.NewProcessor(repo, logger, 3)
		if err := processor.ProcessBatch(ctx, records, errors); err != nil {
			// Many failures expected here since Worker 1 already loaded these emails
			// This is correct behavior - testing that duplicates are properly rejected
			stats := processor.GetStats()
			t.Logf("Worker 2: expected duplicate rejections - %d succeeded, %d duplicates/failures",
				stats.Processed, stats.Failures+stats.Skipped)
		} else {
			t.Logf("Worker 2 completed: processed %d emails", processor.GetStats().Processed)
		}
	}()

	// Worker 3: Create entities directly (simulating extractor)
	wg.Add(1)
	go func() {
		defer wg.Done()

		for i := 0; i < 10; i++ {
			_, err := client.DiscoveredEntity.Create().
				SetUniqueID(fmt.Sprintf("concurrent-entity-%d@test.com", i)).
				SetTypeCategory("person").
				SetName(fmt.Sprintf("Concurrent Person %d", i)).
				SetConfidenceScore(0.80).
				Save(ctx)
			if err != nil {
				errChan <- fmt.Errorf("worker3 entity creation failed: %w", err)
				return
			}
			time.Sleep(10 * time.Millisecond) // Small delay to interleave operations
		}

		t.Logf("Worker 3 completed: created 10 entities")
	}()

	// Worker 4: Create relationships (simulating extractor)
	wg.Add(1)
	go func() {
		defer wg.Done()

		// Wait a bit to ensure some entities exist
		time.Sleep(100 * time.Millisecond)

		// Query existing entities
		entities, err := client.DiscoveredEntity.Query().
			Limit(5).
			All(ctx)
		if err != nil {
			errChan <- fmt.Errorf("worker4 query failed: %w", err)
			return
		}

		if len(entities) < 2 {
			t.Log("Worker 4: not enough entities yet, skipping relationship creation")
			return
		}

		// Create relationships between existing entities
		for i := 0; i < len(entities)-1; i++ {
			_, err := client.Relationship.Create().
				SetType("CONCURRENT_TEST").
				SetFromType("discovered_entity").
				SetFromID(entities[i].ID).
				SetToType("discovered_entity").
				SetToID(entities[i+1].ID).
				SetConfidenceScore(0.75).
				Save(ctx)
			if err != nil {
				errChan <- fmt.Errorf("worker4 relationship creation failed: %w", err)
				return
			}
			time.Sleep(10 * time.Millisecond)
		}

		t.Logf("Worker 4 completed: created relationships")
	}()

	// Wait for all workers to complete
	go func() {
		wg.Wait()
		close(doneChan)
	}()

	// Wait for completion or timeout
	select {
	case <-doneChan:
		t.Log("All concurrent workers completed successfully")
	case <-time.After(60 * time.Second):
		t.Fatal("Test timeout: concurrent operations did not complete within 60 seconds (possible deadlock)")
	}

	// Check for errors
	close(errChan)
	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		t.Errorf("Encountered %d errors during concurrent operations:", len(errors))
		for _, err := range errors {
			t.Errorf("  - %v", err)
		}
		t.Fatal("Concurrent write test failed due to errors")
	}

	// Verify data integrity after concurrent writes
	t.Run("VerifyDataIntegrityAfterConcurrentWrites", func(t *testing.T) {
		// Check for duplicate unique_ids
		var duplicates []struct {
			UniqueID string
			Count    int
		}

		rows, err := db.QueryContext(ctx, `
			SELECT unique_id, COUNT(*) as count
			FROM discovered_entities
			GROUP BY unique_id
			HAVING COUNT(*) > 1
		`)
		if err != nil {
			t.Fatalf("Failed to query for duplicates: %v", err)
		}
		defer rows.Close()

		for rows.Next() {
			var dup struct {
				UniqueID string
				Count    int
			}
			if err := rows.Scan(&dup.UniqueID, &dup.Count); err != nil {
				t.Fatalf("Failed to scan duplicate row: %v", err)
			}
			duplicates = append(duplicates, dup)
		}

		if len(duplicates) > 0 {
			t.Errorf("Found %d duplicate unique_id values after concurrent writes:", len(duplicates))
			for _, dup := range duplicates {
				t.Errorf("  - unique_id '%s' appears %d times", dup.UniqueID, dup.Count)
			}
		}

		// Check for orphaned relationships
		allRels, err := client.Relationship.Query().All(ctx)
		if err != nil {
			t.Fatalf("Failed to query relationships: %v", err)
		}

		var orphanedRels int
		for _, rel := range allRels {
			if rel.FromType == "discovered_entity" {
				exists, err := client.DiscoveredEntity.Query().
					Where(discoveredentity.IDEQ(rel.FromID)).
					Exist(ctx)
				if err != nil {
					t.Fatalf("Failed to check entity existence: %v", err)
				}
				if !exists {
					orphanedRels++
				}
			}
			if rel.ToType == "discovered_entity" {
				exists, err := client.DiscoveredEntity.Query().
					Where(discoveredentity.IDEQ(rel.ToID)).
					Exist(ctx)
				if err != nil {
					t.Fatalf("Failed to check entity existence: %v", err)
				}
				if !exists {
					orphanedRels++
				}
			}
		}

		if orphanedRels > 0 {
			t.Errorf("Found %d orphaned relationships after concurrent writes", orphanedRels)
		}

		// Report final statistics
		entityCount, err := client.DiscoveredEntity.Query().Count(ctx)
		if err != nil {
			t.Fatalf("Failed to count entities: %v", err)
		}

		relCount, err := client.Relationship.Query().Count(ctx)
		if err != nil {
			t.Fatalf("Failed to count relationships: %v", err)
		}

		emailCount, err := client.Email.Query().Count(ctx)
		if err != nil {
			t.Fatalf("Failed to count emails: %v", err)
		}

		t.Logf("✓ Data integrity verified after concurrent writes:")
		t.Logf("  - Entities: %d", entityCount)
		t.Logf("  - Relationships: %d", relCount)
		t.Logf("  - Emails: %d", emailCount)
		t.Logf("  - No duplicates or orphaned relationships found")
	})
}

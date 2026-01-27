package integration

import (
	"encoding/csv"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Blogem/enron-graph/internal/sampler"
)

// TestPerformance_10kEmails verifies that the sampler can extract 10,000 emails
// in less than 10 seconds (Success Criteria SC-001).
func TestPerformance_10kEmails(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	// Arrange: Create temporary CSV with 10,000 test emails
	tmpDir := t.TempDir()
	sourceCSV := filepath.Join(tmpDir, "large-source.csv")

	// Create CSV with 10k records
	file, err := os.Create(sourceCSV)
	if err != nil {
		t.Fatalf("Failed to create test CSV: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)

	// Write header
	if err := writer.Write([]string{"file", "message"}); err != nil {
		t.Fatalf("Failed to write header: %v", err)
	}

	// Write 10,000 test records
	for i := 0; i < 10000; i++ {
		record := []string{
			fmt.Sprintf("test-emails/email-%d", i),
			"Message-ID: <test@enron.com>\nFrom: sender@enron.com\nTo: recipient@enron.com\nSubject: Test Email\n\nThis is a test email body.",
		}
		if err := writer.Write(record); err != nil {
			t.Fatalf("Failed to write record %d: %v", i, err)
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		t.Fatalf("CSV writer error: %v", err)
	}
	file.Close()

	// Act: Measure extraction time
	startTime := time.Now()

	// Parse CSV
	records, errs, err := sampler.ParseCSV(sourceCSV)
	if err != nil {
		t.Fatalf("ParseCSV failed: %v", err)
	}

	// Consume errors in background
	go func() {
		for range errs {
		}
	}()

	// Collect all records
	var emails []sampler.EmailRecord
	for record := range records {
		emails = append(emails, record)
	}

	// Count available
	registry := sampler.NewTrackingRegistry()
	count := sampler.CountAvailable(emails, registry)

	// Generate indices for all 10k
	rng := rand.New(rand.NewSource(12345))
	indices := sampler.GenerateIndices(rng, count, count)

	// Extract emails
	extracted := sampler.ExtractEmails(emails, indices, registry)

	// Write output
	outputPath := filepath.Join(tmpDir, "output.csv")
	err = sampler.WriteCSV(outputPath, extracted)
	if err != nil {
		t.Fatalf("WriteCSV failed: %v", err)
	}

	elapsed := time.Since(startTime)

	// Assert: Should complete in less than 10 seconds (SC-001)
	if elapsed > 10*time.Second {
		t.Errorf("Performance requirement not met: extraction took %v, expected < 10s", elapsed)
	}

	// Assert: Should have extracted all 10k emails
	if len(extracted) != 10000 {
		t.Errorf("Expected 10000 extracted emails, got %d", len(extracted))
	}

	t.Logf("Performance test passed: extracted 10,000 emails in %v", elapsed)
}

// TestPerformance_LargeTrackingRegistry tests performance with a large tracking registry.
func TestPerformance_LargeTrackingRegistry(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	// Arrange: Create registry with 50,000 entries
	registry := sampler.NewTrackingRegistry()
	for i := 0; i < 50000; i++ {
		registry.Add(fmt.Sprintf("email-%d", i))
	}

	// Create 10,000 new emails
	emails := make([]sampler.EmailRecord, 10000)
	for i := 0; i < 10000; i++ {
		emails[i] = sampler.EmailRecord{
			File:    fmt.Sprintf("new-email-%d", i),
			Message: "Test message",
		}
	}

	// Act: Measure counting performance
	startTime := time.Now()
	count := sampler.CountAvailable(emails, registry)
	elapsed := time.Since(startTime)

	// Assert: Should complete quickly (< 1 second for 10k emails with 50k registry)
	if elapsed > 1*time.Second {
		t.Errorf("Counting with large registry took %v, expected < 1s", elapsed)
	}

	// Assert: All 10k should be available (different IDs)
	if count != 10000 {
		t.Errorf("Expected count=10000, got %d", count)
	}

	t.Logf("Large registry test passed: counted 10,000 emails against 50,000 registry in %v", elapsed)
}

// TestPerformance_MultipleExtractions tests performance across multiple extraction runs.
func TestPerformance_MultipleExtractions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	// Arrange: Create CSV with 1,000 test emails
	tmpDir := t.TempDir()
	sourceCSV := filepath.Join(tmpDir, "source.csv")

	file, err := os.Create(sourceCSV)
	if err != nil {
		t.Fatalf("Failed to create test CSV: %v", err)
	}

	writer := csv.NewWriter(file)
	writer.Write([]string{"file", "message"})

	for i := 0; i < 1000; i++ {
		writer.Write([]string{
			fmt.Sprintf("email-%d", i),
			fmt.Sprintf("Test message %d", i),
		})
	}
	writer.Flush()
	file.Close()

	// Act: Perform 10 extraction runs of 100 emails each
	registry := sampler.NewTrackingRegistry()
	startTime := time.Now()

	for run := 0; run < 10; run++ {
		// Parse
		records, errs, err := sampler.ParseCSV(sourceCSV)
		if err != nil {
			t.Fatalf("Run %d: ParseCSV failed: %v", run, err)
		}

		go func() {
			for range errs {
			}
		}()

		var emails []sampler.EmailRecord
		for record := range records {
			emails = append(emails, record)
		}

		// Filter to available emails only
		var availableEmails []sampler.EmailRecord
		for _, email := range emails {
			if !registry.Contains(email.File) {
				availableEmails = append(availableEmails, email)
			}
		}

		// Count available
		count := len(availableEmails)

		// Generate indices for 100 emails
		requestCount := 100
		if count < requestCount {
			requestCount = count
		}

		rng := rand.New(rand.NewSource(int64(run)))
		indices := sampler.GenerateIndices(rng, count, requestCount)

		// Extract from available emails
		var extracted []sampler.EmailRecord
		for _, idx := range indices {
			if idx >= 0 && idx < len(availableEmails) {
				extracted = append(extracted, availableEmails[idx])
			}
		}

		// Update registry
		for _, email := range extracted {
			registry.Add(email.File)
		}

		// Write output
		outputPath := filepath.Join(tmpDir, fmt.Sprintf("output-%d.csv", run))
		sampler.WriteCSV(outputPath, extracted)
	}

	elapsed := time.Since(startTime)

	// Assert: 10 runs should complete in reasonable time (< 5 seconds)
	if elapsed > 5*time.Second {
		t.Errorf("Multiple extractions took %v, expected < 5s", elapsed)
	}

	// Assert: Registry should have 1000 entries (all emails extracted)
	if registry.Count() != 1000 {
		t.Errorf("Expected registry count=1000, got %d", registry.Count())
	}

	t.Logf("Multiple extractions test passed: 10 runs of 100 emails each in %v", elapsed)
}

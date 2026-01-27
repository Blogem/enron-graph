package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Blogem/enron-graph/internal/sampler"
)

// T026: Integration tests for multi-run uniqueness verification (US2)

// TestSamplerDuplicatePrevention_MultipleRuns tests that running the sampler
// multiple times never extracts the same email twice across runs.
func TestSamplerDuplicatePrevention_MultipleRuns(t *testing.T) {
	// Arrange: Setup test environment with fixture
	fixturePath := filepath.Join("..", "fixtures", "sample-small.csv")
	outputDir := t.TempDir()

	// Parse source CSV once (reuse for all runs)
	records, errs, err := sampler.ParseCSV(fixturePath)
	if err != nil {
		t.Fatalf("Failed to parse CSV: %v", err)
	}

	var emails []sampler.EmailRecord
	done := make(chan bool)
	go func() {
		for range errs {
		}
		done <- true
	}()

	for record := range records {
		emails = append(emails, record)
	}
	<-done

	if len(emails) < 10 {
		t.Fatalf("Need at least 10 emails in fixture for multi-run test, got %d", len(emails))
	}

	// Track all extracted email IDs across runs
	allExtractedIDs := make(map[string]bool)

	// Run 1: Extract 3 emails
	t.Run("FirstRun", func(t *testing.T) {
		registry, _, err := sampler.LoadTracking(outputDir)
		if err != nil {
			t.Fatalf("Failed to load tracking: %v", err)
		}

		availableCount := sampler.CountAvailable(emails, registry)
		// Generate indices into the full emails array
		indices := sampler.GenerateIndices(nil, len(emails), 3)
		extracted := sampler.ExtractEmails(emails, indices, registry)

		if len(extracted) == 0 {
			t.Fatal("First run: expected at least 1 extracted email, got 0")
		}

		// Save extracted IDs
		extractedIDs := make([]string, len(extracted))
		for i, email := range extracted {
			extractedIDs[i] = email.File
			allExtractedIDs[email.File] = true
		}

		// Create tracking file
		err = sampler.CreateTrackingFile(outputDir, "run1", extractedIDs)
		if err != nil {
			t.Fatalf("Failed to create tracking file for run 1: %v", err)
		}

		t.Logf("First run extracted %d emails (available: %d)", len(extracted), availableCount)
	})

	firstRunCount := len(allExtractedIDs)

	// Run 2: Extract 3 more emails
	t.Run("SecondRun", func(t *testing.T) {
		registry, _, err := sampler.LoadTracking(outputDir)
		if err != nil {
			t.Fatalf("Failed to load tracking: %v", err)
		}

		// Registry should now have emails from run 1
		if registry.Count() != firstRunCount {
			t.Errorf("Second run: expected registry count=%d, got %d", firstRunCount, registry.Count())
		}

		availableCount := sampler.CountAvailable(emails, registry)
		indices := sampler.GenerateIndices(nil, len(emails), 3)
		extracted := sampler.ExtractEmails(emails, indices, registry)

		if len(extracted) == 0 {
			t.Fatal("Second run: expected at least 1 extracted email, got 0")
		}

		// Verify no duplicates with run 1
		extractedIDs := make([]string, len(extracted))
		for i, email := range extracted {
			extractedIDs[i] = email.File
			if allExtractedIDs[email.File] {
				t.Errorf("Second run: duplicate email detected: %s", email.File)
			}
			allExtractedIDs[email.File] = true
		}

		// Create tracking file
		err = sampler.CreateTrackingFile(outputDir, "run2", extractedIDs)
		if err != nil {
			t.Fatalf("Failed to create tracking file for run 2: %v", err)
		}

		t.Logf("Second run extracted %d emails (available: %d)", len(extracted), availableCount)
	})

	secondRunTotal := len(allExtractedIDs)

	// Run 3: Extract more emails
	t.Run("ThirdRun", func(t *testing.T) {
		registry, _, err := sampler.LoadTracking(outputDir)
		if err != nil {
			t.Fatalf("Failed to load tracking: %v", err)
		}

		// Registry should have emails from runs 1 and 2
		if registry.Count() != secondRunTotal {
			t.Errorf("Third run: expected registry count=%d, got %d", secondRunTotal, registry.Count())
		}

		availableCount := sampler.CountAvailable(emails, registry)
		if availableCount == 0 {
			t.Skip("Third run: no emails available (all extracted in previous runs)")
		}

		indices := sampler.GenerateIndices(nil, len(emails), 4)
		extracted := sampler.ExtractEmails(emails, indices, registry)

		// Verify no duplicates with runs 1 and 2
		for _, email := range extracted {
			if allExtractedIDs[email.File] {
				t.Errorf("Third run: duplicate email detected: %s", email.File)
			}
			allExtractedIDs[email.File] = true
		}

		t.Logf("Third run extracted %d emails (available: %d)", len(extracted), availableCount)
	})

	// Final verification: All extracted emails should be unique
	t.Logf("Total unique emails extracted across all runs: %d", len(allExtractedIDs))
}

// TestSamplerDuplicatePrevention_TrackingFileCreation tests that tracking files
// are created correctly after each run.
func TestSamplerDuplicatePrevention_TrackingFileCreation(t *testing.T) {
	// Arrange: Setup test environment
	outputDir := t.TempDir()

	// Create first tracking file
	run1IDs := []string{"email-1", "email-2", "email-3"}
	err := sampler.CreateTrackingFile(outputDir, "20260127-120000", run1IDs)
	if err != nil {
		t.Fatalf("Failed to create first tracking file: %v", err)
	}

	// Create second tracking file
	run2IDs := []string{"email-4", "email-5"}
	err = sampler.CreateTrackingFile(outputDir, "20260127-130000", run2IDs)
	if err != nil {
		t.Fatalf("Failed to create second tracking file: %v", err)
	}

	// Verify both tracking files exist
	file1 := filepath.Join(outputDir, "extracted-20260127-120000.txt")
	file2 := filepath.Join(outputDir, "extracted-20260127-130000.txt")

	if _, err := os.Stat(file1); os.IsNotExist(err) {
		t.Error("First tracking file should exist")
	}
	if _, err := os.Stat(file2); os.IsNotExist(err) {
		t.Error("Second tracking file should exist")
	}

	// Load tracking and verify all IDs are present
	registry, _, err := sampler.LoadTracking(outputDir)
	if err != nil {
		t.Fatalf("Failed to load tracking: %v", err)
	}

	expectedCount := len(run1IDs) + len(run2IDs)
	if registry.Count() != expectedCount {
		t.Errorf("Expected registry count=%d, got %d", expectedCount, registry.Count())
	}

	// Verify all IDs are in registry
	allIDs := append(run1IDs, run2IDs...)
	for _, id := range allIDs {
		if !registry.Contains(id) {
			t.Errorf("Registry should contain '%s'", id)
		}
	}
}

// TestSamplerDuplicatePrevention_EmptyAfterExtractingAll tests behavior when
// all emails have been extracted in previous runs.
func TestSamplerDuplicatePrevention_EmptyAfterExtractingAll(t *testing.T) {
	// Arrange: Create a small dataset
	emails := []sampler.EmailRecord{
		{File: "email-1", Message: "Message 1"},
		{File: "email-2", Message: "Message 2"},
		{File: "email-3", Message: "Message 3"},
	}

	// Create registry with all emails already extracted
	registry := sampler.NewTrackingRegistry()
	registry.Add("email-1")
	registry.Add("email-2")
	registry.Add("email-3")

	// Act: Try to extract more emails
	availableCount := sampler.CountAvailable(emails, registry)

	// Assert: Should have 0 available
	if availableCount != 0 {
		t.Errorf("Expected 0 available emails (all extracted), got %d", availableCount)
	}

	// Try to extract anyway
	indices := sampler.GenerateIndices(nil, availableCount, 5)
	extracted := sampler.ExtractEmails(emails, indices, registry)

	// Should extract nothing
	if len(extracted) != 0 {
		t.Errorf("Expected 0 extracted emails (all already extracted), got %d", len(extracted))
	}
}

// TestSamplerDuplicatePrevention_PartialAvailability tests extraction when
// some but not all requested emails are available.
func TestSamplerDuplicatePrevention_PartialAvailability(t *testing.T) {
	// Arrange: Create dataset with some already extracted
	emails := []sampler.EmailRecord{
		{File: "email-1", Message: "Message 1"},
		{File: "email-2", Message: "Message 2"},
		{File: "email-3", Message: "Message 3"},
		{File: "email-4", Message: "Message 4"},
		{File: "email-5", Message: "Message 5"},
	}

	// Registry with 3 emails already extracted
	registry := sampler.NewTrackingRegistry()
	registry.Add("email-1")
	registry.Add("email-3")
	registry.Add("email-5")

	// Act: Request 5 emails but only 2 are available
	availableCount := sampler.CountAvailable(emails, registry)

	// Assert: Should have 2 available
	if availableCount != 2 {
		t.Errorf("Expected 2 available emails, got %d", availableCount)
	}

	// Generate indices - pass full array length, not available count
	indices := sampler.GenerateIndices(nil, len(emails), 5)
	extracted := sampler.ExtractEmails(emails, indices, registry)

	// Should extract at most 2 (email-2 and email-4)
	if len(extracted) > 2 {
		t.Errorf("Expected at most 2 extracted emails (only available ones), got %d", len(extracted))
	}

	// Verify only available emails were extracted
	for _, email := range extracted {
		if email.File == "email-1" || email.File == "email-3" || email.File == "email-5" {
			t.Errorf("Should not extract already-extracted email: %s", email.File)
		}
	}
}

// TestSamplerDuplicatePrevention_TrackingFileFormat tests that tracking files
// use the correct format (one ID per line, plain text).
func TestSamplerDuplicatePrevention_TrackingFileFormat(t *testing.T) {
	// Arrange: Create tracking file
	outputDir := t.TempDir()
	emailIDs := []string{"email-1", "email-2", "email-3"}

	err := sampler.CreateTrackingFile(outputDir, "20260127-140000", emailIDs)
	if err != nil {
		t.Fatalf("Failed to create tracking file: %v", err)
	}

	// Act: Read tracking file
	trackingPath := filepath.Join(outputDir, "extracted-20260127-140000.txt")
	content, err := os.ReadFile(trackingPath)
	if err != nil {
		t.Fatalf("Failed to read tracking file: %v", err)
	}

	// Assert: Verify format (one ID per line)
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != len(emailIDs) {
		t.Errorf("Expected %d lines, got %d", len(emailIDs), len(lines))
	}

	for i, line := range lines {
		if line != emailIDs[i] {
			t.Errorf("Line %d: expected '%s', got '%s'", i, emailIDs[i], line)
		}
	}
}

// TestSamplerDuplicatePrevention_ConcurrentReads tests that tracking files
// can be safely read by multiple processes (read-only access).
func TestSamplerDuplicatePrevention_ConcurrentReads(t *testing.T) {
	// Arrange: Create tracking files
	outputDir := t.TempDir()
	err := sampler.CreateTrackingFile(outputDir, "20260127-140000", []string{"email-1", "email-2"})
	if err != nil {
		t.Fatalf("Failed to create tracking file: %v", err)
	}

	// Act: Load tracking concurrently (simulates multiple processes)
	results := make(chan int, 5)
	errors := make(chan error, 5)

	for i := 0; i < 5; i++ {
		go func() {
			registry, _, err := sampler.LoadTracking(outputDir)
			if err != nil {
				errors <- err
				return
			}
			results <- registry.Count()
		}()
	}

	// Assert: All concurrent reads should succeed
	for i := 0; i < 5; i++ {
		select {
		case count := <-results:
			if count != 2 {
				t.Errorf("Expected count=2 from concurrent read, got %d", count)
			}
		case err := <-errors:
			t.Errorf("Concurrent read failed: %v", err)
		}
	}
}

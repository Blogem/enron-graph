package integration

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Blogem/enron-graph/internal/sampler"
)

// T038: Integration tests for various count scenarios (US3)

// TestSamplerConfig_Count10 tests extracting exactly 10 emails
func TestSamplerConfig_Count10(t *testing.T) {
	// Arrange: Setup test environment
	fixturePath := filepath.Join("..", "fixtures", "sample-small.csv")
	outputDir := t.TempDir()
	requestedCount := 10

	// Act: Run full extraction workflow
	registry, extractedCount := runFullExtraction(t, fixturePath, outputDir, requestedCount)

	// Assert: Should extract exactly 10 emails
	if extractedCount != requestedCount {
		t.Errorf("Expected %d extracted emails, got %d", requestedCount, extractedCount)
	}

	// Assert: Tracking registry should contain 10 emails
	if registry.Count() != requestedCount {
		t.Errorf("Expected registry to contain %d emails, got %d", requestedCount, registry.Count())
	}
}

// TestSamplerConfig_Count100 tests extracting 100 emails
func TestSamplerConfig_Count100(t *testing.T) {
	// Note: This test requires a larger fixture or will test the "exceeds available" scenario
	// Using the small fixture, this will test the edge case handling
	fixturePath := filepath.Join("..", "fixtures", "sample-small.csv")
	outputDir := t.TempDir()
	requestedCount := 100

	// Load fixture to determine actual available count
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

	availableCount := len(emails)

	// Act: Run extraction (will be capped at available count)
	registry, extractedCount := runFullExtraction(t, fixturePath, outputDir, requestedCount)

	// Assert: Should extract min(requested, available)
	expectedCount := requestedCount
	if requestedCount > availableCount {
		expectedCount = availableCount
	}

	if extractedCount != expectedCount {
		t.Errorf("Expected %d extracted emails (capped at available), got %d", expectedCount, extractedCount)
	}

	// Assert: Registry should match extracted count
	if registry.Count() != extractedCount {
		t.Errorf("Expected registry to contain %d emails, got %d", extractedCount, registry.Count())
	}
}

// TestSamplerConfig_Count1000 tests extracting 1000 emails
func TestSamplerConfig_Count1000(t *testing.T) {
	// Note: This will be limited by fixture size
	fixturePath := filepath.Join("..", "fixtures", "sample-small.csv")
	outputDir := t.TempDir()
	requestedCount := 1000

	// Load fixture to determine actual available count
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

	availableCount := len(emails)

	// Act: Run extraction
	registry, extractedCount := runFullExtraction(t, fixturePath, outputDir, requestedCount)

	// Assert: Should extract all available (since 1000 > fixture size)
	if extractedCount != availableCount {
		t.Errorf("Expected %d extracted emails (all available), got %d", availableCount, extractedCount)
	}

	// Assert: Registry should contain all extracted emails
	if registry.Count() != extractedCount {
		t.Errorf("Expected registry to contain %d emails, got %d", extractedCount, registry.Count())
	}
}

// TestSamplerConfig_CountExceedsAvailable tests requesting more than available
func TestSamplerConfig_CountExceedsAvailable(t *testing.T) {
	// Arrange: Use small fixture
	fixturePath := filepath.Join("..", "fixtures", "sample-small.csv")
	outputDir := t.TempDir()

	// First, determine how many emails are in the fixture
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

	availableCount := len(emails)

	// Request more than available
	requestedCount := availableCount + 50

	// Act: Run extraction
	registry, extractedCount := runFullExtraction(t, fixturePath, outputDir, requestedCount)

	// Assert: Should extract only what's available
	if extractedCount != availableCount {
		t.Errorf("Expected %d extracted (all available), got %d", availableCount, extractedCount)
	}

	// Assert: Registry should contain all extracted
	if registry.Count() != extractedCount {
		t.Errorf("Expected registry to contain %d emails, got %d", extractedCount, registry.Count())
	}
}

// TestSamplerConfig_MultipleExtractions tests sequential extractions with different counts
func TestSamplerConfig_MultipleExtractions(t *testing.T) {
	// Arrange: Shared output directory for tracking
	fixturePath := filepath.Join("..", "fixtures", "sample-small.csv")
	outputDir := t.TempDir()

	// Determine total available
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

	totalAvailable := len(emails)

	// Act: Run multiple extractions
	counts := []int{3, 5, 2}
	totalExtracted := 0

	for i, count := range counts {
		t.Logf("Extraction %d: Requesting %d emails", i+1, count)

		// Sleep to ensure different timestamps (format is YYYYMMDD-HHMMSS)
		if i > 0 {
			time.Sleep(time.Second * 1)
		}

		registry, extractedCount := runFullExtraction(t, fixturePath, outputDir, count)

		// Update total
		totalExtracted += extractedCount

		// Assert: Registry should grow with each extraction
		if registry.Count() != totalExtracted {
			t.Errorf("After extraction %d: Expected registry=%d, got %d",
				i+1, totalExtracted, registry.Count())
		}

		// Assert: No more than total available
		if totalExtracted > totalAvailable {
			t.Errorf("Total extracted (%d) exceeds total available (%d)",
				totalExtracted, totalAvailable)
		}
	}

	// Final assert: Total should be sum of all extractions (or capped at available)
	expectedTotal := 3 + 5 + 2
	if expectedTotal > totalAvailable {
		expectedTotal = totalAvailable
	}

	if totalExtracted != expectedTotal {
		t.Errorf("Expected total extracted=%d, got %d", expectedTotal, totalExtracted)
	}
}

// TestSamplerConfig_ExactCount tests when requested equals available
func TestSamplerConfig_ExactCount(t *testing.T) {
	// Arrange: Determine exact count available
	fixturePath := filepath.Join("..", "fixtures", "sample-small.csv")
	outputDir := t.TempDir()

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

	exactCount := len(emails)

	// Act: Request exactly what's available
	registry, extractedCount := runFullExtraction(t, fixturePath, outputDir, exactCount)

	// Assert: Should extract all
	if extractedCount != exactCount {
		t.Errorf("Expected %d extracted (exact match), got %d", exactCount, extractedCount)
	}

	// Assert: Registry should contain all
	if registry.Count() != exactCount {
		t.Errorf("Expected registry=%d, got %d", exactCount, registry.Count())
	}

	// Act: Try to extract again (should get 0)
	registry2, extractedCount2 := runFullExtraction(t, fixturePath, outputDir, 10)

	// Assert: Should extract 0 (all already taken)
	if extractedCount2 != 0 {
		t.Errorf("Expected 0 extracted on second run (all taken), got %d", extractedCount2)
	}

	// Assert: Registry unchanged
	if registry2.Count() != exactCount {
		t.Errorf("Expected registry still=%d, got %d", exactCount, registry2.Count())
	}
}

// TestSamplerConfig_SingleEmail tests extracting just one email
func TestSamplerConfig_SingleEmail(t *testing.T) {
	// Arrange
	fixturePath := filepath.Join("..", "fixtures", "sample-small.csv")
	outputDir := t.TempDir()
	requestedCount := 1

	// Act: Extract single email
	registry, extractedCount := runFullExtraction(t, fixturePath, outputDir, requestedCount)

	// Assert: Should extract exactly 1
	if extractedCount != 1 {
		t.Errorf("Expected 1 extracted email, got %d", extractedCount)
	}

	// Assert: Registry should contain 1
	if registry.Count() != 1 {
		t.Errorf("Expected registry=1, got %d", registry.Count())
	}

	// Verify output file exists and is valid
	outputFiles := findSampledFiles(t, outputDir)
	if len(outputFiles) != 1 {
		t.Fatalf("Expected 1 output file, got %d", len(outputFiles))
	}

	// Read and verify CSV
	file, err := os.Open(outputFiles[0])
	if err != nil {
		t.Fatalf("Failed to open output file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("Failed to read CSV: %v", err)
	}

	// Should have header + 1 record
	if len(records) != 2 {
		t.Errorf("Expected 2 CSV rows (header + 1 record), got %d", len(records))
	}
}

// TestSamplerConfig_VariableCounts tests different count values in sequence
func TestSamplerConfig_VariableCounts(t *testing.T) {
	testCases := []struct {
		name  string
		count int
	}{
		{"Count_1", 1},
		{"Count_2", 2},
		{"Count_5", 5},
		{"Count_10", 10},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange: Fresh directory for each test
			fixturePath := filepath.Join("..", "fixtures", "sample-small.csv")
			outputDir := t.TempDir()

			// Determine available
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

			availableCount := len(emails)

			// Act: Run extraction
			registry, extractedCount := runFullExtraction(t, fixturePath, outputDir, tc.count)

			// Assert: Should extract min(requested, available)
			expectedCount := tc.count
			if tc.count > availableCount {
				expectedCount = availableCount
			}

			if extractedCount != expectedCount {
				t.Errorf("Expected %d extracted, got %d", expectedCount, extractedCount)
			}

			if registry.Count() != extractedCount {
				t.Errorf("Expected registry=%d, got %d", extractedCount, registry.Count())
			}
		})
	}
}

// Helper function to run full extraction workflow
func runFullExtraction(t *testing.T, fixturePath, outputDir string, requestedCount int) (*sampler.TrackingRegistry, int) {
	t.Helper()

	timestamp := time.Now()

	// Step 1: Load tracking registry
	registry, _, err := sampler.LoadTracking(outputDir)
	if err != nil {
		t.Fatalf("Failed to load tracking: %v", err)
	}

	// Step 2: Parse source CSV
	records, errs, err := sampler.ParseCSV(fixturePath)
	if err != nil {
		t.Fatalf("Failed to parse CSV: %v", err)
	}

	// Collect all records
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

	// Step 3: Count available
	availableCount := sampler.CountAvailable(emails, registry)

	// Handle edge case: requested exceeds available
	actualCount := requestedCount
	if requestedCount > availableCount {
		actualCount = availableCount
	}

	// Handle edge case: no emails available
	if availableCount == 0 {
		return registry, 0
	}

	// Step 4: Filter to available emails
	var availableEmails []sampler.EmailRecord
	for _, email := range emails {
		if !registry.Contains(email.File) {
			availableEmails = append(availableEmails, email)
		}
	}

	// Step 5: Generate random indices
	indices := sampler.GenerateIndices(nil, len(availableEmails), actualCount)

	// Step 6: Extract emails
	var extracted []sampler.EmailRecord
	for _, idx := range indices {
		if idx >= 0 && idx < len(availableEmails) {
			extracted = append(extracted, availableEmails[idx])
		}
	}

	// Step 7: Write output CSV
	outputFilename := "sampled-emails-" + timestamp.Format("20060102-150405") + ".csv"
	outputPath := filepath.Join(outputDir, outputFilename)

	err = sampler.WriteCSV(outputPath, extracted)
	if err != nil {
		t.Fatalf("Failed to write output CSV: %v", err)
	}

	// Step 8: Create tracking file
	extractedIDs := make([]string, len(extracted))
	for i, email := range extracted {
		extractedIDs[i] = email.File
	}

	err = sampler.CreateTrackingFile(outputDir, timestamp.Format("20060102-150405"), extractedIDs)
	if err != nil {
		t.Fatalf("Failed to create tracking file: %v", err)
	}

	// Reload registry to verify it was updated
	updatedRegistry, _, err := sampler.LoadTracking(outputDir)
	if err != nil {
		t.Fatalf("Failed to reload tracking: %v", err)
	}

	return updatedRegistry, len(extracted)
}

// Helper function to find sampled files in output directory
func findSampledFiles(t *testing.T, outputDir string) []string {
	t.Helper()

	entries, err := os.ReadDir(outputDir)
	if err != nil {
		t.Fatalf("Failed to read output directory: %v", err)
	}

	var sampledFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".csv" {
			sampledFiles = append(sampledFiles, filepath.Join(outputDir, entry.Name()))
		}
	}

	return sampledFiles
}

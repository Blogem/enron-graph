package integration

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Blogem/enron-graph/internal/sampler"
)

// TestSamplerBasicWorkflow tests the complete end-to-end extraction workflow:
// 1. Load tracking registry (empty initially)
// 2. Parse source CSV file
// 3. Count available emails
// 4. Generate random indices
// 5. Extract selected emails
// 6. Write output CSV
// 7. Create tracking file
func TestSamplerBasicWorkflow(t *testing.T) {
	// Arrange: Setup test environment
	fixturePath := filepath.Join("..", "fixtures", "sample-small.csv")
	outputDir := t.TempDir() // Use temp directory for test outputs
	requestedCount := 3

	// Create a new extraction session
	session := &sampler.ExtractionSession{
		RequestedCount: requestedCount,
		Timestamp:      time.Now(),
	}

	// Step 1: Load tracking registry (should be empty initially)
	registry, err := sampler.LoadTracking(outputDir)
	if err != nil {
		t.Fatalf("Failed to load tracking: %v", err)
	}

	if registry.Count() != 0 {
		t.Errorf("Expected empty registry, got %d tracked emails", registry.Count())
	}

	// Step 2: Parse source CSV file
	records, errs, err := sampler.ParseCSV(fixturePath)
	if err != nil {
		t.Fatalf("Failed to parse CSV: %v", err)
	}

	// Collect all records
	var emails []sampler.EmailRecord
	done := make(chan bool)
	go func() {
		for range errs {
			// Consume errors (shouldn't be any for valid fixture)
		}
		done <- true
	}()

	for record := range records {
		emails = append(emails, record)
	}
	<-done

	// Verify we got emails from fixture
	if len(emails) == 0 {
		t.Fatal("No emails parsed from fixture file")
	}

	// Step 3: Count available emails
	availableCount := sampler.CountAvailable(emails, registry)
	session.AvailableCount = availableCount

	if availableCount != len(emails) {
		t.Errorf("Expected available=%d (all emails), got %d", len(emails), availableCount)
	}

	// Step 4: Generate random indices
	indices := sampler.GenerateIndices(nil, availableCount, requestedCount)
	session.SelectedIndices = indices

	if len(indices) != requestedCount {
		t.Errorf("Expected %d indices, got %d", requestedCount, len(indices))
	}

	// Step 5: Extract selected emails
	extracted := sampler.ExtractEmails(emails, indices, registry)
	session.ExtractedCount = len(extracted)

	if len(extracted) != requestedCount {
		t.Errorf("Expected %d extracted emails, got %d", requestedCount, len(extracted))
	}

	// Step 6: Write output CSV
	outputPath := filepath.Join(outputDir, "sampled-emails-test.csv")
	session.OutputPath = outputPath

	err = sampler.WriteCSV(outputPath, extracted)
	if err != nil {
		t.Fatalf("Failed to write output CSV: %v", err)
	}

	// Verify output file was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Errorf("Output CSV file was not created at %s", outputPath)
	}

	// Step 7: Create tracking file
	timestamp := "test"
	trackingFilename := "extracted-test.txt"
	trackingPath := filepath.Join(outputDir, trackingFilename)
	session.TrackingPath = trackingPath

	// Extract file IDs for tracking
	extractedIDs := make([]string, len(extracted))
	for i, email := range extracted {
		extractedIDs[i] = email.File
	}

	err = sampler.CreateTrackingFile(outputDir, timestamp, extractedIDs)
	if err != nil {
		t.Fatalf("Failed to create tracking file: %v", err)
	}

	// Verify tracking file was created
	if _, err := os.Stat(trackingPath); os.IsNotExist(err) {
		t.Errorf("Tracking file was not created at %s", trackingPath)
	}

	// Validate: Read back the output CSV and verify content
	outputFile, err := os.Open(outputPath)
	if err != nil {
		t.Fatalf("Failed to open output CSV for validation: %v", err)
	}
	defer outputFile.Close()

	reader := csv.NewReader(outputFile)
	outputRecords, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("Failed to read output CSV: %v", err)
	}

	// Check header + data rows
	expectedRows := requestedCount + 1 // +1 for header
	if len(outputRecords) != expectedRows {
		t.Errorf("Expected %d rows in output CSV (header+data), got %d",
			expectedRows, len(outputRecords))
	}

	// Verify header
	if len(outputRecords) > 0 {
		header := outputRecords[0]
		if len(header) != 2 || header[0] != "file" || header[1] != "message" {
			t.Errorf("Invalid output CSV header: %v", header)
		}
	}

	// Verify data rows have 2 columns
	for i := 1; i < len(outputRecords); i++ {
		if len(outputRecords[i]) != 2 {
			t.Errorf("Row %d has %d columns, expected 2", i, len(outputRecords[i]))
		}
	}
}

// TestSamplerWorkflow_SmallCount tests extraction with minimal count
func TestSamplerWorkflow_SmallCount(t *testing.T) {
	// Arrange: Setup for extracting just 1 email
	fixturePath := filepath.Join("..", "fixtures", "sample-small.csv")
	outputDir := t.TempDir()
	requestedCount := 1

	// Parse source
	records, errs, err := sampler.ParseCSV(fixturePath)
	if err != nil {
		t.Fatalf("Failed to parse CSV: %v", err)
	}

	var emails []sampler.EmailRecord
	go func() {
		for range errs {
		}
	}()
	for record := range records {
		emails = append(emails, record)
	}

	// Extract and write
	registry := sampler.NewTrackingRegistry()
	availableCount := sampler.CountAvailable(emails, registry)
	indices := sampler.GenerateIndices(nil, availableCount, requestedCount)
	extracted := sampler.ExtractEmails(emails, indices, registry)

	outputPath := filepath.Join(outputDir, "sampled-1.csv")
	err = sampler.WriteCSV(outputPath, extracted)
	if err != nil {
		t.Fatalf("Failed to write output: %v", err)
	}

	// Assert: Exactly 1 email extracted
	if len(extracted) != 1 {
		t.Errorf("Expected 1 extracted email, got %d", len(extracted))
	}

	// Verify file exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("Output file not created")
	}
}

// TestSamplerWorkflow_OutputFormat tests that output CSV preserves
// the exact format of the source CSV (headers, quoting, multi-line content)
func TestSamplerWorkflow_OutputFormat(t *testing.T) {
	// Arrange: Extract from fixture
	fixturePath := filepath.Join("..", "fixtures", "sample-small.csv")
	outputDir := t.TempDir()
	requestedCount := 5

	// Parse, extract, write
	records, errs, err := sampler.ParseCSV(fixturePath)
	if err != nil {
		t.Fatalf("Failed to parse CSV: %v", err)
	}

	var emails []sampler.EmailRecord
	go func() {
		for range errs {
		}
	}()
	for record := range records {
		emails = append(emails, record)
	}

	registry := sampler.NewTrackingRegistry()
	availableCount := sampler.CountAvailable(emails, registry)
	indices := sampler.GenerateIndices(nil, availableCount, requestedCount)
	extracted := sampler.ExtractEmails(emails, indices, registry)

	outputPath := filepath.Join(outputDir, "format-test.csv")
	err = sampler.WriteCSV(outputPath, extracted)
	if err != nil {
		t.Fatalf("Failed to write output: %v", err)
	}

	// Read back and validate format
	outputFile, err := os.Open(outputPath)
	if err != nil {
		t.Fatalf("Failed to open output file: %v", err)
	}
	defer outputFile.Close()

	reader := csv.NewReader(outputFile)
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true

	outputRecords, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("Failed to parse output CSV: %v", err)
	}

	// Assert: Header must be present and correct
	if len(outputRecords) < 1 {
		t.Fatal("Output CSV has no header")
	}

	header := outputRecords[0]
	if len(header) != 2 || header[0] != "file" || header[1] != "message" {
		t.Errorf("Invalid header format: %v", header)
	}

	// Assert: All data rows have exactly 2 fields
	for i := 1; i < len(outputRecords); i++ {
		if len(outputRecords[i]) != 2 {
			t.Errorf("Row %d has %d fields, expected 2", i, len(outputRecords[i]))
		}
	}

	// Assert: Output should be valid input to loader.ParseCSV
	// (This is SC-003: output format compatibility)
	reRecords, reErrs, err := sampler.ParseCSV(outputPath)
	if err != nil {
		t.Fatalf("Output CSV is not compatible with ParseCSV: %v", err)
	}

	var reEmails []sampler.EmailRecord
	go func() {
		for range reErrs {
		}
	}()
	for record := range reRecords {
		reEmails = append(reEmails, record)
	}

	if len(reEmails) != len(extracted) {
		t.Errorf("Re-parsing output CSV produced %d emails, expected %d",
			len(reEmails), len(extracted))
	}
}

// TestSamplerWorkflow_NoSourceFile tests error handling when source
// CSV file does not exist
func TestSamplerWorkflow_NoSourceFile(t *testing.T) {
	// Arrange: Non-existent source file
	nonExistentPath := "/tmp/does-not-exist-12345.csv"

	// Act: Attempt to parse
	_, _, err := sampler.ParseCSV(nonExistentPath)

	// Assert: Should return error
	if err == nil {
		t.Error("Expected error for non-existent source file, got nil")
	}
}

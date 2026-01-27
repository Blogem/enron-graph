package sampler

import (
	"path/filepath"
	"testing"
)

// TestParseCSV_ValidFile tests that ParseCSV correctly wraps loader.ParseCSV
// and returns valid EmailRecord channel from a well-formed CSV file.
func TestParseCSV_ValidFile(t *testing.T) {
	// Arrange: Use sample-small.csv fixture with 10 test emails
	fixturePath := filepath.Join("..", "..", "tests", "fixtures", "sample-small.csv")

	// Act: Parse the CSV file
	records, errs, err := ParseCSV(fixturePath)
	if err != nil {
		t.Fatalf("ParseCSV() returned error: %v", err)
	}

	// Collect all records
	var emails []EmailRecord
	var parseErrors []error

	done := make(chan bool)
	go func() {
		for err := range errs {
			parseErrors = append(parseErrors, err)
		}
		done <- true
	}()

	for record := range records {
		emails = append(emails, record)
	}
	<-done

	// Assert: Should have 10 emails from fixture
	if len(emails) != 10 {
		t.Errorf("Expected 10 emails, got %d", len(emails))
	}

	// Assert: No parsing errors
	if len(parseErrors) > 0 {
		t.Errorf("Expected no parsing errors, got %d: %v", len(parseErrors), parseErrors)
	}

	// Assert: First email should have expected file ID
	if len(emails) > 0 && emails[0].File != "test-email-1" {
		t.Errorf("Expected first email File='test-email-1', got '%s'", emails[0].File)
	}

	// Assert: Messages should not be empty
	if len(emails) > 0 && emails[0].Message == "" {
		t.Error("Expected first email Message to be non-empty")
	}
}

// TestParseCSV_FileNotFound tests that ParseCSV returns an error
// when the specified CSV file does not exist.
func TestParseCSV_FileNotFound(t *testing.T) {
	// Arrange: Non-existent file path
	nonExistentPath := "/tmp/nonexistent-file-12345.csv"

	// Act: Attempt to parse non-existent file
	_, _, err := ParseCSV(nonExistentPath)

	// Assert: Should return an error
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

// TestParseCSV_EmptyFile tests that ParseCSV handles empty CSV files
// gracefully (only header, no data rows).
func TestParseCSV_EmptyFile(t *testing.T) {
	// This test would require creating a temporary empty CSV file
	// or having a fixture for empty files. Skipping for now.
	t.Skip("Empty file test requires temporary file creation")
}

// TestParseCSV_RecordStructure tests that EmailRecord fields
// are correctly populated from CSV columns.
func TestParseCSV_RecordStructure(t *testing.T) {
	// Arrange: Use sample-small.csv fixture
	fixturePath := filepath.Join("..", "..", "tests", "fixtures", "sample-small.csv")

	// Act: Parse and get first record
	records, errs, err := ParseCSV(fixturePath)
	if err != nil {
		t.Fatalf("ParseCSV() returned error: %v", err)
	}

	// Consume errors channel
	go func() {
		for range errs {
		}
	}()

	// Get first record
	var firstRecord EmailRecord
	for record := range records {
		firstRecord = record
		break
	}

	// Drain remaining records
	for range records {
	}

	// Assert: File field should be populated
	if firstRecord.File == "" {
		t.Error("Expected File field to be non-empty")
	}

	// Assert: Message field should be populated
	if firstRecord.Message == "" {
		t.Error("Expected Message field to be non-empty")
	}

	// Assert: Message should contain email-like content
	if len(firstRecord.Message) < 10 {
		t.Errorf("Expected Message to have substantial content, got length %d", len(firstRecord.Message))
	}
}

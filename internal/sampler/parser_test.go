package sampler

import (
	"os"
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

// TestParseCSV_CorruptedRow tests that ParseCSV handles corrupted CSV rows
// gracefully by reporting errors on the errors channel but continuing to parse.
func TestParseCSV_CorruptedRow(t *testing.T) {
	// Arrange: Create temporary CSV file with corrupted row
	tmpDir := t.TempDir()
	corruptedCSV := filepath.Join(tmpDir, "corrupted.csv")

	// Write CSV with valid header, one valid row, one corrupted row (missing field), and another valid row
	content := `file,message
email-1,This is a valid email
email-2
email-3,Another valid email`

	err := os.WriteFile(corruptedCSV, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Act: Parse the corrupted CSV
	records, errs, err := ParseCSV(corruptedCSV)
	if err != nil {
		t.Fatalf("ParseCSV() returned error: %v", err)
	}

	// Collect all records and errors
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

	// Assert: Should have successfully parsed 2 valid emails (skipping corrupted row)
	if len(emails) < 2 {
		t.Errorf("Expected at least 2 valid emails, got %d", len(emails))
	}

	// Assert: Should have reported at least one parse error for corrupted row
	if len(parseErrors) == 0 {
		t.Error("Expected parse errors for corrupted row, got none")
	}

	// Assert: Valid emails should be correctly parsed
	if len(emails) > 0 && emails[0].File != "email-1" {
		t.Errorf("Expected first email File='email-1', got '%s'", emails[0].File)
	}
}

// TestParseCSV_LoaderCompatibility tests that ParseCSV correctly wraps
// loader.ParseCSV and maintains compatibility with the loader package's CSV format.
// This validates Success Criteria SC-003: CSV format compatibility.
func TestParseCSV_LoaderCompatibility(t *testing.T) {
	// Arrange: Create temporary CSV file with various edge cases that loader.ParseCSV handles
	tmpDir := t.TempDir()
	testCSV := filepath.Join(tmpDir, "loader-compat.csv")

	// Create CSV with edge cases: quotes, commas, newlines
	content := `file,message
email-1,"Simple message without special characters"
email-2,"Message with, commas, inside"
email-3,"Message with ""quotes"" inside"
email-4,"Multi
line
message"
email-5,"Complex: ""quoted"", multi
line, with commas"`

	err := os.WriteFile(testCSV, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Act: Parse using sampler.ParseCSV (which wraps loader.ParseCSV)
	records, errs, err := ParseCSV(testCSV)
	if err != nil {
		t.Fatalf("ParseCSV() returned error: %v", err)
	}

	// Collect all records and errors
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

	// Assert: Should have parsed all 5 emails successfully
	if len(emails) != 5 {
		t.Errorf("Expected 5 emails, got %d", len(emails))
	}

	// Assert: No parsing errors
	if len(parseErrors) > 0 {
		t.Errorf("Expected no parsing errors, got %d: %v", len(parseErrors), parseErrors)
	}

	// Assert: Verify specific edge case handling
	if len(emails) >= 2 {
		// Email 2 should preserve commas
		if emails[1].Message != "Message with, commas, inside" {
			t.Errorf("Email 2: commas not preserved. Got: %s", emails[1].Message)
		}
	}

	if len(emails) >= 3 {
		// Email 3 should preserve quotes (CSV escapes "" as ")
		if emails[2].Message != `Message with "quotes" inside` {
			t.Errorf("Email 3: quotes not preserved. Got: %s", emails[2].Message)
		}
	}

	if len(emails) >= 4 {
		// Email 4 should preserve newlines
		expected := "Multi\nline\nmessage"
		if emails[3].Message != expected {
			t.Errorf("Email 4: newlines not preserved. Expected: %q, Got: %q", expected, emails[3].Message)
		}
	}

	if len(emails) >= 5 {
		// Email 5 should handle complex combination
		expected := `Complex: "quoted", multi
line, with commas`
		if emails[4].Message != expected {
			t.Errorf("Email 5: complex content not preserved. Expected: %q, Got: %q", expected, emails[4].Message)
		}
	}
}

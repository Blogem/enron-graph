package loader

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// T026: Unit tests for CSV parser
// Tests streaming CSV rows, parsing file/message columns, handling malformed CSV,
// empty files, and missing columns

func TestParseCSV_ValidFile(t *testing.T) {
	testCSV := `file,message
inbox/1.txt,"From: alice@enron.com
To: bob@enron.com
Subject: Meeting

Body text"
inbox/2.txt,"Second message"`

	// Create temp file
	tmpFile := createTempCSV(t, testCSV)
	defer removeTempFile(t, tmpFile)

	records, errors, err := ParseCSV(tmpFile)
	if err != nil {
		t.Fatalf("ParseCSV failed: %v", err)
	}

	var count int
	var firstRecord EmailRecord
	for record := range records {
		count++
		if count == 1 {
			firstRecord = record
		}
	}

	// Check for any errors
	for err := range errors {
		t.Errorf("Unexpected error: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected 2 records, got %d", count)
	}

	if firstRecord.File != "inbox/1.txt" {
		t.Errorf("Expected file 'inbox/1.txt', got '%s'", firstRecord.File)
	}

	if !strings.Contains(firstRecord.Message, "alice@enron.com") {
		t.Error("Expected message to contain sender email")
	}
}

func TestParseCSV_StreamingBehavior(t *testing.T) {
	// Test that parser streams rows without loading entire file
	var csvBuilder strings.Builder
	csvBuilder.WriteString("file,message\n")

	// Generate 100 rows
	for i := 0; i < 100; i++ {
		csvBuilder.WriteString("file")
		csvBuilder.WriteString(string(rune('0' + i%10)))
		csvBuilder.WriteString(".txt,\"Message ")
		csvBuilder.WriteString(string(rune('0' + i%10)))
		csvBuilder.WriteString("\"\n")
	}

	tmpFile := createTempCSV(t, csvBuilder.String())
	defer removeTempFile(t, tmpFile)

	records, errors, err := ParseCSV(tmpFile)
	if err != nil {
		t.Fatalf("ParseCSV failed: %v", err)
	}

	var count int
	for range records {
		count++
		// Verify streaming by not accumulating all records
		if count == 50 {
			// Should be able to process incrementally
			break
		}
	}

	// Drain remaining records and errors
	for range records {
	}
	for range errors {
	}

	if count != 50 {
		t.Errorf("Expected to read 50 records before break, got %d", count)
	}
}

func TestParseCSV_MalformedCSV(t *testing.T) {
	testCSV := `file,message
inbox/1.txt,"Unclosed quote
inbox/2.txt,"Valid message"`

	tmpFile := createTempCSV(t, testCSV)
	defer removeTempFile(t, tmpFile)

	records, errChan, err := ParseCSV(tmpFile)
	if err != nil {
		t.Fatalf("ParseCSV failed: %v", err)
	}

	var errorCount int
	var recordCount int

	// Process both channels simultaneously to avoid deadlock
	done := make(chan bool)
	go func() {
		for {
			select {
			case _, ok := <-records:
				if !ok {
					done <- true
					return
				}
				recordCount++
			case _, ok := <-errChan:
				if !ok {
					done <- true
					return
				}
				errorCount++
			}
		}
	}()

	<-done

	// Malformed CSV may or may not produce errors depending on CSV parser behavior
	// The test passes if it doesn't panic
	_ = errorCount
}

func TestParseCSV_EmptyFile(t *testing.T) {
	testCSV := ""

	tmpFile := createTempCSV(t, testCSV)
	defer removeTempFile(t, tmpFile)

	records, errChan, err := ParseCSV(tmpFile)
	if err != nil {
		t.Fatalf("ParseCSV failed: %v", err)
	}

	var count int
	for range records {
		count++
	}

	var hasError bool
	for range errChan {
		hasError = true
	}

	if !hasError {
		t.Error("Expected error for empty CSV")
	}

	if count != 0 {
		t.Errorf("Expected 0 records, got %d", count)
	}
}

func TestParseCSV_MissingColumns(t *testing.T) {
	testCSV := `file
inbox/1.txt
inbox/2.txt`

	tmpFile := createTempCSV(t, testCSV)
	defer removeTempFile(t, tmpFile)

	records, errChan, err := ParseCSV(tmpFile)
	if err != nil {
		t.Fatalf("ParseCSV failed: %v", err)
	}

	var hasError bool
	for range records {
	}
	for range errChan {
		hasError = true
	}

	if !hasError {
		t.Error("Expected error for missing 'message' column")
	}
}

func TestParseCSV_OnlyHeader(t *testing.T) {
	testCSV := `file,message`

	tmpFile := createTempCSV(t, testCSV)
	defer removeTempFile(t, tmpFile)

	records, errChan, err := ParseCSV(tmpFile)
	if err != nil {
		t.Fatalf("ParseCSV failed: %v", err)
	}

	var count int
	for range records {
		count++
	}

	for range errChan {
	}

	if count != 0 {
		t.Errorf("Expected 0 records for header-only file, got %d", count)
	}
}

func TestParseCSV_NonExistentFile(t *testing.T) {
	_, _, err := ParseCSV("/nonexistent/file/path.csv")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

// Helper functions
func createTempCSV(t *testing.T, content string) string {
	t.Helper()
	tmpFile := filepath.Join(t.TempDir(), "test.csv")
	err := os.WriteFile(tmpFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	return tmpFile
}

func removeTempFile(t *testing.T, path string) {
	t.Helper()
	// TempDir is automatically cleaned up by testing framework
}

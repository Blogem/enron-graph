package sampler

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestWriteCSV_EmptyRecords tests writing zero records
func TestWriteCSV_EmptyRecords(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "output.csv")

	var records []EmailRecord

	err := WriteCSV(outputPath, records)
	if err != nil {
		t.Fatalf("WriteCSV failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatal("Output file was not created")
	}

	// Read and verify content
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	// Should only have header
	expected := "file,message\n"
	if string(content) != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, string(content))
	}
}

// TestWriteCSV_SingleRecord tests writing one email record
func TestWriteCSV_SingleRecord(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "output.csv")

	records := []EmailRecord{
		{
			File:    "test-email-1",
			Message: "Subject: Test\n\nThis is a test email.",
		},
	}

	err := WriteCSV(outputPath, records)
	if err != nil {
		t.Fatalf("WriteCSV failed: %v", err)
	}

	// Parse output file using encoding/csv to verify format
	file, err := os.Open(outputPath)
	if err != nil {
		t.Fatalf("Failed to open output file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	rows, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("Failed to parse output CSV: %v", err)
	}

	// Verify header + 1 data row
	if len(rows) != 2 {
		t.Fatalf("Expected 2 rows (header + data), got %d", len(rows))
	}

	// Verify header
	if rows[0][0] != "file" || rows[0][1] != "message" {
		t.Errorf("Invalid header: got %v", rows[0])
	}

	// Verify data row
	if rows[1][0] != "test-email-1" {
		t.Errorf("Expected file='test-email-1', got '%s'", rows[1][0])
	}
	if rows[1][1] != "Subject: Test\n\nThis is a test email." {
		t.Errorf("Message content mismatch, got: '%s'", rows[1][1])
	}
}

// TestWriteCSV_MultipleRecords tests writing multiple email records
func TestWriteCSV_MultipleRecords(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "output.csv")

	records := []EmailRecord{
		{File: "email-1", Message: "First email"},
		{File: "email-2", Message: "Second email"},
		{File: "email-3", Message: "Third email"},
	}

	err := WriteCSV(outputPath, records)
	if err != nil {
		t.Fatalf("WriteCSV failed: %v", err)
	}

	// Parse output file
	file, err := os.Open(outputPath)
	if err != nil {
		t.Fatalf("Failed to open output file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	rows, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("Failed to parse output CSV: %v", err)
	}

	// Verify row count (header + 3 data rows)
	if len(rows) != 4 {
		t.Fatalf("Expected 4 rows, got %d", len(rows))
	}

	// Verify each record
	for i, record := range records {
		rowIdx := i + 1 // Skip header row
		if rows[rowIdx][0] != record.File {
			t.Errorf("Row %d: expected file='%s', got '%s'", rowIdx, record.File, rows[rowIdx][0])
		}
		if rows[rowIdx][1] != record.Message {
			t.Errorf("Row %d: expected message='%s', got '%s'", rowIdx, record.Message, rows[rowIdx][1])
		}
	}
}

// TestWriteCSV_MultilineContent tests preserving multi-line email content
func TestWriteCSV_MultilineContent(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "output.csv")

	multilineMessage := `Message-ID: <12345@enron.com>
From: test@enron.com
To: recipient@enron.com
Subject: Multi-line test

This is a multi-line email.
It has several lines.
Each should be preserved.`

	records := []EmailRecord{
		{
			File:    "test-multiline",
			Message: multilineMessage,
		},
	}

	err := WriteCSV(outputPath, records)
	if err != nil {
		t.Fatalf("WriteCSV failed: %v", err)
	}

	// Parse output file
	file, err := os.Open(outputPath)
	if err != nil {
		t.Fatalf("Failed to open output file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.LazyQuotes = true // Same as loader
	rows, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("Failed to parse output CSV: %v", err)
	}

	if len(rows) != 2 {
		t.Fatalf("Expected 2 rows, got %d", len(rows))
	}

	// Verify multi-line content is preserved exactly
	if rows[1][1] != multilineMessage {
		t.Errorf("Multi-line content not preserved.\nExpected:\n%s\n\nGot:\n%s", multilineMessage, rows[1][1])
	}
}

// TestWriteCSV_SpecialCharacters tests handling special characters and quotes
func TestWriteCSV_SpecialCharacters(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "output.csv")

	records := []EmailRecord{
		{
			File:    "test-quotes",
			Message: `He said, "Hello, world!" and left.`,
		},
		{
			File:    "test-commas",
			Message: "Contains, commas, in, content",
		},
		{
			File:    "test-newlines",
			Message: "Line 1\nLine 2\nLine 3",
		},
	}

	err := WriteCSV(outputPath, records)
	if err != nil {
		t.Fatalf("WriteCSV failed: %v", err)
	}

	// Parse output file
	file, err := os.Open(outputPath)
	if err != nil {
		t.Fatalf("Failed to open output file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.LazyQuotes = true
	rows, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("Failed to parse output CSV: %v", err)
	}

	// Verify each special case
	for i, record := range records {
		rowIdx := i + 1
		if rows[rowIdx][1] != record.Message {
			t.Errorf("Row %d: special characters not preserved.\nExpected: %s\nGot: %s",
				rowIdx, record.Message, rows[rowIdx][1])
		}
	}
}

// TestWriteCSV_FormatCompatibility tests that output can be read by loader.ParseCSV
func TestWriteCSV_FormatCompatibility(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "output.csv")

	originalRecords := []EmailRecord{
		{File: "email-1", Message: "Simple message"},
		{File: "email-2", Message: "Multi\nline\nmessage"},
		{File: "email-3", Message: `Message with "quotes" and commas, here`},
	}

	// Write using WriteCSV
	err := WriteCSV(outputPath, originalRecords)
	if err != nil {
		t.Fatalf("WriteCSV failed: %v", err)
	}

	// Read back using standard CSV parser (simulating loader behavior)
	file, err := os.Open(outputPath)
	if err != nil {
		t.Fatalf("Failed to open output file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.LazyQuotes = true
	reader.FieldsPerRecord = 2

	// Read header
	header, err := reader.Read()
	if err != nil {
		t.Fatalf("Failed to read header: %v", err)
	}
	if header[0] != "file" || header[1] != "message" {
		t.Errorf("Invalid header: got %v", header)
	}

	// Read and verify all records
	readRecords := []EmailRecord{}
	for {
		row, err := reader.Read()
		if err != nil {
			break // EOF expected
		}
		readRecords = append(readRecords, EmailRecord{
			File:    row[0],
			Message: row[1],
		})
	}

	// Verify count
	if len(readRecords) != len(originalRecords) {
		t.Fatalf("Expected %d records, got %d", len(originalRecords), len(readRecords))
	}

	// Verify each record matches
	for i, original := range originalRecords {
		if readRecords[i].File != original.File {
			t.Errorf("Record %d: file mismatch. Expected '%s', got '%s'",
				i, original.File, readRecords[i].File)
		}
		if readRecords[i].Message != original.Message {
			t.Errorf("Record %d: message mismatch. Expected '%s', got '%s'",
				i, original.Message, readRecords[i].Message)
		}
	}
}

// TestWriteCSV_InvalidPath tests error handling for invalid output path
func TestWriteCSV_InvalidPath(t *testing.T) {
	// Try to write to a non-existent directory without creating it
	invalidPath := "/nonexistent/directory/output.csv"

	records := []EmailRecord{
		{File: "test", Message: "test"},
	}

	err := WriteCSV(invalidPath, records)
	if err == nil {
		t.Fatal("Expected error for invalid path, got nil")
	}

	if !strings.Contains(err.Error(), "failed to create output file") {
		t.Errorf("Expected 'failed to create output file' error, got: %v", err)
	}
}

// TestWriteCSV_EmptyFields tests handling empty file or message fields
func TestWriteCSV_EmptyFields(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "output.csv")

	records := []EmailRecord{
		{File: "", Message: "Message without file ID"},
		{File: "file-only", Message: ""},
		{File: "", Message: ""},
	}

	err := WriteCSV(outputPath, records)
	if err != nil {
		t.Fatalf("WriteCSV failed: %v", err)
	}

	// Read back
	file, err := os.Open(outputPath)
	if err != nil {
		t.Fatalf("Failed to open output file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	rows, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("Failed to parse output CSV: %v", err)
	}

	// Verify empty fields are preserved
	for i, record := range records {
		rowIdx := i + 1
		if rows[rowIdx][0] != record.File {
			t.Errorf("Row %d: file field mismatch", rowIdx)
		}
		if rows[rowIdx][1] != record.Message {
			t.Errorf("Row %d: message field mismatch", rowIdx)
		}
	}
}

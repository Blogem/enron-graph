package loader

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
)

// EmailRecord represents a raw email record from CSV
type EmailRecord struct {
	File    string // Email file identifier
	Message string // Full email message including headers and body
}

// ParseCSV streams CSV rows and returns a channel of EmailRecord
func ParseCSV(filePath string) (<-chan EmailRecord, <-chan error, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open CSV file: %w", err)
	}

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = 2 // Expect exactly 2 columns: file, message
	reader.LazyQuotes = true   // Be lenient with quotes
	reader.TrimLeadingSpace = true

	records := make(chan EmailRecord, 100) // Buffer for smoother streaming
	errors := make(chan error, 10)

	go func() {
		defer close(records)
		defer close(errors)
		defer file.Close()

		// Read header row
		header, err := reader.Read()
		if err != nil {
			errors <- fmt.Errorf("failed to read CSV header: %w", err)
			return
		}

		// Validate header
		if len(header) != 2 || header[0] != "file" || header[1] != "message" {
			errors <- fmt.Errorf("invalid CSV header: expected [file, message], got %v", header)
			return
		}

		// Stream rows
		lineNum := 1 // Header is line 0
		for {
			row, err := reader.Read()
			if err == io.EOF {
				break
			}
			lineNum++

			if err != nil {
				// Log error but continue processing
				errors <- fmt.Errorf("error reading line %d: %w", lineNum, err)
				continue
			}

			if len(row) != 2 {
				errors <- fmt.Errorf("invalid row at line %d: expected 2 columns, got %d", lineNum, len(row))
				continue
			}

			records <- EmailRecord{
				File:    row[0],
				Message: row[1],
			}
		}
	}()

	return records, errors, nil
}

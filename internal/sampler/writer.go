package sampler

import (
	"encoding/csv"
	"fmt"
	"os"
)

// WriteCSV writes email records to a CSV file with proper formatting.
// The output format is compatible with loader.ParseCSV() for verification.
func WriteCSV(outputPath string, records []EmailRecord) error {
	// Create output file
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	// Create CSV writer
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header row
	if err := writer.Write([]string{"file", "message"}); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Write data rows
	for _, record := range records {
		row := []string{record.File, record.Message}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write record: %w", err)
		}
	}

	// Flush ensures all buffered data is written
	writer.Flush()
	if err := writer.Error(); err != nil {
		return fmt.Errorf("failed to flush CSV writer: %w", err)
	}

	return nil
}

package sampler

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// LoadTracking reads all extracted-*.txt tracking files from the specified directory
// and aggregates the email identifiers into a TrackingRegistry.
// Corrupted files are skipped with a warning logged.
// Returns the registry and the number of successfully loaded tracking files.
func LoadTracking(dir string) (*TrackingRegistry, int, error) {
	registry := NewTrackingRegistry()

	// Find all tracking files matching pattern extracted-*.txt
	pattern := filepath.Join(dir, "extracted-*.txt")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find tracking files: %w", err)
	}

	// No tracking files is valid (first run)
	if len(matches) == 0 {
		return registry, 0, nil
	}

	// Load each tracking file
	loadedCount := 0
	for _, filePath := range matches {
		if err := loadTrackingFile(filePath, registry); err != nil {
			// Log warning but continue with other files
			log.Printf("WARNING: Skipping corrupted tracking file %s: %v", filePath, err)
			continue
		}
		loadedCount++
	}

	return registry, loadedCount, nil
}

// loadTrackingFile reads a single tracking file and adds IDs to the registry
func loadTrackingFile(filePath string, registry *TrackingRegistry) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines
		if line == "" {
			continue
		}

		// Add to registry (duplicates across files are tolerated)
		registry.Add(line)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file at line %d: %w", lineNum, err)
	}

	return nil
}

// CreateTrackingFile writes the extracted email IDs to a new timestamped tracking file
func CreateTrackingFile(dir string, timestamp string, emailIDs []string) error {
	filename := fmt.Sprintf("extracted-%s.txt", timestamp)
	filePath := filepath.Join(dir, filename)

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create tracking file: %w", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, id := range emailIDs {
		if _, err := fmt.Fprintln(writer, id); err != nil {
			return fmt.Errorf("failed to write tracking entry: %w", err)
		}
	}

	if err := writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush tracking file: %w", err)
	}

	return nil
}

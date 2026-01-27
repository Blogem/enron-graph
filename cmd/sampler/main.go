package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/Blogem/enron-graph/internal/sampler"
)

const (
	// Default source CSV file path (relative to repo root)
	defaultSourcePath = "assets/enron-emails/emails.csv"
	// Output directory for sampled files and tracking files
	outputDir = "assets/enron-emails"
)

func main() {
	// Command line flags
	count := flag.Int("count", 0, "Number of emails to extract (required)")
	help := flag.Bool("help", false, "Show usage information")

	flag.Parse()

	// Show help
	if *help {
		printUsage()
		os.Exit(0)
	}

	// Validate flags
	if *count <= 0 {
		log.Fatal("--count must be a positive integer")
	}

	fmt.Printf("Random Email Sampler - extracting %d emails\n", *count)

	// Run extraction workflow
	if err := runExtraction(*count); err != nil {
		log.Fatalf("Extraction failed: %v", err)
	}
}

// runExtraction performs the complete email extraction workflow
func runExtraction(requestedCount int) error {
	timestamp := time.Now()

	// Step 1: Load tracking registry
	fmt.Println("Loading tracking registry...")
	registry, fileCount, err := sampler.LoadTracking(outputDir)
	if err != nil {
		return fmt.Errorf("failed to load tracking: %w", err)
	}

	if registry.Count() > 0 {
		fmt.Printf("Found %d previously extracted emails (from %d tracking files)\n", registry.Count(), fileCount)
	}

	// Step 2: Parse source CSV file (T021: error handling for missing file)
	fmt.Printf("Parsing source CSV: %s\n", defaultSourcePath)
	recordsChan, errsChan, err := sampler.ParseCSV(defaultSourcePath)
	if err != nil {
		return fmt.Errorf("failed to open source CSV file '%s': %w\nPlease ensure the file exists and is accessible", defaultSourcePath, err)
	}

	// Collect all records (T019: progress logging)
	var emails []sampler.EmailRecord
	done := make(chan bool)

	// Error collector goroutine
	var parseErrors []error
	go func() {
		for err := range errsChan {
			parseErrors = append(parseErrors, err)
			log.Printf("WARNING: CSV parsing error: %v", err)
		}
		done <- true
	}()

	// Record collector with progress tracking
	recordCount := 0
	for record := range recordsChan {
		emails = append(emails, record)
		recordCount++

		// Progress logging every 100 records (T019)
		if recordCount%100 == 0 {
			fmt.Printf("Processing %d emails...\n", recordCount)
		}
	}
	<-done

	fmt.Printf("Loaded %d total emails from source\n", len(emails))

	// Report parsing errors if any occurred
	if len(parseErrors) > 0 {
		fmt.Printf("WARNING: Encountered %d parsing errors (continuing with valid records)\n", len(parseErrors))
	}

	// Step 3: Count available emails
	availableCount := sampler.CountAvailable(emails, registry)
	fmt.Printf("Available for extraction: %d emails (excluding previously extracted)\n", availableCount)

	// Handle edge case: requested count exceeds available
	actualCount := requestedCount
	if requestedCount > availableCount {
		fmt.Printf("WARNING: Only %d emails available, extracting all remaining\n", availableCount)
		actualCount = availableCount
	}

	// Handle edge case: no emails available (T054)
	// Continue with extraction of 0 emails and create tracking file
	if availableCount == 0 {
		fmt.Println("WARNING: No unextracted emails available")
		actualCount = 0
	}

	// Step 4: Filter to only available emails
	var availableEmails []sampler.EmailRecord
	for _, email := range emails {
		if !registry.Contains(email.File) {
			availableEmails = append(availableEmails, email)
		}
	}

	// Step 5: Generate random indices from available emails
	fmt.Printf("Selecting %d random emails...\n", actualCount)
	indices := sampler.GenerateIndices(nil, len(availableEmails), actualCount)

	// Step 6: Extract emails at selected indices from available emails
	var extracted []sampler.EmailRecord
	for _, idx := range indices {
		if idx >= 0 && idx < len(availableEmails) {
			extracted = append(extracted, availableEmails[idx])
		}
	}
	fmt.Printf("Extracted %d emails\n", len(extracted))

	// Step 7: Generate timestamped output filename (T020)
	outputFilename := fmt.Sprintf("sampled-emails-%s.csv", timestamp.Format("20060102-150405"))
	outputPath := filepath.Join(outputDir, outputFilename)

	// Write output CSV
	fmt.Printf("Writing output to: %s\n", outputPath)
	if err := sampler.WriteCSV(outputPath, extracted); err != nil {
		return fmt.Errorf("failed to write output CSV: %w", err)
	}

	// Step 8: Create tracking file (T020: same timestamp format)
	trackingFilename := fmt.Sprintf("extracted-%s.txt", timestamp.Format("20060102-150405"))

	// Extract file IDs for tracking
	extractedIDs := make([]string, len(extracted))
	for i, email := range extracted {
		extractedIDs[i] = email.File
	}

	if err := sampler.CreateTrackingFile(outputDir, timestamp.Format("20060102-150405"), extractedIDs); err != nil {
		return fmt.Errorf("failed to create tracking file: %w", err)
	}

	// Final summary (T019)
	fmt.Printf("\n✓ Successfully extracted %d emails to %s\n", len(extracted), outputPath)
	fmt.Printf("✓ Tracking file created: %s\n", trackingFilename)

	return nil
}

func printUsage() {
	fmt.Println("Random Email Sampler")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  go run cmd/sampler/main.go --count <number>")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  --count  Number of emails to extract (required, must be positive)")
	fmt.Println("  --help   Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  go run cmd/sampler/main.go --count 10")
	fmt.Println("  go run cmd/sampler/main.go --count 1000")
}

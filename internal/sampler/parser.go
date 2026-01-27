package sampler

import (
	"github.com/Blogem/enron-graph/internal/loader"
)

// ParseCSV wraps loader.ParseCSV to parse the source CSV file
// and convert loader.EmailRecord to sampler.EmailRecord.
// Returns channels for streaming records and errors, plus initial error if file can't be opened.
func ParseCSV(filePath string) (<-chan EmailRecord, <-chan error, error) {
	// Call loader.ParseCSV to get loader.EmailRecord stream
	loaderRecords, loaderErrors, err := loader.ParseCSV(filePath)
	if err != nil {
		return nil, nil, err
	}

	// Create output channels for sampler.EmailRecord
	records := make(chan EmailRecord, 100)
	errors := make(chan error, 10)

	// Convert loader.EmailRecord to sampler.EmailRecord
	go func() {
		defer close(records)
		for loaderRecord := range loaderRecords {
			samplerRecord := EmailRecord{
				File:    loaderRecord.File,
				Message: loaderRecord.Message,
			}
			records <- samplerRecord
		}
	}()

	// Forward errors from loader
	go func() {
		defer close(errors)
		for loaderError := range loaderErrors {
			errors <- loaderError
		}
	}()

	return records, errors, nil
}

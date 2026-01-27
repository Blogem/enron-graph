package sampler

import "time"

// EmailRecord represents a single email entry from the source CSV file.
// Maps to the "file" and "message" columns in the Enron emails dataset.
type EmailRecord struct {
	File    string // Unique identifier from "file" column
	Message string // Full email content from "message" column
}

// ExtractionSession represents a single execution of the sampler utility
// with runtime state tracking the extraction process.
type ExtractionSession struct {
	RequestedCount  int       // Number of emails requested via --count flag
	AvailableCount  int       // Count of emails not in TrackingRegistry
	ExtractedCount  int       // Actual number of emails extracted
	SelectedIndices []int     // Random indices to extract (sorted for sequential access)
	OutputPath      string    // Full path to generated CSV file
	TrackingPath    string    // Full path to tracking file for this session
	Timestamp       time.Time // When extraction started (used for filenames)
}

// TrackingRegistry holds the set of previously extracted email identifiers.
// Loaded from all extracted-*.txt files at session start.
type TrackingRegistry struct {
	ExtractedIDs map[string]bool // Set of email file identifiers
}

// NewTrackingRegistry creates an empty tracking registry.
func NewTrackingRegistry() *TrackingRegistry {
	return &TrackingRegistry{
		ExtractedIDs: make(map[string]bool),
	}
}

// Contains checks if the given email identifier has been extracted before.
func (tr *TrackingRegistry) Contains(fileID string) bool {
	return tr.ExtractedIDs[fileID]
}

// Add marks an email identifier as extracted.
func (tr *TrackingRegistry) Add(fileID string) {
	tr.ExtractedIDs[fileID] = true
}

// Count returns the total number of tracked email identifiers.
func (tr *TrackingRegistry) Count() int {
	return len(tr.ExtractedIDs)
}

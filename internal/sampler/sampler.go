package sampler

import (
	"math/rand"
	"sort"
	"time"
)

// CountAvailable counts how many emails are available for extraction
// by excluding emails that exist in the tracking registry.
func CountAvailable(emails []EmailRecord, registry *TrackingRegistry) int {
	availableCount := 0
	for _, email := range emails {
		if !registry.Contains(email.File) {
			availableCount++
		}
	}
	return availableCount
}

// GenerateIndices generates random indices for selecting emails.
// Returns a sorted slice of unique indices within [0, totalAvailable).
// If requestedCount exceeds totalAvailable, returns all available indices.
// If rng is nil, creates a new random generator with current time as seed.
func GenerateIndices(rng *rand.Rand, totalAvailable int, requestedCount int) []int {
	// Cap requested count at available
	if requestedCount > totalAvailable {
		requestedCount = totalAvailable
	}

	// Create random generator if not provided
	if rng == nil {
		rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	}

	// Generate random indices using Perm for guaranteed uniqueness
	perm := rng.Perm(totalAvailable)
	indices := perm[:requestedCount]

	// Sort indices for sequential file access optimization
	sort.Ints(indices)

	return indices
}

// ExtractEmails extracts emails at the specified indices,
// filtering out any that exist in the tracking registry.
// Returns a slice of extracted EmailRecords.
func ExtractEmails(emails []EmailRecord, indices []int, registry *TrackingRegistry) []EmailRecord {
	extracted := make([]EmailRecord, 0, len(indices))

	for _, idx := range indices {
		// Bounds check
		if idx < 0 || idx >= len(emails) {
			continue
		}

		email := emails[idx]

		// Skip if already in tracking registry
		if registry.Contains(email.File) {
			continue
		}

		extracted = append(extracted, email)
	}

	return extracted
}

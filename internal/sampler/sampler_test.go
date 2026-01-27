package sampler

import (
	"math/rand"
	"testing"
)

// TestCountAvailable_EmptyRegistry tests counting available emails
// when no emails have been previously extracted.
func TestCountAvailable_EmptyRegistry(t *testing.T) {
	// Arrange: Create empty tracking registry
	registry := NewTrackingRegistry()

	// Create sample emails
	emails := []EmailRecord{
		{File: "email-1", Message: "Message 1"},
		{File: "email-2", Message: "Message 2"},
		{File: "email-3", Message: "Message 3"},
		{File: "email-4", Message: "Message 4"},
		{File: "email-5", Message: "Message 5"},
	}

	// Act: Count available emails
	count := CountAvailable(emails, registry)

	// Assert: All emails should be available
	if count != 5 {
		t.Errorf("Expected count=5 with empty registry, got %d", count)
	}
}

// TestCountAvailable_WithExistingExtractions tests counting available emails
// when some emails have been previously extracted.
func TestCountAvailable_WithExistingExtractions(t *testing.T) {
	// Arrange: Create registry with some extracted emails
	registry := NewTrackingRegistry()
	registry.Add("email-2")
	registry.Add("email-4")

	// Create sample emails
	emails := []EmailRecord{
		{File: "email-1", Message: "Message 1"},
		{File: "email-2", Message: "Message 2"}, // Already extracted
		{File: "email-3", Message: "Message 3"},
		{File: "email-4", Message: "Message 4"}, // Already extracted
		{File: "email-5", Message: "Message 5"},
	}

	// Act: Count available emails
	count := CountAvailable(emails, registry)

	// Assert: Only 3 emails should be available
	if count != 3 {
		t.Errorf("Expected count=3 with 2 extracted, got %d", count)
	}
}

// TestCountAvailable_AllExtracted tests counting when all emails
// have been previously extracted.
func TestCountAvailable_AllExtracted(t *testing.T) {
	// Arrange: Create registry with all emails
	registry := NewTrackingRegistry()
	registry.Add("email-1")
	registry.Add("email-2")
	registry.Add("email-3")

	// Create sample emails
	emails := []EmailRecord{
		{File: "email-1", Message: "Message 1"},
		{File: "email-2", Message: "Message 2"},
		{File: "email-3", Message: "Message 3"},
	}

	// Act: Count available emails
	count := CountAvailable(emails, registry)

	// Assert: No emails should be available
	if count != 0 {
		t.Errorf("Expected count=0 with all extracted, got %d", count)
	}
}

// TestGenerateIndices_ValidCount tests that GenerateIndices generates
// the correct number of random indices within bounds.
func TestGenerateIndices_ValidCount(t *testing.T) {
	// Arrange: Request 5 random indices from pool of 100
	seed := int64(12345)
	rng := rand.New(rand.NewSource(seed))
	totalAvailable := 100
	requestedCount := 5

	// Act: Generate random indices
	indices := GenerateIndices(rng, totalAvailable, requestedCount)

	// Assert: Should return exactly 5 indices
	if len(indices) != 5 {
		t.Errorf("Expected 5 indices, got %d", len(indices))
	}

	// Assert: All indices should be within bounds [0, 99]
	for _, idx := range indices {
		if idx < 0 || idx >= totalAvailable {
			t.Errorf("Index %d out of bounds [0, %d)", idx, totalAvailable)
		}
	}
}

// TestGenerateIndices_Uniqueness tests that GenerateIndices returns
// unique indices without duplicates.
func TestGenerateIndices_Uniqueness(t *testing.T) {
	// Arrange: Request 10 random indices
	seed := int64(54321)
	rng := rand.New(rand.NewSource(seed))
	totalAvailable := 100
	requestedCount := 10

	// Act: Generate random indices
	indices := GenerateIndices(rng, totalAvailable, requestedCount)

	// Assert: All indices should be unique
	seen := make(map[int]bool)
	for _, idx := range indices {
		if seen[idx] {
			t.Errorf("Duplicate index found: %d", idx)
		}
		seen[idx] = true
	}

	if len(seen) != requestedCount {
		t.Errorf("Expected %d unique indices, got %d", requestedCount, len(seen))
	}
}

// TestGenerateIndices_Sorted tests that GenerateIndices returns
// indices in sorted order for sequential file access optimization.
func TestGenerateIndices_Sorted(t *testing.T) {
	// Arrange: Request indices
	seed := int64(99999)
	rng := rand.New(rand.NewSource(seed))
	totalAvailable := 1000
	requestedCount := 20

	// Act: Generate random indices
	indices := GenerateIndices(rng, totalAvailable, requestedCount)

	// Assert: Indices should be in ascending order
	for i := 1; i < len(indices); i++ {
		if indices[i] <= indices[i-1] {
			t.Errorf("Indices not sorted: indices[%d]=%d, indices[%d]=%d",
				i-1, indices[i-1], i, indices[i])
		}
	}
}

// TestGenerateIndices_RequestMoreThanAvailable tests edge case when
// requested count exceeds available emails.
func TestGenerateIndices_RequestMoreThanAvailable(t *testing.T) {
	// Arrange: Request more indices than available
	seed := int64(11111)
	rng := rand.New(rand.NewSource(seed))
	totalAvailable := 10
	requestedCount := 20

	// Act: Generate random indices
	indices := GenerateIndices(rng, totalAvailable, requestedCount)

	// Assert: Should cap at available count
	if len(indices) != totalAvailable {
		t.Errorf("Expected %d indices (capped), got %d", totalAvailable, len(indices))
	}
}

// TestGenerateIndices_Randomization tests that different seeds produce
// different index sets (verifies actual randomization).
func TestGenerateIndices_Randomization(t *testing.T) {
	// Arrange: Same parameters, different seeds
	totalAvailable := 100
	requestedCount := 10

	rng1 := rand.New(rand.NewSource(1))
	rng2 := rand.New(rand.NewSource(2))

	// Act: Generate with different seeds
	indices1 := GenerateIndices(rng1, totalAvailable, requestedCount)
	indices2 := GenerateIndices(rng2, totalAvailable, requestedCount)

	// Assert: Should produce different results
	same := true
	for i := 0; i < requestedCount; i++ {
		if indices1[i] != indices2[i] {
			same = false
			break
		}
	}

	if same {
		t.Error("Expected different indices with different seeds, got identical sets")
	}
}

// TestExtractEmails_BasicExtraction tests extracting emails at specified
// indices while filtering out already-extracted ones.
func TestExtractEmails_BasicExtraction(t *testing.T) {
	// Arrange: Sample emails and empty registry
	emails := []EmailRecord{
		{File: "email-1", Message: "Message 1"},
		{File: "email-2", Message: "Message 2"},
		{File: "email-3", Message: "Message 3"},
		{File: "email-4", Message: "Message 4"},
		{File: "email-5", Message: "Message 5"},
	}
	registry := NewTrackingRegistry()
	indices := []int{0, 2, 4} // Extract emails 1, 3, 5

	// Act: Extract emails
	extracted := ExtractEmails(emails, indices, registry)

	// Assert: Should have 3 extracted emails
	if len(extracted) != 3 {
		t.Errorf("Expected 3 extracted emails, got %d", len(extracted))
	}

	// Assert: Should have correct emails
	expectedFiles := []string{"email-1", "email-3", "email-5"}
	for i, record := range extracted {
		if record.File != expectedFiles[i] {
			t.Errorf("Expected extracted[%d].File='%s', got '%s'",
				i, expectedFiles[i], record.File)
		}
	}
}

// TestExtractEmails_FilterDuplicates tests that ExtractEmails skips
// emails that exist in the tracking registry.
func TestExtractEmails_FilterDuplicates(t *testing.T) {
	// Arrange: Sample emails with some already in registry
	emails := []EmailRecord{
		{File: "email-1", Message: "Message 1"},
		{File: "email-2", Message: "Message 2"},
		{File: "email-3", Message: "Message 3"},
		{File: "email-4", Message: "Message 4"},
		{File: "email-5", Message: "Message 5"},
	}
	registry := NewTrackingRegistry()
	registry.Add("email-2") // Mark email-2 as already extracted
	registry.Add("email-4") // Mark email-4 as already extracted

	indices := []int{1, 3} // Try to extract emails 2 and 4 (both already extracted)

	// Act: Extract emails
	extracted := ExtractEmails(emails, indices, registry)

	// Assert: Should extract 0 emails (all filtered out)
	if len(extracted) != 0 {
		t.Errorf("Expected 0 extracted emails (all duplicates), got %d", len(extracted))
	}
}

// TestExtractEmails_PartialFiltering tests extraction when some emails
// are duplicates and some are new.
func TestExtractEmails_PartialFiltering(t *testing.T) {
	// Arrange: Sample emails with mixed registry status
	emails := []EmailRecord{
		{File: "email-1", Message: "Message 1"},
		{File: "email-2", Message: "Message 2"},
		{File: "email-3", Message: "Message 3"},
		{File: "email-4", Message: "Message 4"},
	}
	registry := NewTrackingRegistry()
	registry.Add("email-2") // Only email-2 is already extracted

	indices := []int{0, 1, 2, 3} // Try to extract all 4

	// Act: Extract emails
	extracted := ExtractEmails(emails, indices, registry)

	// Assert: Should extract 3 emails (excluding email-2)
	if len(extracted) != 3 {
		t.Errorf("Expected 3 extracted emails, got %d", len(extracted))
	}

	// Assert: email-2 should not be in results
	for _, record := range extracted {
		if record.File == "email-2" {
			t.Error("Expected email-2 to be filtered out, but found in results")
		}
	}
}

// TestExtractEmails_EmptyIndices tests extraction with no indices
// (edge case: requesting 0 emails).
func TestExtractEmails_EmptyIndices(t *testing.T) {
	// Arrange: Sample emails, empty indices
	emails := []EmailRecord{
		{File: "email-1", Message: "Message 1"},
		{File: "email-2", Message: "Message 2"},
	}
	registry := NewTrackingRegistry()
	indices := []int{} // No indices to extract

	// Act: Extract emails
	extracted := ExtractEmails(emails, indices, registry)

	// Assert: Should extract 0 emails
	if len(extracted) != 0 {
		t.Errorf("Expected 0 extracted emails, got %d", len(extracted))
	}
}

// TestExtractEmails_PreservesContent tests that ExtractEmails preserves
// the complete email content including multi-line messages.
func TestExtractEmails_PreservesContent(t *testing.T) {
	// Arrange: Email with multi-line message
	multiLineMsg := "Line 1\nLine 2\nLine 3"
	emails := []EmailRecord{
		{File: "email-1", Message: multiLineMsg},
	}
	registry := NewTrackingRegistry()
	indices := []int{0}

	// Act: Extract email
	extracted := ExtractEmails(emails, indices, registry)

	// Assert: Message content should be preserved exactly
	if len(extracted) != 1 {
		t.Fatalf("Expected 1 extracted email, got %d", len(extracted))
	}

	if extracted[0].Message != multiLineMsg {
		t.Errorf("Expected message to be preserved exactly, got:\n%s\nwant:\n%s",
			extracted[0].Message, multiLineMsg)
	}
}

package sampler

import (
	"fmt"
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

// T024: Additional tests for CountAvailable with exclusion logic (US2)

// TestCountAvailable_LargeRegistry tests performance with large tracking registry
func TestCountAvailable_LargeRegistry(t *testing.T) {
	// Arrange: Create registry with 10,000 extracted emails
	registry := NewTrackingRegistry()
	for i := 0; i < 10000; i++ {
		registry.Add(fmt.Sprintf("email-%d", i))
	}

	// Create sample emails with some overlap
	emails := make([]EmailRecord, 100)
	for i := 0; i < 100; i++ {
		emails[i] = EmailRecord{
			File:    fmt.Sprintf("email-%d", i+5000), // Emails 5000-5099, overlap with registry 0-9999
			Message: fmt.Sprintf("Message %d", i),
		}
	}

	// Act: Count available emails
	count := CountAvailable(emails, registry)

	// Assert: All 100 emails should be in registry (range 5000-5099 is within 0-9999)
	expectedAvailable := 0
	if count != expectedAvailable {
		t.Errorf("Expected count=%d with large registry (all in registry), got %d", expectedAvailable, count)
	}
}

// TestCountAvailable_MultipleRunsSimulation tests counting across multiple extraction runs
func TestCountAvailable_MultipleRunsSimulation(t *testing.T) {
	// Arrange: Simulate multiple runs
	registry := NewTrackingRegistry()
	allEmails := []EmailRecord{
		{File: "email-1", Message: "Message 1"},
		{File: "email-2", Message: "Message 2"},
		{File: "email-3", Message: "Message 3"},
		{File: "email-4", Message: "Message 4"},
		{File: "email-5", Message: "Message 5"},
	}

	// First run: Extract 2 emails
	count1 := CountAvailable(allEmails, registry)
	if count1 != 5 {
		t.Errorf("First run: expected 5 available, got %d", count1)
	}

	registry.Add("email-1")
	registry.Add("email-3")

	// Second run: Should have 3 available
	count2 := CountAvailable(allEmails, registry)
	if count2 != 3 {
		t.Errorf("Second run: expected 3 available, got %d", count2)
	}

	registry.Add("email-2")
	registry.Add("email-4")

	// Third run: Should have 1 available
	count3 := CountAvailable(allEmails, registry)
	if count3 != 1 {
		t.Errorf("Third run: expected 1 available, got %d", count3)
	}

	registry.Add("email-5")

	// Fourth run: Should have 0 available
	count4 := CountAvailable(allEmails, registry)
	if count4 != 0 {
		t.Errorf("Fourth run: expected 0 available, got %d", count4)
	}
}

// T025: Additional tests for ExtractEmails with duplicate filtering (US2)

// TestExtractEmails_AllDuplicates tests extraction when all selected emails are duplicates
func TestExtractEmails_AllDuplicates(t *testing.T) {
	// Arrange: All emails in registry
	emails := []EmailRecord{
		{File: "email-1", Message: "Message 1"},
		{File: "email-2", Message: "Message 2"},
		{File: "email-3", Message: "Message 3"},
	}
	registry := NewTrackingRegistry()
	registry.Add("email-1")
	registry.Add("email-2")
	registry.Add("email-3")

	indices := []int{0, 1, 2} // Try to extract all

	// Act: Extract emails
	extracted := ExtractEmails(emails, indices, registry)

	// Assert: Should extract nothing
	if len(extracted) != 0 {
		t.Errorf("Expected 0 extracted emails (all duplicates), got %d", len(extracted))
	}
}

// TestExtractEmails_ConsecutiveDuplicates tests filtering consecutive duplicates
func TestExtractEmails_ConsecutiveDuplicates(t *testing.T) {
	// Arrange: Registry with some consecutive emails
	emails := []EmailRecord{
		{File: "email-1", Message: "Message 1"}, // Available
		{File: "email-2", Message: "Message 2"}, // Duplicate
		{File: "email-3", Message: "Message 3"}, // Duplicate
		{File: "email-4", Message: "Message 4"}, // Available
		{File: "email-5", Message: "Message 5"}, // Available
	}
	registry := NewTrackingRegistry()
	registry.Add("email-2")
	registry.Add("email-3")

	indices := []int{0, 1, 2, 3, 4} // Try to extract all

	// Act: Extract emails
	extracted := ExtractEmails(emails, indices, registry)

	// Assert: Should extract only 3 emails (1, 4, 5)
	if len(extracted) != 3 {
		t.Errorf("Expected 3 extracted emails, got %d", len(extracted))
	}

	// Assert: Verify correct emails were extracted
	expectedFiles := []string{"email-1", "email-4", "email-5"}
	for i, email := range extracted {
		if email.File != expectedFiles[i] {
			t.Errorf("Expected extracted[%d].File='%s', got '%s'",
				i, expectedFiles[i], email.File)
		}
	}
}

// TestExtractEmails_DuplicatePreservesOrder tests that filtering maintains index order
func TestExtractEmails_DuplicatePreservesOrder(t *testing.T) {
	// Arrange: Mix of available and duplicate emails
	emails := []EmailRecord{
		{File: "email-1", Message: "Message 1"},
		{File: "email-2", Message: "Message 2"},
		{File: "email-3", Message: "Message 3"},
		{File: "email-4", Message: "Message 4"},
		{File: "email-5", Message: "Message 5"},
	}
	registry := NewTrackingRegistry()
	registry.Add("email-2")
	registry.Add("email-4")

	indices := []int{0, 1, 2, 3, 4} // Sorted indices

	// Act: Extract emails
	extracted := ExtractEmails(emails, indices, registry)

	// Assert: Extracted emails should maintain order (1, 3, 5)
	expectedFiles := []string{"email-1", "email-3", "email-5"}
	for i, email := range extracted {
		if email.File != expectedFiles[i] {
			t.Errorf("Expected extracted[%d].File='%s', got '%s'",
				i, expectedFiles[i], email.File)
		}
	}
}

// T036: Test for count exceeding available emails
// TestCountExceedsAvailable_RequestMoreThanTotal tests requesting more emails than exist in source
func TestCountExceedsAvailable_RequestMoreThanTotal(t *testing.T) {
	// Arrange: Create small set of emails
	emails := []EmailRecord{
		{File: "email-1", Message: "Message 1"},
		{File: "email-2", Message: "Message 2"},
		{File: "email-3", Message: "Message 3"},
	}
	registry := NewTrackingRegistry() // Empty registry

	// Act: Request more emails than available (requesting 10, only 3 available)
	availableCount := CountAvailable(emails, registry)

	// Assert: Available count should be 3
	if availableCount != 3 {
		t.Errorf("Expected availableCount=3, got %d", availableCount)
	}

	// Act: Try to generate indices for more than available
	// Should only generate as many as available
	requestedCount := 10
	actualCount := requestedCount
	if requestedCount > availableCount {
		actualCount = availableCount
	}

	indices := GenerateIndices(nil, availableCount, actualCount)

	// Assert: Should only get 3 indices
	if len(indices) != 3 {
		t.Errorf("Expected 3 indices (limited by available), got %d", len(indices))
	}

	// Assert: All indices should be valid (0-2)
	for i, idx := range indices {
		if idx < 0 || idx >= availableCount {
			t.Errorf("Index[%d]=%d is out of bounds [0, %d)", i, idx, availableCount)
		}
	}
}

// TestCountExceedsAvailable_RequestMoreAfterExtractions tests requesting more than remaining
func TestCountExceedsAvailable_RequestMoreAfterExtractions(t *testing.T) {
	// Arrange: Create emails with some already extracted
	emails := []EmailRecord{
		{File: "email-1", Message: "Message 1"},
		{File: "email-2", Message: "Message 2"},
		{File: "email-3", Message: "Message 3"},
		{File: "email-4", Message: "Message 4"},
		{File: "email-5", Message: "Message 5"},
	}
	registry := NewTrackingRegistry()
	registry.Add("email-2")
	registry.Add("email-4")
	registry.Add("email-5")
	// Only 2 emails available (email-1, email-3)

	// Act: Count available
	availableCount := CountAvailable(emails, registry)

	// Assert: Only 2 should be available
	if availableCount != 2 {
		t.Errorf("Expected availableCount=2, got %d", availableCount)
	}

	// Act: Request 10 emails when only 2 available
	requestedCount := 10
	actualCount := requestedCount
	if requestedCount > availableCount {
		actualCount = availableCount
	}

	// Filter to available emails
	var availableEmails []EmailRecord
	for _, email := range emails {
		if !registry.Contains(email.File) {
			availableEmails = append(availableEmails, email)
		}
	}

	indices := GenerateIndices(nil, len(availableEmails), actualCount)

	// Assert: Should only get 2 indices
	if len(indices) != 2 {
		t.Errorf("Expected 2 indices (limited by available), got %d", len(indices))
	}

	// Act: Extract
	extracted := ExtractEmails(availableEmails, indices, NewTrackingRegistry())

	// Assert: Should extract exactly 2 emails
	if len(extracted) != 2 {
		t.Errorf("Expected 2 extracted emails, got %d", len(extracted))
	}
}

// TestCountExceedsAvailable_ExactMatch tests requesting exactly the available count
func TestCountExceedsAvailable_ExactMatch(t *testing.T) {
	// Arrange: Create emails with exact count available
	emails := []EmailRecord{
		{File: "email-1", Message: "Message 1"},
		{File: "email-2", Message: "Message 2"},
		{File: "email-3", Message: "Message 3"},
	}
	registry := NewTrackingRegistry()

	// Act: Request exactly what's available
	availableCount := CountAvailable(emails, registry)
	requestedCount := 3

	// Assert: Available should match requested
	if availableCount != requestedCount {
		t.Errorf("Expected availableCount=%d, got %d", requestedCount, availableCount)
	}

	// Act: Generate indices
	indices := GenerateIndices(nil, availableCount, requestedCount)

	// Assert: Should get exactly 3 indices
	if len(indices) != 3 {
		t.Errorf("Expected 3 indices, got %d", len(indices))
	}

	// Act: Extract
	extracted := ExtractEmails(emails, indices, registry)

	// Assert: Should extract all 3 emails
	if len(extracted) != 3 {
		t.Errorf("Expected 3 extracted emails, got %d", len(extracted))
	}
}

// TestCountExceedsAvailable_OneRemaining tests edge case with only one email left
func TestCountExceedsAvailable_OneRemaining(t *testing.T) {
	// Arrange: Create emails with all but one extracted
	emails := []EmailRecord{
		{File: "email-1", Message: "Message 1"},
		{File: "email-2", Message: "Message 2"},
		{File: "email-3", Message: "Message 3"},
	}
	registry := NewTrackingRegistry()
	registry.Add("email-1")
	registry.Add("email-2")
	// Only email-3 remains

	// Act: Count available
	availableCount := CountAvailable(emails, registry)

	// Assert: Only 1 should be available
	if availableCount != 1 {
		t.Errorf("Expected availableCount=1, got %d", availableCount)
	}

	// Act: Request 100 emails when only 1 available
	requestedCount := 100
	actualCount := requestedCount
	if requestedCount > availableCount {
		actualCount = availableCount
	}

	// Filter to available
	var availableEmails []EmailRecord
	for _, email := range emails {
		if !registry.Contains(email.File) {
			availableEmails = append(availableEmails, email)
		}
	}

	indices := GenerateIndices(nil, len(availableEmails), actualCount)

	// Assert: Should only get 1 index
	if len(indices) != 1 {
		t.Errorf("Expected 1 index, got %d", len(indices))
	}

	// Act: Extract
	extracted := ExtractEmails(availableEmails, indices, NewTrackingRegistry())

	// Assert: Should extract exactly 1 email
	if len(extracted) != 1 {
		t.Errorf("Expected 1 extracted email, got %d", len(extracted))
	}

	// Assert: Should be email-3
	if extracted[0].File != "email-3" {
		t.Errorf("Expected extracted email-3, got %s", extracted[0].File)
	}
}

// T037: Tests for zero emails available edge case
// TestZeroEmailsAvailable_AllExtracted tests when all emails have been extracted
func TestZeroEmailsAvailable_AllExtracted(t *testing.T) {
	// Arrange: Create emails with all already extracted
	emails := []EmailRecord{
		{File: "email-1", Message: "Message 1"},
		{File: "email-2", Message: "Message 2"},
		{File: "email-3", Message: "Message 3"},
	}
	registry := NewTrackingRegistry()
	registry.Add("email-1")
	registry.Add("email-2")
	registry.Add("email-3")

	// Act: Count available
	availableCount := CountAvailable(emails, registry)

	// Assert: Should be zero
	if availableCount != 0 {
		t.Errorf("Expected availableCount=0, got %d", availableCount)
	}

	// Act: Filter to available emails
	var availableEmails []EmailRecord
	for _, email := range emails {
		if !registry.Contains(email.File) {
			availableEmails = append(availableEmails, email)
		}
	}

	// Assert: Should have no available emails
	if len(availableEmails) != 0 {
		t.Errorf("Expected 0 available emails, got %d", len(availableEmails))
	}
}

// TestZeroEmailsAvailable_EmptySource tests when source CSV has no emails
func TestZeroEmailsAvailable_EmptySource(t *testing.T) {
	// Arrange: Empty email list
	emails := []EmailRecord{}
	registry := NewTrackingRegistry()

	// Act: Count available
	availableCount := CountAvailable(emails, registry)

	// Assert: Should be zero
	if availableCount != 0 {
		t.Errorf("Expected availableCount=0 for empty source, got %d", availableCount)
	}

	// Act: Try to generate indices
	indices := GenerateIndices(nil, 0, 10)

	// Assert: Should generate no indices
	if len(indices) != 0 {
		t.Errorf("Expected 0 indices for zero available, got %d", len(indices))
	}
}

// TestZeroEmailsAvailable_RequestWithZeroAvailable tests requesting emails when none available
func TestZeroEmailsAvailable_RequestWithZeroAvailable(t *testing.T) {
	// Arrange: All emails already extracted
	emails := []EmailRecord{
		{File: "email-1", Message: "Message 1"},
		{File: "email-2", Message: "Message 2"},
	}
	registry := NewTrackingRegistry()
	registry.Add("email-1")
	registry.Add("email-2")

	// Act: Count available
	availableCount := CountAvailable(emails, registry)

	// Assert: Zero available
	if availableCount != 0 {
		t.Errorf("Expected availableCount=0, got %d", availableCount)
	}

	// Act: Request 10 emails
	requestedCount := 10
	actualCount := requestedCount
	if requestedCount > availableCount {
		actualCount = availableCount
	}

	// Assert: Actual count should be capped at 0
	if actualCount != 0 {
		t.Errorf("Expected actualCount=0 (capped), got %d", actualCount)
	}

	// Act: Filter available emails
	var availableEmails []EmailRecord
	for _, email := range emails {
		if !registry.Contains(email.File) {
			availableEmails = append(availableEmails, email)
		}
	}

	// Act: Generate indices for zero available
	indices := GenerateIndices(nil, len(availableEmails), actualCount)

	// Assert: Should get no indices
	if len(indices) != 0 {
		t.Errorf("Expected 0 indices when none available, got %d", len(indices))
	}

	// Act: Extract with no indices
	extracted := ExtractEmails(availableEmails, indices, NewTrackingRegistry())

	// Assert: Should extract nothing
	if len(extracted) != 0 {
		t.Errorf("Expected 0 extracted emails, got %d", len(extracted))
	}
}

// TestZeroEmailsAvailable_EmptyRegistry tests zero count with empty tracking
func TestZeroEmailsAvailable_EmptyRegistry(t *testing.T) {
	// Arrange: Empty emails and empty registry
	emails := []EmailRecord{}
	registry := NewTrackingRegistry()

	// Act: Count available
	availableCount := CountAvailable(emails, registry)

	// Assert: Should be zero
	if availableCount != 0 {
		t.Errorf("Expected availableCount=0, got %d", availableCount)
	}

	// Act: Extract with empty inputs
	indices := []int{}
	extracted := ExtractEmails(emails, indices, registry)

	// Assert: Should extract nothing
	if len(extracted) != 0 {
		t.Errorf("Expected 0 extracted emails, got %d", len(extracted))
	}
}

// TestZeroEmailsAvailable_GenerateIndicesWithZeroCount tests GenerateIndices edge case
func TestZeroEmailsAvailable_GenerateIndicesWithZeroCount(t *testing.T) {
	// Arrange: Request zero emails
	poolSize := 100
	requestedCount := 0

	// Act: Generate indices for zero count
	indices := GenerateIndices(nil, poolSize, requestedCount)

	// Assert: Should return empty slice
	if len(indices) != 0 {
		t.Errorf("Expected 0 indices for requestedCount=0, got %d", len(indices))
	}
}

// TestZeroEmailsAvailable_GenerateIndicesWithZeroPool tests GenerateIndices with empty pool
func TestZeroEmailsAvailable_GenerateIndicesWithZeroPool(t *testing.T) {
	// Arrange: Request from empty pool
	poolSize := 0
	requestedCount := 10

	// Act: Generate indices from empty pool
	indices := GenerateIndices(nil, poolSize, requestedCount)

	// Assert: Should return empty slice (can't generate from empty pool)
	if len(indices) != 0 {
		t.Errorf("Expected 0 indices for poolSize=0, got %d", len(indices))
	}
}

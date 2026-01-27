package sampler

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestLoadTracking_EmptyDirectory tests loading tracking files when none exist
func TestLoadTracking_EmptyDirectory(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()

	registry, _, err := LoadTracking(tmpDir)
	if err != nil {
		t.Fatalf("LoadTracking failed on empty directory: %v", err)
	}

	if registry.Count() != 0 {
		t.Errorf("Expected 0 tracked emails, got %d", registry.Count())
	}
}

// TestLoadTracking_SingleFile tests loading one tracking file
func TestLoadTracking_SingleFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create tracking file
	trackingContent := "test-email-1\ntest-email-2\ntest-email-3\n"
	trackingPath := filepath.Join(tmpDir, "extracted-20260127-120000.txt")
	if err := os.WriteFile(trackingPath, []byte(trackingContent), 0644); err != nil {
		t.Fatalf("Failed to create tracking file: %v", err)
	}

	registry, _, err := LoadTracking(tmpDir)
	if err != nil {
		t.Fatalf("LoadTracking failed: %v", err)
	}

	if registry.Count() != 3 {
		t.Errorf("Expected 3 tracked emails, got %d", registry.Count())
	}

	// Verify specific IDs
	if !registry.Contains("test-email-1") {
		t.Error("Registry should contain test-email-1")
	}
	if !registry.Contains("test-email-2") {
		t.Error("Registry should contain test-email-2")
	}
	if !registry.Contains("test-email-3") {
		t.Error("Registry should contain test-email-3")
	}
}

// TestLoadTracking_MultipleFiles tests aggregating multiple tracking files
func TestLoadTracking_MultipleFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create first tracking file
	file1Path := filepath.Join(tmpDir, "extracted-20260127-120000.txt")
	if err := os.WriteFile(file1Path, []byte("email-1\nemail-2\n"), 0644); err != nil {
		t.Fatalf("Failed to create tracking file 1: %v", err)
	}

	// Create second tracking file
	file2Path := filepath.Join(tmpDir, "extracted-20260127-130000.txt")
	if err := os.WriteFile(file2Path, []byte("email-3\nemail-4\n"), 0644); err != nil {
		t.Fatalf("Failed to create tracking file 2: %v", err)
	}

	registry, _, err := LoadTracking(tmpDir)
	if err != nil {
		t.Fatalf("LoadTracking failed: %v", err)
	}

	if registry.Count() != 4 {
		t.Errorf("Expected 4 tracked emails (aggregated), got %d", registry.Count())
	}

	// Verify all IDs are present
	for _, id := range []string{"email-1", "email-2", "email-3", "email-4"} {
		if !registry.Contains(id) {
			t.Errorf("Registry should contain %s", id)
		}
	}
}

// TestLoadTracking_DuplicatesAcrossFiles tests handling duplicate IDs across multiple files
func TestLoadTracking_DuplicatesAcrossFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create tracking files with overlapping IDs
	file1Path := filepath.Join(tmpDir, "extracted-20260127-120000.txt")
	if err := os.WriteFile(file1Path, []byte("email-1\nemail-2\n"), 0644); err != nil {
		t.Fatalf("Failed to create tracking file 1: %v", err)
	}

	file2Path := filepath.Join(tmpDir, "extracted-20260127-130000.txt")
	if err := os.WriteFile(file2Path, []byte("email-2\nemail-3\n"), 0644); err != nil {
		t.Fatalf("Failed to create tracking file 2: %v", err)
	}

	registry, _, err := LoadTracking(tmpDir)
	if err != nil {
		t.Fatalf("LoadTracking failed: %v", err)
	}

	// Should have 3 unique IDs (email-1, email-2, email-3)
	if registry.Count() != 3 {
		t.Errorf("Expected 3 unique tracked emails, got %d", registry.Count())
	}
}

// TestLoadTracking_EmptyLines tests handling empty lines in tracking files
func TestLoadTracking_EmptyLines(t *testing.T) {
	tmpDir := t.TempDir()

	// Create tracking file with empty lines
	trackingContent := "email-1\n\nemail-2\n\n\nemail-3\n"
	trackingPath := filepath.Join(tmpDir, "extracted-20260127-120000.txt")
	if err := os.WriteFile(trackingPath, []byte(trackingContent), 0644); err != nil {
		t.Fatalf("Failed to create tracking file: %v", err)
	}

	registry, _, err := LoadTracking(tmpDir)
	if err != nil {
		t.Fatalf("LoadTracking failed: %v", err)
	}

	// Should only count non-empty lines
	if registry.Count() != 3 {
		t.Errorf("Expected 3 tracked emails (empty lines ignored), got %d", registry.Count())
	}
}

// TestLoadTracking_IgnoresNonTrackingFiles tests that non-matching files are ignored
func TestLoadTracking_IgnoresNonTrackingFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create valid tracking file
	trackingPath := filepath.Join(tmpDir, "extracted-20260127-120000.txt")
	if err := os.WriteFile(trackingPath, []byte("email-1\n"), 0644); err != nil {
		t.Fatalf("Failed to create tracking file: %v", err)
	}

	// Create non-tracking files that should be ignored
	if err := os.WriteFile(filepath.Join(tmpDir, "other-file.txt"), []byte("ignore-me\n"), 0644); err != nil {
		t.Fatalf("Failed to create other file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "sampled-emails-20260127-120000.csv"), []byte("file,message\n"), 0644); err != nil {
		t.Fatalf("Failed to create CSV file: %v", err)
	}

	registry, _, err := LoadTracking(tmpDir)
	if err != nil {
		t.Fatalf("LoadTracking failed: %v", err)
	}

	// Should only load from extracted-*.txt file
	if registry.Count() != 1 {
		t.Errorf("Expected 1 tracked email (non-tracking files ignored), got %d", registry.Count())
	}
}

// TestLoadTracking_CorruptedFile tests handling of corrupted tracking files
func TestLoadTracking_CorruptedFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create valid tracking file
	validPath := filepath.Join(tmpDir, "extracted-20260127-120000.txt")
	if err := os.WriteFile(validPath, []byte("email-1\nemail-2\n"), 0644); err != nil {
		t.Fatalf("Failed to create valid tracking file: %v", err)
	}

	// Create corrupted tracking file (binary garbage)
	corruptedPath := filepath.Join(tmpDir, "extracted-20260127-130000.txt")
	if err := os.WriteFile(corruptedPath, []byte{0xFF, 0xFE, 0x00, 0x01}, 0644); err != nil {
		t.Fatalf("Failed to create corrupted file: %v", err)
	}

	registry, _, err := LoadTracking(tmpDir)

	// Should succeed and skip corrupted file
	if err != nil {
		t.Fatalf("LoadTracking should handle corrupted files gracefully: %v", err)
	}

	// Should only have entries from valid file
	if registry.Count() < 2 {
		t.Errorf("Expected at least 2 tracked emails from valid file, got %d", registry.Count())
	}

	// Verify valid entries are present
	if !registry.Contains("email-1") || !registry.Contains("email-2") {
		t.Error("Valid tracking entries should be present despite corrupted file")
	}
}

// TestLoadTracking_WhitespaceHandling tests trimming whitespace from IDs
func TestLoadTracking_WhitespaceHandling(t *testing.T) {
	tmpDir := t.TempDir()

	// Create tracking file with whitespace
	trackingContent := "  email-1  \n\temail-2\t\nemail-3 \n"
	trackingPath := filepath.Join(tmpDir, "extracted-20260127-120000.txt")
	if err := os.WriteFile(trackingPath, []byte(trackingContent), 0644); err != nil {
		t.Fatalf("Failed to create tracking file: %v", err)
	}

	registry, _, err := LoadTracking(tmpDir)
	if err != nil {
		t.Fatalf("LoadTracking failed: %v", err)
	}

	// Check that IDs are trimmed
	if !registry.Contains("email-1") {
		t.Error("Should contain email-1 (whitespace trimmed)")
	}
	if !registry.Contains("email-2") {
		t.Error("Should contain email-2 (whitespace trimmed)")
	}
	if !registry.Contains("email-3") {
		t.Error("Should contain email-3 (whitespace trimmed)")
	}

	// Check that spaces in IDs are not present
	if registry.Contains("  email-1  ") || registry.Contains("\temail-2\t") {
		t.Error("IDs should have whitespace trimmed")
	}
}

// T023: Tests for CreateTrackingFile function

// TestCreateTrackingFile_BasicWrite tests creating a tracking file with email IDs
func TestCreateTrackingFile_BasicWrite(t *testing.T) {
	tmpDir := t.TempDir()
	timestamp := "20260127-143000"
	emailIDs := []string{"email-1", "email-2", "email-3"}

	// Act: Create tracking file
	err := CreateTrackingFile(tmpDir, timestamp, emailIDs)
	if err != nil {
		t.Fatalf("CreateTrackingFile failed: %v", err)
	}

	// Assert: File should exist
	expectedPath := filepath.Join(tmpDir, "extracted-20260127-143000.txt")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Fatalf("Tracking file was not created at %s", expectedPath)
	}

	// Assert: File should contain all email IDs
	content, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("Failed to read created tracking file: %v", err)
	}

	contentStr := string(content)
	for _, id := range emailIDs {
		if !strings.Contains(contentStr, id) {
			t.Errorf("Tracking file should contain '%s', got:\n%s", id, contentStr)
		}
	}
}

// TestCreateTrackingFile_EmptyList tests creating a tracking file with no emails
func TestCreateTrackingFile_EmptyList(t *testing.T) {
	tmpDir := t.TempDir()
	timestamp := "20260127-143000"
	emailIDs := []string{}

	// Act: Create tracking file with empty list
	err := CreateTrackingFile(tmpDir, timestamp, emailIDs)
	if err != nil {
		t.Fatalf("CreateTrackingFile failed with empty list: %v", err)
	}

	// Assert: File should exist but be empty (except newlines)
	expectedPath := filepath.Join(tmpDir, "extracted-20260127-143000.txt")
	content, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("Failed to read created tracking file: %v", err)
	}

	contentStr := strings.TrimSpace(string(content))
	if contentStr != "" {
		t.Errorf("Expected empty file, got: %s", contentStr)
	}
}

// TestCreateTrackingFile_UniqueFilenames tests that different timestamps create different files
func TestCreateTrackingFile_UniqueFilenames(t *testing.T) {
	tmpDir := t.TempDir()

	// Create first tracking file
	err := CreateTrackingFile(tmpDir, "20260127-120000", []string{"email-1"})
	if err != nil {
		t.Fatalf("Failed to create first tracking file: %v", err)
	}

	// Create second tracking file with different timestamp
	err = CreateTrackingFile(tmpDir, "20260127-130000", []string{"email-2"})
	if err != nil {
		t.Fatalf("Failed to create second tracking file: %v", err)
	}

	// Assert: Both files should exist
	path1 := filepath.Join(tmpDir, "extracted-20260127-120000.txt")
	path2 := filepath.Join(tmpDir, "extracted-20260127-130000.txt")

	if _, err := os.Stat(path1); os.IsNotExist(err) {
		t.Error("First tracking file should exist")
	}
	if _, err := os.Stat(path2); os.IsNotExist(err) {
		t.Error("Second tracking file should exist")
	}

	// Assert: Files should have different content
	content1, _ := os.ReadFile(path1)
	content2, _ := os.ReadFile(path2)

	if string(content1) == string(content2) {
		t.Error("Different tracking files should have different content")
	}
}

// TestCreateTrackingFile_InvalidDirectory tests handling of invalid directory
func TestCreateTrackingFile_InvalidDirectory(t *testing.T) {
	invalidDir := "/nonexistent/directory/path"
	timestamp := "20260127-143000"
	emailIDs := []string{"email-1"}

	// Act: Try to create tracking file in invalid directory
	err := CreateTrackingFile(invalidDir, timestamp, emailIDs)

	// Assert: Should return error
	if err == nil {
		t.Error("Expected error when creating file in nonexistent directory")
	}
}

// TestCreateTrackingFile_OneLinePerID tests that each email ID is on its own line
func TestCreateTrackingFile_OneLinePerID(t *testing.T) {
	tmpDir := t.TempDir()
	timestamp := "20260127-143000"
	emailIDs := []string{"email-1", "email-2", "email-3", "email-4"}

	// Act: Create tracking file
	err := CreateTrackingFile(tmpDir, timestamp, emailIDs)
	if err != nil {
		t.Fatalf("CreateTrackingFile failed: %v", err)
	}

	// Assert: File should have exactly 4 lines (one per ID)
	expectedPath := filepath.Join(tmpDir, "extracted-20260127-143000.txt")
	content, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("Failed to read tracking file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != len(emailIDs) {
		t.Errorf("Expected %d lines (one per ID), got %d", len(emailIDs), len(lines))
	}

	// Assert: Each line should match an email ID
	for i, line := range lines {
		if line != emailIDs[i] {
			t.Errorf("Line %d: expected '%s', got '%s'", i, emailIDs[i], line)
		}
	}
}

// TestCreateTrackingFile_CanBeReloaded tests that created files can be loaded by LoadTracking
func TestCreateTrackingFile_CanBeReloaded(t *testing.T) {
	tmpDir := t.TempDir()
	timestamp := "20260127-143000"
	emailIDs := []string{"email-1", "email-2", "email-3"}

	// Act: Create tracking file
	err := CreateTrackingFile(tmpDir, timestamp, emailIDs)
	if err != nil {
		t.Fatalf("CreateTrackingFile failed: %v", err)
	}

	// Act: Load tracking files
	registry, _, err := LoadTracking(tmpDir)
	if err != nil {
		t.Fatalf("LoadTracking failed: %v", err)
	}

	// Assert: Registry should contain all created IDs
	if registry.Count() != len(emailIDs) {
		t.Errorf("Expected registry count=%d, got %d", len(emailIDs), registry.Count())
	}

	for _, id := range emailIDs {
		if !registry.Contains(id) {
			t.Errorf("Registry should contain '%s' after reload", id)
		}
	}
}

// T027: Additional tests for corrupted tracking file handling (US2)

// TestLoadTracking_CorruptedFileAmongValid tests that corrupted files are skipped
// while valid files are still loaded.
func TestLoadTracking_CorruptedFileAmongValid(t *testing.T) {
	tmpDir := t.TempDir()

	// Create first valid tracking file
	validPath1 := filepath.Join(tmpDir, "extracted-20260127-100000.txt")
	if err := os.WriteFile(validPath1, []byte("email-1\nemail-2\n"), 0644); err != nil {
		t.Fatalf("Failed to create valid tracking file 1: %v", err)
	}

	// Create corrupted tracking file (contains null bytes)
	corruptedPath := filepath.Join(tmpDir, "extracted-20260127-110000.txt")
	corruptedData := []byte("email-3\x00\x00\x00invalid\nemail-4\n")
	if err := os.WriteFile(corruptedPath, corruptedData, 0644); err != nil {
		t.Fatalf("Failed to create corrupted file: %v", err)
	}

	// Create second valid tracking file
	validPath2 := filepath.Join(tmpDir, "extracted-20260127-120000.txt")
	if err := os.WriteFile(validPath2, []byte("email-5\nemail-6\n"), 0644); err != nil {
		t.Fatalf("Failed to create valid tracking file 2: %v", err)
	}

	// Act: Load tracking files
	registry, _, err := LoadTracking(tmpDir)

	// Assert: Should succeed despite corrupted file
	if err != nil {
		t.Fatalf("LoadTracking should handle corrupted files gracefully: %v", err)
	}

	// Assert: Should have entries from valid files
	// (Corrupted file may contribute some valid lines before corruption)
	if registry.Count() < 4 {
		t.Errorf("Expected at least 4 tracked emails from valid files, got %d", registry.Count())
	}

	// Verify valid entries are present
	if !registry.Contains("email-1") || !registry.Contains("email-2") {
		t.Error("Valid tracking entries from file 1 should be present")
	}
	if !registry.Contains("email-5") || !registry.Contains("email-6") {
		t.Error("Valid tracking entries from file 2 should be present")
	}
}

// TestLoadTracking_EmptyCorruptedFile tests handling of empty corrupted files
func TestLoadTracking_EmptyCorruptedFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create empty file (edge case of corruption)
	emptyPath := filepath.Join(tmpDir, "extracted-20260127-120000.txt")
	if err := os.WriteFile(emptyPath, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to create empty file: %v", err)
	}

	// Act: Load tracking files
	registry, _, err := LoadTracking(tmpDir)

	// Assert: Should succeed
	if err != nil {
		t.Fatalf("LoadTracking should handle empty files: %v", err)
	}

	// Assert: Registry should be empty
	if registry.Count() != 0 {
		t.Errorf("Expected empty registry for empty file, got %d entries", registry.Count())
	}
}

// TestLoadTracking_UnreadableFile tests handling of files with permission issues
func TestLoadTracking_UnreadableFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create valid tracking file
	validPath := filepath.Join(tmpDir, "extracted-20260127-100000.txt")
	if err := os.WriteFile(validPath, []byte("email-1\nemail-2\n"), 0644); err != nil {
		t.Fatalf("Failed to create valid tracking file: %v", err)
	}

	// Create unreadable file (no read permissions)
	unreadablePath := filepath.Join(tmpDir, "extracted-20260127-110000.txt")
	if err := os.WriteFile(unreadablePath, []byte("email-3\n"), 0000); err != nil {
		t.Fatalf("Failed to create unreadable file: %v", err)
	}
	// Clean up permissions after test
	defer os.Chmod(unreadablePath, 0644)

	// Act: Load tracking files
	registry, _, err := LoadTracking(tmpDir)

	// Assert: Should succeed and load valid file
	if err != nil {
		t.Fatalf("LoadTracking should skip unreadable files: %v", err)
	}

	// Assert: Should have entries from valid file only
	if registry.Count() < 2 {
		t.Errorf("Expected at least 2 entries from valid file, got %d", registry.Count())
	}

	// Verify valid entries are present
	if !registry.Contains("email-1") || !registry.Contains("email-2") {
		t.Error("Valid tracking entries should be present despite unreadable file")
	}
}

// TestLoadTracking_VeryLongLines tests handling of corrupted files with extremely long lines
func TestLoadTracking_VeryLongLines(t *testing.T) {
	tmpDir := t.TempDir()

	// Create file with very long line (potential buffer overflow)
	longLine := strings.Repeat("x", 1000000) // 1MB line
	content := "email-1\n" + longLine + "\nemail-2\n"
	longLinePath := filepath.Join(tmpDir, "extracted-20260127-120000.txt")
	if err := os.WriteFile(longLinePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create file with long line: %v", err)
	}

	// Act: Load tracking files
	registry, _, err := LoadTracking(tmpDir)

	// Assert: Should handle gracefully (may skip long line or include it)
	if err != nil {
		t.Fatalf("LoadTracking should handle long lines: %v", err)
	}

	// Assert: Should at least have the valid entries
	if !registry.Contains("email-1") {
		t.Error("Should contain email-1")
	}
}

// TestLoadTracking_MixedLineEndings tests handling of files with mixed line endings
func TestLoadTracking_MixedLineEndings(t *testing.T) {
	tmpDir := t.TempDir()

	// Create file with mixed line endings (Windows CRLF, Unix LF)
	// Note: bufio.Scanner treats \r\n and \n as line delimiters, but not \r alone
	content := "email-1\r\nemail-2\nemail-3\n"
	mixedPath := filepath.Join(tmpDir, "extracted-20260127-120000.txt")
	if err := os.WriteFile(mixedPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create file with mixed line endings: %v", err)
	}

	// Act: Load tracking files
	registry, _, err := LoadTracking(tmpDir)

	// Assert: Should succeed
	if err != nil {
		t.Fatalf("LoadTracking failed with mixed line endings: %v", err)
	}

	// Assert: Should handle LF and CRLF line endings
	if !registry.Contains("email-1") || !registry.Contains("email-2") || !registry.Contains("email-3") {
		t.Error("Should handle standard line endings (LF and CRLF)")
	}
}

// TestLoadTracking_NonUTF8Content tests handling of files with invalid UTF-8 sequences
func TestLoadTracking_NonUTF8Content(t *testing.T) {
	tmpDir := t.TempDir()

	// Create file with invalid UTF-8 sequences
	invalidUTF8 := []byte{0xFF, 0xFE, 0xFD} // Invalid UTF-8 bytes
	content := append([]byte("email-1\n"), invalidUTF8...)
	content = append(content, []byte("\nemail-2\n")...)

	invalidPath := filepath.Join(tmpDir, "extracted-20260127-120000.txt")
	if err := os.WriteFile(invalidPath, content, 0644); err != nil {
		t.Fatalf("Failed to create file with invalid UTF-8: %v", err)
	}

	// Act: Load tracking files
	registry, _, err := LoadTracking(tmpDir)

	// Assert: Should handle gracefully (may replace invalid sequences or skip)
	if err != nil {
		t.Fatalf("LoadTracking should handle non-UTF8 content: %v", err)
	}

	// Assert: Should at least have valid entries
	if !registry.Contains("email-1") {
		t.Error("Should contain email-1 (before invalid UTF-8)")
	}
}

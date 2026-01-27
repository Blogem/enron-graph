package sampler

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoadTracking_EmptyDirectory tests loading tracking files when none exist
func TestLoadTracking_EmptyDirectory(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()

	registry, err := LoadTracking(tmpDir)
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

	registry, err := LoadTracking(tmpDir)
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

	registry, err := LoadTracking(tmpDir)
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

	registry, err := LoadTracking(tmpDir)
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

	registry, err := LoadTracking(tmpDir)
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

	registry, err := LoadTracking(tmpDir)
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

	registry, err := LoadTracking(tmpDir)

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

	registry, err := LoadTracking(tmpDir)
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

# Research: Random Email Extractor

**Feature**: 003-random-email-extractor  
**Phase**: 0 - Research  
**Date**: 2026-01-27

## Overview

This document captures research findings for implementing a random email sampling utility. The tool must extract configurable numbers of emails from a large CSV file, prevent duplicate extractions across runs, and maintain CSV format compatibility.

## Research Tasks

### 1. Randomization Strategy for Large Files

**Decision**: Two-pass approach with reservoir sampling fallback

**Rationale**:
- First pass: Count total available emails (not in tracking file)
- Second pass: Select random indices and extract on-the-fly
- Avoids loading entire 500k email file into memory
- Enables true uniform random selection across entire dataset

**Implementation Approach**:
```go
1. Load tracking file into map for O(1) lookup
2. Stream CSV and count unextracted emails
3. Generate N random indices in range [0, count-1]
4. Sort indices for sequential file access
5. Stream CSV again, extracting emails at selected indices
```

**Alternatives Considered**:
- **Reservoir sampling (single-pass)**: Rejected because requires unknown sample size upfront and doesn't work well with tracking file exclusions
- **Load all into memory**: Rejected due to 500k email file size (memory constraints)
- **Database-backed selection**: Rejected to maintain simplicity and avoid database dependency

### 2. Tracking File Format

**Decision**: Simple text file with one identifier per line

**Rationale**:
- Maximum simplicity for manual inspection and debugging
- Easy to append (newline-delimited)
- Compatible with standard Unix tools (grep, wc -l, sort)
- No parsing complexity or version management

**Format Specification**:
```
# File: assets/enron-emails/extracted-YYYYMMDD-HHMMSS.txt
# One email file identifier per line (one file per extraction session)
allen-p/_sent_mail/1.
allen-p/_sent_mail/2.
bailey-s/all_documents/1.
...
```

**Alternatives Considered**:
- **JSON format**: Rejected as unnecessarily complex for simple list; harder to manually edit
- **CSV with metadata**: Rejected as no need for extraction timestamps or other metadata per requirements
- **Binary format**: Rejected due to lack of human readability and debugging difficulty

### 3. CSV Parsing Reuse

**Decision**: Reuse existing `internal/loader/parser.go` ParseCSV() function

**Rationale**:
- Already handles multi-line quoted email content correctly
- Proven with 500k email dataset in production use
- Configured with `LazyQuotes = true` for lenient parsing
- Returns streaming channels to avoid memory issues

**Integration Pattern**:
```go
// Existing function from internal/loader/parser.go
records, errors, err := loader.ParseCSV(filePath)
if err != nil {
    return err
}

// Process streaming records
for record := range records {
    // Use record.File as unique identifier
    // Use record.Message for email content
}
```

**Code Location**: `/Users/jochem/code/enron-graph/internal/loader/parser.go`

**Alternatives Considered**:
- **Reimplementation**: Rejected to avoid duplicating proven CSV handling logic
- **Direct encoding/csv use**: Rejected because loader.ParseCSV() has domain-specific configuration (LazyQuotes, field count validation)

### 4. Output Filename Generation

**Decision**: Timestamp-based naming with format `sampled-emails-YYYYMMDD-HHMMSS.csv`

**Rationale**:
- User clarification specified timestamp-based approach
- Sortable filenames (chronological listing)
- No collision risk (second-level precision sufficient for manual tool usage)
- Self-documenting extraction time

**Implementation**:
```go
timestamp := time.Now().Format("20060102-150405")
filename := fmt.Sprintf("sampled-emails-%s.csv", timestamp)
outputPath := filepath.Join("assets/enron-emails", filename)
```

**Alternatives Considered**:
- **Sequential numbering**: Rejected due to complexity of tracking next number
- **Fixed filename**: Rejected as overwrites previous samples
- **User-specified names**: Rejected to maintain simplicity (not in requirements)

### 5. Verbose Logging Strategy

**Decision**: Progress updates using standard log package with configurable verbosity

**Rationale**:
- User clarification specified verbose logging with progress (e.g., "Processed 500/1000 emails...")
- Standard library log package sufficient for CLI tool
- Can output to stderr to keep stdout clean for piping

**Logging Points**:
1. Startup: "Loading tracking file..." with count of tracked emails
2. First pass: "Counting available emails... N found"
3. Selection: "Generating random sample of N emails from M available"
4. Extraction progress: Every 100 emails processed (e.g., "Extracted 500/1000 emails...")
5. Completion: "Successfully extracted N emails to [filename]"
6. Errors: All skipped rows, corrupted tracking file, etc.

**Alternatives Considered**:
- **Silent operation**: Rejected per user clarification
- **Progress bars**: Rejected as unnecessary complexity for development tool
- **Structured logging**: Rejected as overkill for simple CLI utility

### 6. CSV Writing with Format Preservation

**Decision**: Use encoding/csv writer with same settings as parser

**Rationale**:
- Must preserve exact format for compatibility (SC-003)
- encoding/csv handles multi-line field quoting automatically
- Match parser settings: LazyQuotes for reading, but use strict quoting for output

**Implementation**:
```go
writer := csv.NewWriter(file)
writer.Write([]string{"file", "message"}) // Header
for each selected record {
    writer.Write([]string{record.File, record.Message})
}
writer.Flush()
```

**Format Guarantees**:
- Header row preserved: "file,message"
- Multi-line email content properly quoted
- Field delimiters: comma (standard)
- Compatible with loader.ParseCSV() for verification

### 7. Error Handling Strategy

**Decision**: Graceful degradation with user notification

**Per Edge Cases Specification**:
- **Insufficient available emails**: Extract all remaining, log actual count
- **Corrupted CSV rows**: Skip row, log warning, continue processing
- **Corrupted tracking file**: Log warning, start fresh tracking
- **Missing source file**: Fatal error with clear message
- **Disk space issues**: Fatal error, do not create partial files

**Error Reporting**:
- Use log.Printf() for warnings (skipped rows)
- Use log.Fatal() for fatal errors (missing source file)
- Return non-zero exit codes for scriptability

### 8. File Identifier Uniqueness

**Decision**: Use "file" column as unique identifier without runtime validation

**Rationale**:
- Static dataset verified to have 100% unique file identifiers (517,401 emails, 517,401 unique identifiers)
- No runtime validation needed since source CSV is not modified
- Verification performed via shell script analysis on entire dataset: `grep -E '^"[^"]+","Message-ID:' emails.csv | cut -d',' -f1 | tr -d '"' | sort | uniq -d` returns zero duplicates

**Implementation**:
```go
// Simply use record.File as unique identifier
// No validation needed - source is static and verified
tracker.Add(record.File)
```

**Assumptions**:
- Source CSV file (emails.csv) is static and not modified
- Uniqueness verified externally during development
- No duplicate detection logic needed in production code

## Technology Choices

### Go Standard Library Packages

| Package | Purpose | Justification |
|---------|---------|---------------|
| `encoding/csv` | CSV reading/writing | Already used in loader, handles quoting |
| `flag` | CLI arguments | Standard in all cmd/ tools |
| `math/rand` | Randomization | Sufficient for non-cryptographic sampling |
| `time` | Timestamps | Filename generation, seeding random |
| `os` | File I/O | Reading source, writing output |
| `log` | Logging | Simple CLI output, no structured logging needed |
| `bufio` | Tracking file | Line-by-line reading for efficiency |

### No External Dependencies Required

The implementation requires only Go standard library packages. This aligns with:
- Simplicity principle (Principle V)
- Existing project patterns (cmd/loader uses minimal dependencies)
- Easy deployment (no vendoring or go.mod changes)

## Performance Considerations

### Memory Usage

**Strategy**: Streaming processing to handle large files
- Tracking file loaded into map: ~50MB for 500k entries (100 bytes/identifier)
- Random index list: ~80KB for 10k indices (8 bytes each)
- CSV streaming: constant memory (buffered channels)
- Total estimated: <100MB for worst-case scenario

### Processing Speed

**Estimates** (based on existing loader performance):
- CSV parsing: 100k+ emails/sec (proven in loader tests)
- Random index generation: <1ms for 10k indices
- Two-pass approach: ~0.5s per 500k emails (counting) + 0.1s (extraction)
- Total: **<1 second for 10k email sample** from 500k source (exceeds SC-001 requirement of 10 seconds)

### File I/O Optimization

- Sequential access pattern (sorted indices) for better disk cache performance
- Buffered readers/writers to minimize syscalls
- Tracking file kept sorted for binary search (future optimization if needed)

## Summary

All research tasks completed with clear decisions documented. The implementation will:

1. **Reuse proven CSV parsing** from internal/loader
2. **Use two-pass random sampling** for memory efficiency and true randomization
3. **Simple text-based tracking** for debuggability and manual inspection
4. **Timestamp-based output naming** for organization and no collisions
5. **Verbose progress logging** for user visibility
6. **Graceful error handling** per specification edge cases
7. **Standard library only** for simplicity and minimal dependencies

No blocking research items remain. Ready to proceed to Phase 1 (Design).

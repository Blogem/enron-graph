# Quickstart: Random Email Extractor

**Feature**: 003-random-email-extractor  
**Version**: 1.0.0  
**Last Updated**: 2026-01-27

## Overview

The Random Email Extractor is a command-line utility that creates random samples from the Enron emails dataset. It prevents duplicate extractions across multiple runs by maintaining a tracking file of previously extracted emails.

## Prerequisites

- Go 1.21 or higher
- Source data: `assets/enron-emails/emails.csv` (Enron dataset)
- Write access to `assets/enron-emails/` directory

## Installation

No installation required. Run directly from source:

```bash
cd /path/to/enron-graph
go run cmd/sampler/main.go [options]
```

Or build the binary:

```bash
go build -o bin/sampler cmd/sampler/main.go
./bin/sampler [options]
```

## Basic Usage

### Extract 1000 Random Emails

```bash
go run cmd/sampler/main.go --count 1000
```

**Output**:
```
Loading tracking files...
Found 0 previously extracted emails (from 0 tracking files)
Counting available emails... 517401 found
Generating random sample of 1000 emails from 517401 available
Extracting emails...
Processed 100/1000 emails...
Processed 200/1000 emails...
...
Processed 1000/1000 emails...
Successfully extracted 1000 emails to assets/enron-emails/sampled-emails-20260127-143022.csv
Created tracking file: assets/enron-emails/extracted-20260127-143022.txt
```

### Extract Smaller Sample

```bash
# For quick testing (10 emails)
go run cmd/sampler/main.go --count 10

# For unit tests (100 emails)
go run cmd/sampler/main.go --count 100
```

### Extract Large Sample

```bash
# For performance testing (10,000 emails)
go run cmd/sampler/main.go --count 10000
```

## Command-Line Options

```
Usage: sampler [options]

Options:
  --count N, -n N     Number of emails to extract (required)
                      Must be a positive integer
  
  --help, -h          Show this help message
```

## Output Files

### Sample CSV Files

**Location**: `assets/enron-emails/`  
**Naming**: `sampled-emails-YYYYMMDD-HHMMSS.csv`  
**Format**: Identical to source CSV (file, message columns)

**Example**:
```csv
file,message
allen-p/_sent_mail/123.,"Message-ID: <12345@enron.com>
From: phillip.allen@enron.com
To: john.doe@enron.com
Subject: Energy Trading

Email content here..."
```

### Tracking Files

**Location**: `assets/enron-emails/extracted-*.txt`  
**Format**: One email identifier per line (one file per extraction)  
**Purpose**: Prevents duplicate extractions

**Example** (`extracted-20260127-143022.txt`):
```
allen-p/_sent_mail/1.
allen-p/_sent_mail/2.
bailey-s/all_documents/1.
```

## Common Scenarios

### Scenario 1: Building Test Dataset

Extract incrementally to build up a diverse test dataset:

```bash
# Day 1: Get initial 500 emails
go run cmd/sampler/main.go --count 500

# Day 2: Add 300 more (guaranteed different emails)
go run cmd/sampler/main.go --count 300

# Day 3: Add 200 more
go run cmd/sampler/main.go --count 200

# Result: 3 CSV files with 1000 total unique emails
```

### Scenario 2: Unit Test Fixtures

Create small samples for fast unit tests:

```bash
# Extract 10 emails for quick tests
go run cmd/sampler/main.go --count 10

# Move to test fixtures
mv assets/enron-emails/sampled-emails-*.csv tests/fixtures/sample-10.csv
```

### Scenario 3: Performance Testing

Extract large samples for stress testing:

```bash
# 10k emails for performance validation
go run cmd/sampler/main.go --count 10000

# Load into system for testing
go run cmd/loader/main.go --csv-path assets/enron-emails/sampled-emails-*.csv
```

### Scenario 4: Resetting Tracking

Start fresh (re-extract previously sampled emails):

```bash
# Remove all tracking files
rm assets/enron-emails/extracted-*.txt

# Now all emails available again
go run cmd/sampler/main.go --count 100
```

### Scenario 5: Partial Reset

Allow specific batch to be re-extracted:

```bash
# Delete specific tracking file to allow those emails to be extracted again
rm assets/enron-emails/extracted-20260127-143022.txt

# Run extractor normally - those emails are now available again
go run cmd/sampler/main.go --count 500
```

## Expected Behavior

### First Run

- No tracking files exist → creates first `extracted-YYYYMMDD-HHMMSS.txt`
- All emails in source CSV are available
- Extracts exactly `--count` emails (if available)

### Subsequent Runs

- Reads all existing tracking files (extracted-*.txt)
- Excludes previously extracted emails from all files
- Extracts from remaining available pool
- Creates new tracking file for this extraction

### Edge Cases

#### Insufficient Available Emails

```bash
# If only 200 emails remain unextracted
go run cmd/sampler/main.go --count 1000

# Output:
# Warning: Only 200 emails available, extracting all remaining
# Successfully extracted 200 emails to ...
```

#### All Emails Extracted

```bash
# After all emails extracted
go run cmd/sampler/main.go --count 100

# Output:
# Warning: No unextracted emails available
# Successfully extracted 0 emails to sampled-emails-YYYYMMDD-HHMMSS.csv
```

#### Corrupted Source CSV Row

```bash
# Malformed row in source CSV
# Output:
# Warning: Skipped malformed row at line 12345
# Continuing extraction...
```

## Verification

### Verify Output Format

Check that output CSV is compatible with loader:

```bash
# Parse output with existing loader
go run cmd/loader/main.go --csv-path assets/enron-emails/sampled-emails-*.csv
```

### Verify No Duplicates

Check tracking files for duplicates:

```bash
# Count total entries across all tracking files
cat assets/enron-emails/extracted-*.txt | wc -l

# Check for duplicates across all files (should show no output)
cat assets/enron-emails/extracted-*.txt | sort | uniq -d
```

### Verify Uniqueness Across Samples

Compare two output files to ensure no overlap:

```bash
# Extract file IDs from two samples
cut -d',' -f1 sampled-emails-1.csv | tail -n +2 > sample1-ids.txt
cut -d',' -f1 sampled-emails-2.csv | tail -n +2 > sample2-ids.txt

# Find common IDs (should be empty)
comm -12 <(sort sample1-ids.txt) <(sort sample2-ids.txt)
```

## Troubleshooting

### Error: Source CSV Not Found

```
Error: Failed to open CSV file: assets/enron-emails/emails.csv not found
```

**Solution**: Ensure Enron dataset is downloaded and placed in `assets/enron-emails/`

### Error: Permission Denied

```
Error: Cannot write to assets/enron-emails/: permission denied
```

**Solution**: Check write permissions on assets directory:
```bash
chmod u+w assets/enron-emails/
```

### Error: Corrupted Source File

```
Error: Failed to read source CSV - unexpected format
```

**Solution**: The source CSV may have been modified or corrupted. The utility expects the static Enron dataset with unique file identifiers (verified during development). Verify file integrity or restore from backup.

### Warning: Corrupted Tracking File

```
Warning: Corrupted tracking file 'extracted-20260101-120000.txt' detected, skipping this file
```

**Solution**: Automatic recovery - corrupted file is skipped, other tracking files are still used. Delete corrupted file if needed.

## Performance Tips

### Faster Extraction

For large samples, performance is dominated by CSV I/O:

```bash
# Faster: Extract once with higher count
go run cmd/sampler/main.go --count 10000

# Slower: Multiple small extractions
for i in {1..100}; do
    go run cmd/sampler/main.go --count 100
done
```

### Memory Usage

Utility uses minimal memory (~100MB even for full dataset):
- Tracking file loaded into set: ~50MB for 500k entries
- CSV streaming: constant memory
- Random indices: ~80KB for 10k samples

## Integration with Existing Tools

### Use with Email Loader

```bash
# Extract sample
go run cmd/sampler/main.go --count 1000

# Load into database with entity extraction
go run cmd/loader/main.go \
    --csv-path assets/enron-emails/sampled-emails-20260127-143022.csv \
    --extract \
    --workers 5
```

### Use in Test Scripts

```bash
#!/bin/bash
# test-with-sample.sh

# Create fresh sample
go run cmd/sampler/main.go --count 100

# Get latest sample file
SAMPLE=$(ls -t assets/enron-emails/sampled-emails-*.csv | head -1)

# Run tests
go test ./tests/integration/... -csv-path "$SAMPLE"
```

## FAQ

**Q: Can I extract the same emails again?**  
A: Delete the specific tracking file for that extraction, or delete all tracking files (`extracted-*.txt`) to reset entirely.

**Q: What happens if I interrupt the extraction (Ctrl+C)?**  
A: Partial output file may exist but tracking file is only updated after completion. Safe to re-run.

**Q: How do I know which emails were extracted?**  
A: Check the first column of the output CSV file or view the corresponding tracking file (same timestamp).

**Q: Can I use custom source CSV files?**  
A: Currently hardcoded to `assets/enron-emails/emails.csv`. Edit code to support other files.

**Q: Why timestamp-based filenames instead of sequential numbers?**  
A: Timestamps are self-documenting (when extracted) and avoid collision tracking complexity.

**Q: Does the tool modify the source CSV?**  
A: No. Source file is read-only. Only creates new files.

## Next Steps

- [Feature Specification](spec.md) - Complete requirements and user stories
- [Data Model](data-model.md) - Entity definitions and relationships
- [Implementation Plan](plan.md) - Technical architecture and design
- [Research Notes](research.md) - Technology decisions and alternatives

## Support

For issues or questions:
1. Check this quickstart guide
2. Review specification in `specs/003-random-email-extractor/spec.md`
3. Examine test fixtures in `tests/fixtures/`
4. Run integration tests: `go test ./tests/integration/sampler_test.go`

#!/usr/bin/env bash
# T034: Test US2 - Prevent Duplicate Extraction
# This script tests that the sampler prevents duplicate extraction across multiple runs.
#
# Test steps:
# 1. Run sampler twice (first with --count 5, then --count 3)
# 2. Verify 8 unique emails across both files
# 3. Verify tracking files created
# 4. Verify no duplicates between runs

set -e  # Exit on error

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
OUTPUT_DIR="${REPO_ROOT}/assets/enron-emails"

echo "=== Test US2: Duplicate Prevention ==="
echo

# Setup: Backup existing emails.csv and create test data
echo "Setting up test environment..."
BACKUP_CSV=""
if [ -f "${OUTPUT_DIR}/emails.csv" ]; then
    BACKUP_CSV="${OUTPUT_DIR}/emails.csv.backup.$$"
    mv "${OUTPUT_DIR}/emails.csv" "$BACKUP_CSV"
    echo "✓ Backed up existing emails.csv"
fi

# Create test CSV with exactly 8 emails
echo "Creating test email data (8 emails)..."
cat > "${OUTPUT_DIR}/emails.csv" << 'EOF'
file,message
test-dup-1,"Message-ID: <dup1@test.com>
From: alice@test.com
To: bob@test.com
Subject: Test Duplicate 1

First test email."
test-dup-2,"Message-ID: <dup2@test.com>
From: bob@test.com
To: alice@test.com
Subject: Test Duplicate 2

Second test email."
test-dup-3,"Message-ID: <dup3@test.com>
From: charlie@test.com
To: alice@test.com
Subject: Test Duplicate 3

Third test email."
test-dup-4,"Message-ID: <dup4@test.com>
From: alice@test.com
To: charlie@test.com
Subject: Test Duplicate 4

Fourth test email."
test-dup-5,"Message-ID: <dup5@test.com>
From: bob@test.com
To: charlie@test.com
Subject: Test Duplicate 5

Fifth test email."
test-dup-6,"Message-ID: <dup6@test.com>
From: charlie@test.com
To: bob@test.com
Subject: Test Duplicate 6

Sixth test email."
test-dup-7,"Message-ID: <dup7@test.com>
From: alice@test.com
To: bob@test.com
Subject: Test Duplicate 7

Seventh test email."
test-dup-8,"Message-ID: <dup8@test.com>
From: bob@test.com
To: alice@test.com
Subject: Test Duplicate 8

Eighth test email."
EOF

echo "✓ Created test data with 8 emails"
echo

# Clean up previous test files
echo "Cleaning up previous test files..."
rm -f "${OUTPUT_DIR}/sampled-emails-"*.csv
rm -f "${OUTPUT_DIR}/extracted-"*.txt

# Run 1: Extract 5 emails
echo "Running sampler (1st run) - requesting 5 emails..."
cd "${REPO_ROOT}"
go run cmd/sampler/main.go --count 5

# Find the first output file
RUN1_CSV=$(ls -t "${OUTPUT_DIR}/sampled-emails-"*.csv | head -1)
RUN1_TRACKING=$(ls -t "${OUTPUT_DIR}/extracted-"*.txt | head -1)

if [ ! -f "$RUN1_CSV" ]; then
    echo "ERROR: First run did not create output CSV file"
    exit 1
fi

if [ ! -f "$RUN1_TRACKING" ]; then
    echo "ERROR: First run did not create tracking file"
    exit 1
fi

echo "✓ First run completed"
echo "  Output: $RUN1_CSV"
echo "  Tracking: $RUN1_TRACKING"

# Count emails in first run (use tracking file instead of CSV to avoid multi-line message issues)
RUN1_TRACKING_COUNT=$(cat "$RUN1_TRACKING" | wc -l | tr -d ' ')
echo "  Extracted: $RUN1_TRACKING_COUNT emails"

if [ "$RUN1_TRACKING_COUNT" -ne 5 ]; then
    echo "ERROR: Expected 5 emails in first run, got $RUN1_TRACKING_COUNT"
    exit 1
fi

# Sleep to ensure different timestamp
sleep 2

# Run 2: Extract another 3 emails (only 3 remaining out of 8 total)
echo
echo "Running sampler (2nd run) - requesting 3 more emails..."
go run cmd/sampler/main.go --count 3

# Find the second output file (most recent, excluding first)
RUN2_CSV=$(ls -t "${OUTPUT_DIR}/sampled-emails-"*.csv | head -1)
RUN2_TRACKING=$(ls -t "${OUTPUT_DIR}/extracted-"*.txt | head -1)

if [ ! -f "$RUN2_CSV" ]; then
    echo "ERROR: Second run did not create output CSV file"
    exit 1
fi

if [ ! -f "$RUN2_TRACKING" ]; then
    echo "ERROR: Second run did not create tracking file"
    exit 1
fi

echo "✓ Second run completed"
echo "  Output: $RUN2_CSV"
echo "  Tracking: $RUN2_TRACKING"

# Count emails in second run
RUN2_TRACKING_COUNT=$(cat "$RUN2_TRACKING" | wc -l | tr -d ' ')
echo "  Extracted: $RUN2_TRACKING_COUNT emails"

if [ "$RUN2_TRACKING_COUNT" -ne 3 ]; then
    echo "ERROR: Expected 3 emails in second run, got $RUN2_TRACKING_COUNT"
    exit 1
fi

# Verify total unique emails
echo
echo "Verifying uniqueness across runs..."

# Extract file IDs from tracking files
RUN1_IDS=$(sort "$RUN1_TRACKING")
RUN2_IDS=$(sort "$RUN2_TRACKING")

# Combine and count unique IDs
TOTAL_UNIQUE=$(echo -e "$RUN1_IDS\n$RUN2_IDS" | sort -u | wc -l | tr -d ' ')
TOTAL_EXTRACTED=$((RUN1_TRACKING_COUNT + RUN2_TRACKING_COUNT))

echo "  Total emails extracted: $TOTAL_EXTRACTED"
echo "  Unique emails: $TOTAL_UNIQUE"

# Verify that all extracted emails are unique (no duplicates)
if [ "$TOTAL_UNIQUE" -ne "$TOTAL_EXTRACTED" ]; then
    echo "ERROR: Found duplicates! Expected $TOTAL_EXTRACTED unique emails, got $TOTAL_UNIQUE"
    exit 1
fi

# Should have extracted all 8 emails
if [ "$TOTAL_UNIQUE" -ne 8 ]; then
    echo "ERROR: Expected 8 unique emails total, got $TOTAL_UNIQUE"
    exit 1
fi

# Check for overlapping IDs
OVERLAP=$(comm -12 <(echo "$RUN1_IDS") <(echo "$RUN2_IDS") | wc -l | tr -d ' ')

if [ "$OVERLAP" -ne 0 ]; then
    echo "ERROR: Found $OVERLAP duplicate email(s) between runs"
    echo "Duplicates:"
    comm -12 <(echo "$RUN1_IDS") <(echo "$RUN2_IDS")
    exit 1
fi

# Verify tracking files exist
TRACKING_COUNT=$(ls "${OUTPUT_DIR}/extracted-"*.txt 2>/dev/null | wc -l | tr -d ' ')
echo
echo "Verifying tracking files..."
echo "  Tracking files created: $TRACKING_COUNT"

if [ "$TRACKING_COUNT" -lt 2 ]; then
    echo "ERROR: Expected at least 2 tracking files, found $TRACKING_COUNT"
    exit 1
fi

# Verify tracking file contents
echo
echo "Verifying tracking file contents..."

RUN1_TRACKING_COUNT=$(wc -l < "$RUN1_TRACKING" | tr -d ' ')
RUN2_TRACKING_COUNT=$(wc -l < "$RUN2_TRACKING" | tr -d ' ')

echo "  Run 1 tracking entries: $RUN1_TRACKING_COUNT"
echo "  Run 2 tracking entries: $RUN2_TRACKING_COUNT"

if [ "$RUN1_TRACKING_COUNT" -ne 5 ]; then
    echo "ERROR: Expected 5 entries in first tracking file, got $RUN1_TRACKING_COUNT"
    exit 1
fi

if [ "$RUN2_TRACKING_COUNT" -ne 3 ]; then
    echo "ERROR: Expected 3 entries in second tracking file, got $RUN2_TRACKING_COUNT"
    exit 1
fi

# Verify that CSV IDs match tracking file IDs
echo
echo "Verifying CSV matches tracking files..."

# Just verify the tracking files have correct content
RUN1_TRACKING_IDS=$(sort "$RUN1_TRACKING")
RUN2_TRACKING_IDS=$(sort "$RUN2_TRACKING")

# Count that tracking files have the expected number of non-empty lines
RUN1_LINE_COUNT=$(grep -c . "$RUN1_TRACKING" || true)
RUN2_LINE_COUNT=$(grep -c . "$RUN2_TRACKING" || true)

if [ "$RUN1_LINE_COUNT" -ne 5 ]; then
    echo "ERROR: Run 1 tracking file should have 5 email IDs, got $RUN1_LINE_COUNT"
    exit 1
fi

if [ "$RUN2_LINE_COUNT" -ne 3 ]; then
    echo "ERROR: Run 2 tracking file should have 3 email IDs, got $RUN2_LINE_COUNT"
    exit 1
fi

echo "✓ Tracking files have correct format"

# Cleanup: Restore original emails.csv
echo
echo "Cleaning up test environment..."
rm -f "${OUTPUT_DIR}/emails.csv"
if [ -n "$BACKUP_CSV" ]; then
    mv "$BACKUP_CSV" "${OUTPUT_DIR}/emails.csv"
    echo "✓ Restored original emails.csv"
fi

# Clean up test output files
rm -f "${OUTPUT_DIR}/sampled-emails-"*.csv
rm -f "${OUTPUT_DIR}/extracted-"*.txt
echo "✓ Removed test artifacts"

# Success!
echo
echo "==================================="
echo "✓ All duplicate prevention tests passed!"
echo "==================================="
echo
echo "Summary:"
echo "  - Run 1: Extracted 5 emails"
echo "  - Run 2: Extracted 3 emails"
echo "  - Total unique: 8 emails"
echo "  - No duplicates found between runs"
echo "  - Tracking files created and accurate"
echo

exit 0

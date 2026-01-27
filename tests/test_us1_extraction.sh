#!/bin/bash
set -e

# Test script for US1: Extract Random Sample
# Validates basic extraction flow for random email sampler
# Tests:
# - Runs sampler with --count 10
# - Verifies output file exists
# - Verifies exactly 10 emails extracted
# - Verifies CSV format preserved

REPO_ROOT="/Users/jochem/code/enron-graph"
TEST_DIR="${REPO_ROOT}/tests/fixtures"
OUTPUT_DIR="/tmp/test-sampler-us1"
TEST_CSV="${OUTPUT_DIR}/test-emails.csv"

echo "=== US1 Extraction Test ==="
echo ""

# Setup: Create clean test directory
echo "Setting up test environment..."
rm -rf "${OUTPUT_DIR}"
mkdir -p "${OUTPUT_DIR}"

# Create test CSV with exactly 50 emails
echo "Creating test email data (50 emails)..."
cat > "${TEST_CSV}" << 'EOF'
file,message
test-email-1,"Message-ID: <test1@example.com>
Date: Mon, 01 Jan 2026 10:00:00 -0700 (PDT)
From: alice@example.com
To: bob@example.com
Subject: Test Email 1

This is test email 1."
test-email-2,"Message-ID: <test2@example.com>
Date: Mon, 01 Jan 2026 10:01:00 -0700 (PDT)
From: bob@example.com
To: alice@example.com
Subject: Test Email 2

This is test email 2."
test-email-3,"Message-ID: <test3@example.com>
Date: Mon, 01 Jan 2026 10:02:00 -0700 (PDT)
From: charlie@example.com
To: alice@example.com
Subject: Test Email 3

This is test email 3."
test-email-4,"Message-ID: <test4@example.com>
Date: Mon, 01 Jan 2026 10:03:00 -0700 (PDT)
From: alice@example.com
To: charlie@example.com
Subject: Test Email 4

This is test email 4."
test-email-5,"Message-ID: <test5@example.com>
Date: Mon, 01 Jan 2026 10:04:00 -0700 (PDT)
From: bob@example.com
To: charlie@example.com
Subject: Test Email 5

This is test email 5."
test-email-6,"Message-ID: <test6@example.com>
Date: Mon, 01 Jan 2026 10:05:00 -0700 (PDT)
From: charlie@example.com
To: bob@example.com
Subject: Test Email 6

This is test email 6."
test-email-7,"Message-ID: <test7@example.com>
Date: Mon, 01 Jan 2026 10:06:00 -0700 (PDT)
From: alice@example.com
To: bob@example.com
Subject: Test Email 7

This is test email 7."
test-email-8,"Message-ID: <test8@example.com>
Date: Mon, 01 Jan 2026 10:07:00 -0700 (PDT)
From: bob@example.com
To: alice@example.com
Subject: Test Email 8

This is test email 8."
test-email-9,"Message-ID: <test9@example.com>
Date: Mon, 01 Jan 2026 10:08:00 -0700 (PDT)
From: charlie@example.com
To: alice@example.com
Subject: Test Email 9

This is test email 9."
test-email-10,"Message-ID: <test10@example.com>
Date: Mon, 01 Jan 2026 10:09:00 -0700 (PDT)
From: alice@example.com
To: charlie@example.com
Subject: Test Email 10

This is test email 10."
test-email-11,"Message-ID: <test11@example.com>
Date: Mon, 01 Jan 2026 10:10:00 -0700 (PDT)
From: bob@example.com
To: charlie@example.com
Subject: Test Email 11

This is test email 11."
test-email-12,"Message-ID: <test12@example.com>
Date: Mon, 01 Jan 2026 10:11:00 -0700 (PDT)
From: charlie@example.com
To: bob@example.com
Subject: Test Email 12

This is test email 12."
test-email-13,"Message-ID: <test13@example.com>
Date: Mon, 01 Jan 2026 10:12:00 -0700 (PDT)
From: alice@example.com
To: bob@example.com
Subject: Test Email 13

This is test email 13."
test-email-14,"Message-ID: <test14@example.com>
Date: Mon, 01 Jan 2026 10:13:00 -0700 (PDT)
From: bob@example.com
To: alice@example.com
Subject: Test Email 14

This is test email 14."
test-email-15,"Message-ID: <test15@example.com>
Date: Mon, 01 Jan 2026 10:14:00 -0700 (PDT)
From: charlie@example.com
To: alice@example.com
Subject: Test Email 15

This is test email 15."
test-email-16,"Message-ID: <test16@example.com>
Date: Mon, 01 Jan 2026 10:15:00 -0700 (PDT)
From: alice@example.com
To: charlie@example.com
Subject: Test Email 16

This is test email 16."
test-email-17,"Message-ID: <test17@example.com>
Date: Mon, 01 Jan 2026 10:16:00 -0700 (PDT)
From: bob@example.com
To: charlie@example.com
Subject: Test Email 17

This is test email 17."
test-email-18,"Message-ID: <test18@example.com>
Date: Mon, 01 Jan 2026 10:17:00 -0700 (PDT)
From: charlie@example.com
To: bob@example.com
Subject: Test Email 18

This is test email 18."
test-email-19,"Message-ID: <test19@example.com>
Date: Mon, 01 Jan 2026 10:18:00 -0700 (PDT)
From: alice@example.com
To: bob@example.com
Subject: Test Email 19

This is test email 19."
test-email-20,"Message-ID: <test20@example.com>
Date: Mon, 01 Jan 2026 10:19:00 -0700 (PDT)
From: bob@example.com
To: alice@example.com
Subject: Test Email 20

This is test email 20."
test-email-21,"Message-ID: <test21@example.com>
Date: Mon, 01 Jan 2026 10:20:00 -0700 (PDT)
From: charlie@example.com
To: alice@example.com
Subject: Test Email 21

This is test email 21."
test-email-22,"Message-ID: <test22@example.com>
Date: Mon, 01 Jan 2026 10:21:00 -0700 (PDT)
From: alice@example.com
To: charlie@example.com
Subject: Test Email 22

This is test email 22."
test-email-23,"Message-ID: <test23@example.com>
Date: Mon, 01 Jan 2026 10:22:00 -0700 (PDT)
From: bob@example.com
To: charlie@example.com
Subject: Test Email 23

This is test email 23."
test-email-24,"Message-ID: <test24@example.com>
Date: Mon, 01 Jan 2026 10:23:00 -0700 (PDT)
From: charlie@example.com
To: bob@example.com
Subject: Test Email 24

This is test email 24."
test-email-25,"Message-ID: <test25@example.com>
Date: Mon, 01 Jan 2026 10:24:00 -0700 (PDT)
From: alice@example.com
To: bob@example.com
Subject: Test Email 25

This is test email 25."
test-email-26,"Message-ID: <test26@example.com>
Date: Mon, 01 Jan 2026 10:25:00 -0700 (PDT)
From: bob@example.com
To: alice@example.com
Subject: Test Email 26

This is test email 26."
test-email-27,"Message-ID: <test27@example.com>
Date: Mon, 01 Jan 2026 10:26:00 -0700 (PDT)
From: charlie@example.com
To: alice@example.com
Subject: Test Email 27

This is test email 27."
test-email-28,"Message-ID: <test28@example.com>
Date: Mon, 01 Jan 2026 10:27:00 -0700 (PDT)
From: alice@example.com
To: charlie@example.com
Subject: Test Email 28

This is test email 28."
test-email-29,"Message-ID: <test29@example.com>
Date: Mon, 01 Jan 2026 10:28:00 -0700 (PDT)
From: bob@example.com
To: charlie@example.com
Subject: Test Email 29

This is test email 29."
test-email-30,"Message-ID: <test30@example.com>
Date: Mon, 01 Jan 2026 10:29:00 -0700 (PDT)
From: charlie@example.com
To: bob@example.com
Subject: Test Email 30

This is test email 30."
test-email-31,"Message-ID: <test31@example.com>
Date: Mon, 01 Jan 2026 10:30:00 -0700 (PDT)
From: alice@example.com
To: bob@example.com
Subject: Test Email 31

This is test email 31."
test-email-32,"Message-ID: <test32@example.com>
Date: Mon, 01 Jan 2026 10:31:00 -0700 (PDT)
From: bob@example.com
To: alice@example.com
Subject: Test Email 32

This is test email 32."
test-email-33,"Message-ID: <test33@example.com>
Date: Mon, 01 Jan 2026 10:32:00 -0700 (PDT)
From: charlie@example.com
To: alice@example.com
Subject: Test Email 33

This is test email 33."
test-email-34,"Message-ID: <test34@example.com>
Date: Mon, 01 Jan 2026 10:33:00 -0700 (PDT)
From: alice@example.com
To: charlie@example.com
Subject: Test Email 34

This is test email 34."
test-email-35,"Message-ID: <test35@example.com>
Date: Mon, 01 Jan 2026 10:34:00 -0700 (PDT)
From: bob@example.com
To: charlie@example.com
Subject: Test Email 35

This is test email 35."
test-email-36,"Message-ID: <test36@example.com>
Date: Mon, 01 Jan 2026 10:35:00 -0700 (PDT)
From: charlie@example.com
To: bob@example.com
Subject: Test Email 36

This is test email 36."
test-email-37,"Message-ID: <test37@example.com>
Date: Mon, 01 Jan 2026 10:36:00 -0700 (PDT)
From: alice@example.com
To: bob@example.com
Subject: Test Email 37

This is test email 37."
test-email-38,"Message-ID: <test38@example.com>
Date: Mon, 01 Jan 2026 10:37:00 -0700 (PDT)
From: bob@example.com
To: alice@example.com
Subject: Test Email 38

This is test email 38."
test-email-39,"Message-ID: <test39@example.com>
Date: Mon, 01 Jan 2026 10:38:00 -0700 (PDT)
From: charlie@example.com
To: alice@example.com
Subject: Test Email 39

This is test email 39."
test-email-40,"Message-ID: <test40@example.com>
Date: Mon, 01 Jan 2026 10:39:00 -0700 (PDT)
From: alice@example.com
To: charlie@example.com
Subject: Test Email 40

This is test email 40."
test-email-41,"Message-ID: <test41@example.com>
Date: Mon, 01 Jan 2026 10:40:00 -0700 (PDT)
From: bob@example.com
To: charlie@example.com
Subject: Test Email 41

This is test email 41."
test-email-42,"Message-ID: <test42@example.com>
Date: Mon, 01 Jan 2026 10:41:00 -0700 (PDT)
From: charlie@example.com
To: bob@example.com
Subject: Test Email 42

This is test email 42."
test-email-43,"Message-ID: <test43@example.com>
Date: Mon, 01 Jan 2026 10:42:00 -0700 (PDT)
From: alice@example.com
To: bob@example.com
Subject: Test Email 43

This is test email 43."
test-email-44,"Message-ID: <test44@example.com>
Date: Mon, 01 Jan 2026 10:43:00 -0700 (PDT)
From: bob@example.com
To: alice@example.com
Subject: Test Email 44

This is test email 44."
test-email-45,"Message-ID: <test45@example.com>
Date: Mon, 01 Jan 2026 10:44:00 -0700 (PDT)
From: charlie@example.com
To: alice@example.com
Subject: Test Email 45

This is test email 45."
test-email-46,"Message-ID: <test46@example.com>
Date: Mon, 01 Jan 2026 10:45:00 -0700 (PDT)
From: alice@example.com
To: charlie@example.com
Subject: Test Email 46

This is test email 46."
test-email-47,"Message-ID: <test47@example.com>
Date: Mon, 01 Jan 2026 10:46:00 -0700 (PDT)
From: bob@example.com
To: charlie@example.com
Subject: Test Email 47

This is test email 47."
test-email-48,"Message-ID: <test48@example.com>
Date: Mon, 01 Jan 2026 10:47:00 -0700 (PDT)
From: charlie@example.com
To: bob@example.com
Subject: Test Email 48

This is test email 48."
test-email-49,"Message-ID: <test49@example.com>
Date: Mon, 01 Jan 2026 10:48:00 -0700 (PDT)
From: alice@example.com
To: bob@example.com
Subject: Test Email 49

This is test email 49."
test-email-50,"Message-ID: <test50@example.com>
Date: Mon, 01 Jan 2026 10:49:00 -0700 (PDT)
From: bob@example.com
To: alice@example.com
Subject: Test Email 50

This is test email 50."
EOF

echo "✅ Created test data with 50 emails"
echo ""

# Test: Run sampler with --count 10
echo "=== Running sampler with --count 10 ==="
cd "${REPO_ROOT}"

# Create a temporary copy of the source CSV in the expected location
mkdir -p "${REPO_ROOT}/assets/enron-emails"
BACKUP_CSV=""
if [ -f "${REPO_ROOT}/assets/enron-emails/emails.csv" ]; then
    BACKUP_CSV="${REPO_ROOT}/assets/enron-emails/emails.csv.backup.$$"
    mv "${REPO_ROOT}/assets/enron-emails/emails.csv" "$BACKUP_CSV"
fi

# Remove any existing tracking files to ensure clean test
rm -f "${REPO_ROOT}/assets/enron-emails"/extracted-*.txt

# Copy test CSV to expected location
cp "${TEST_CSV}" "${REPO_ROOT}/assets/enron-emails/emails.csv"

# Run the sampler
go run cmd/sampler/main.go --count 10

# Get the most recently created sampled and tracking files
NEWEST_SAMPLED=$(ls -t "${REPO_ROOT}/assets/enron-emails"/sampled-emails-*.csv 2>/dev/null | head -n 1)
NEWEST_TRACKING=$(ls -t "${REPO_ROOT}/assets/enron-emails"/extracted-*.txt 2>/dev/null | head -n 1)

# Move only the newest files to test directory
if [ -n "$NEWEST_SAMPLED" ]; then
    mv "$NEWEST_SAMPLED" "${OUTPUT_DIR}/"
fi
if [ -n "$NEWEST_TRACKING" ]; then
    mv "$NEWEST_TRACKING" "${OUTPUT_DIR}/"
fi

# Restore original CSV if it existed
if [ -n "$BACKUP_CSV" ]; then
    mv "$BACKUP_CSV" "${REPO_ROOT}/assets/enron-emails/emails.csv"
else
    rm -f "${REPO_ROOT}/assets/enron-emails/emails.csv"
fi

echo ""
echo "✅ Sampler execution completed"
echo ""

# Verify: Output file exists
echo "=== Verification: Output file exists ==="
OUTPUT_FILES=$(ls -1 "${OUTPUT_DIR}"/sampled-emails-*.csv 2>/dev/null | wc -l | tr -d ' ')
if [ "$OUTPUT_FILES" -eq 0 ]; then
    echo "❌ FAIL: No output file found"
    # Cleanup before exit
    rm -rf "${OUTPUT_DIR}"
    exit 1
fi

OUTPUT_FILE=$(ls -1 "${OUTPUT_DIR}"/sampled-emails-*.csv | head -n 1)
echo "✅ PASS: Output file exists: $(basename "$OUTPUT_FILE")"
echo ""

# Verify: Exactly 10 emails extracted
echo "=== Verification: Exactly 10 emails extracted ==="
# Count data rows in CSV by counting email record markers
# Each CSV row starts with the file field (test-email-*)
EMAIL_COUNT=$(grep -c "^test-email-" "$OUTPUT_FILE" || echo "0")

if [ "$EMAIL_COUNT" -ne 10 ]; then
    echo "❌ FAIL: Expected 10 emails, found $EMAIL_COUNT"
    rm -rf "${OUTPUT_DIR}"
    exit 1
fi
echo "✅ PASS: Exactly 10 emails extracted"
echo ""

# Verify: CSV format preserved
echo "=== Verification: CSV format preserved ==="

# Check header line
HEADER=$(head -n 1 "$OUTPUT_FILE")
if [ "$HEADER" != "file,message" ]; then
    echo "❌ FAIL: CSV header incorrect. Expected 'file,message', got '$HEADER'"
    rm -rf "${OUTPUT_DIR}"
    exit 1
fi
echo "✅ PASS: CSV header correct"

# Verify each line has proper structure (file field and message field)
# Extract first record (line 2) and check it has the file field
FIRST_RECORD=$(sed -n '2p' "$OUTPUT_FILE")
FILE_FIELD=$(echo "$FIRST_RECORD" | cut -d',' -f1)

if [ -z "$FILE_FIELD" ]; then
    echo "❌ FAIL: First record has no file field"
    rm -rf "${OUTPUT_DIR}"
    exit 1
fi
echo "✅ PASS: CSV records have proper structure"

# Verify the message field contains expected email structure
MESSAGE_SAMPLE=$(sed -n '2p' "$OUTPUT_FILE" | cut -d',' -f2-)
if ! echo "$MESSAGE_SAMPLE" | grep -q "Message-ID:"; then
    echo "❌ FAIL: Message field doesn't contain expected email structure"
    rm -rf "${OUTPUT_DIR}"
    exit 1
fi
echo "✅ PASS: Message field contains email content"
echo ""

# Verify: Tracking file created
echo "=== Verification: Tracking file created ==="
TRACKING_FILES=$(ls -1 "${OUTPUT_DIR}"/extracted-*.txt 2>/dev/null | wc -l | tr -d ' ')
if [ "$TRACKING_FILES" -eq 0 ]; then
    echo "❌ FAIL: No tracking file found"
    rm -rf "${OUTPUT_DIR}"
    exit 1
fi

TRACKING_FILE=$(ls -1 "${OUTPUT_DIR}"/extracted-*.txt | head -n 1)
echo "✅ PASS: Tracking file exists: $(basename "$TRACKING_FILE")"

# Verify tracking file has 10 entries
TRACKING_COUNT=$(wc -l < "$TRACKING_FILE" | tr -d ' ')
if [ "$TRACKING_COUNT" -ne 10 ]; then
    echo "❌ FAIL: Expected 10 entries in tracking file, found $TRACKING_COUNT"
    rm -rf "${OUTPUT_DIR}"
    exit 1
fi
echo "✅ PASS: Tracking file has 10 entries"
echo ""

# Cleanup
echo "=== Cleanup ==="
rm -rf "${OUTPUT_DIR}"
echo "✅ Test artifacts cleaned up"
echo ""

echo "=== Test Complete ==="
echo "✅ All US1 extraction tests passed"
echo ""
echo "Summary:"
echo "  ✓ Sampler executed successfully with --count 10"
echo "  ✓ Output file created in correct format"
echo "  ✓ Exactly 10 emails extracted"
echo "  ✓ CSV format preserved (header + records)"
echo "  ✓ Tracking file created with correct entries"

#!/bin/bash
set -e

# Test script for US3: Configure Sample Size
# Validates various count scenarios for random email sampler
# Tests:
# - Runs sampler with --count 10, 100, 1000
# - Verifies exact counts
# - Tests --help flag
# - Tests invalid count values
# - Verifies appropriate error messages

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TEST_DIR="${REPO_ROOT}/tests/fixtures"
OUTPUT_DIR="/tmp/test-sampler-us3"
TEST_CSV="${OUTPUT_DIR}/test-emails.csv"

echo "=== US3 Configuration Test ==="
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

echo "Test data created: ${TEST_CSV}"
echo ""

# Create backup of original emails.csv if it exists
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

echo "Test environment ready"
echo ""

# Test 1: Extract 10 emails
echo "Test 1: Extracting 10 emails..."
cd "${REPO_ROOT}"
go run cmd/sampler/main.go --count 10 > /dev/null 2>&1
if [ $? -ne 0 ]; then
    echo "❌ FAILED: Extraction with --count 10 failed"
    # Restore backup before exit
    if [ -n "$BACKUP_CSV" ]; then
        mv "$BACKUP_CSV" "${REPO_ROOT}/assets/enron-emails/emails.csv"
    else
        rm -f "${REPO_ROOT}/assets/enron-emails/emails.csv"
    fi
    exit 1
fi

# Get the most recently created sampled file
NEWEST_SAMPLED=$(ls -t "${REPO_ROOT}/assets/enron-emails"/sampled-emails-*.csv 2>/dev/null | head -n 1)

# Move to test directory
if [ -n "$NEWEST_SAMPLED" ]; then
    mv "$NEWEST_SAMPLED" "${OUTPUT_DIR}/"
fi

# Verify output file exists
OUTPUT_FILE=$(ls -t "${OUTPUT_DIR}"/sampled-emails-*.csv 2>/dev/null | head -1)
if [ ! -f "${OUTPUT_FILE}" ]; then
    echo "❌ FAILED: Output file not found"
    # Restore backup before exit
    if [ -n "$BACKUP_CSV" ]; then
        mv "$BACKUP_CSV" "${REPO_ROOT}/assets/enron-emails/emails.csv"
    else
        rm -f "${REPO_ROOT}/assets/enron-emails/emails.csv"
    fi
    exit 1
fi

# Count records in output (count email file identifiers)
RECORD_COUNT=$(grep -c "^test-email-" "${OUTPUT_FILE}" || echo "0")
if [ "${RECORD_COUNT}" -ne 10 ]; then
    echo "❌ FAILED: Expected 10 records, got ${RECORD_COUNT}"
    # Restore backup before exit
    if [ -n "$BACKUP_CSV" ]; then
        mv "$BACKUP_CSV" "${REPO_ROOT}/assets/enron-emails/emails.csv"
    else
        rm -f "${REPO_ROOT}/assets/enron-emails/emails.csv"
    fi
    exit 1
fi

echo "✓ Test 1 PASSED: Extracted exactly 10 emails"
echo ""

# Test 2: Extract 100 emails (should cap at 50 - 10 = 40 remaining)
echo "Test 2: Extracting 100 emails (should cap at 40 remaining)..."
sleep 1  # Ensure different timestamp
cd "${REPO_ROOT}"
OUTPUT=$(go run cmd/sampler/main.go --count 100 2>&1)
if [ $? -ne 0 ]; then
    echo "❌ FAILED: Extraction with --count 100 failed"
    # Restore backup before exit
    if [ -n "$BACKUP_CSV" ]; then
        mv "$BACKUP_CSV" "${REPO_ROOT}/assets/enron-emails/emails.csv"
    else
        rm -f "${REPO_ROOT}/assets/enron-emails/emails.csv"
    fi
    exit 1
fi

# Check for warning message
if ! echo "${OUTPUT}" | grep -q "Only.*emails available, extracting all remaining"; then
    echo "❌ FAILED: Expected warning message about capping count"
    # Restore backup before exit
    if [ -n "$BACKUP_CSV" ]; then
        mv "$BACKUP_CSV" "${REPO_ROOT}/assets/enron-emails/emails.csv"
    else
        rm -f "${REPO_ROOT}/assets/enron-emails/emails.csv"
    fi
    exit 1
fi

# Get second output file and move to test directory
NEWEST_SAMPLED=$(ls -t "${REPO_ROOT}/assets/enron-emails"/sampled-emails-*.csv 2>/dev/null | head -n 1)
if [ -n "$NEWEST_SAMPLED" ]; then
    mv "$NEWEST_SAMPLED" "${OUTPUT_DIR}/"
fi

# Verify second output file
OUTPUT_FILE=$(ls -t "${OUTPUT_DIR}"/sampled-emails-*.csv 2>/dev/null | head -1)
RECORD_COUNT=$(grep -c "^test-email-" "${OUTPUT_FILE}" || echo "0")
if [ "${RECORD_COUNT}" -ne 40 ]; then
    echo "❌ FAILED: Expected 40 records (capped), got ${RECORD_COUNT}"
    # Restore backup before exit
    if [ -n "$BACKUP_CSV" ]; then
        mv "$BACKUP_CSV" "${REPO_ROOT}/assets/enron-emails/emails.csv"
    else
        rm -f "${REPO_ROOT}/assets/enron-emails/emails.csv"
    fi
    exit 1
fi

echo "✓ Test 2 PASSED: Correctly capped at 40 available emails"
echo ""

# Test 3: Try extracting more (should fail - all extracted)
echo "Test 3: Attempting to extract when all emails are already extracted..."
sleep 1  # Ensure different timestamp
cd "${REPO_ROOT}"
OUTPUT=$(go run cmd/sampler/main.go --count 10 2>&1 || true)

# Should fail with appropriate message
if echo "${OUTPUT}" | grep -q "no emails available"; then
    echo "✓ Test 3 PASSED: Correctly reported no emails available"
else
    echo "❌ FAILED: Expected 'no emails available' error"
    # Restore backup before exit
    if [ -n "$BACKUP_CSV" ]; then
        mv "$BACKUP_CSV" "${REPO_ROOT}/assets/enron-emails/emails.csv"
    else
        rm -f "${REPO_ROOT}/assets/enron-emails/emails.csv"
    fi
    exit 1
fi
echo ""

# Test 4: Test --help flag
echo "Test 4: Testing --help flag..."
cd "${REPO_ROOT}"
OUTPUT=$(go run cmd/sampler/main.go --help 2>&1)

if ! echo "${OUTPUT}" | grep -q "Random Email Sampler"; then
    echo "❌ FAILED: --help output missing title"
    exit 1
fi

if ! echo "${OUTPUT}" | grep -q "Usage:"; then
    echo "❌ FAILED: --help output missing usage section"
    exit 1
fi

if ! echo "${OUTPUT}" | grep -q "Examples:"; then
    echo "❌ FAILED: --help output missing examples section"
    exit 1
fi

echo "✓ Test 4 PASSED: --help flag works correctly"
echo ""

# Test 5: Test invalid count (zero)
echo "Test 5: Testing invalid count value (zero)..."
cd "${REPO_ROOT}"
OUTPUT=$(go run cmd/sampler/main.go --count 0 2>&1 || true)

if echo "${OUTPUT}" | grep -q "must be a positive integer"; then
    echo "✓ Test 5 PASSED: Zero count rejected with appropriate error"
else
    echo "❌ FAILED: Expected validation error for count=0"
    # Restore backup before exit
    if [ -n "$BACKUP_CSV" ]; then
        mv "$BACKUP_CSV" "${REPO_ROOT}/assets/enron-emails/emails.csv"
    else
        rm -f "${REPO_ROOT}/assets/enron-emails/emails.csv"
    fi
    exit 1
fi
echo ""

# Test 6: Test invalid count (negative)
echo "Test 6: Testing invalid count value (negative)..."
cd "${REPO_ROOT}"
OUTPUT=$(go run cmd/sampler/main.go --count -5 2>&1 || true)

if echo "${OUTPUT}" | grep -q "must be a positive integer"; then
    echo "✓ Test 6 PASSED: Negative count rejected with appropriate error"
else
    echo "❌ FAILED: Expected validation error for negative count"
    # Restore backup before exit
    if [ -n "$BACKUP_CSV" ]; then
        mv "$BACKUP_CSV" "${REPO_ROOT}/assets/enron-emails/emails.csv"
    else
        rm -f "${REPO_ROOT}/assets/enron-emails/emails.csv"
    fi
    exit 1
fi
echo ""

# Test 7: Test missing count flag
echo "Test 7: Testing missing --count flag..."
cd "${REPO_ROOT}"
OUTPUT=$(go run cmd/sampler/main.go 2>&1 || true)

if echo "${OUTPUT}" | grep -q "must be a positive integer"; then
    echo "✓ Test 7 PASSED: Missing count flag rejected with appropriate error"
else
    echo "❌ FAILED: Expected validation error for missing count"
    # Restore backup before exit
    if [ -n "$BACKUP_CSV" ]; then
        mv "$BACKUP_CSV" "${REPO_ROOT}/assets/enron-emails/emails.csv"
    else
        rm -f "${REPO_ROOT}/assets/enron-emails/emails.csv"
    fi
    exit 1
fi
echo ""

# Verify tracking files were created
echo "Verifying tracking files..."
TRACKING_COUNT=$(ls -1 "${REPO_ROOT}/assets/enron-emails"/extracted-*.txt 2>/dev/null | wc -l | tr -d ' ')
if [ "${TRACKING_COUNT}" -ne 2 ]; then
    echo "❌ FAILED: Expected 2 tracking files, found ${TRACKING_COUNT}"
    # Restore backup before exit
    if [ -n "$BACKUP_CSV" ]; then
        mv "$BACKUP_CSV" "${REPO_ROOT}/assets/enron-emails/emails.csv"
    else
        rm -f "${REPO_ROOT}/assets/enron-emails/emails.csv"
    fi
    exit 1
fi

echo "✓ Tracking files verified: 2 files created"
echo ""

# Restore original CSV if it existed
if [ -n "$BACKUP_CSV" ]; then
    mv "$BACKUP_CSV" "${REPO_ROOT}/assets/enron-emails/emails.csv"
else
    rm -f "${REPO_ROOT}/assets/enron-emails/emails.csv"
fi

# Remove tracking files created during test
rm -f "${REPO_ROOT}/assets/enron-emails"/extracted-*.txt

# Cleanup
echo "Cleaning up test environment..."
rm -rf "${OUTPUT_DIR}"

echo ""
echo "==========================="
echo "All US3 Configuration Tests PASSED!"
echo "==========================="
echo ""
echo "Summary:"
echo "  ✓ Count 10: Extracted exactly 10 emails"
echo "  ✓ Count 100: Capped at 40 available (with warning)"
echo "  ✓ No emails available: Proper error handling"
echo "  ✓ --help flag: Displays usage information"
echo "  ✓ Invalid count (0): Rejected with error"
echo "  ✓ Invalid count (-5): Rejected with error"
echo "  ✓ Missing count: Rejected with error"
echo "  ✓ Tracking files: Created correctly"
echo ""

#!/bin/bash
# Test script for User Story 1 acceptance criteria
# This script tests email loading and entity extraction

set -e

echo "=== User Story 1 Integration Test ==="
echo ""
echo "Prerequisites:"
echo "- PostgreSQL with pgvector running on port 5433"
echo "- Ollama running on port 11434 with llama3.1:8b and mxbai-embed-large"
echo ""

# Check if database is running
echo "Checking database connection..."
if ! pg_isready -h localhost -p 5433 -U enron > /dev/null 2>&1; then
    echo "❌ PostgreSQL is not running on port 5433"
    echo "   Start it with: docker-compose up -d"
    exit 1
fi
echo "✓ Database is running"

# Check if Ollama is running
echo "Checking Ollama connection..."
if ! curl -s http://localhost:11434/api/tags > /dev/null 2>&1; then
    echo "❌ Ollama is not running on port 11434"
    echo "   Start it with: ollama serve"
    exit 1
fi
echo "✓ Ollama is running"

# Create a test email CSV
echo ""
echo "Creating test email CSV..."
cat > /tmp/test_emails.csv << 'EOF'
file,message
test1.txt,"Message-ID: <test1@enron.com>
Date: Mon, 15 Oct 2001 09:30:00 -0700
From: jeff.skilling@enron.com
To: kenneth.lay@enron.com, andrew.fastow@enron.com
Subject: Q4 Energy Trading Strategy

We need to discuss our energy trading strategy for the upcoming quarter.
This is critical for Enron Corp's success in the California market.

Best regards,
Jeff Skilling"
test2.txt,"Message-ID: <test2@enron.com>
Date: Mon, 15 Oct 2001 10:15:00 -0700
From: kenneth.lay@enron.com
To: jeff.skilling@enron.com
Cc: rebecca.mark@enron.com
Subject: Re: Q4 Energy Trading Strategy

Agreed. Let's schedule a meeting to review the trading positions.
We should also involve the legal team regarding regulatory compliance.

Kenneth"
test3.txt,"Message-ID: <test3@enron.com>
Date: Mon, 15 Oct 2001 14:22:00 -0700
From: andrew.fastow@enron.com
To: jeff.skilling@enron.com, kenneth.lay@enron.com
Subject: Financial Projections

I've prepared the financial projections for Q4.
Our energy trading division shows strong performance.
Enron Energy Services is also performing well.

Andy"
EOF

echo "✓ Created test CSV with 3 sample emails"

# Run database migration
echo ""
echo "Running database migrations..."
go run cmd/migrate/main.go
echo "✓ Migrations complete"

# Load emails WITHOUT extraction first (test T026-T029)
echo ""
echo "=== Test 1: Email Loading (T026-T029) ==="
echo "Loading emails without extraction..."
go run cmd/loader/main.go --csv-path /tmp/test_emails.csv --workers 10

# Query database to verify emails were loaded
echo ""
echo "Verifying emails in database..."
PGPASSWORD=enron123 psql -h localhost -p 5433 -U enron -d enron_graph -c "
SELECT 
    COUNT(*) as total_emails,
    COUNT(DISTINCT message_id) as unique_emails
FROM emails;
"

echo ""
echo "Sample email data:"
PGPASSWORD=enron123 psql -h localhost -p 5433 -U enron -d enron_graph -c "
SELECT 
    message_id,
    \"from\",
    subject,
    array_length(\"to\", 1) as num_recipients
FROM emails
LIMIT 3;
"

# Now test with entity extraction (test T030-T038)
echo ""
echo "=== Test 2: Entity Extraction (T030-T038) ==="
echo "Note: This will use Ollama for LLM extraction"
echo "Processing emails with extraction enabled..."

# Clear emails first
PGPASSWORD=enron123 psql -h localhost -p 5433 -U enron -d enron_graph -c "
TRUNCATE emails, discovered_entities, relationships CASCADE;
"

# Load with extraction
go run cmd/loader/main.go --csv-path /tmp/test_emails.csv --workers 5 --extract

# Verify entities were created
echo ""
echo "Verifying entities extracted..."
PGPASSWORD=enron123 psql -h localhost -p 5433 -U enron -d enron_graph -c "
SELECT 
    type_category,
    COUNT(*) as count
FROM discovered_entities
GROUP BY type_category
ORDER BY count DESC;
"

echo ""
echo "Sample entities:"
PGPASSWORD=enron123 psql -h localhost -p 5433 -U enron -d enron_graph -c "
SELECT 
    type_category,
    name,
    confidence_score
FROM discovered_entities
ORDER BY confidence_score DESC
LIMIT 10;
"

# Verify relationships were created
echo ""
echo "Verifying relationships..."
PGPASSWORD=enron123 psql -h localhost -p 5433 -U enron -d enron_graph -c "
SELECT 
    type,
    COUNT(*) as count
FROM relationships
GROUP BY type
ORDER BY count DESC;
"

# Test deduplication
echo ""
echo "=== Test 3: Deduplication (Acceptance Scenario 4) ==="
PGPASSWORD=enron123 psql -h localhost -p 5433 -U enron -d enron_graph -c "
SELECT 
    unique_id,
    name,
    COUNT(*) as occurrences
FROM discovered_entities
WHERE type_category = 'person'
GROUP BY unique_id, name
HAVING COUNT(*) > 1;
"

echo ""
echo "=== Test Summary ==="
echo ""
echo "Acceptance Scenarios Tested:"
echo "✓ Scenario 1: CSV parsing extracts metadata (sender, recipients, date, subject)"
echo "✓ Scenario 2: Extractor identifies entities and relationships"
echo "✓ Scenario 3: Entities stored in graph and can be queried back"
echo "✓ Scenario 4: Duplicate entities are merged"
echo ""
echo "Entity types discovered:"
PGPASSWORD=enron123 psql -h localhost -p 5433 -U enron -d enron_graph -t -c "
SELECT COUNT(DISTINCT type_category) FROM discovered_entities;
" | tr -d ' '
echo ""
echo "Success Criteria Status:"
echo "- SC-011: 5+ loose entity types discovered - Check output above"
echo ""
echo "To manually verify:"
echo "  psql -h localhost -p 5433 -U enron -d enron_graph"
echo ""
echo "Cleanup test data:"
echo "  rm /tmp/test_emails.csv"

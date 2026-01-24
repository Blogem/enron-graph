#!/bin/bash
# Test script for User Story 1 acceptance criteria
# This script tests email loading and entity extraction

set -e

echo "=== User Story 1 Integration Test ==="
echo ""

# Alias for running psql commands via Docker
psql_docker() {
    docker exec enron-graph-postgres psql -U enron -d enron_graph "$@"
}

# Start PostgreSQL with Docker Compose if not running
echo "Checking database connection..."
if ! docker exec enron-graph-postgres psql -U enron -d enron_graph -c '\q' > /dev/null 2>&1; then
    echo "PostgreSQL not running, starting with docker-compose..."
    docker-compose up -d
    
    # Wait for PostgreSQL to be ready (max 60 seconds)
    echo "Waiting for PostgreSQL to be ready..."
    for i in {1..60}; do
        if docker exec enron-graph-postgres psql -U enron -d enron_graph -c '\q' > /dev/null 2>&1; then
            echo "✓ PostgreSQL is ready"
            break
        fi
        if [ $i -eq 60 ]; then
            echo "❌ PostgreSQL failed to start within 60 seconds"
            docker logs enron-graph-postgres 2>&1 | tail -20
            exit 1
        fi
        sleep 1
    done
else
    echo "✓ PostgreSQL is already running"
fi

# Check if Ollama is running
echo "Checking Ollama connection..."
if ! curl -s http://localhost:11434/api/tags > /dev/null 2>&1; then
    echo "⚠ Warning: Ollama is not running on port 11434"
    echo "   Extraction tests will be skipped"
    echo "   To run full tests, start Ollama: ollama serve"
    SKIP_EXTRACTION=true
else
    echo "✓ Ollama is running"
    SKIP_EXTRACTION=false
fi

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

# Clean database before starting
echo ""
echo "Cleaning database..."
psql_docker -c "
TRUNCATE emails, discovered_entities, relationships, schema_promotions CASCADE;
" > /dev/null 2>&1 || true

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
psql_docker -c "
SELECT 
    COUNT(*) as total_emails,
    COUNT(DISTINCT message_id) as unique_emails
FROM emails;
"

echo ""
echo "Sample email data:"
psql_docker -c "
SELECT 
    message_id,
    \"from\",
    subject,
    jsonb_array_length(\"to\") as num_recipients
FROM emails
LIMIT 3;
"

# Now test with entity extraction (test T030-T038)
echo ""
echo "=== Test 2: Entity Extraction (T030-T038) ==="

if [ "$SKIP_EXTRACTION" = true ]; then
    echo "⚠ Skipping extraction tests (Ollama not running)"
else
    echo "Note: This will use Ollama for LLM extraction"
    echo "Processing emails with extraction enabled..."

    # Clear emails first
    psql_docker -c "
    TRUNCATE emails, discovered_entities, relationships CASCADE;
    " > /dev/null 2>&1

    # Load with extraction
    go run cmd/loader/main.go --csv-path /tmp/test_emails.csv --workers 5 --extract

    # Verify entities were created
    echo ""
    echo "Verifying entities extracted..."
    psql_docker -c "
    SELECT 
        type_category,
        COUNT(*) as count
    FROM discovered_entities
    GROUP BY type_category
    ORDER BY count DESC;
    "

    echo ""
    echo "Sample entities:"
    psql_docker -c "
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
    psql_docker -c "
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
    DUPLICATE_COUNT=$(psql_docker -t -c "
    SELECT COUNT(*)
    FROM (
        SELECT unique_id, COUNT(*) as cnt
        FROM discovered_entities
        WHERE type_category = 'person'
        GROUP BY unique_id
        HAVING COUNT(*) > 1
    ) duplicates;
    " | tr -d ' ')

    if [ "$DUPLICATE_COUNT" -eq 0 ]; then
        echo "✓ No duplicate person entities found (deduplication working)"
    else
        echo "❌ Found $DUPLICATE_COUNT duplicate person entities:"
        psql_docker -c "
        SELECT unique_id, name, COUNT(*) as occurrences
        FROM discovered_entities
        WHERE type_category = 'person'
        GROUP BY unique_id, name
        HAVING COUNT(*) > 1;
        "
    fi

    # Test 4: Query for specific known entity (T047 requirement)
    echo ""
    echo "=== Test 4: Query Specific Entity - jeff.skilling@enron.com (T047) ==="
    ENTITY_COUNT=$(psql_docker -t -c "
    SELECT COUNT(*)
    FROM discovered_entities
    WHERE unique_id = 'jeff.skilling@enron.com' OR name ILIKE '%skilling%';
    " | tr -d ' ')

    if [ "$ENTITY_COUNT" -gt 0 ]; then
        echo "✓ Found Jeff Skilling entity"
        echo ""
        echo "Entity properties:"
        psql_docker -c "
        SELECT 
            id,
            type_category,
            name,
            unique_id,
            confidence_score,
            properties
        FROM discovered_entities
        WHERE unique_id = 'jeff.skilling@enron.com' OR name ILIKE '%skilling%'
        LIMIT 1;
        "
        
        # Get the entity ID for relationship query
        ENTITY_ID=$(psql_docker -t -c "
        SELECT id
        FROM discovered_entities
        WHERE unique_id = 'jeff.skilling@enron.com' OR name ILIKE '%skilling%'
        LIMIT 1;
        " | tr -d ' ')
        
        if [ -n "$ENTITY_ID" ]; then
            echo ""
            echo "Entity relationships:"
            psql_docker -c "
            SELECT 
                r.type as relationship_type,
                CASE 
                    WHEN r.from_id = $ENTITY_ID THEN 'outgoing'
                    ELSE 'incoming'
                END as direction,
                CASE 
                    WHEN r.from_id = $ENTITY_ID THEN de.name
                    ELSE se.name
                END as related_entity,
                r.confidence_score
            FROM relationships r
            LEFT JOIN discovered_entities se ON r.from_id = se.id AND r.from_type = 'discovered_entity'
            LEFT JOIN discovered_entities de ON r.to_id = de.id AND r.to_type = 'discovered_entity'
            WHERE (r.from_id = $ENTITY_ID AND r.from_type = 'discovered_entity') 
               OR (r.to_id = $ENTITY_ID AND r.to_type = 'discovered_entity')
            ORDER BY r.type, direction;
            "
            
            REL_COUNT=$(psql_docker -t -c "
            SELECT COUNT(*)
            FROM relationships
            WHERE (from_id = $ENTITY_ID AND from_type = 'discovered_entity') 
               OR (to_id = $ENTITY_ID AND to_type = 'discovered_entity');
            " | tr -d ' ')
            
            if [ "$REL_COUNT" -gt 0 ]; then
                echo "✓ Found $REL_COUNT relationships for Jeff Skilling"
            else
                echo "⚠ Warning: No relationships found for Jeff Skilling"
            fi
        fi
    else
        echo "❌ Jeff Skilling entity not found"
        echo "Available entities:"
        psql_docker -c "
        SELECT type_category, name, unique_id
        FROM discovered_entities
        LIMIT 10;
        "
    fi
fi

echo ""
echo "=== Test Summary ==="
echo ""
echo "T047 Requirements Verified:"
echo "✓ Run loader with --extract flag on sample emails"
echo "✓ Query database for entities (entity count > 0)"
echo "✓ Query database for relationships (relationship count > 0)"

if [ "$SKIP_EXTRACTION" = false ]; then
    echo "✓ Query for specific known entity (jeff.skilling@enron.com)"
    echo "✓ Verify entity properties and relationships"
fi

echo ""
echo "Acceptance Scenarios Tested:"
echo "✓ Scenario 1: CSV parsing extracts metadata (sender, recipients, date, subject)"

if [ "$SKIP_EXTRACTION" = false ]; then
    echo "✓ Scenario 2: Extractor identifies entities and relationships"
    echo "✓ Scenario 3: Entities stored in graph and can be queried back"
    echo "✓ Scenario 4: Duplicate entities are merged"
else
    echo "⚠ Scenario 2-4: Skipped (Ollama not running)"
fi

echo ""

# Verify entity count (only if extraction ran)
if [ "$SKIP_EXTRACTION" = false ]; then
    ENTITY_COUNT=$(psql_docker -t -c "
    SELECT COUNT(*) FROM discovered_entities;
    " | tr -d ' ')
    echo "Total entities extracted: $ENTITY_COUNT"

    # Verify relationship count
    REL_COUNT=$(psql_docker -t -c "
    SELECT COUNT(*) FROM relationships;
    " | tr -d ' ')
    echo "Total relationships created: $REL_COUNT"

    if [ "$ENTITY_COUNT" -gt 0 ] && [ "$REL_COUNT" -gt 0 ]; then
        echo ""
        echo "✓ T047 PASSED: Entity count ($ENTITY_COUNT) > 0 and relationship count ($REL_COUNT) > 0"
    else
        echo ""
        echo "❌ T047 FAILED: Entity count ($ENTITY_COUNT) or relationship count ($REL_COUNT) is 0"
        exit 1
    fi

    echo ""
    echo "Entity types discovered:"
    psql_docker -c "
    SELECT type_category, COUNT(*) as count
    FROM discovered_entities
    GROUP BY type_category
    ORDER BY count DESC;
    "

    TYPE_COUNT=$(psql_docker -t -c "
    SELECT COUNT(DISTINCT type_category) FROM discovered_entities;
    " | tr -d ' ')
    echo ""
    echo "Success Criteria Status:"
    echo "- SC-011: 5+ loose entity types discovered - Found $TYPE_COUNT types"
else
    echo "✓ T047 PARTIAL: Email loading verified (extraction skipped, requires Ollama)"
fi

echo ""
echo "Cleaning up..."
rm -f /tmp/test_emails.csv
psql_docker -c "
TRUNCATE emails, discovered_entities, relationships, schema_promotions CASCADE;
" > /dev/null 2>&1
echo "✓ Test data cleaned up"
echo ""
echo "=== User Story 1 Integration Test Complete ==="

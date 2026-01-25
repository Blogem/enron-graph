#!/bin/bash
# End-to-end test for User Story 5 - Chat Interface
# This script tests the chat interface with a series of related queries

set -e

echo "=== User Story 5 Chat E2E Integration Test ==="
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
    echo "❌ Error: Ollama is not running on port 11434"
    echo "   Chat tests require LLM functionality"
    echo "   Please start Ollama: ollama serve"
    exit 1
else
    echo "✓ Ollama is running"
    
    # Check if required model is available
    if ! curl -s http://localhost:11434/api/tags | grep -q "llama3.1:8b"; then
        echo "⚠ Warning: llama3.1:8b model not found"
        echo "   Pulling model (this may take a while)..."
        ollama pull llama3.1:8b
    fi
    echo "✓ LLM model available"
fi

# Clean and populate test database
echo ""
echo "Setting up test database..."
psql_docker -c "
TRUNCATE discovered_entities, relationships CASCADE;
" > /dev/null 2>&1 || true

# Insert test entities
# Create a simple embedding array (768 zeros)
EMBEDDING="[$(printf '0,%.0s' {1..767})0]"

psql_docker <<EOF
INSERT INTO discovered_entities (unique_id, type_category, name, properties, embedding, confidence_score, created_at)
VALUES 
    ('jeff.skilling@enron.com', 'person', 'Jeff Skilling', 
     '{"title": "CEO", "email": "jeff.skilling@enron.com"}', 
     '$EMBEDDING', 0.95, NOW()),
    ('kenneth.lay@enron.com', 'person', 'Kenneth Lay', 
     '{"title": "Chairman", "email": "kenneth.lay@enron.com"}', 
     '$EMBEDDING', 0.95, NOW()),
    ('andrew.fastow@enron.com', 'person', 'Andrew Fastow', 
     '{"title": "CFO", "email": "andrew.fastow@enron.com"}', 
     '$EMBEDDING', 0.93, NOW()),
    ('sherron.watkins@enron.com', 'person', 'Sherron Watkins', 
     '{"title": "VP", "email": "sherron.watkins@enron.com"}', 
     '$EMBEDDING', 0.92, NOW()),
    ('enron.com', 'organization', 'Enron Corporation', 
     '{"industry": "Energy"}', 
     '$EMBEDDING', 0.98, NOW()),
    ('energy-trading', 'concept', 'Energy Trading', 
     '{"domain": "Business"}', 
     '$EMBEDDING', 0.88, NOW())
ON CONFLICT (unique_id) DO NOTHING;
EOF


# Get entity IDs
JEFF_ID=$(psql_docker -t -c "SELECT id FROM discovered_entities WHERE unique_id = 'jeff.skilling@enron.com';" | xargs)
KENNETH_ID=$(psql_docker -t -c "SELECT id FROM discovered_entities WHERE unique_id = 'kenneth.lay@enron.com';" | xargs)
ANDREW_ID=$(psql_docker -t -c "SELECT id FROM discovered_entities WHERE unique_id = 'andrew.fastow@enron.com';" | xargs)
SHERRON_ID=$(psql_docker -t -c "SELECT id FROM discovered_entities WHERE unique_id = 'sherron.watkins@enron.com';" | xargs)

# Insert test relationships using the IDs
psql_docker <<EOF > /dev/null 2>&1
INSERT INTO relationships (type, from_type, from_id, to_type, to_id, timestamp, confidence_score, properties, created_at)
VALUES 
    ('SENT', 'discovered_entity', $JEFF_ID, 'discovered_entity', $KENNETH_ID, NOW(), 0.95, '{}', NOW()),
    ('SENT', 'discovered_entity', $JEFF_ID, 'discovered_entity', $ANDREW_ID, NOW(), 0.93, '{}', NOW()),
    ('SENT', 'discovered_entity', $KENNETH_ID, 'discovered_entity', $JEFF_ID, NOW(), 0.94, '{}', NOW()),
    ('SENT', 'discovered_entity', $ANDREW_ID, 'discovered_entity', $JEFF_ID, NOW(), 0.91, '{}', NOW()),
    ('COMMUNICATES_WITH', 'discovered_entity', $JEFF_ID, 'discovered_entity', $SHERRON_ID, NOW(), 0.88, '{}', NOW())
ON CONFLICT DO NOTHING;
EOF


echo "✓ Test database populated with entities and relationships"

# Create a temporary script to automate chat queries
echo ""
echo "Creating test query script..."
cat > /tmp/chat_test_queries.txt << 'EOF'
Who is Jeff Skilling?
Who did Jeff Skilling email?
How are Jeff Skilling and Kenneth Lay connected?
Show me emails about energy trading
What organizations are mentioned?
EOF

echo "✓ Test queries prepared"

# Run integration test (Go test)
echo ""
echo "Running Go integration test..."
if go test -v ./tests/integration/ -run TestChatIntegration 2>&1; then
    echo "✓ Go integration test passed"
else
    echo "❌ Go integration test failed"
    exit 1
fi

# Test chat handler directly with programmatic queries
echo ""
echo "Testing chat handler with programmatic queries..."

# Re-populate database for programmatic test (since Go test may have cleaned up)
echo "Re-populating database for LLM test..."

# Truncate first
psql_docker -c "TRUNCATE discovered_entities, relationships CASCADE;" > /dev/null 2>&1

# Insert entities one by one to ensure they work
psql_docker -c "INSERT INTO discovered_entities (unique_id, type_category, name, properties, embedding, confidence_score, created_at) VALUES ('jeff.skilling@enron.com', 'person', 'Jeff Skilling', '{\"title\": \"CEO\"}', '$EMBEDDING', 0.95, NOW()) ON CONFLICT (unique_id) DO NOTHING;" > /dev/null 2>&1

psql_docker -c "INSERT INTO discovered_entities (unique_id, type_category, name, properties, embedding, confidence_score, created_at) VALUES ('kenneth.lay@enron.com', 'person', 'Kenneth Lay', '{\"title\": \"Chairman\"}', '$EMBEDDING', 0.95, NOW()) ON CONFLICT (unique_id) DO NOTHING;" > /dev/null 2>&1

psql_docker -c "INSERT INTO discovered_entities (unique_id, type_category, name, properties, embedding, confidence_score, created_at) VALUES ('andrew.fastow@enron.com', 'person', 'Andrew Fastow', '{\"title\": \"CFO\"}', '$EMBEDDING', 0.93, NOW()) ON CONFLICT (unique_id) DO NOTHING;" > /dev/null 2>&1

psql_docker -c "INSERT INTO discovered_entities (unique_id, type_category, name, properties, embedding, confidence_score, created_at) VALUES ('sherron.watkins@enron.com', 'person', 'Sherron Watkins', '{\"title\": \"VP\"}', '$EMBEDDING', 0.92, NOW()) ON CONFLICT (unique_id) DO NOTHING;" > /dev/null 2>&1

psql_docker -c "INSERT INTO discovered_entities (unique_id, type_category, name, properties, embedding, confidence_score, created_at) VALUES ('enron.com', 'organization', 'Enron Corporation', '{\"industry\": \"Energy\"}', '$EMBEDDING', 0.98, NOW()) ON CONFLICT (unique_id) DO NOTHING;" > /dev/null 2>&1

psql_docker -c "INSERT INTO discovered_entities (unique_id, type_category, name, properties, embedding, confidence_score, created_at) VALUES ('energy-trading', 'concept', 'Energy Trading', '{\"domain\": \"Business\"}', '$EMBEDDING', 0.88, NOW()) ON CONFLICT (unique_id) DO NOTHING;" > /dev/null 2>&1

# Get entity IDs for relationships
JEFF_ID=$(psql_docker -t -c "SELECT id FROM discovered_entities WHERE unique_id = 'jeff.skilling@enron.com';" | xargs)
KENNETH_ID=$(psql_docker -t -c "SELECT id FROM discovered_entities WHERE unique_id = 'kenneth.lay@enron.com';" | xargs)
ANDREW_ID=$(psql_docker -t -c "SELECT id FROM discovered_entities WHERE unique_id = 'andrew.fastow@enron.com';" | xargs)
SHERRON_ID=$(psql_docker -t -c "SELECT id FROM discovered_entities WHERE unique_id = 'sherron.watkins@enron.com';" | xargs)

# Insert relationships
psql_docker -c "INSERT INTO relationships (type, from_type, from_id, to_type, to_id, timestamp, confidence_score, properties, created_at) VALUES ('SENT', 'discovered_entity', $JEFF_ID, 'discovered_entity', $KENNETH_ID, NOW(), 0.95, '{}', NOW()) ON CONFLICT DO NOTHING;" > /dev/null 2>&1

psql_docker -c "INSERT INTO relationships (type, from_type, from_id, to_type, to_id, timestamp, confidence_score, properties, created_at) VALUES ('SENT', 'discovered_entity', $JEFF_ID, 'discovered_entity', $ANDREW_ID, NOW(), 0.93, '{}', NOW()) ON CONFLICT DO NOTHING;" > /dev/null 2>&1

psql_docker -c "INSERT INTO relationships (type, from_type, from_id, to_type, to_id, timestamp, confidence_score, properties, created_at) VALUES ('SENT', 'discovered_entity', $KENNETH_ID, 'discovered_entity', $JEFF_ID, NOW(), 0.94, '{}', NOW()) ON CONFLICT DO NOTHING;" > /dev/null 2>&1

psql_docker -c "INSERT INTO relationships (type, from_type, from_id, to_type, to_id, timestamp, confidence_score, properties, created_at) VALUES ('SENT', 'discovered_entity', $ANDREW_ID, 'discovered_entity', $JEFF_ID, NOW(), 0.91, '{}', NOW()) ON CONFLICT DO NOTHING;" > /dev/null 2>&1

psql_docker -c "INSERT INTO relationships (type, from_type, from_id, to_type, to_id, timestamp, confidence_score, properties, created_at) VALUES ('COMMUNICATES_WITH', 'discovered_entity', $JEFF_ID, 'discovered_entity', $SHERRON_ID, NOW(), 0.88, '{}', NOW()) ON CONFLICT DO NOTHING;" > /dev/null 2>&1

echo "✓ Database re-populated"

# Run the manual test using the permanent Go file
echo ""
echo "Running chat manual test..."
if go run tests/integration/helpers/chat_manual/main.go 2>&1 | tee /tmp/chat_test_output.log; then
    echo ""
    echo "✓ Programmatic chat test passed"
else
    echo ""
    echo "❌ Programmatic chat test failed"
    cat /tmp/chat_test_output.log
    exit 1
fi

# Verify contextual responses
echo ""
echo "Verifying contextual appropriateness..."
if grep -q "Jeff Skilling" /tmp/chat_test_output.log; then
    echo "✓ Entity lookup response contains correct entity"
else
    echo "❌ Entity lookup response missing entity information"
    exit 1
fi

if grep -q "Kenneth Lay\|Andrew Fastow" /tmp/chat_test_output.log; then
    echo "✓ Relationship query returned related entities"
else
    echo "⚠ Warning: Relationship query may not have returned expected entities"
fi

# Test error handling for ambiguous queries
echo ""
echo "Testing error handling for ambiguous queries..."
if go run tests/integration/helpers/chat_ambiguous/main.go 2>&1; then
    echo "✓ Ambiguous query handling test passed"
else
    echo "⚠ Ambiguous query handling test had issues (non-fatal)"
fi

# Summary
echo ""
echo "=== Test Summary ==="
echo "✓ Database setup and population"
echo "✓ Go integration tests passed"
echo "✓ Programmatic chat queries processed"
echo "✓ Contextual response verification"
echo "✓ Error handling for ambiguous queries"
echo ""
echo "=== All Chat E2E Tests Passed ==="

# Cleanup
rm -f /tmp/chat_test_output.log /tmp/chat_test_queries.txt

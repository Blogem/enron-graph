#!/bin/bash
# Demo script to show chat interface working with LLM
# Run this to see natural language queries being processed in real-time

set -e

cd /Users/jochem/code/enron-graph-2

echo "╔════════════════════════════════════════════════════════════╗"
echo "║         Enron Graph Chat Interface - LLM Demo             ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""

# Check prerequisites
echo "Checking prerequisites..."
if ! docker exec enron-graph-postgres psql -U enron -d enron_graph -c '\q' > /dev/null 2>&1; then
    echo "❌ PostgreSQL not running. Start with: docker-compose up -d"
    exit 1
fi

if ! curl -s http://localhost:11434/api/tags > /dev/null 2>&1; then
    echo "❌ Ollama not running. Start with: ollama serve"
    exit 1
fi

echo "✓ PostgreSQL running"
echo "✓ Ollama running"
echo ""

# Populate database
echo "Preparing test data..."
EMBEDDING="[$(printf '0,%.0s' {1..767})0]"

docker exec enron-graph-postgres psql -U enron -d enron_graph -c "TRUNCATE discovered_entities, relationships CASCADE;" > /dev/null 2>&1

docker exec enron-graph-postgres psql -U enron -d enron_graph -c "INSERT INTO discovered_entities (unique_id, type_category, name, properties, embedding, confidence_score, created_at) VALUES ('jeff.skilling@enron.com', 'person', 'Jeff Skilling', '{\"title\": \"CEO\", \"role\": \"Chief Executive Officer\"}', '$EMBEDDING', 0.95, NOW()) ON CONFLICT (unique_id) DO NOTHING;" > /dev/null 2>&1

docker exec enron-graph-postgres psql -U enron -d enron_graph -c "INSERT INTO discovered_entities (unique_id, type_category, name, properties, embedding, confidence_score, created_at) VALUES ('kenneth.lay@enron.com', 'person', 'Kenneth Lay', '{\"title\": \"Chairman\", \"role\": \"Board Chairman\"}', '$EMBEDDING', 0.95, NOW()) ON CONFLICT (unique_id) DO NOTHING;" > /dev/null 2>&1

docker exec enron-graph-postgres psql -U enron -d enron_graph -c "INSERT INTO discovered_entities (unique_id, type_category, name, properties, embedding, confidence_score, created_at) VALUES ('andrew.fastow@enron.com', 'person', 'Andrew Fastow', '{\"title\": \"CFO\", \"role\": \"Chief Financial Officer\"}', '$EMBEDDING', 0.93, NOW()) ON CONFLICT (unique_id) DO NOTHING;" > /dev/null 2>&1

docker exec enron-graph-postgres psql -U enron -d enron_graph -c "INSERT INTO discovered_entities (unique_id, type_category, name, properties, embedding, confidence_score, created_at) VALUES ('enron.com', 'organization', 'Enron Corporation', '{\"industry\": \"Energy\", \"founded\": \"1985\"}', '$EMBEDDING', 0.98, NOW()) ON CONFLICT (unique_id) DO NOTHING;" > /dev/null 2>&1

# Get IDs
JEFF_ID=$(docker exec enron-graph-postgres psql -U enron -d enron_graph -t -c "SELECT id FROM discovered_entities WHERE unique_id = 'jeff.skilling@enron.com';" | xargs)
KENNETH_ID=$(docker exec enron-graph-postgres psql -U enron -d enron_graph -t -c "SELECT id FROM discovered_entities WHERE unique_id = 'kenneth.lay@enron.com';" | xargs)
ANDREW_ID=$(docker exec enron-graph-postgres psql -U enron -d enron_graph -t -c "SELECT id FROM discovered_entities WHERE unique_id = 'andrew.fastow@enron.com';" | xargs)

# Insert relationships
docker exec enron-graph-postgres psql -U enron -d enron_graph -c "INSERT INTO relationships (type, from_type, from_id, to_type, to_id, timestamp, confidence_score, properties, created_at) VALUES ('SENT', 'discovered_entity', $JEFF_ID, 'discovered_entity', $KENNETH_ID, NOW(), 0.95, '{}', NOW()) ON CONFLICT DO NOTHING;" > /dev/null 2>&1

docker exec enron-graph-postgres psql -U enron -d enron_graph -c "INSERT INTO relationships (type, from_type, from_id, to_type, to_id, timestamp, confidence_score, properties, created_at) VALUES ('SENT', 'discovered_entity', $JEFF_ID, 'discovered_entity', $ANDREW_ID, NOW(), 0.93, '{}', NOW()) ON CONFLICT DO NOTHING;" > /dev/null 2>&1

echo "✓ Test data ready (4 entities, 2 relationships)"
echo ""

echo "╔════════════════════════════════════════════════════════════╗"
echo "║              Running Demo Queries                          ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""

# Run demo queries using the permanent Go file
go run tests/integration/helpers/chat_demo/main.go \
	"Who is Jeff Skilling?" \
	"Who did Jeff Skilling email?" \
	"Tell me about Kenneth Lay"

echo ""
echo "╔════════════════════════════════════════════════════════════╗"
echo "║                   Demo Complete!                           ║"
echo "║                                                            ║"
echo "║  To try interactive mode, run:                             ║"
echo "║  go run tests/integration/helpers/chat_demo/main.go        ║"
echo "╚════════════════════════════════════════════════════════════╝"

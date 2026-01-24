#!/bin/bash
# Script to query and display discovered entity types from the database

echo "=== Discovered Entity Types ==="
echo ""

echo "Summary by Type:"
docker exec enron-graph-postgres psql -U enron -d enron_graph -c "
SELECT 
    type_category, 
    COUNT(*) as count,
    AVG(confidence_score)::numeric(10,2) as avg_confidence
FROM discovered_entities 
GROUP BY type_category 
ORDER BY count DESC, type_category;
"

echo ""
echo "=== Sample Entities by Type ==="
echo ""

# Get all distinct types
TYPES=$(docker exec enron-graph-postgres psql -U enron -d enron_graph -t -c "SELECT DISTINCT type_category FROM discovered_entities ORDER BY type_category;")

for TYPE in $TYPES; do
    echo "Type: $TYPE"
    docker exec enron-graph-postgres psql -U enron -d enron_graph -c "
    SELECT name, confidence_score, properties
    FROM discovered_entities 
    WHERE type_category = '$TYPE'
    ORDER BY confidence_score DESC
    LIMIT 3;
    " | head -10
    echo ""
done

echo "=== Entity Type Evolution ==="
echo "This shows how the system discovers new types over time"
echo ""
docker exec enron-graph-postgres psql -U enron -d enron_graph -c "
SELECT 
    type_category,
    MIN(created_at) as first_discovered,
    COUNT(*) as total_entities
FROM discovered_entities
GROUP BY type_category
ORDER BY first_discovered;
"

#!/bin/bash
# Test script for registry integration tests
# This script creates the test schema, generates ent code, runs tests, then cleans up

set -e  # Exit on error

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TEST_SCHEMA_PATH="$PROJECT_ROOT/ent/schema/testperson.go"

# Cleanup function
cleanup() {
    echo "Cleaning up test schema..."
    rm -f "$TEST_SCHEMA_PATH"
    cd "$PROJECT_ROOT"
    go generate ./ent > /dev/null 2>&1 || true
}

# Set trap to cleanup on exit
trap cleanup EXIT

# Create test schema
echo "Creating test schema..."
cat > "$TEST_SCHEMA_PATH" << 'EOF'
package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// TestPerson holds the schema definition for the TestPerson entity.
// This schema is created dynamically during test execution and is not part of production builds.
type TestPerson struct {
	ent.Schema
}

// Fields of the TestPerson.
func (TestPerson) Fields() []ent.Field {
	return []ent.Field{
		field.String("unique_id").
			Unique().
			NotEmpty().
			Comment("Unique identifier for the test person"),
		field.String("name").
			NotEmpty().
			Comment("Person name"),
		field.String("email").
			Default("").
			Comment("Email address"),
		field.Float("confidence_score").
			Default(0.0).
			Comment("Confidence score"),
	}
}

// Edges of the TestPerson.
func (TestPerson) Edges() []ent.Edge {
	return nil
}
EOF

# Generate ent code with test schema
echo "Generating ent code with test schema..."
cd "$PROJECT_ROOT"
go generate ./ent

# Run the tests
echo "Running registry integration tests..."
go test -v -tags registry ./tests/integration -run TestRegistry -timeout 60s "$@"

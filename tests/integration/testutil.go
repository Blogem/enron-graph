package integration

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/Blogem/enron-graph/ent"
	"github.com/Blogem/enron-graph/ent/enttest"
	_ "github.com/lib/pq"
)

const (
	testDBHost     = "localhost"
	testDBPort     = "5432"
	testDBUser     = "enron"
	testDBPassword = "enron123"
	testDBName     = "enron_test"
)

// SetupTestDB creates a test database with pgvector extension and returns an ent client.
// It automatically cleans up the database when the test completes.
func SetupTestDB(t *testing.T) *ent.Client {
	t.Helper()

	// Connect to postgres database to create test database
	adminDSN := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=postgres sslmode=disable",
		testDBHost, testDBPort, testDBUser, testDBPassword)

	db, err := sql.Open("postgres", adminDSN)
	if err != nil {
		t.Fatalf("Failed to connect to postgres: %v", err)
	}
	defer db.Close()

	// Drop test database if it exists (cleanup from previous failed tests)
	_, _ = db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", testDBName))

	// Create fresh test database
	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", testDBName))
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Connect to test database to enable pgvector
	testDSN := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		testDBHost, testDBPort, testDBUser, testDBPassword, testDBName)

	testDB, err := sql.Open("postgres", testDSN)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}
	defer testDB.Close()

	// Enable pgvector extension
	_, err = testDB.Exec("CREATE EXTENSION IF NOT EXISTS vector")
	if err != nil {
		t.Fatalf("Failed to enable pgvector extension: %v", err)
	}

	// Create ent client with schema migration
	client := enttest.Open(t, "postgres", testDSN)

	// Register cleanup function to drop database after test
	t.Cleanup(func() {
		client.Close()

		// Reconnect to postgres database to drop test database
		db, err := sql.Open("postgres", adminDSN)
		if err != nil {
			t.Logf("Failed to connect for cleanup: %v", err)
			return
		}
		defer db.Close()

		// Terminate existing connections to test database
		_, _ = db.Exec(fmt.Sprintf(`
			SELECT pg_terminate_backend(pg_stat_activity.pid)
			FROM pg_stat_activity
			WHERE pg_stat_activity.datname = '%s'
			AND pid <> pg_backend_pid()
		`, testDBName))

		// Drop test database
		_, err = db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", testDBName))
		if err != nil {
			t.Logf("Failed to drop test database: %v", err)
		}
	})

	return client
}

// SetupTestDBWithSQL creates a test database and returns both an ent client and SQL connection.
// This is needed for tests that require direct SQL access (e.g., pgvector similarity search).
func SetupTestDBWithSQL(t *testing.T) (*ent.Client, *sql.DB) {
	t.Helper()

	// Connect to postgres database to create test database
	adminDSN := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=postgres sslmode=disable",
		testDBHost, testDBPort, testDBUser, testDBPassword)

	adminDB, err := sql.Open("postgres", adminDSN)
	if err != nil {
		t.Fatalf("Failed to connect to postgres: %v", err)
	}
	defer adminDB.Close()

	// Drop test database if it exists (cleanup from previous failed tests)
	_, _ = adminDB.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", testDBName))

	// Create fresh test database
	_, err = adminDB.Exec(fmt.Sprintf("CREATE DATABASE %s", testDBName))
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Connect to test database to enable pgvector
	testDSN := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		testDBHost, testDBPort, testDBUser, testDBPassword, testDBName)

	testDB, err := sql.Open("postgres", testDSN)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Enable pgvector extension
	_, err = testDB.Exec("CREATE EXTENSION IF NOT EXISTS vector")
	if err != nil {
		testDB.Close()
		t.Fatalf("Failed to enable pgvector extension: %v", err)
	}

	// Create ent client with schema migration
	client := enttest.Open(t, "postgres", testDSN)

	// Register cleanup function
	t.Cleanup(func() {
		client.Close()
		testDB.Close()

		// Reconnect to postgres database to drop test database
		db, err := sql.Open("postgres", adminDSN)
		if err != nil {
			t.Logf("Failed to connect for cleanup: %v", err)
			return
		}
		defer db.Close()

		// Terminate existing connections
		_, _ = db.Exec(fmt.Sprintf(`
			SELECT pg_terminate_backend(pg_stat_activity.pid)
			FROM pg_stat_activity
			WHERE pg_stat_activity.datname = '%s'
			AND pid <> pg_backend_pid()
		`, testDBName))

		// Drop test database
		_, err = db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", testDBName))
		if err != nil {
			t.Logf("Failed to drop test database: %v", err)
		}
	})

	return client, testDB
}

// CleanupTestData removes all data from the test database tables while preserving the schema.
// Useful for cleaning up between subtests.
func CleanupTestData(t *testing.T, client *ent.Client) {
	t.Helper()

	ctx := context.Background()

	// Delete in order respecting foreign key constraints
	if _, err := client.Relationship.Delete().Exec(ctx); err != nil {
		t.Logf("Warning: Failed to delete relationships: %v", err)
	}

	if _, err := client.DiscoveredEntity.Delete().Exec(ctx); err != nil {
		t.Logf("Warning: Failed to delete discovered entities: %v", err)
	}

	if _, err := client.Email.Delete().Exec(ctx); err != nil {
		t.Logf("Warning: Failed to delete emails: %v", err)
	}

	if _, err := client.SchemaPromotion.Delete().Exec(ctx); err != nil {
		t.Logf("Warning: Failed to delete schema promotions: %v", err)
	}
}

package test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// Force context import usage check
var _ context.Context

// TestContainer holds test database instance
type TestContainer struct {
	DB      *sqlx.DB
	Cleanup func()
}

// SetupTestDB creates a test database using testcontainers
func SetupTestDB(t *testing.T) *TestContainer {
	// Skip if testcontainers not available (CI might not support)
	if os.Getenv("SKIP_DOCKER_TESTS") == "1" {
		t.Skip("Skipping integration test - SKIP_DOCKER_TESTS=1")
	}

	t.Helper()

	ctx := t.Context()

	// Start PostgreSQL container
	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: "postgres:16-alpine",
			Env: map[string]string{
				"POSTGRES_USER":     "test",
				"POSTGRES_PASSWORD": "test",
				"POSTGRES_DB":       "test_db",
			},
			ExposedPorts: []string{"5432/tcp"},
			WaitingFor:   wait.ForLog("database system is ready to accept connections").WithStartupTimeout(60 * time.Second),
		},
		Started: true,
	}

	container, err := testcontainers.GenericContainer(ctx, req)
	require.NoError(t, err)

	t.Cleanup(func() {
		container.Terminate(ctx)
	})

	// Get connection details
	host, err := container.Host(ctx)
	require.NoError(t, err)
	port, err := container.MappedPort(ctx, "5432")
	require.NoError(t, err)

	// Connect to database
	connStr := fmt.Sprintf(
		"host=%s port=%s user=test password=test dbname=test_db sslmode=disable",
		host, port.Port(),
	)

	db, err := sqlx.Open("postgres", connStr)
	require.NoError(t, err)

	// Wait for DB to be ready
	require.Eventually(t, func() bool {
		return db.Ping() == nil
	}, 30*time.Second, 1*time.Second)

	return &TestContainer{
		DB: db,
		Cleanup: func() {
			db.Close()
		},
	}
}

// SetupSchema creates all tables for testing
func (tc *TestContainer) SetupSchema(t *testing.T, migrationsDir string) {
	// Apply migrations
	ctx := t.Context()
	if err := goose.Up(tc.DB.DB, migrationsDir); err != nil {
		t.Fatalf("Failed to apply migrations: %v", err)
	}

	t.Cleanup(func() {
		// Drop all tables
		tc.DB.MustExecContext(ctx, `
			DO $$ DECLARE
				r RECORD;
			BEGIN
				FOR r IN (SELECT tablename FROM pg_tables WHERE schemaname = 'public') LOOP
					EXECUTE 'DROP TABLE IF EXISTS ' || quote_ident(r.tablename) || ' CASCADE';
				END LOOP;
			END $$;
		`)
	})
}

// Helper to create a test user
func CreateTestUser(t *testing.T, db *sqlx.DB) uuid.UUID {
	t.Helper()

	userID := uuid.New()
	now := time.Now()

	_, err := db.ExecContext(
		t.Context(),
		`INSERT INTO users (user_id, username, email, password_hash, is_active, email_verified, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		userID, fmt.Sprintf("user_%s", userID[:8]), fmt.Sprintf("email_%s", userID[:8]), "hashed_password",
		true, false, now, now,
	)
	require.NoError(t, err)

	return userID
}

// Helper to create a test couple
func CreateTestCouple(t *testing.T, db *sqlx.DB, user1ID, user2ID uuid.UUID) uuid.UUID {
	t.Helper()

	coupleID := uuid.New()
	now := time.Now()

	_, err := db.ExecContext(
		t.Context(),
		`INSERT INTO couples (couple_id, user1_id, user2_id, status, relationship_type, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		coupleID, user1ID, user2ID, "active", "dating", now, now,
	)
	require.NoError(t, err)

	return coupleID
}

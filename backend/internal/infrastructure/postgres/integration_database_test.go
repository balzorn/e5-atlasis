package postgres

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

const defaultIntegrationDatabaseURL = "postgres://e5_atlasis:e5_atlasis_dev@127.0.0.1:54329/e5_atlasis_test?sslmode=disable"

func ensureIntegrationDatabase(t *testing.T, databaseURL string) {
	t.Helper()

	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		t.Fatalf("parse test database URL: %v", err)
	}

	databaseName := cfg.ConnConfig.Database
	if err := validateIntegrationDatabaseName(databaseName); err != nil {
		t.Fatal(err)
	}

	adminConfig := cfg.Copy()
	adminConfig.ConnConfig.Database = "postgres"

	adminDB, err := pgxpool.NewWithConfig(context.Background(), adminConfig)
	if err != nil {
		t.Fatalf("connect to PostgreSQL maintenance database: %v", err)
	}
	defer adminDB.Close()

	var exists bool
	if err := adminDB.QueryRow(
		context.Background(),
		"SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = $1)",
		databaseName,
	).Scan(&exists); err != nil {
		t.Fatalf("check integration database: %v", err)
	}

	if exists {
		return
	}

	quotedName := `"` + strings.ReplaceAll(databaseName, `"`, `""`) + `"`
	if _, err := adminDB.Exec(
		context.Background(),
		fmt.Sprintf("CREATE DATABASE %s", quotedName),
	); err != nil {
		t.Fatalf("create integration database %q: %v", databaseName, err)
	}
}

func TestValidateIntegrationDatabaseName(t *testing.T) {
	for _, databaseName := range []string{"e5_atlasis_test", "e5_atlasis_integration_test"} {
		if err := validateIntegrationDatabaseName(databaseName); err != nil {
			t.Fatalf("validateIntegrationDatabaseName(%q) error = %v", databaseName, err)
		}
	}

	for _, databaseName := range []string{"", "e5_atlasis", "e5_atlasis_dev", "postgres"} {
		if err := validateIntegrationDatabaseName(databaseName); err == nil {
			t.Fatalf("validateIntegrationDatabaseName(%q) error = nil", databaseName)
		}
	}
}

func validateIntegrationDatabaseName(databaseName string) error {
	if databaseName == "" {
		return fmt.Errorf("integration database name must not be empty")
	}

	if databaseName == "e5_atlasis" {
		return fmt.Errorf(
			"integration tests must not use development database %q; configure E5_ATLASIS_TEST_DATABASE_URL",
			databaseName,
		)
	}

	if !strings.HasSuffix(databaseName, "_test") {
		return fmt.Errorf("integration database %q must end with _test", databaseName)
	}

	return nil
}

func ensureIntegrationSchema(t *testing.T, db *DB) {
	t.Helper()

	ctx := context.Background()

	if _, err := db.pool.Exec(
		ctx,
		`CREATE TABLE IF NOT EXISTS e5_atlasis_test_schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
	); err != nil {
		t.Fatalf("create test migration table: %v", err)
	}

	_, sourceFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("locate integration test source")
	}

	migrationsDir := filepath.Clean(filepath.Join(filepath.Dir(sourceFile), "../../../migrations"))
	paths, err := filepath.Glob(filepath.Join(migrationsDir, "*.sql"))
	if err != nil {
		t.Fatalf("list migrations: %v", err)
	}

	sort.Strings(paths)
	if len(paths) == 0 {
		t.Fatalf("no migrations found in %s", migrationsDir)
	}

	for _, path := range paths {
		version := filepath.Base(path)

		var applied bool
		if err := db.pool.QueryRow(
			ctx,
			"SELECT EXISTS (SELECT 1 FROM e5_atlasis_test_schema_migrations WHERE version = $1)",
			version,
		).Scan(&applied); err != nil {
			t.Fatalf("check migration %s: %v", version, err)
		}

		if applied {
			continue
		}

		sql, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read migration %s: %v", version, err)
		}

		tx, err := db.pool.Begin(ctx)
		if err != nil {
			t.Fatalf("begin migration %s: %v", version, err)
		}

		if _, err := tx.Exec(ctx, string(sql)); err != nil {
			_ = tx.Rollback(ctx)
			t.Fatalf("apply migration %s: %v", version, err)
		}

		if _, err := tx.Exec(
			ctx,
			"INSERT INTO e5_atlasis_test_schema_migrations(version) VALUES ($1)",
			version,
		); err != nil {
			_ = tx.Rollback(ctx)
			t.Fatalf("record migration %s: %v", version, err)
		}

		if err := tx.Commit(ctx); err != nil {
			t.Fatalf("commit migration %s: %v", version, err)
		}
	}
}

func newIntegrationTestURL() string {
	if value := os.Getenv("E5_ATLASIS_TEST_DATABASE_URL"); value != "" {
		return value
	}

	return defaultIntegrationDatabaseURL
}

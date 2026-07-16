//#go:build integration

package handler

// Integration tests for the Apache Iceberg REST catalog driver.
// These tests require a running iceberg-rest + MinIO stack (see docker-compose.yml).
// Run with: docker compose exec -T app go test ./internal/application/handler/ -run TestIntegration_Iceberg -v
//
// DSN is read from the ICEBERG_DSN environment variable (loaded from .env).
// All tests guard themselves with testing.Short() so that they are skipped
// during unit-test runs (make test-unit / go test -short).
//
// Count semantics (identical to SQL drivers):
//   count = 1 (base "000000_000000_base" record) + N (user migrations applied)
// After "down all", count = 1 (base remains).

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/raoptimus/db-migrator.go/internal/infrastructure/dal/connection"
	"github.com/raoptimus/db-migrator.go/internal/infrastructure/dal/repository"
	infralog "github.com/raoptimus/db-migrator.go/internal/infrastructure/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// icebergDSN returns the Iceberg DSN from the environment.
func icebergDSN() string {
	return os.Getenv("ICEBERG_DSN")
}

// newIcebergOptions builds Options for the given fixtures directory and history table name.
// Using different table names per test group avoids shared state in the history namespace.
func newIcebergOptions(migrationsPath, tableName string) *Options {
	return &Options{
		DSN:         icebergDSN(),
		Directory:   migrationsPathAbs(migrationsPath),
		TableName:   tableName,
		Compact:     true,
		Interactive: false,
	}
}

// cleanupIceberg reverts all applied migrations. Errors are ignored because cleanup
// runs inside defer where t.FailNow() cannot propagate safely.
func cleanupIceberg(handlers *Handlers, createCmd func(string) *Command) {
	_ = handlers.Downgrade.Handle(createCmd("all"))
}

// assertIcebergMigrationsCount checks the number of migration records.
// The count includes the base migration record ("000000_000000_base") that is always
// inserted by the migration service when the history namespace is first created.
// Convention: count = 1 (base) + N (applied user migrations) — same as SQL drivers.
func assertIcebergMigrationsCount(
	t *testing.T,
	ctx context.Context,
	repo repository.Repository,
	expected int,
) {
	t.Helper()
	count, err := repo.MigrationsCount(ctx)
	require.NoError(t, err)
	assert.Equal(t, expected, count, "expected %d migration records (1 base + applied), got %d", expected, count)
}

// loadIcebergEnv loads .env once; silently ignores missing file (env may already be set).
func loadIcebergEnv(t *testing.T) {
	t.Helper()
	_ = godotenv.Load("../../../.env")
	dsn := icebergDSN()
	if dsn == "" {
		t.Skip("ICEBERG_DSN not set — skipping iceberg integration tests")
	}
}

// TestIntegration_Iceberg_UpDown verifies core up/down migration behaviour:
//   - up N applies N migrations
//   - down N reverts N migrations
//   - history / new reflect state correctly
//   - repeated up is idempotent
func TestIntegration_Iceberg_UpDown(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	loadIcebergEnv(t)

	opts := newIcebergOptions(os.Getenv("ICEBERG_MIGRATIONS_PATH"), "mig_updown")
	handlers := NewHandlers(opts, &infralog.NopLogger{})

	createCommand := func(arg string) *Command {
		args := NewMockArgs(t)
		args.EXPECT().First().Return(arg).Maybe()
		args.EXPECT().Present().Return(true).Maybe()
		return &Command{Args: args}
	}

	conn, err := connection.Try(opts.DSN, 1)
	require.NoError(t, err)
	defer conn.Close()

	repo, err := repository.New(conn, &repository.Options{TableName: opts.TableName})
	require.NoError(t, err)

	ctx := context.Background()

	// up 2 applies first 2 migrations.
	t.Run("up_2_and_down_2", func(t *testing.T) {
		cleanupIceberg(handlers, createCommand)
		defer func() { cleanupIceberg(handlers, createCommand) }()

		err = handlers.Upgrade.Handle(createCommand("2"))
		require.NoError(t, err)
		assertIcebergMigrationsCount(t, ctx, repo, 3) // base + 2 applied

		err = handlers.Downgrade.Handle(createCommand("2"))
		require.NoError(t, err)
		assertIcebergMigrationsCount(t, ctx, repo, 1) // base only
	})

	// history shows applied migrations; new shows pending ones.
	t.Run("history_and_new", func(t *testing.T) {
		cleanupIceberg(handlers, createCommand)
		defer func() { cleanupIceberg(handlers, createCommand) }()

		err = handlers.Upgrade.Handle(createCommand("3"))
		require.NoError(t, err)
		assertIcebergMigrationsCount(t, ctx, repo, 4) // base + 3 applied

		// history should not return error.
		err = handlers.History.Handle(createCommand("all"))
		require.NoError(t, err)

		// new should not return error.
		err = handlers.HistoryNew.Handle(createCommand("all"))
		require.NoError(t, err)
	})

	// Idempotency — applying all then calling up again is a no-op.
	t.Run("idempotent_up", func(t *testing.T) {
		cleanupIceberg(handlers, createCommand)
		defer func() { cleanupIceberg(handlers, createCommand) }()

		// Apply all migrations.
		err = handlers.Upgrade.Handle(createCommand(""))
		require.NoError(t, err)

		count1, err := repo.MigrationsCount(ctx)
		require.NoError(t, err)

		// Apply again — should be a no-op.
		err = handlers.Upgrade.Handle(createCommand(""))
		require.NoError(t, err)

		count2, err := repo.MigrationsCount(ctx)
		require.NoError(t, err)

		assert.Equal(t, count1, count2, "repeated up must be idempotent")
	})

	// down all reverts all user migrations (base record remains).
	t.Run("down_all", func(t *testing.T) {
		cleanupIceberg(handlers, createCommand)
		defer func() { cleanupIceberg(handlers, createCommand) }()

		err = handlers.Upgrade.Handle(createCommand("3"))
		require.NoError(t, err)

		err = handlers.Downgrade.Handle(createCommand("all"))
		require.NoError(t, err)

		assertIcebergMigrationsCount(t, ctx, repo, 1) // base only
	})
}

// TestIntegration_Iceberg_Redo verifies the redo command (revert + reapply).
func TestIntegration_Iceberg_Redo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	loadIcebergEnv(t)

	opts := newIcebergOptions(os.Getenv("ICEBERG_MIGRATIONS_PATH"), "mig_redo")
	handlers := NewHandlers(opts, &infralog.NopLogger{})

	createCommand := func(arg string) *Command {
		args := NewMockArgs(t)
		args.EXPECT().First().Return(arg).Maybe()
		args.EXPECT().Present().Return(true).Maybe()
		return &Command{Args: args}
	}

	conn, err := connection.Try(opts.DSN, 1)
	require.NoError(t, err)
	defer conn.Close()

	repo, err := repository.New(conn, &repository.Options{TableName: opts.TableName})
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("redo_1", func(t *testing.T) {
		cleanupIceberg(handlers, createCommand)
		defer func() { cleanupIceberg(handlers, createCommand) }()

		err = handlers.Upgrade.Handle(createCommand("2"))
		require.NoError(t, err)
		assertIcebergMigrationsCount(t, ctx, repo, 3) // base + 2

		// redo 1: revert last migration and reapply it.
		err = handlers.Redo.Handle(createCommand("1"))
		require.NoError(t, err)

		// Count should remain the same after redo.
		assertIcebergMigrationsCount(t, ctx, repo, 3) // base + 2 (unchanged)
	})
}

// TestIntegration_Iceberg_To verifies the to command in both upgrade and downgrade directions.
func TestIntegration_Iceberg_To(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	loadIcebergEnv(t)

	opts := newIcebergOptions(os.Getenv("ICEBERG_MIGRATIONS_PATH"), "mig_to")
	handlers := NewHandlers(opts, &infralog.NopLogger{})

	createCommand := func(arg string) *Command {
		args := NewMockArgs(t)
		args.EXPECT().First().Return(arg).Maybe()
		args.EXPECT().Present().Return(true).Maybe()
		return &Command{Args: args}
	}

	conn, err := connection.Try(opts.DSN, 1)
	require.NoError(t, err)
	defer conn.Close()

	repo, err := repository.New(conn, &repository.Options{TableName: opts.TableName})
	require.NoError(t, err)

	ctx := context.Background()

	// Fixture versions (from fixtures/iceberg/):
	//  250101_100000 - create_namespace
	//  250101_100100 - create_events
	//  250101_100200 - alter_add_column
	const firstVersion = "250101_100000"
	const thirdVersion = "250101_100200"

	// to target version (upgrade direction).
	t.Run("to_upgrade", func(t *testing.T) {
		cleanupIceberg(handlers, createCommand)
		defer func() { cleanupIceberg(handlers, createCommand) }()

		// Apply first migration only.
		err = handlers.Upgrade.Handle(createCommand("1"))
		require.NoError(t, err)
		assertIcebergMigrationsCount(t, ctx, repo, 2) // base + 1

		// Use 'to' to reach the third migration (upgrade).
		err = handlers.To.Handle(createCommand(thirdVersion))
		require.NoError(t, err)
		assertIcebergMigrationsCount(t, ctx, repo, 4) // base + 3
	})

	// to target version (downgrade direction).
	t.Run("to_downgrade", func(t *testing.T) {
		cleanupIceberg(handlers, createCommand)
		defer func() { cleanupIceberg(handlers, createCommand) }()

		// Apply first 3 migrations.
		err = handlers.Upgrade.Handle(createCommand("3"))
		require.NoError(t, err)
		assertIcebergMigrationsCount(t, ctx, repo, 4) // base + 3

		// Use 'to' to revert back to the first migration (downgrade).
		err = handlers.To.Handle(createCommand(firstVersion))
		require.NoError(t, err)
		assertIcebergMigrationsCount(t, ctx, repo, 2) // base + 1
	})

	// already at target — no changes.
	t.Run("to_already_at_target", func(t *testing.T) {
		cleanupIceberg(handlers, createCommand)
		defer func() { cleanupIceberg(handlers, createCommand) }()

		err = handlers.Upgrade.Handle(createCommand("1"))
		require.NoError(t, err)
		assertIcebergMigrationsCount(t, ctx, repo, 2) // base + 1

		err = handlers.To.Handle(createCommand(firstVersion))
		require.NoError(t, err)
		assertIcebergMigrationsCount(t, ctx, repo, 2) // no change
	})
}

// TestIntegration_Iceberg_Create verifies that the create command generates a migration file pair.
func TestIntegration_Iceberg_Create(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	loadIcebergEnv(t)

	// Use a temporary directory so we don't pollute the fixtures directory.
	tmpDir, err := os.MkdirTemp("", "iceberg-create-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	opts := &Options{
		DSN:         icebergDSN(),
		Directory:   tmpDir,
		TableName:   "migration",
		Interactive: false,
	}
	handlers := NewHandlers(opts, &infralog.NopLogger{})

	args := NewMockArgs(t)
	args.EXPECT().First().Return("add_events_table").Maybe()
	args.EXPECT().Present().Return(true).Maybe()
	cmd := &Command{Args: args}

	err = handlers.Create.Handle(cmd)
	require.NoError(t, err)

	// Both .up.sql and .down.sql files must exist.
	entries, err := os.ReadDir(tmpDir)
	require.NoError(t, err)

	var upFile, downFile string
	for _, e := range entries {
		name := e.Name()
		if strings.Contains(name, "add_events_table") && strings.HasSuffix(name, ".up.sql") {
			upFile = name
		}
		if strings.Contains(name, "add_events_table") && strings.HasSuffix(name, ".down.sql") {
			downFile = name
		}
	}

	assert.NotEmpty(t, upFile, "expected a .up.sql file to be created")
	assert.NotEmpty(t, downFile, "expected a .down.sql file to be created")
	// The create handler may generate .safe.up.sql or .up.sql depending on the builder.
	assert.True(t,
		strings.HasSuffix(upFile, "_add_events_table.up.sql") ||
			strings.HasSuffix(upFile, "_add_events_table.safe.up.sql"),
		"up file name format mismatch: %s", upFile)
	assert.True(t,
		strings.HasSuffix(downFile, "_add_events_table.down.sql") ||
			strings.HasSuffix(downFile, "_add_events_table.safe.down.sql"),
		"down file name format mismatch: %s", downFile)
}

// TestIntegration_Iceberg_ReferenceTable verifies a full-featured CREATE TABLE migration:
//   - Reference CREATE TABLE with namespace/columns/doc/partition/TBLPROPERTIES/COMMENT is applied.
//   - Leading catalog segment is stripped from qualified identifiers.
//   - Partition transform days(event_time) is parsed and applied.
//   - TIMESTAMP columns are mapped to timestamptz (with timezone, UTC).
//   - Column COMMENT clauses are preserved as field doc strings.
func TestIntegration_Iceberg_ReferenceTable(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	loadIcebergEnv(t)

	// Uses fixtures/iceberg/ — first 2 migrations create namespace + reference table.
	opts := newIcebergOptions(os.Getenv("ICEBERG_MIGRATIONS_PATH"), "mig_reftable")
	handlers := NewHandlers(opts, &infralog.NopLogger{})

	createCommand := func(arg string) *Command {
		args := NewMockArgs(t)
		args.EXPECT().First().Return(arg).Maybe()
		args.EXPECT().Present().Return(true).Maybe()
		return &Command{Args: args}
	}

	conn, err := connection.Try(opts.DSN, 1)
	require.NoError(t, err)
	defer conn.Close()

	repo, err := repository.New(conn, &repository.Options{TableName: opts.TableName})
	require.NoError(t, err)

	ctx := context.Background()

	// Apply 2 migrations (namespace + reference CREATE TABLE).
	t.Run("apply_and_revert_reference_table", func(t *testing.T) {
		cleanupIceberg(handlers, createCommand)
		defer func() { cleanupIceberg(handlers, createCommand) }()

		// Apply 2 migrations: create_namespace + create_events.
		err = handlers.Upgrade.Handle(createCommand("2"))
		require.NoError(t, err)
		assertIcebergMigrationsCount(t, ctx, repo, 3) // base + 2

		// Revert both.
		err = handlers.Downgrade.Handle(createCommand("2"))
		require.NoError(t, err)
		assertIcebergMigrationsCount(t, ctx, repo, 1) // base only
	})

	// Apply all fixtures (including ALTER ops and partition field changes).
	t.Run("apply_all_fixtures", func(t *testing.T) {
		cleanupIceberg(handlers, createCommand)
		defer func() { cleanupIceberg(handlers, createCommand) }()

		err = handlers.Upgrade.Handle(createCommand(""))
		require.NoError(t, err)

		// All 8 fixture migrations must be recorded (base + 8 = 9).
		assertIcebergMigrationsCount(t, ctx, repo, 9) // base + 8 fixtures

		// Revert all.
		err = handlers.Downgrade.Handle(createCommand("all"))
		require.NoError(t, err)
		assertIcebergMigrationsCount(t, ctx, repo, 1) // base only
	})
}

// TestIntegration_Iceberg_ReleaseRollback verifies the release and rollback commands.
func TestIntegration_Iceberg_ReleaseRollback(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	loadIcebergEnv(t)

	opts := newIcebergOptions(os.Getenv("ICEBERG_RELEASE_MIGRATIONS_PATH"), "mig_release")
	handlers := NewHandlers(opts, &infralog.NopLogger{})

	createCommand := func(arg string) *Command {
		args := NewMockArgs(t)
		args.EXPECT().First().Return(arg).Maybe()
		args.EXPECT().Present().Return(true).Maybe()
		return &Command{Args: args}
	}

	conn, err := connection.Try(opts.DSN, 1)
	require.NoError(t, err)
	defer conn.Close()

	repo, err := repository.New(conn, &repository.Options{TableName: opts.TableName})
	require.NoError(t, err)

	ctx := context.Background()

	// release applies all pending migrations with a shared apply_time.
	t.Run("release_applies_all", func(t *testing.T) {
		cleanupIceberg(handlers, createCommand)
		defer func() { cleanupIceberg(handlers, createCommand) }()

		err = handlers.Release.Handle(createCommand(""))
		require.NoError(t, err)
		assertIcebergMigrationsCount(t, ctx, repo, 3) // base + 2 release fixtures
	})

	// rollback reverts the latest batch (identified by max apply_time).
	t.Run("rollback_latest_release", func(t *testing.T) {
		cleanupIceberg(handlers, createCommand)
		defer func() { cleanupIceberg(handlers, createCommand) }()

		err = handlers.Release.Handle(createCommand(""))
		require.NoError(t, err)
		assertIcebergMigrationsCount(t, ctx, repo, 3) // base + 2

		err = handlers.Rollback.Handle(createCommand(""))
		require.NoError(t, err)
		assertIcebergMigrationsCount(t, ctx, repo, 1) // base only
	})

	// rollback on empty state is a no-op (no error).
	t.Run("rollback_empty", func(t *testing.T) {
		cleanupIceberg(handlers, createCommand)
		defer func() { cleanupIceberg(handlers, createCommand) }()

		err = handlers.Rollback.Handle(createCommand(""))
		require.NoError(t, err)
		assertIcebergMigrationsCount(t, ctx, repo, 1) // base only
	})

	// release → rollback → release cycle works correctly.
	t.Run("release_rollback_release", func(t *testing.T) {
		cleanupIceberg(handlers, createCommand)
		defer func() { cleanupIceberg(handlers, createCommand) }()

		err = handlers.Release.Handle(createCommand(""))
		require.NoError(t, err)

		err = handlers.Rollback.Handle(createCommand(""))
		require.NoError(t, err)

		err = handlers.Release.Handle(createCommand(""))
		require.NoError(t, err)

		assertIcebergMigrationsCount(t, ctx, repo, 3) // base + 2
	})
}

// TestIntegration_Iceberg_ReleaseBestEffortFail verifies that a mid-release failure leaves
// the successfully applied migrations recorded (best-effort, per-table atomicity).
func TestIntegration_Iceberg_ReleaseBestEffortFail(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	loadIcebergEnv(t)

	opts := newIcebergOptions(os.Getenv("ICEBERG_RELEASE_FAIL_MIGRATIONS_PATH"), "mig_relfail")
	handlers := NewHandlers(opts, &infralog.NopLogger{})

	createCommand := func(arg string) *Command {
		args := NewMockArgs(t)
		args.EXPECT().First().Return(arg).Maybe()
		args.EXPECT().Present().Return(true).Maybe()
		return &Command{Args: args}
	}

	conn, err := connection.Try(opts.DSN, 1)
	require.NoError(t, err)
	defer conn.Close()

	repo, err := repository.New(conn, &repository.Options{TableName: opts.TableName})
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("release_fails_at_second_migration", func(t *testing.T) {
		cleanupIceberg(handlers, createCommand)
		defer func() { cleanupIceberg(handlers, createCommand) }()

		// Release should fail because the second migration has TRUNCATE (unsupported DDL).
		err = handlers.Release.Handle(createCommand(""))
		require.Error(t, err, "release must fail when a migration contains unsupported DDL")

		// The first migration (CREATE NAMESPACE) was applied before the failure.
		// Iceberg is best-effort (per-table), so the first record must remain.
		// count = 1 base + 1 applied = 2
		assertIcebergMigrationsCount(t, ctx, repo, 2) // base + 1 applied before failure
	})
}

// TestIntegration_Iceberg_IrreversibleDown verifies that a down migration containing a
// type-narrowing operation is rejected by the catalog and the migration record remains applied.
func TestIntegration_Iceberg_IrreversibleDown(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	loadIcebergEnv(t)

	opts := newIcebergOptions(os.Getenv("ICEBERG_IRREVERSIBLE_MIGRATIONS_PATH"), "mig_irrev")
	handlers := NewHandlers(opts, &infralog.NopLogger{})

	createCommand := func(arg string) *Command {
		args := NewMockArgs(t)
		args.EXPECT().First().Return(arg).Maybe()
		args.EXPECT().Present().Return(true).Maybe()
		return &Command{Args: args}
	}

	conn, err := connection.Try(opts.DSN, 1)
	require.NoError(t, err)
	defer conn.Close()

	repo, err := repository.New(conn, &repository.Options{TableName: opts.TableName})
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("down_irreversible_type_narrowing", func(t *testing.T) {
		// Cleanup at start.
		cleanupIceberg(handlers, createCommand)
		// Cleanup at end: best-effort.
		defer func() { cleanupIceberg(handlers, createCommand) }()

		// Apply both up migrations: create namespace/table + widen int→long.
		err = handlers.Upgrade.Handle(createCommand(""))
		require.NoError(t, err)

		// base + 2 applied = 3
		assertIcebergMigrationsCount(t, ctx, repo, 3)

		// down attempt must fail (long→int narrowing is rejected by the catalog).
		err = handlers.Downgrade.Handle(createCommand("1"))
		require.Error(t, err, "down with type narrowing must be rejected by the catalog")

		// The migration record must remain (not removed on error) — still 3.
		assertIcebergMigrationsCount(t, ctx, repo, 3)
	})
}

// TestIntegration_Iceberg_Negative_UnsupportedDDL verifies that a migration containing
// TRUNCATE (unsupported DDL) fails with a clear error and is NOT marked as applied.
func TestIntegration_Iceberg_Negative_UnsupportedDDL(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	loadIcebergEnv(t)

	// Use a temporary directory with an inline migration containing TRUNCATE.
	tmpDir := t.TempDir()

	// First create a namespace so TRUNCATE has something to reference.
	require.NoError(t, os.WriteFile(
		filepath.Join(tmpDir, "260101_100000_create_ns.up.sql"),
		[]byte("CREATE NAMESPACE raw_neg_ddl;\n"),
		0o600,
	))
	require.NoError(t, os.WriteFile(
		filepath.Join(tmpDir, "260101_100000_create_ns.down.sql"),
		[]byte("DROP NAMESPACE raw_neg_ddl;\n"),
		0o600,
	))
	require.NoError(t, os.WriteFile(
		filepath.Join(tmpDir, "260101_100100_truncate.up.sql"),
		[]byte("TRUNCATE TABLE raw_neg_ddl.something;\n"),
		0o600,
	))
	require.NoError(t, os.WriteFile(
		filepath.Join(tmpDir, "260101_100100_truncate.down.sql"),
		[]byte("-- no-op\n"),
		0o600,
	))

	opts := &Options{
		DSN:         icebergDSN(),
		Directory:   tmpDir,
		TableName:   "mig_neg_ddl",
		Interactive: false,
	}
	handlers := NewHandlers(opts, &infralog.NopLogger{})

	createCommand := func(arg string) *Command {
		args := NewMockArgs(t)
		args.EXPECT().First().Return(arg).Maybe()
		args.EXPECT().Present().Return(true).Maybe()
		return &Command{Args: args}
	}

	conn, err := connection.Try(opts.DSN, 1)
	require.NoError(t, err)
	defer conn.Close()

	repo, err := repository.New(conn, &repository.Options{TableName: opts.TableName})
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("unsupported_ddl_fails", func(t *testing.T) {
		cleanupIceberg(handlers, createCommand)
		defer func() { cleanupIceberg(handlers, createCommand) }()

		// up should fail on the TRUNCATE statement.
		err = handlers.Upgrade.Handle(createCommand(""))
		require.Error(t, err, "up with unsupported DDL must return an error")

		// Only the first migration (CREATE NAMESPACE) should be applied.
		// base + 1 = 2
		assertIcebergMigrationsCount(t, ctx, repo, 2)
	})
}

// TestIntegration_Iceberg_Negative_UnknownTransform verifies that a migration with an
// unknown partition transform fails with a clear error.
func TestIntegration_Iceberg_Negative_UnknownTransform(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	loadIcebergEnv(t)

	tmpDir := t.TempDir()

	require.NoError(t, os.WriteFile(
		filepath.Join(tmpDir, "260101_100000_ns.up.sql"),
		[]byte("CREATE NAMESPACE raw_neg_transform;\n"),
		0o600,
	))
	require.NoError(t, os.WriteFile(
		filepath.Join(tmpDir, "260101_100000_ns.down.sql"),
		[]byte("DROP NAMESPACE raw_neg_transform;\n"),
		0o600,
	))
	// weeks(ts) is not a supported Iceberg partition transform.
	require.NoError(t, os.WriteFile(
		filepath.Join(tmpDir, "260101_100100_bad_transform.up.sql"),
		[]byte("CREATE TABLE raw_neg_transform.t (ts timestamp) USING iceberg PARTITIONED BY (weeks(ts));\n"),
		0o600,
	))
	require.NoError(t, os.WriteFile(
		filepath.Join(tmpDir, "260101_100100_bad_transform.down.sql"),
		[]byte("DROP TABLE raw_neg_transform.t;\n"),
		0o600,
	))

	opts := &Options{
		DSN:         icebergDSN(),
		Directory:   tmpDir,
		TableName:   "mig_neg_transform",
		Interactive: false,
	}
	handlers := NewHandlers(opts, &infralog.NopLogger{})

	createCommand := func(arg string) *Command {
		args := NewMockArgs(t)
		args.EXPECT().First().Return(arg).Maybe()
		args.EXPECT().Present().Return(true).Maybe()
		return &Command{Args: args}
	}

	conn, err := connection.Try(opts.DSN, 1)
	require.NoError(t, err)
	defer conn.Close()

	repo, err := repository.New(conn, &repository.Options{TableName: opts.TableName})
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("unknown_transform_fails", func(t *testing.T) {
		cleanupIceberg(handlers, createCommand)
		defer func() { cleanupIceberg(handlers, createCommand) }()

		err = handlers.Upgrade.Handle(createCommand(""))
		require.Error(t, err, "up with unknown partition transform must return an error")

		// Only the first (namespace) migration applied — base + 1 = 2.
		assertIcebergMigrationsCount(t, ctx, repo, 2)
	})
}

// TestIntegration_Iceberg_Negative_NoNamespace verifies that a migration with a
// single-part table identifier (no namespace) fails with a clear error.
func TestIntegration_Iceberg_Negative_NoNamespace(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	loadIcebergEnv(t)

	tmpDir := t.TempDir()

	// Table identifier with no namespace: "events" alone is invalid.
	require.NoError(t, os.WriteFile(
		filepath.Join(tmpDir, "260101_100000_no_ns.up.sql"),
		[]byte("CREATE TABLE events (id long);\n"),
		0o600,
	))
	require.NoError(t, os.WriteFile(
		filepath.Join(tmpDir, "260101_100000_no_ns.down.sql"),
		[]byte("DROP TABLE events;\n"),
		0o600,
	))

	opts := &Options{
		DSN:         icebergDSN(),
		Directory:   tmpDir,
		TableName:   "mig_neg_nons",
		Interactive: false,
	}
	handlers := NewHandlers(opts, &infralog.NopLogger{})

	createCommand := func(arg string) *Command {
		args := NewMockArgs(t)
		args.EXPECT().First().Return(arg).Maybe()
		args.EXPECT().Present().Return(true).Maybe()
		return &Command{Args: args}
	}

	conn, err := connection.Try(opts.DSN, 1)
	require.NoError(t, err)
	defer conn.Close()

	repo, err := repository.New(conn, &repository.Options{TableName: opts.TableName})
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("no_namespace_fails", func(t *testing.T) {
		cleanupIceberg(handlers, createCommand)
		defer func() { cleanupIceberg(handlers, createCommand) }()

		err = handlers.Upgrade.Handle(createCommand(""))
		require.Error(t, err, "CREATE TABLE without namespace must return an error")

		// No user migrations applied — only the base record.
		assertIcebergMigrationsCount(t, ctx, repo, 1) // base only
	})
}

// TestIntegration_Iceberg_Connection verifies connectivity behaviour:
//   - successful ping with a valid DSN
//   - connection failure with an unreachable host
func TestIntegration_Iceberg_Connection(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	loadIcebergEnv(t)

	// valid DSN connects successfully.
	t.Run("valid_dsn_connects", func(t *testing.T) {
		conn, err := connection.Try(icebergDSN(), 1)
		require.NoError(t, err, "connection with valid Iceberg DSN must succeed")
		require.NoError(t, conn.Ping(), "Ping must succeed for valid Iceberg connection")
		_ = conn.Close()
	})

	// unreachable host returns a connection error.
	t.Run("unreachable_host_fails", func(t *testing.T) {
		badDSN := "iceberg://127.0.0.1:19999/iceberg"
		_, err := connection.Try(badDSN, 1)
		assert.Error(t, err, "connection to unreachable host must return an error")
	})
}

// TestIntegration_Iceberg_SecretMasking verifies that s3.secret-access-key is masked in log output.
func TestIntegration_Iceberg_SecretMasking(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	loadIcebergEnv(t)

	// capturingLogger records all log output in a buffer.
	capture := &bytes.Buffer{}

	logFn := func(format string, args ...any) {
		if len(args) > 0 {
			capture.WriteString(fmt.Sprintf(format, args...))
		} else {
			capture.WriteString(format)
		}
		capture.WriteByte('\n')
	}

	cl := &testCapturingLogger{fn: logFn}

	opts := newIcebergOptions(os.Getenv("ICEBERG_MIGRATIONS_PATH"), "mig_masking")
	handlers := NewHandlers(opts, cl)

	createCommand := func(arg string) *Command {
		args := NewMockArgs(t)
		args.EXPECT().First().Return(arg).Maybe()
		args.EXPECT().Present().Return(true).Maybe()
		return &Command{Args: args}
	}

	cleanupIceberg(handlers, createCommand)
	defer func() { cleanupIceberg(handlers, createCommand) }()

	// Apply a couple of migrations to generate log output.
	_ = handlers.Upgrade.Handle(createCommand("2"))

	logOutput := capture.String()

	// The S3 secret access key must not appear in logs.
	const secret = "minioadmin"
	assert.NotContains(t, logOutput, secret,
		"s3.secret-access-key must be masked in log output")
}

// testCapturingLogger implements Logger by forwarding all messages to a capture function.
type testCapturingLogger struct {
	fn func(format string, args ...any)
}

func (l *testCapturingLogger) Infof(format string, args ...any)    { l.fn(format, args...) }
func (l *testCapturingLogger) Info(args ...any)                    { l.fn(fmt.Sprint(args...)) }
func (l *testCapturingLogger) Successf(format string, args ...any) { l.fn(format, args...) }
func (l *testCapturingLogger) Success(args ...any)                 { l.fn(fmt.Sprint(args...)) }
func (l *testCapturingLogger) Warnf(format string, args ...any)    { l.fn(format, args...) }
func (l *testCapturingLogger) Warn(args ...any)                    { l.fn(fmt.Sprint(args...)) }
func (l *testCapturingLogger) Errorf(format string, args ...any)   { l.fn(format, args...) }
func (l *testCapturingLogger) Error(args ...any)                   { l.fn(fmt.Sprint(args...)) }
func (l *testCapturingLogger) Fatalf(format string, args ...any)   { l.fn(format, args...) }
func (l *testCapturingLogger) Fatal(args ...any)                   { l.fn(fmt.Sprint(args...)) }

// TestIntegration_Iceberg_CatalogPrefixStripped verifies that when the leading segment of a
// table identifier equals the warehouse name, it is stripped:
// "iceberg.raw.events" → namespace=[raw], table=events.
func TestIntegration_Iceberg_CatalogPrefixStripped(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	loadIcebergEnv(t)

	tmpDir := t.TempDir()

	// The DSN warehouse is "iceberg". The leading "iceberg" in "iceberg.raw.catalog_test"
	// must be treated as the catalog name and stripped.
	require.NoError(t, os.WriteFile(
		filepath.Join(tmpDir, "260101_100000_ns.up.sql"),
		[]byte("CREATE NAMESPACE raw;\n"),
		0o600,
	))
	require.NoError(t, os.WriteFile(
		filepath.Join(tmpDir, "260101_100000_ns.down.sql"),
		[]byte("DROP NAMESPACE raw;\n"),
		0o600,
	))
	require.NoError(t, os.WriteFile(
		filepath.Join(tmpDir, "260101_100100_catalog_table.up.sql"),
		// "iceberg" matches warehouse → stripped; table goes into namespace "raw".
		[]byte("CREATE TABLE iceberg.raw.catalog_test (id long);\n"),
		0o600,
	))
	require.NoError(t, os.WriteFile(
		filepath.Join(tmpDir, "260101_100100_catalog_table.down.sql"),
		[]byte("DROP TABLE iceberg.raw.catalog_test;\n"),
		0o600,
	))

	opts := &Options{
		DSN:         icebergDSN(),
		Directory:   tmpDir,
		TableName:   "mig_catalog_prefix",
		Interactive: false,
	}
	handlers := NewHandlers(opts, &infralog.NopLogger{})

	createCommand := func(arg string) *Command {
		args := NewMockArgs(t)
		args.EXPECT().First().Return(arg).Maybe()
		args.EXPECT().Present().Return(true).Maybe()
		return &Command{Args: args}
	}

	conn, err := connection.Try(opts.DSN, 1)
	require.NoError(t, err)
	defer conn.Close()

	repo, err := repository.New(conn, &repository.Options{TableName: opts.TableName})
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("catalog_prefix_stripped", func(t *testing.T) {
		cleanupIceberg(handlers, createCommand)
		defer func() { cleanupIceberg(handlers, createCommand) }()

		err = handlers.Upgrade.Handle(createCommand(""))
		require.NoError(t, err, "table with leading warehouse prefix must be created successfully")

		assertIcebergMigrationsCount(t, ctx, repo, 3) // base + 2

		err = handlers.Downgrade.Handle(createCommand("all"))
		require.NoError(t, err)
		assertIcebergMigrationsCount(t, ctx, repo, 1) // base only
	})
}

// TestIntegration_Iceberg_MultiLevelNamespace verifies multi-level namespace support.
// Note: the iceberg-rest implementation in the test stack may not support multi-level
// namespaces (returns an HTML error page). This test documents the current behaviour
// and skips gracefully when multi-level namespaces are unsupported.
func TestIntegration_Iceberg_MultiLevelNamespace(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	loadIcebergEnv(t)

	tmpDir := t.TempDir()

	// Multi-level namespace: bronze.raw.
	require.NoError(t, os.WriteFile(
		filepath.Join(tmpDir, "260101_100000_ns_bronze.up.sql"),
		[]byte("CREATE NAMESPACE bronze;\n"),
		0o600,
	))
	require.NoError(t, os.WriteFile(
		filepath.Join(tmpDir, "260101_100000_ns_bronze.down.sql"),
		[]byte("DROP NAMESPACE bronze;\n"),
		0o600,
	))
	require.NoError(t, os.WriteFile(
		filepath.Join(tmpDir, "260101_100100_ns_bronze_raw.up.sql"),
		[]byte("CREATE NAMESPACE bronze.raw;\n"),
		0o600,
	))
	require.NoError(t, os.WriteFile(
		filepath.Join(tmpDir, "260101_100100_ns_bronze_raw.down.sql"),
		[]byte("DROP NAMESPACE bronze.raw;\n"),
		0o600,
	))
	require.NoError(t, os.WriteFile(
		filepath.Join(tmpDir, "260101_100200_create_table.up.sql"),
		[]byte("CREATE TABLE bronze.raw.events (id long);\n"),
		0o600,
	))
	require.NoError(t, os.WriteFile(
		filepath.Join(tmpDir, "260101_100200_create_table.down.sql"),
		[]byte("DROP TABLE bronze.raw.events;\n"),
		0o600,
	))

	opts := &Options{
		DSN:         icebergDSN(),
		Directory:   tmpDir,
		TableName:   "mig_multilevel",
		Interactive: false,
	}
	handlers := NewHandlers(opts, &infralog.NopLogger{})

	createCommand := func(arg string) *Command {
		args := NewMockArgs(t)
		args.EXPECT().First().Return(arg).Maybe()
		args.EXPECT().Present().Return(true).Maybe()
		return &Command{Args: args}
	}

	conn, err := connection.Try(opts.DSN, 1)
	require.NoError(t, err)
	defer conn.Close()

	repo, err := repository.New(conn, &repository.Options{TableName: opts.TableName})
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("multi_level_namespace", func(t *testing.T) {
		cleanupIceberg(handlers, createCommand)
		defer func() { cleanupIceberg(handlers, createCommand) }()

		err = handlers.Upgrade.Handle(createCommand(""))
		if err != nil {
			// Multi-level namespaces may not be supported by the test catalog implementation.
			// Document the finding and skip rather than fail hard.
			t.Logf("multi-level namespace: catalog returned error (may be unsupported by test stack): %v", err)
			t.Skip("multi-level namespace not supported by this catalog implementation")
		}

		assertIcebergMigrationsCount(t, ctx, repo, 4) // base + 3

		// Revert everything.
		err = handlers.Downgrade.Handle(createCommand("all"))
		require.NoError(t, err)
		assertIcebergMigrationsCount(t, ctx, repo, 1) // base only
	})
}

// TestIntegration_Iceberg_IfNotExistsIdempotent reproduces the reported bug where
// `release` with a duplicate `CREATE TABLE IF NOT EXISTS` failed with AlreadyExistsException,
// and verifies the whole IF [NOT] EXISTS family is now idempotent end-to-end against the catalog:
//   - CREATE TABLE IF NOT EXISTS on an existing table is skipped (not an error);
//   - DROP TABLE IF EXISTS on an already-dropped table is skipped;
//   - DROP NAMESPACE IF EXISTS drops the namespace.
func TestIntegration_Iceberg_IfNotExistsIdempotent(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	loadIcebergEnv(t)

	tmpDir := t.TempDir()
	write := func(name, content string) {
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, name), []byte(content), 0o600))
	}

	// Namespace (bare name, like the reference fixtures).
	write("260301_100000_ns.up.sql", "CREATE NAMESPACE IF NOT EXISTS idemp;\n")
	write("260301_100000_ns.down.sql", "DROP NAMESPACE IF EXISTS idemp;\n")
	// First table creation. DSN warehouse = iceberg, so the leading segment is stripped:
	// iceberg.idemp.widgets => namespace=[idemp], table=widgets.
	write("260301_100100_tbl.up.sql", "CREATE TABLE IF NOT EXISTS iceberg.idemp.widgets (id long);\n")
	write("260301_100100_tbl.down.sql", "DROP TABLE IF EXISTS iceberg.idemp.widgets;\n")
	// Duplicate creation of the SAME table: must be skipped idempotently, not fail.
	write("260301_100200_tbl_dup.up.sql", "CREATE TABLE IF NOT EXISTS iceberg.idemp.widgets (id long);\n")
	// Its down drops the table; the previous migration's down then hits DROP TABLE IF EXISTS
	// on an already-dropped table, exercising the idempotent-drop path.
	write("260301_100200_tbl_dup.down.sql", "DROP TABLE IF EXISTS iceberg.idemp.widgets;\n")

	opts := &Options{
		DSN:         icebergDSN(),
		Directory:   tmpDir,
		TableName:   "mig_ifnotexists",
		Compact:     true,
		Interactive: false,
	}
	handlers := NewHandlers(opts, &infralog.NopLogger{})

	createCommand := func(arg string) *Command {
		args := NewMockArgs(t)
		args.EXPECT().First().Return(arg).Maybe()
		args.EXPECT().Present().Return(true).Maybe()
		return &Command{Args: args}
	}

	conn, err := connection.Try(opts.DSN, 1)
	require.NoError(t, err)
	defer conn.Close()

	repo, err := repository.New(conn, &repository.Options{TableName: opts.TableName})
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("release_with_duplicate_create_table_is_idempotent", func(t *testing.T) {
		cleanupIceberg(handlers, createCommand)
		defer func() { cleanupIceberg(handlers, createCommand) }()

		// release applies all three migrations in one batch. The duplicate
		// CREATE TABLE IF NOT EXISTS must be skipped, not raise AlreadyExistsException.
		err = handlers.Release.Handle(createCommand(""))
		require.NoError(t, err, "duplicate CREATE TABLE IF NOT EXISTS must be skipped, not fail")
		assertIcebergMigrationsCount(t, ctx, repo, 4) // base + 3

		// down all reverts in reverse order; the second DROP TABLE IF EXISTS lands on an
		// already-dropped table (idempotent skip), then DROP NAMESPACE IF EXISTS removes it.
		err = handlers.Downgrade.Handle(createCommand("all"))
		require.NoError(t, err)
		assertIcebergMigrationsCount(t, ctx, repo, 1) // base only
	})
}

// TestIntegration_Iceberg_WriteOrderedBy verifies the ALTER TABLE … WRITE ORDERED BY / WRITE
// UNORDERED sort-order operations end-to-end against the real catalog: a sort order with a plain
// column and a transform is committed via CommitTable, and the down migration resets it with
// WRITE UNORDERED.
func TestIntegration_Iceberg_WriteOrderedBy(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	loadIcebergEnv(t)

	tmpDir := t.TempDir()
	write := func(name, content string) {
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, name), []byte(content), 0o600))
	}

	write("260401_100000_ns.up.sql", "CREATE NAMESPACE IF NOT EXISTS sortns;\n")
	write("260401_100000_ns.down.sql", "DROP NAMESPACE IF EXISTS sortns;\n")
	write("260401_100100_tbl.up.sql",
		"CREATE TABLE iceberg.sortns.orders (id long, amount long, created_at timestamp);\n")
	write("260401_100100_tbl.down.sql", "DROP TABLE IF EXISTS iceberg.sortns.orders;\n")
	// Sort order over a plain column (with explicit null ordering) and a bucket transform.
	write("260401_100200_sort.up.sql",
		"ALTER TABLE iceberg.sortns.orders WRITE ORDERED BY created_at DESC NULLS LAST, bucket(8, id);\n")
	// Reverting a sort order is not automatic — reset to unsorted.
	write("260401_100200_sort.down.sql", "ALTER TABLE iceberg.sortns.orders WRITE UNORDERED;\n")

	opts := &Options{
		DSN:         icebergDSN(),
		Directory:   tmpDir,
		TableName:   "mig_sortorder",
		Compact:     true,
		Interactive: false,
	}
	handlers := NewHandlers(opts, &infralog.NopLogger{})

	createCommand := func(arg string) *Command {
		args := NewMockArgs(t)
		args.EXPECT().First().Return(arg).Maybe()
		args.EXPECT().Present().Return(true).Maybe()
		return &Command{Args: args}
	}

	conn, err := connection.Try(opts.DSN, 1)
	require.NoError(t, err)
	defer conn.Close()

	repo, err := repository.New(conn, &repository.Options{TableName: opts.TableName})
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("apply_sort_order_and_reset", func(t *testing.T) {
		cleanupIceberg(handlers, createCommand)
		defer func() { cleanupIceberg(handlers, createCommand) }()

		// up applies namespace + table + WRITE ORDERED BY (commits sort order via CommitTable).
		err = handlers.Upgrade.Handle(createCommand(""))
		require.NoError(t, err, "WRITE ORDERED BY must commit the sort order successfully")
		assertIcebergMigrationsCount(t, ctx, repo, 4) // base + 3

		// down all reverts: WRITE UNORDERED (reset), DROP TABLE, DROP NAMESPACE.
		err = handlers.Downgrade.Handle(createCommand("all"))
		require.NoError(t, err)
		assertIcebergMigrationsCount(t, ctx, repo, 1) // base only
	})
}

// Compile-time verification that *testCapturingLogger satisfies Logger.
// Uses the Logger type alias defined in dependency.go (= log.Logger interface).
var _ Logger = (*testCapturingLogger)(nil)

// Keep time import used.
var _ = time.Second

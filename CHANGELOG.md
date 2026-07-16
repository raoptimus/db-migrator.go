# Changelog

## v1.8.2

### Fixes
- **iceberg**: use GET instead of HEAD for namespace existence (some REST servers reject HEAD with 400).

### Improvements
- **iceberg**: support `CREATE NAMESPACE IF NOT EXISTS` (idempotent namespace creation).

## v1.6.0

### New Commands

#### `release` - Atomic Batch Apply
Applies ALL pending migrations atomically within a single database transaction.

- All migrations in a release share the same `apply_time` value, enabling batch identification for later rollback
- If any migration fails, the entire batch is rolled back automatically
- Individual `.safe` migrations skip their inner transaction when inside a release (outer transaction provides atomicity)

```bash
DSN="postgres://user:pass@localhost:5432/db" db-migrator release
```

#### `rollback` - Atomic Batch Revert
Reverts all migrations from the latest release batch, identified by `MAX(apply_time)`.

- Pre-checks that all `.down.sql` files exist before starting the rollback
- Wraps all reverts in a single transaction for atomicity
- If no release batch is found, shows an informational message

```bash
DSN="postgres://user:pass@localhost:5432/db" db-migrator rollback
```

### New Repository Methods
- `InsertMigrationWithApplyTime` - insert migration record with explicit apply time (used by `release` to assign shared batch timestamp)
- `MigrationsByMaxApplyTime` - query migrations belonging to the latest release batch

Implemented for all 4 database drivers: PostgreSQL, MySQL, ClickHouse, Tarantool.

### New Domain Service Methods
- `ApplyFileWithApplyTime` - apply migration file with explicit apply time (reuses `applyFileCore` extracted from `ApplyFile`)
- `LatestReleaseMigrations` - retrieve and map latest release batch migrations
- `ExecInTransaction` - execute a function within a database transaction
- `FileExists` - check whether a migration file exists

### Internal Improvements
- Refactored `ApplyFile` into `applyFileCore` + wrapper for code reuse between `ApplyFile` and `ApplyFileWithApplyTime`
- Added 23 new unit tests covering release handler, rollback handler, and new domain service methods
- Updated documentation in CLAUDE.md and README.md

## v1.5.0

- Implementation of `to` command - bidirectional migration to specific version
- Supports 4 version formats: timestamp, full name, datetime string, UNIX timestamp

## v1.4.0

- Implementation of dry run mode (`DRY_RUN=true`)

## v1.3.0

- Refactor to Clean Architecture
- Credential masking in log output
- SQL identifier validation
- Comprehensive unit test coverage

## v1.2.0

- Tarantool database driver support

## v1.1.0

- ClickHouse cluster and replication support
- MySQL driver support

## v1.0.0

- Initial release
- PostgreSQL and ClickHouse support
- Migration file management (up, down, redo, create, history, new)

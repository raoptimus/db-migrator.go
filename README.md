[![Test](https://github.com/raoptimus/db-migrator.go/workflows/Test/badge.svg)](https://github.com/raoptimus/db-migrator.go/actions)
[![Coverage](https://github.com/raoptimus/db-migrator.go/wiki/coverage.svg)](https://raw.githack.com/wiki/raoptimus/db-migrator.go/coverage.html)
[![GitHub Release](https://img.shields.io/github/release/raoptimus/db-migrator.go.svg)](https://github.com/raoptimus/db-migrator.go/releases)
[![Go Reference](https://pkg.go.dev/badge/github.com/raoptimus/db-migrator.go.svg)](https://pkg.go.dev/github.com/raoptimus/db-migrator.go)
[![Go Report Card](https://goreportcard.com/badge/github.com/raoptimus/db-migrator.go)](https://goreportcard.com/report/github.com/raoptimus/db-migrator.go)
[![License: BSD-3-Clause](https://img.shields.io/badge/License-BSD_3--Clause-blue.svg)](LICENSE)

# db-migrator.go
Database Migration tool in CLI on Golang that allows you to keep track of database changes in terms of database migrations which are version-controlled.
The db migration tool currently supports the following db drivers:
- clickhouse
- postgres
- mysql
- tarantool db
- apache iceberg (via REST Catalog)

## Installation

### Homebrew (macOS / Linux)
```bash
brew install --cask raoptimus/tap/db-migrator
```

### Debian / Ubuntu (.deb)
```bash
# Replace VERSION and ARCH (amd64 or arm64) with the desired release
curl -sLO https://github.com/raoptimus/db-migrator.go/releases/latest/download/db-migrator_VERSION_linux_amd64.deb
sudo dpkg -i db-migrator_*.deb
```

### RHEL / Fedora (.rpm) and Alpine (.apk)
Download the matching `.rpm` / `.apk` from the
[latest release](https://github.com/raoptimus/db-migrator.go/releases/latest) and install with
`rpm -i` / `apk add --allow-untrusted`.

### Pre-built binaries
Grab a `tar.gz` (or `.zip` for Windows) for your OS/arch from the
[releases page](https://github.com/raoptimus/db-migrator.go/releases/latest), extract it, and put
`db-migrator` on your `PATH`.

### Go install
```bash
go install github.com/raoptimus/db-migrator.go/cmd/db-migrator@latest
```

### Docker
```bash
docker run --rm -e DSN -e MIGRATION_PATH -v "$PWD/migrations:/migrations" raoptimus/db-migrator:latest up
```

### Helm (Kubernetes)
See the reusable chart in [`charts/db-migrator`](charts/db-migrator).

## Database Connection Examples

### PostgreSQL
```bash
DSN="postgres://username:password@localhost:5432/mydb?sslmode=disable" \
MIGRATION_PATH=./migrations \
db-migrator up
```

### MySQL
```bash
DSN="mysql://username:password@localhost:3306/mydb" \
MIGRATION_PATH=./migrations \
db-migrator up
```

### ClickHouse
```bash
# Single node
DSN="clickhouse://username:password@localhost:9000/mydb?sslmode=disable&compress=true" \
MIGRATION_PATH=./migrations \
db-migrator up

# Cluster mode
DSN="clickhouse://username:password@localhost:9000/mydb?sslmode=disable&compress=true" \
MIGRATION_PATH=./migrations \
MIGRATION_CLUSTER_NAME=my_cluster \
MIGRATION_REPLICATED=true \
db-migrator up
```

### Tarantool
```bash
DSN="tarantool://username:password@localhost:3301/mydb" \
MIGRATION_PATH=./migrations \
db-migrator up
```

### Apache Iceberg (REST Catalog)

Iceberg uses a REST Catalog endpoint instead of a direct database connection.
The DSN format is `iceberg://host:port/<warehouse>?<auth-params>[&<storage-params>]`.

**Bearer token authentication:**
```bash
DSN="iceberg://localhost:8181/warehouse?token=my-bearer-token&s3.endpoint=http://localhost:9000&s3.access-key-id=admin&s3.secret-access-key=password&s3.region=us-east-1&s3.force-virtual-addressing=false" \
MIGRATION_PATH=./migrations \
db-migrator up
```

**OAuth2 client-credentials authentication:**
```bash
DSN="iceberg://localhost:8181/warehouse?credential=client-id:client-secret&oauth2_server_uri=https://auth.example.com/oauth/token&scope=catalog&s3.endpoint=http://localhost:9000&s3.access-key-id=admin&s3.secret-access-key=password&s3.region=us-east-1&s3.force-virtual-addressing=false" \
MIGRATION_PATH=./migrations \
db-migrator up
```

**DSN parameters:**

| Parameter | Description |
|-----------|-------------|
| `token=<bearer>` | Bearer token auth (mutually exclusive with `credential`) |
| `credential=<id:secret>` | OAuth2 client credentials (requires `oauth2_server_uri`) |
| `oauth2_server_uri=<url>` | OAuth2 token endpoint URL |
| `scope=<s>` | OAuth2 scope (optional, used with `credential`) |
| `prefix=<p>` | Catalog path prefix (optional) |
| `secure=true` or `sslmode=require` | Use HTTPS (default: HTTP) |
| `s3.endpoint=<url>` | S3/MinIO endpoint URL |
| `s3.access-key-id=<k>` | S3/MinIO access key |
| `s3.secret-access-key=<sk>` | S3/MinIO secret key |
| `s3.session-token=<t>` | S3/MinIO session token (optional) |
| `s3.region=<r>` | S3/MinIO region |
| `s3.force-virtual-addressing=false` | Use path-style URLs (required for MinIO) |

> **Why S3 parameters are needed:** Iceberg schema-evolution commands (`ALTER TABLE`) require the
> client to load the table's `metadata.json` directly from object storage. The REST server handles
> `CREATE TABLE` and `DROP TABLE` on its own, but `ALTER` operations call `iceberg-go` which reads
> S3/MinIO storage client-side. Without S3 credentials the driver returns an I/O error. Provide S3
> parameters whenever you run `ALTER TABLE` migrations.

---

## Creating Migrations
To create a new migration, run the following command:  
`MIGRATION_PATH=./migrations db-migrator create <name>`

The required name argument gives a brief description about the new migration.  
For example, if the migration is about creating a new table named news, you may use the name `create_news_table` and run the following command:  
`db-migrator create create_news_table`

The above command will create a new sql file named 200101_232501_create_news_table.safe.up.sql in the ./migrations directory.

The migration file name is automatically generated in the format of <YYMMDD_HHMMSS>_<Name>.<Safe>.<Action>.sql, where
- <YYMMDD_HHMMSS> refers to the UTC datetime at which the migration creation command is executed.
- <Name> is the same as the value of the name argument that you provide to the command.
- <Safe> is the safely sql. Migration will be executed in one transaction.
- <Action> is the action like up or down.

### Applying Migrations
To upgrade a database to its latest structure, you should apply all available new migrations using the following command:  
`db-migrator` or `db-migrator up`

For each migration that has been successfully applied, the command will insert a row into 
a database table named migration to record the successful application of the migration. 
This will allow the migration tool to identify which migrations have been applied and which have not.

Sometimes, you may only want to apply one or a few new migrations, instead of all available migrations. 
You can do so by specifying the number of migrations that you want to apply when running the command. 
For example, the following command will try to apply the next three available migrations:  
`db-migrator up 3`

You can also explicitly specify a particular migration to which the database should be migrated 
by using the migrate/to command in one of the following formats:
```bash
db-migrator to 150101_185401                      # using timestamp to specify the migration
db-migrator to "2015-01-01 18:54:01"              # using a string that can be parsed by strtotime()
db-migrator to 150101_185401_create_news_table    # using full name
db-migrator to 1392853618                         # using UNIX timestamp
```
If there are any unapplied migrations earlier than the specified one, 
they will all be applied before the specified migration is applied.

If the specified migration has already been applied before, any later applied migrations will be reverted.

### Releasing Migrations (Atomic Batch Apply)
To apply ALL pending migrations atomically in a single transaction, use the `release` command:
```bash
db-migrator release
```
All migrations in a release share the same `apply_time`, allowing batch identification for later rollback.
If any migration fails, the entire batch is rolled back automatically.

> **Iceberg note:** `release` is **best-effort per-table** on Iceberg. The REST Catalog does not
> support cross-table DDL transactions. Each migration is applied individually; if one fails, already
> applied migrations remain in history and are not automatically reverted. The error message makes
> the partial-apply state explicit.

### Rolling Back a Release
To revert the latest release batch atomically, use the `rollback` command:
```bash
db-migrator rollback
```
This identifies the latest batch by `MAX(apply_time)` and reverts all migrations in that batch within a single transaction.
Before reverting, the command checks that all `.down.sql` files exist.

> **Iceberg note:** `rollback` is also **best-effort per-table** on Iceberg for the same reason.

### Reverting Migrations
To revert (undo) one or multiple migrations that have been applied before, you can run the following command:
```bash
db-migrator down     # revert the most recently applied migration
db-migrator down 3   # revert the most 3 recently applied migrations
```

### Redoing Migrations
Redoing migrations means first reverting the specified migrations and then applying again. This can be done as follows:
```bash
db-migrator redo        # redo the last applied migration
db-migrator redo 3      # redo the last 3 applied migrations
```

### Listing Migrations
To list which migrations have been applied and which are not, you may use the following commands:
```bash
db-migrator history     # showing the last 10 applied migrations
db-migrator history 5   # showing the last 5 applied migrations
db-migrator history all # showing all applied migrations

db-migrator new         # showing the first 10 new migrations
db-migrator new 5       # showing the first 5 new migrations
db-migrator new all     # showing all new migrations
```

### Using Command Line Options
The migration command comes with a few command-line options that can be used to customize its behaviors:

| Option                 | Alias | Env Variable | Default | Description |
|------------------------|-------|--------------|---------|-------------|
| `dsn`                  | `d` | `DSN` | (required) | Database connection string. Format: `driver://username:password@host:port/dbname?options` |
| `migrationPath`        | `p` | `MIGRATION_PATH` | `./migrations` | Directory storing all migration SQL files |
| `migrationTable`       | `t` | `MIGRATION_TABLE` | `migration` | Table name for storing migration history |
| `migrationClusterName` | `cn` | `MIGRATION_CLUSTER_NAME` | (empty) | Cluster name for migration history table. Used only for ClickHouse |
| `migrationReplicated`  | `cr` | `MIGRATION_REPLICATED` | `false` | Use replicated table for migration history. Used only for ClickHouse |
| `placeholderCustom`    | `phc` | `PLACEHOLDER_CUSTOM` | (empty) | Custom placeholder value for `{placeholder_custom}` in migrations |
| `maxConnAttempts`      | `ma` | `MAX_CONN_ATTEMPTS` | `1` | Maximum number of database connection attempts (1-100) |
| `compact`              | `c` | `COMPACT` | `false` | Output in compact mode |
| `interactive`          | `i` | `INTERACTIVE` | `true` | Run in interactive mode with prompts |
| `dryRun`               | `dry` | `DRY_RUN` | `false` | Show SQL that would be executed without running it |

#### Example with env params:
```bash
DSN=clickhouse://default:@localhost:9000/docker?sslmode=disable&compress=true \
MIGRATION_PATH=./migrations \
MIGRATION_TABLE=migration \
MIGRATION_CLUSTER_NAME=test_cluster \
MIGRATION_REPLICATED=true \
PLACEHOLDER_CUSTOM=my_value \
MAX_CONN_ATTEMPTS=3 \
COMPACT=true \
INTERACTIVE=false \
db-migrator up
```

### Dry Run Preview
Use dry run to preview the SQL and migration plan without applying changes. Interactive prompts are disabled.

```bash
DRY_RUN=true \
DSN=postgres://username:password@localhost:5432/mydb?sslmode=disable \
MIGRATION_PATH=./migrations \
db-migrator up
```

### How to build and install?
You can execute the command in root directory `make build` or `build-docker` into docker container.
If you want build the debian package, then you can run the command 
`make build-deb` or `build-deb-docker` into docker container.

#### With Go toolchain
The Latest version:
```bash
go get -u -d github.com/raoptimus/db-migrator.go/cmd/db-migrator
```
or
```bash
go install github.com/raoptimus/db-migrator.go/cmd/db-migrator@latest
```
The custom version:
```bash
go get -u -d github.com/raoptimus/db-migrator.go/cmd/db-migrator@v1.2.0
```
or
```bash
go install github.com/raoptimus/db-migrator.go/cmd/db-migrator@v1.2.0
```

#### With docker
```bash
docker pull raoptimus/db-migrator:latest
```
See [https://hub.docker.com/r/raoptimus/db-migrator](https://hub.docker.com/r/raoptimus/db-migrator)

### Example
```bash
DSN=clickhouse://default:@localhost:9000/docker?sslmode=disable&compress=true&debug=false
MIGRATION_TABLE=migration
MIGRATION_PATH=./migrator/db/clickhouseMigration/test_migrates
```
Runs the command `db-migrator up` and get the response
```
Total 1 new migration to be applied: 
        200905_192800_create_test_table
Apply the above migration? [y/n]: y
2020/09/11 22:02:22 *** applying 200905_192800_create_test_table
2020/09/11 22:02:22     > execute SQL: CREATE TABLE test (
    time DateTime DEFAULT now(),
    value UInt32

) ENGINE = MergeTree
PARTITION BY toYYYYMM(time)
ORDER BY (time, value); ...
2020/09/11 22:02:22     > execute SQL: ALTER TABLE test ADD COLUMN value2 UInt8; ...
2020/09/11 22:02:22 *** applied 200905_192800_create_test_table (time: 0.022s)
2020/09/11 22:02:22 1 migration was applied
Migrated up successfully
```

---

## Running in Kubernetes (Helm)

A reusable Helm chart lives in [`charts/db-migrator`](charts/db-migrator). It runs the tool
as a Kubernetes `Job`:

- **`release`** — applies pending migrations. Rendered as a Helm hook
  (`pre-install,pre-upgrade`), so the schema is ready before application pods roll out.
- **`rollback`** — reverts the latest release batch. Opt-in plain `Job` (disabled by default),
  triggered explicitly, so it works the same under plain Helm and [werf](https://werf.io).

The DSN is referenced from an existing `Secret` (never stored in values), and `INTERACTIVE`
is always forced to `false` because a Job has no TTY.

```bash
kubectl create secret generic db-migrator-dsn \
  --from-literal=dsn='postgres://user:pass@postgres:5432/app?sslmode=disable'

helm install my-migrations ./charts/db-migrator \
  --set image.repository=myregistry/myapp-migrations \
  --set image.tag=1.4.2 \
  --set migrator.dsn.existingSecret=db-migrator-dsn
```

The base image ships no migrations — build your own image with them baked in:

```dockerfile
FROM raoptimus/db-migrator:1.7.0
COPY ./migrations /migrations
```

To roll back the latest batch, enable the rollback Job explicitly:

```bash
helm upgrade --install my-migrations ./charts/db-migrator \
  --reuse-values --set rollback.enabled=true
```

The chart can also be consumed as a subchart dependency. See
[`charts/db-migrator/README.md`](charts/db-migrator/README.md) for the full values reference
and usage examples. Local helpers: `make helm-lint`, `make helm-template`, `make helm-package`.

---

## Placeholders in Migrations

You can use placeholders in your migration SQL files that will be replaced at runtime:

| Placeholder | Description | Environment Variable |
|-------------|-------------|---------------------|
| `{cluster}` | ClickHouse cluster name | `MIGRATION_CLUSTER_NAME` |
| `{placeholder_custom}` | Custom placeholder value | `PLACEHOLDER_CUSTOM` |
| `{username}` | Database username (from DSN) | Extracted from `DSN` |
| `{password}` | Database password (from DSN) | Extracted from `DSN` |

### Example Migration with Placeholders

**ClickHouse cluster migration:**
```sql
CREATE TABLE IF NOT EXISTS events ON CLUSTER {cluster} (
    id UInt64,
    event_time DateTime,
    user_id UInt32
) ENGINE = ReplicatedMergeTree
ORDER BY (event_time, id);

-- Grant permissions using credentials from DSN
GRANT SELECT ON events TO {username};
```

**Custom placeholder usage:**
```bash
PLACEHOLDER_CUSTOM="production_suffix" \
DSN="clickhouse://admin:secret@localhost:9000/mydb" \
db-migrator up
```

```sql
CREATE TABLE IF NOT EXISTS logs_{placeholder_custom} (
    id UInt64,
    message String
) ENGINE = MergeTree()
ORDER BY id;
```

---

## Security: Credential Masking

**Credentials are automatically masked in all log output.** When migration SQL contains `{username}` or `{password}` placeholders, the actual values will be replaced with `****` in logs. This ensures sensitive credentials never appear in console output or log files.

### Example

If your migration contains:
```sql
CREATE USER {username} IDENTIFIED BY '{password}';
```

The log output will show:
```
> execute SQL: CREATE USER **** IDENTIFIED BY '****'; ...
```

This protection applies to:
- Console output during migration execution
- Error messages
- Debug logs

For Iceberg DSN parameters, the following values are also masked: `token`, `credential`,
`s3.secret-access-key`, and `s3.session-token`.

---

## Using as a Go Library

You can use db-migrator directly in your Go code:

```go
package main

import (
    "context"
    "log"

    dbmigrator "github.com/raoptimus/db-migrator.go"
)

func main() {
    ctx := context.Background()
    dsn := "postgres://user:pass@localhost:5432/mydb?sslmode=disable"

    // Create database connection
    conn, err := dbmigrator.NewConnection(dsn)
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    // Or use TryConnection with retries
    // conn, err := dbmigrator.TryConnection(dsn, 3) // 3 attempts

    // Configure migration service
    opts := &dbmigrator.Options{
        DSN:       dsn,
        TableName: "migration",
    }

    // Create migration service (pass nil for default logger)
    service, err := dbmigrator.NewDBService(opts, conn, nil)
    if err != nil {
        log.Fatal(err)
    }

    // Apply a migration
    upSQL := `
        CREATE TABLE users (
            id SERIAL PRIMARY KEY,
            name VARCHAR(255) NOT NULL
        );
    `
    err = service.Upgrade(ctx, "240101_120000", upSQL, true) // true = safe (transactional)
    if err != nil {
        log.Fatal(err)
    }

    // Revert a migration
    downSQL := `DROP TABLE users;`
    err = service.Downgrade(ctx, "240101_120000", downSQL, true)
    if err != nil {
        log.Fatal(err)
    }
}
```

### Options

```go
type Options struct {
    DSN         string // Database connection string
    TableName   string // Migration history table name (default: "migration")
    ClusterName string // ClickHouse cluster name (optional)
    Replicated  bool   // Use replicated tables for ClickHouse (optional)
}
```

### Methods

- `Upgrade(ctx, version, sql, safety)` - Apply a migration
- `Downgrade(ctx, version, sql, safety)` - Revert a migration

The `safety` parameter determines whether the migration runs within a transaction.

---

## Apache Iceberg-Specific Considerations

### Migration Dialect: Spark SQL (Iceberg DDL)

Iceberg migrations are written in **Spark SQL** with Iceberg extensions.
The supported subset covers the most common schema-evolution DDL (v1):

| Statement | Description |
|-----------|-------------|
| `CREATE NAMESPACE <ns>` | Create a namespace (maps to Iceberg namespace) |
| `DROP NAMESPACE <ns>` | Drop a namespace |
| `CREATE TABLE <id> (…) USING iceberg [PARTITIONED BY (…)] [COMMENT '…'] [TBLPROPERTIES (…)]` | Create an Iceberg table |
| `DROP TABLE <id>` | Drop an Iceberg table |
| `RENAME TABLE <from> TO <to>` | Rename an Iceberg table |
| `ALTER TABLE <id> ADD COLUMN <name> <type> [COMMENT '…']` | Add a column |
| `ALTER TABLE <id> DROP COLUMN <name>` | Drop a column |
| `ALTER TABLE <id> RENAME COLUMN <old> TO <new>` | Rename a column |
| `ALTER TABLE <id> ALTER COLUMN <name> TYPE <type>` | Change a column type (widening only) |
| `ALTER TABLE <id> ADD PARTITION FIELD <transform>(<col>)` | Add a partition field |
| `ALTER TABLE <id> DROP PARTITION FIELD <transform>(<col>)` | Drop a partition field |

SQL comments (`--` and `/* */`) are supported inside migration files.

**Supported column types:**

| Spark SQL type | Iceberg type | Notes |
|----------------|--------------|-------|
| `STRING` | string | |
| `INT` / `INTEGER` | int | |
| `BIGINT` / `LONG` | long | |
| `DOUBLE` | double | |
| `DECIMAL(p,s)` | decimal(p,s) | |
| `DATE` | date | |
| `TIMESTAMP` | timestamptz | UTC, with timezone |
| `TIMESTAMP_NTZ` | timestamp | without timezone |
| `UUID` | uuid | |
| `BINARY` | binary | |
| `struct<f1:T1,…>` | struct | Nested struct |
| `array<T>` | list | Nested list |
| `map<K,V>` | map | Nested map |

**Supported partition transforms:**
`identity`, `years`, `months`, `days`, `hours`, `bucket(N)`, `truncate(N)`

**Reference `CREATE TABLE` example** (from `fixtures/iceberg/`):

```sql
CREATE TABLE iceberg.analytics.events (
  event_id   STRING    COMMENT 'unique event identifier',
  user_id    STRING    COMMENT 'external user identifier',
  event_type STRING    COMMENT 'event category',
  event_time TIMESTAMP COMMENT 'canonical event time (UTC)',
  created_at TIMESTAMP COMMENT 'record ingestion time',
  source     STRING    COMMENT 'ingestion source label'
)
USING iceberg
PARTITIONED BY (days(event_time))
COMMENT 'analytics events (bronze layer)'
TBLPROPERTIES (
  'format-version' = '2',
  'write.format.default' = 'parquet',
  'write.parquet.compression-codec' = 'zstd'
);
```

### Identifier Rules (Namespace Qualification)

Iceberg tables live inside namespaces. The identifier format is:
```
[<catalog>.]<namespace…>.<table>
```

- The leading segment is compared to the DSN warehouse name; if they match it is stripped automatically.
- The remaining segments (except the last) form the (possibly multi-level) namespace.
- A bare single-segment identifier (no namespace) returns an error: `namespace required`.

**Example:** DSN warehouse = `iceberg`, identifier = `iceberg.raw.events`
→ catalog stripped → namespace = `[raw]`, table = `events`

> **Known limitation (multi-level namespaces):** Single-level namespaces (e.g. `raw.events`) are
> fully supported. Multi-level namespaces (e.g. `bronze.raw.events`) parse correctly and the namespace
> itself can be created, but **table operations inside a multi-level namespace may fail against some
> REST Catalog servers** (notably the `apache/iceberg-rest-fixture` used in tests, which mishandles the
> `%2E`-encoded namespace separator produced by the `iceberg-go` client). This is an upstream
> client/catalog interaction, not a limitation of the migration tool. Prefer single-level namespaces
> until validated against your production catalog.

### Atomicity Boundaries

The REST Catalog provides **per-table optimistic concurrency**, not cross-table DDL transactions.

| Command | Atomicity on Iceberg |
|---------|----------------------|
| `up` / `down` / `redo` / `to` | Per-statement: each DDL is committed individually |
| `.safe` migrations | No-op wrapper; Iceberg has no intra-table DDL transactions |
| `release` | **Best-effort**: all pending migrations applied in order; failure leaves history in partial-apply state |
| `rollback` | **Best-effort**: each migration reverted individually; failure stops at the first error |

When a `release` fails mid-batch, the history table reflects the migrations that were actually
applied. The error message reports which migration failed.

### Down Migrations and Irreversible Operations

Each migration requires a paired `.down.sql` file written by the user (no automatic inversion).

**Irreversible operations** — the Iceberg catalog rejects these in a `.down.sql`:
- **Type narrowing** (e.g., `long → int`): rejected by catalog.
- **`DROP COLUMN`**: the column's field-id is permanently retired; re-adding a column with the same
  name gets a new field-id, which breaks readers expecting the old id.
- **`RENAME COLUMN`**: similarly loses the original field-id association.

When an irreversible operation fails, the tool returns a clear error and the migration record
**remains marked as applied** in history (fail-fast). It is the user's responsibility to handle
the incompatibility manually.

**Example — irreversible `down`:**

```sql
-- up: widen id int -> long (valid Iceberg promotion)
ALTER TABLE iceberg.raw_demo.orders ALTER COLUMN id TYPE long;
```

```sql
-- down: attempting to narrow long -> int — Iceberg rejects this
ALTER TABLE iceberg.raw_demo.orders ALTER COLUMN id TYPE int;
-- error: incompatible type change; migration stays in history as applied
```

### Migration History Storage

For Iceberg, migration history is stored in **namespace properties** instead of a database table:

- A dedicated namespace named by `MIGRATION_TABLE` (default: `migration`) is created automatically.
- Each applied migration is stored as a property: key `migrate.<version>` → value `<apply_time_unix>`.
- `MAX(apply_time)` across all properties identifies the latest release batch for `rollback`.
- Sorting and aggregation are performed in Go (REST Catalog does not guarantee property order).

**Known limitation:** REST Catalog servers impose a size limit on namespace properties. For a
realistic number of migrations (hundreds) this is well within limits, but if the limit is
approached the catalog returns a clear error.

### Example Migrations

See `fixtures/iceberg/` for a full reversible migration chain and
`fixtures/iceberg-irreversible/` for a negative test case (irreversible type narrowing).

---

## Tarantool-Specific Considerations

When using Tarantool:
- Connection DSN format: `tarantool://username:password@host:port/database`
- Example: `tarantool://guest:@localhost:3301/mydb`
- Tarantool uses Lua-based migrations instead of SQL
- Transactions are supported via streams
- Schema management uses box.schema.space.create() and related APIs

### Example Migration

**Up migration (251002_183908_create_test_space.up.sql)**:
```lua
box.schema.space.create('test', {if_not_exists = true})
box.space.test:format({
    {'name', type = 'string', is_nullable = false},
    {'rank', type = 'unsigned', is_nullable = false}
})
box.space.test:create_index('primary', {parts = {'name'}, if_not_exists = true})
```

**Down migration (251002_183908_create_test_space.down.sql)**:
```lua
box.space.test:drop()
```

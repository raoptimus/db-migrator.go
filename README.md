[![Test](https://github.com/raoptimus/db-migrator.go/workflows/Test/badge.svg)](https://github.com/raoptimus/db-migrator.go/actions)
[![Coverage](https://github.com/raoptimus/db-migrator.go/wiki/coverage.svg)](https://raw.githack.com/wiki/raoptimus/db-migrator.go/coverage.html)
[![GitHub Release](https://img.shields.io/github/release/raoptimus/db-migrator.go.svg)](https://github.com/raoptimus/db-migrator.go/releases)

# db-migrator.go
Database Migration tool in CLI on Golang that allows you to keep track of database changes in terms of database migrations which are version-controlled.
The db migration tool currently supports the following db drivers:
- clickhouse
- postgres
- mysql
- tarantool db

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

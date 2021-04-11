[![Build Status](https://travis-ci.org/raoptimus/db-migrator.go.svg?branch=master)](https://travis-ci.org/raoptimus/db-migrator.go)
[![GitHub Release](https://img.shields.io/github/release/raoptimus/db-migrator.go.svg)](https://github.com/raoptimus/db-migrator.go/releases)

# db-migrator.go
Database Migration tool in CLI on Golang that allows you to keep track of database changes in terms of database migrations which are version-controlled.
The db migration tool currently supports the following db drivers: 
- clickhouse
- postgres

### Creating Migrations
To create a new migration, run the following command:  
`DSN=clickhouse://default:@localhost:9000/docker?sslmode=disable&compress=true MIGRATION_PATH=./migrations db-migrator create <name>`

The required name argument gives a brief description about the new migration.  
For example, if the migration is about creating a new table named news, you may use the name `create_news_table` and run the following command:  
`db-migrator create create_news_table`

The above command will create a new sql file named 200101_232501_create_news_table.safe.up.sql in the ./migrations directory. 

The migration file name is automatically generated in the format of <YYMMDD_HHMMSS>_<Name>.<Safe>.<Action>.sql, where
- <YYMMDD_HHMMSS> refers to the UTC datetime at which the migration creation command is executed.
- <Name> is the same as the value of the name argument that you provide to the command.
- <Safe> is the safely sql. MIgration will be executed in one transaction.
- <Action> is the action like up or down.

### Applying Migrations 
To upgrade a database to its latest structure, you should apply all available new migrations using the following command:  
`db-migrator` or `db-migrator up`

For each migration that has been successfully applied, the command will insert a row into 
a database table named migration to record the successful application of the migration. 
This will allow the migration tool to identify which migrations have been applied and which have not.`

Sometimes, you may only want to apply one or a few new migrations, instead of all available migrations. 
You can do so by specifying the number of migrations that you want to apply when running the command. 
For example, the following command will try to apply the next three available migrations:  
`db-migrator up 3`

You can also explicitly specify a particular migration to which the database should be migrated 
by using the migrate/to command in one of the following formats:
```
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
```
db-migrator down     # revert the most recently applied migration
db-migrator down 3   # revert the most 3 recently applied migrations
```

### Redoing Migrations
Redoing migrations means first reverting the specified migrations and then applying again. This can be done as follows:
```
db-migrator redo        # redo the last applied migration
db-migrator redo 3      # redo the last 3 applied migrations
```
### Listing Migrations
To list which migrations have been applied and which are not, you may use the following commands:
```
db-migrator history     # showing the last 10 applied migrations
db-migrator history 5   # showing the last 5 applied migrations
db-migrator history all # showing all applied migrations

db-migrator new         # showing the first 10 new migrations
db-migrator new 5       # showing the first 5 new migrations
db-migrator new all     # showing all new migrations
```
### Using Command Line Options
The migration command comes with a few command-line options that can be used to customize its behaviors:

- `interactive` or `i`: boolean (defaults to true), specifies whether to perform migrations in an interactive mode. 
When this is true, the user will be prompted before the command performs certain actions. 
You may want to set this to false if the command is being used in a background process.
- `migrationPath` or `p`: string (defaults to ./migrations), specifies the directory storing all migration sql files. 
This can be specified as either a directory path or a path alias. 
Note that the directory must exist, or the command may trigger an error.
- `migrationTable` or `t`: string (defaults to migration), specifies the name of the database table for storing migration history information. 
The table will be automatically created by the command if it does not exist. 
You may also manually create it using the structure version varchar(255) primary key, apply_time integer.
- `migrationClusterName` or `cn`: string (defaults to empty), specifies the name of the database cluster name for storing migration history information. 
The table will be automatically created in cluster by the command if it does not exist.
It uses only for clickhouse
- `dsn` or `d`: string (defaults to empty), Database connection strings are specified via URLs. 
The URL format is driver dependent but generally has the form: driver://username:password@host:port/dbname?option1=true.
- `compact` or `c`: boolean (defaults to false), output in compact mode

#### You can use each option as env params:
```
DSN=clickhouse://default:@localhost:9000/docker?sslmode=disable&compress=true&debug=false
MIGRATION_PATH=./migrations
MIGRATION_TABLE=migration
MIGRATION_CLUSTER_NAME=test_cluster
COMPACT=true
INTERACTIVE=false
```

### How to build and install?
You can execute the command in root directory `make build` or `build-docker` into docker container.
If you want build the debian package, then you can run the command 
`make build-deb` or `build-deb-docker` into docker container.

#### With Go toolchain
The Latest version:
```
go get -u -d github.com/raoptimus/db-migrator.go/cmd/db-migrator
```  
The custom version:  
```
go get -u -d github.com/raoptimus/db-migrator.go/cmd/db-migrator@0.1.1
```

#### With docker
```
docker pull raoptimus/db-migrator:latest
```
See [https://hub.docker.com/r/raoptimus/db-migrator](https://hub.docker.com/r/raoptimus/db-migrator)

### Example
```
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

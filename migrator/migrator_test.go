package migrator

import "os"

func createPostgresMigrator() (*Service, error) {
	return New(Options{
		DSN:         os.Getenv("POSTGRES_DSN"),
		Directory:   os.Getenv("POSTGRES_MIGRATIONS_PATH"),
		TableName:   "migration",
		Compact:     false,
		Interactive: false,
	})
}

func createClickhouseMigrator() (*Service, error) {
	return New(Options{
		DSN:         os.Getenv("CLICKHOUSE_DSN"),
		Directory:   os.Getenv("CLICKHOUSE_MIGRATIONS_PATH"),
		TableName:   "migration",
		Compact:     false,
		Interactive: false,
	})
}

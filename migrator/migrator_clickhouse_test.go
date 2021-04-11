package migrator

import (
	"database/sql"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestMigrateService_ClickHouse_UpDown(t *testing.T) {
	m, err := createClickhouseMigrator()
	assert.NoError(t, err)

	_, err = m.db.Exec("DROP DATABASE docker")
	assert.NoError(t, err)
	_, err = m.db.Exec("CREATE DATABASE docker")
	assert.NoError(t, err)

	err = m.Up("2")
	assert.NoError(t, err)

	assertEqualMigrationsCount(t, m.db, 2+1)
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

func assertEqualMigrationsCount(t *testing.T, db *sql.DB, expected int) {
	var count int
	err := db.QueryRow("SELECT count(*) FROM docker.migration").Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, expected, count)
}

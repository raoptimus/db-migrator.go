package migrator

import (
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

	var count int
	err = m.db.QueryRow("SELECT count(*) FROM docker.migration").Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 2+1, count)
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

package migrator

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestMigrateService_ClickHouse_UpDown(t *testing.T) {
	m, err := createClickhouseMigrator()
	assert.NoError(t, err)

	_, err = m.db.Exec("DROP DATABASE default")
	assert.NoError(t, err)
	_, err = m.db.Exec("CREATE DATABASE default")
	assert.NoError(t, err)

	err = m.Up("2")
	assert.NoError(t, err)
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

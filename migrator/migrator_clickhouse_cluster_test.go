package migrator

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestMigrateService_ClickHouseCluster_UpDown(t *testing.T) {
	m, err := createClickhouseClusterMigrator()
	assert.NoError(t, err)

	_, err = m.db.Exec("DROP DATABASE default")
	assert.NoError(t, err)
	_, err = m.db.Exec("CREATE DATABASE default")
	assert.NoError(t, err)

	err = m.Up("2")
	assert.NoError(t, err)
}

func createClickhouseClusterMigrator() (*Service, error) {
	return New(Options{
		DSN:         os.Getenv("CLICKHOUSE_CLUSTER_DSN1"),
		Directory:   os.Getenv("CLICKHOUSE_CLUSTER_MIGRATIONS_PATH"),
		TableName:   "migration",
		Compact:     false,
		Interactive: false,
	})
}

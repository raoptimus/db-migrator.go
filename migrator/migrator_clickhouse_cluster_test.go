package migrator

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestMigrateService_ClickHouseCluster_UpDown(t *testing.T) {
	var m1 *Service
	var m2 *Service
	var err error
	m1, err = createClickhouse1ClusterMigrator()
	assert.NoError(t, err)

	m2, err = createClickhouse2ClusterMigrator()
	assert.NoError(t, err)

	_, err = m1.db.Exec("DROP DATABASE docker ON CLUSTER test_cluster")
	assert.NoError(t, err)
	_, err = m1.db.Exec("CREATE DATABASE docker ON CLUSTER test_cluster")
	assert.NoError(t, err)

	err = m1.Up("1")
	assert.NoError(t, err)

	assertEqualMigrationsCount(t, m1.db, 1+1)
	assertEqualMigrationsCount(t, m2.db, 1+1)
}

func createClickhouse1ClusterMigrator() (*Service, error) {
	return New(Options{
		DSN:         os.Getenv("CLICKHOUSE_CLUSTER_DSN1"),
		Directory:   os.Getenv("CLICKHOUSE_CLUSTER_MIGRATIONS_PATH"),
		TableName:   "migration",
		ClusterName: os.Getenv("CLICKHOUSE_CLUSTER_NAME"),
		Compact:     false,
		Interactive: false,
	})
}

func createClickhouse2ClusterMigrator() (*Service, error) {
	return New(Options{
		DSN:         os.Getenv("CLICKHOUSE_CLUSTER_DSN2"),
		Directory:   os.Getenv("CLICKHOUSE_CLUSTER_MIGRATIONS_PATH"),
		TableName:   "migration",
		ClusterName: os.Getenv("CLICKHOUSE_CLUSTER_NAME"),
		Compact:     false,
		Interactive: false,
	})
}

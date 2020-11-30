package migrator

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMigrateController_Postgres_Up(t *testing.T) {
	m, err := createPostgresMigrator()
	assert.NoError(t, err)

	err = m.Up("2")
	assert.NoError(t, err)
}

func TestMigrateController_Clickhouse_Up(t *testing.T) {
	m, err := createClickhouseMigrator()
	assert.NoError(t, err)

	err = m.Up("2")
	assert.NoError(t, err)
}

package migrator

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestMigrateService_Mysql_UpDown_Successfully(t *testing.T) {
	m, err := createMysqlMigrator("migration")
	assert.NoError(t, err)

	err = m.Down("all")
	assert.NoError(t, err)

	err = m.Up("1")
	assert.NoError(t, err)

	var c int
	err = m.db.QueryRow("select count(*) from test").Scan(&c)
	assert.NoError(t, err)
	assert.Equal(t, 0, c)

	err = m.db.QueryRow("select count(*) from migration").Scan(&c)
	assert.NoError(t, err)
	assert.Equal(t, 2, c)

	err = m.Down("all")
	assert.NoError(t, err)
}

func TestMigrateService_Mysql_Redo_Successfully(t *testing.T) {
	m, err := createPostgresMigrator("migration")
	assert.NoError(t, err)

	err = m.Down("all")
	assert.NoError(t, err)

	err = m.Up("1")
	assert.NoError(t, err)

	err = m.Redo("1")
	assert.NoError(t, err)

	var c int
	err = m.db.QueryRow("select count(*) from migration").Scan(&c)
	assert.NoError(t, err)
	assert.Equal(t, 2, c)

	err = m.Down("all")
	assert.NoError(t, err)
}

func createMysqlMigrator(migrationTableName string) (*Service, error) {
	return New(Options{
		DSN:         os.Getenv("MYSQL_DSN"),
		Directory:   os.Getenv("MYSQL_MIGRATIONS_PATH"),
		TableName:   migrationTableName,
		Compact:     false,
		Interactive: false,
	})
}

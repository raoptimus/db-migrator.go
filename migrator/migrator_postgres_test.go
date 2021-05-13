package migrator

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestMigrateService_Postgres_UpDown_Successfully(t *testing.T) {
	m, err := createPostgresMigrator("migration")
	assert.NoError(t, err)

	err = m.Down("all")
	assert.NoError(t, err)

	err = m.Up("2")
	assert.NoError(t, err)

	err = m.Up("1")
	assert.Error(t, err)

	var c int
	err = m.db.QueryRow("select count(*) from test").Scan(&c)
	assert.NoError(t, err)
	assert.Equal(t, 0, c)

	err = m.db.QueryRow("select count(*) from public.migration").Scan(&c)
	assert.NoError(t, err)
	assert.Equal(t, 3, c)

	err = m.Down("all")
	assert.NoError(t, err)
}

func TestMigrateService_Postgres_Redo_Successfully(t *testing.T) {
	m, err := createPostgresMigrator("migration")
	assert.NoError(t, err)

	err = m.Down("all")
	assert.NoError(t, err)

	err = m.Up("2")
	assert.NoError(t, err)

	err = m.Redo("1")
	assert.NoError(t, err)

	var c int
	err = m.db.QueryRow("select count(*) from public.migration").Scan(&c)
	assert.NoError(t, err)
	assert.Equal(t, 3, c)

	err = m.Down("all")
	assert.NoError(t, err)
}

func TestMigrateService_Postgres_UpDown_WithSchema_Successfully(t *testing.T) {
	m, err := createPostgresMigrator("docker.migration")
	assert.NoError(t, err)

	err = m.Down("all")
	assert.NoError(t, err)

	err = m.Up("2")
	assert.NoError(t, err)

	err = m.Up("1")
	assert.Error(t, err)

	var c int
	err = m.db.QueryRow("select count(*) from test").Scan(&c)
	assert.NoError(t, err)
	assert.Equal(t, 0, c)

	err = m.db.QueryRow("select count(*) from docker.migration").Scan(&c)
	assert.NoError(t, err)
	assert.Equal(t, 3, c)

	err = m.Down("all")
	assert.NoError(t, err)
}

func TestMigrateService_Postgres_Redo_WithSchema_Successfully(t *testing.T) {
	m, err := createPostgresMigrator("docker.migration")
	assert.NoError(t, err)

	err = m.Down("all")
	assert.NoError(t, err)

	err = m.Up("2")
	assert.NoError(t, err)

	err = m.Redo("1")
	assert.NoError(t, err)

	var c int
	err = m.db.QueryRow("select count(*) from docker.migration").Scan(&c)
	assert.NoError(t, err)
	assert.Equal(t, 3, c)

	err = m.Down("all")
	assert.NoError(t, err)
}

func createPostgresMigrator(migrationTableName string) (*Service, error) {
	return New(Options{
		DSN:         os.Getenv("POSTGRES_DSN"),
		Directory:   os.Getenv("POSTGRES_MIGRATIONS_PATH"),
		TableName:   migrationTableName,
		Compact:     false,
		Interactive: false,
	})
}

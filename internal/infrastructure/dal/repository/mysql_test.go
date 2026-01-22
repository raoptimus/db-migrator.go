/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package repository

import (
	"context"
	"testing"

	"github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
	"github.com/raoptimus/db-migrator.go/internal/infrastructure/sqlex"
	thelp "github.com/raoptimus/db-migrator.go/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestMySQL_ExecQuery_Successfully(t *testing.T) {
	ctx := context.Background()
	conn := NewMockConnection(t)
	conn.EXPECT().
		ExecContext(ctx, "SELECT 1").
		Return(nil, nil)

	repo := NewMySQL(conn, &Options{
		TableName:  "migration",
		SchemaName: "test_db",
	})
	err := repo.ExecQuery(ctx, "SELECT 1")
	require.NoError(t, err)
}

func TestMySQL_ExecQuery_Failure(t *testing.T) {
	ctx := context.Background()
	conn := NewMockConnection(t)
	conn.EXPECT().
		ExecContext(ctx, "SELECT 1").
		Return(nil, &mysql.MySQLError{Number: 1064, Message: "syntax error"})

	repo := NewMySQL(conn, &Options{
		TableName:  "migration",
		SchemaName: "test_db",
	})
	err := repo.ExecQuery(ctx, "SELECT 1")
	require.Error(t, err)
}

func TestMySQL_ExecQueryTransaction_Successfully(t *testing.T) {
	ctx := context.Background()
	conn := NewMockConnection(t)
	txFn := func(ctx context.Context) error {
		return nil
	}
	conn.EXPECT().
		Transaction(ctx, mock.AnythingOfType("func(context.Context) error")).
		Return(nil)

	repo := NewMySQL(conn, &Options{
		TableName:  "migration",
		SchemaName: "test_db",
	})
	err := repo.ExecQueryTransaction(ctx, txFn)
	require.NoError(t, err)
}

func TestMySQL_ExecQueryTransaction_Failure(t *testing.T) {
	ctx := context.Background()
	conn := NewMockConnection(t)
	expectedErr := errors.New("transaction failed")
	conn.EXPECT().
		Transaction(ctx, mock.AnythingOfType("func(context.Context) error")).
		Return(expectedErr)

	repo := NewMySQL(conn, &Options{
		TableName:  "migration",
		SchemaName: "test_db",
	})
	err := repo.ExecQueryTransaction(ctx, func(ctx context.Context) error {
		return nil
	})
	require.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

func TestMySQL_CreateMigrationHistoryTable_Successfully(t *testing.T) {
	ctx := context.Background()

	expectedSQL := `
		CREATE TABLE migration (
		  version VARCHAR(180) PRIMARY KEY,
		  apply_time INT
		)
		ENGINE=InnoDB
	`

	conn := NewMockConnection(t)
	conn.EXPECT().
		ExecContext(ctx, mock.MatchedBy(thelp.CompareSQL(expectedSQL))).
		Return(nil, nil).
		Once()

	repo := NewMySQL(conn, &Options{
		TableName:  "migration",
		SchemaName: "test_db",
	})
	err := repo.CreateMigrationHistoryTable(ctx)
	require.NoError(t, err)
}

func TestMySQL_CreateMigrationHistoryTable_Failure(t *testing.T) {
	ctx := context.Background()

	conn := NewMockConnection(t)
	conn.EXPECT().
		ExecContext(ctx, mock.AnythingOfType("string")).
		Return(nil, &mysql.MySQLError{Number: 1050, Message: "Table already exists"}).
		Once()

	repo := NewMySQL(conn, &Options{
		TableName:  "migration",
		SchemaName: "test_db",
	})
	err := repo.CreateMigrationHistoryTable(ctx)
	require.Error(t, err)
	var dbErr *DBError
	require.ErrorAs(t, err, &dbErr)
	assert.Equal(t, "1050", dbErr.Code)
}

func TestMySQL_DropMigrationHistoryTable_Successfully(t *testing.T) {
	ctx := context.Background()

	expectedSQL := `DROP TABLE migration`

	conn := NewMockConnection(t)
	conn.EXPECT().
		ExecContext(ctx, mock.MatchedBy(thelp.CompareSQL(expectedSQL))).
		Return(nil, nil).
		Once()

	repo := NewMySQL(conn, &Options{
		TableName:  "migration",
		SchemaName: "test_db",
	})
	err := repo.DropMigrationHistoryTable(ctx)
	require.NoError(t, err)
}

func TestMySQL_DropMigrationHistoryTable_Failure(t *testing.T) {
	ctx := context.Background()

	conn := NewMockConnection(t)
	conn.EXPECT().
		ExecContext(ctx, mock.AnythingOfType("string")).
		Return(nil, &mysql.MySQLError{Number: 1051, Message: "Unknown table"}).
		Once()

	repo := NewMySQL(conn, &Options{
		TableName:  "migration",
		SchemaName: "test_db",
	})
	err := repo.DropMigrationHistoryTable(ctx)
	require.Error(t, err)
	var dbErr *DBError
	require.ErrorAs(t, err, &dbErr)
	assert.Equal(t, "1051", dbErr.Code)
}

func TestMySQL_InsertMigration_Successfully(t *testing.T) {
	ctx := context.Background()

	conn := NewMockConnection(t)
	conn.EXPECT().
		ExecContext(ctx, mock.AnythingOfType("string"), "210328_221600_test", mock.AnythingOfType("uint32")).
		Return(nil, nil).
		Once()

	repo := NewMySQL(conn, &Options{
		TableName:  "migration",
		SchemaName: "test_db",
	})
	err := repo.InsertMigration(ctx, "210328_221600_test")
	require.NoError(t, err)
}

func TestMySQL_InsertMigration_Failure(t *testing.T) {
	ctx := context.Background()

	conn := NewMockConnection(t)
	conn.EXPECT().
		ExecContext(ctx, mock.AnythingOfType("string"), "210328_221600_test", mock.AnythingOfType("uint32")).
		Return(nil, &mysql.MySQLError{Number: 1062, Message: "Duplicate entry"}).
		Once()

	repo := NewMySQL(conn, &Options{
		TableName:  "migration",
		SchemaName: "test_db",
	})
	err := repo.InsertMigration(ctx, "210328_221600_test")
	require.Error(t, err)
	var dbErr *DBError
	require.ErrorAs(t, err, &dbErr)
	assert.Equal(t, "1062", dbErr.Code)
}

func TestMySQL_RemoveMigration_Successfully(t *testing.T) {
	ctx := context.Background()

	expectedSQL := `DELETE FROM migration WHERE version = ?`

	conn := NewMockConnection(t)
	conn.EXPECT().
		ExecContext(ctx, mock.MatchedBy(thelp.CompareSQL(expectedSQL)), "210328_221600_test").
		Return(nil, nil).
		Once()

	repo := NewMySQL(conn, &Options{
		TableName:  "migration",
		SchemaName: "test_db",
	})
	err := repo.RemoveMigration(ctx, "210328_221600_test")
	require.NoError(t, err)
}

func TestMySQL_RemoveMigration_Failure(t *testing.T) {
	ctx := context.Background()

	conn := NewMockConnection(t)
	conn.EXPECT().
		ExecContext(ctx, mock.AnythingOfType("string"), "210328_221600_test").
		Return(nil, &mysql.MySQLError{Number: 1146, Message: "Table doesn't exist"}).
		Once()

	repo := NewMySQL(conn, &Options{
		TableName:  "migration",
		SchemaName: "test_db",
	})
	err := repo.RemoveMigration(ctx, "210328_221600_test")
	require.Error(t, err)
	var dbErr *DBError
	require.ErrorAs(t, err, &dbErr)
	assert.Equal(t, "1146", dbErr.Code)
}

func TestMySQL_Migrations_Failure(t *testing.T) {
	ctx := context.Background()

	expectedSQL := `
		SELECT version, apply_time
		FROM migration
		ORDER BY apply_time DESC, version DESC
		LIMIT ?
	`

	conn := NewMockConnection(t)
	conn.EXPECT().
		QueryContext(ctx, mock.MatchedBy(thelp.CompareSQL(expectedSQL)), 10).
		Return(nil, &mysql.MySQLError{Number: 1146, Message: "Table doesn't exist"}).
		Once()

	repo := NewMySQL(conn, &Options{
		TableName:  "migration",
		SchemaName: "test_db",
	})
	_, err := repo.Migrations(ctx, 10)
	require.Error(t, err)
}

func TestMySQL_HasMigrationHistoryTable_Failure(t *testing.T) {
	ctx := context.Background()

	expectedSQL := `
		SELECT EXISTS(
		    SELECT *
			FROM information_schema.tables
			WHERE table_name = ? AND table_schema = ?
		)
	`

	conn := NewMockConnection(t)
	conn.EXPECT().
		QueryContext(ctx, mock.MatchedBy(thelp.CompareSQL(expectedSQL)), "migration", "test_db").
		Return(nil, errors.New("connection error")).
		Once()

	repo := NewMySQL(conn, &Options{
		TableName:  "migration",
		SchemaName: "test_db",
	})
	_, err := repo.HasMigrationHistoryTable(ctx)
	require.Error(t, err)
}

func TestMySQL_MigrationsCount_Failure(t *testing.T) {
	ctx := context.Background()

	expectedSQL := `SELECT count(*) FROM migration`

	conn := NewMockConnection(t)
	conn.EXPECT().
		QueryContext(ctx, mock.MatchedBy(thelp.CompareSQL(expectedSQL))).
		Return(nil, errors.New("connection error")).
		Once()

	repo := NewMySQL(conn, &Options{
		TableName:  "migration",
		SchemaName: "test_db",
	})
	_, err := repo.MigrationsCount(ctx)
	require.Error(t, err)
}

func TestMySQL_ExistsMigration_Failure(t *testing.T) {
	ctx := context.Background()

	expectedSQL := `SELECT 1 FROM test_db.migration WHERE version = ?`

	conn := NewMockConnection(t)
	conn.EXPECT().
		QueryContext(ctx, mock.MatchedBy(thelp.CompareSQL(expectedSQL)), "210328_221600_test").
		Return(nil, errors.New("connection error")).
		Once()

	repo := NewMySQL(conn, &Options{
		TableName:  "migration",
		SchemaName: "test_db",
	})
	_, err := repo.ExistsMigration(ctx, "210328_221600_test")
	require.Error(t, err)
}

func TestMySQL_TableNameWithSchema(t *testing.T) {
	tests := []struct {
		name       string
		tableName  string
		schemaName string
		want       string
	}{
		{
			name:       "standard table name",
			tableName:  "migration",
			schemaName: "test_db",
			want:       "test_db.migration",
		},
		{
			name:       "custom table name",
			tableName:  "custom_migrations",
			schemaName: "mydb",
			want:       "mydb.custom_migrations",
		},
		{
			name:       "empty schema name",
			tableName:  "migration",
			schemaName: "",
			want:       ".migration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMySQL(nil, &Options{
				TableName:  tt.tableName,
				SchemaName: tt.schemaName,
			})
			got := repo.TableNameWithSchema()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMySQL_dbError_MySQLError(t *testing.T) {
	repo := NewMySQL(nil, &Options{
		TableName:  "migration",
		SchemaName: "test_db",
	})

	mysqlErr := &mysql.MySQLError{Number: 1045, Message: "Access denied"}
	query := "SELECT * FROM migration"

	err := repo.dbError(mysqlErr, query)

	var dbErr *DBError
	require.ErrorAs(t, err, &dbErr)
	assert.Equal(t, "1045", dbErr.Code)
	assert.Equal(t, "Access denied", dbErr.Message)
	assert.Equal(t, query, dbErr.InternalQuery)
}

func TestMySQL_dbError_NonMySQLError(t *testing.T) {
	repo := NewMySQL(nil, &Options{
		TableName:  "migration",
		SchemaName: "test_db",
	})

	regularErr := errors.New("some other error")
	query := "SELECT * FROM migration"

	err := repo.dbError(regularErr, query)

	assert.Equal(t, regularErr, err)

	var dbErr *DBError
	assert.False(t, errors.As(err, &dbErr))
}

func TestMySQL_Migrations_Successfully(t *testing.T) {
	ctx := context.Background()

	expectedSQL := `
		SELECT version, apply_time
		FROM migration
		ORDER BY apply_time DESC, version DESC
		LIMIT ?
	`

	rows := sqlex.NewRowsWithSlice([]interface{}{
		[]any{"210328_221600_create_users", int64(1616968560)},
		[]any{"210329_121500_add_index", int64(1617020100)},
	})

	conn := NewMockConnection(t)
	conn.EXPECT().
		QueryContext(ctx, mock.MatchedBy(thelp.CompareSQL(expectedSQL)), 10).
		Return(rows, nil).
		Once()

	repo := NewMySQL(conn, &Options{
		TableName:  "migration",
		SchemaName: "test_db",
	})
	migrations, err := repo.Migrations(ctx, 10)

	require.NoError(t, err)
	require.Len(t, migrations, 2)
	assert.Equal(t, "210328_221600_create_users", migrations[0].Version)
	assert.Equal(t, int64(1616968560), migrations[0].ApplyTime)
}

func TestMySQL_Migrations_EmptyResult_Successfully(t *testing.T) {
	ctx := context.Background()

	rows := sqlex.NewRowsWithSlice([]interface{}{})

	conn := NewMockConnection(t)
	conn.EXPECT().
		QueryContext(ctx, mock.AnythingOfType("string"), 10).
		Return(rows, nil).
		Once()

	repo := NewMySQL(conn, &Options{
		TableName:  "migration",
		SchemaName: "test_db",
	})
	migrations, err := repo.Migrations(ctx, 10)

	require.NoError(t, err)
	assert.Empty(t, migrations)
}

func TestMySQL_HasMigrationHistoryTable_TableExists_Successfully(t *testing.T) {
	ctx := context.Background()

	rows := sqlex.NewRowsWithSlice([]interface{}{true})

	conn := NewMockConnection(t)
	conn.EXPECT().
		QueryContext(ctx, mock.AnythingOfType("string"), "migration", "test_db").
		Return(rows, nil).
		Once()

	repo := NewMySQL(conn, &Options{
		TableName:  "migration",
		SchemaName: "test_db",
	})
	exists, err := repo.HasMigrationHistoryTable(ctx)

	require.NoError(t, err)
	assert.True(t, exists)
}

func TestMySQL_HasMigrationHistoryTable_TableNotExists_Successfully(t *testing.T) {
	ctx := context.Background()

	rows := sqlex.NewRowsWithSlice([]interface{}{false})

	conn := NewMockConnection(t)
	conn.EXPECT().
		QueryContext(ctx, mock.AnythingOfType("string"), "migration", "test_db").
		Return(rows, nil).
		Once()

	repo := NewMySQL(conn, &Options{
		TableName:  "migration",
		SchemaName: "test_db",
	})
	exists, err := repo.HasMigrationHistoryTable(ctx)

	require.NoError(t, err)
	assert.False(t, exists)
}

func TestMySQL_MigrationsCount_Successfully(t *testing.T) {
	ctx := context.Background()

	rows := sqlex.NewRowsWithSlice([]interface{}{5})

	conn := NewMockConnection(t)
	conn.EXPECT().
		QueryContext(ctx, mock.AnythingOfType("string")).
		Return(rows, nil).
		Once()

	repo := NewMySQL(conn, &Options{
		TableName:  "migration",
		SchemaName: "test_db",
	})
	count, err := repo.MigrationsCount(ctx)

	require.NoError(t, err)
	assert.Equal(t, 5, count)
}

func TestMySQL_QueryScalar_Successfully(t *testing.T) {
	ctx := context.Background()

	rows := sqlex.NewRowsWithSlice([]interface{}{42})

	conn := NewMockConnection(t)
	conn.EXPECT().
		QueryContext(ctx, "SELECT count(*)").
		Return(rows, nil).
		Once()

	repo := NewMySQL(conn, &Options{
		TableName:  "migration",
		SchemaName: "test_db",
	})
	var result int
	err := repo.QueryScalar(ctx, "SELECT count(*)", &result)

	require.NoError(t, err)
	assert.Equal(t, 42, result)
}

func TestMySQL_QueryScalar_InvalidPtr_Failure(t *testing.T) {
	ctx := context.Background()

	conn := NewMockConnection(t)

	repo := NewMySQL(conn, &Options{
		TableName:  "migration",
		SchemaName: "test_db",
	})
	var result []int
	err := repo.QueryScalar(ctx, "SELECT count(*)", &result)

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrPtrValueMustBeAPointerAndScalar)
}

func TestMySQL_ExistsMigration_Exists_Successfully(t *testing.T) {
	ctx := context.Background()

	rows := sqlex.NewRowsWithSlice([]interface{}{1})

	conn := NewMockConnection(t)
	conn.EXPECT().
		QueryContext(ctx, mock.AnythingOfType("string"), "210328_221600_test").
		Return(rows, nil).
		Once()

	repo := NewMySQL(conn, &Options{
		TableName:  "migration",
		SchemaName: "test_db",
	})
	exists, err := repo.ExistsMigration(ctx, "210328_221600_test")

	require.NoError(t, err)
	assert.True(t, exists)
}

func TestMySQL_ExistsMigration_NotExists_Successfully(t *testing.T) {
	ctx := context.Background()

	rows := sqlex.NewRowsWithSlice([]interface{}{})

	conn := NewMockConnection(t)
	conn.EXPECT().
		QueryContext(ctx, mock.AnythingOfType("string"), "210328_221600_test").
		Return(rows, nil).
		Once()

	repo := NewMySQL(conn, &Options{
		TableName:  "migration",
		SchemaName: "test_db",
	})
	exists, err := repo.ExistsMigration(ctx, "210328_221600_test")

	require.NoError(t, err)
	assert.False(t, exists)
}

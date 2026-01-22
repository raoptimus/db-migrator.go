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

	"github.com/pkg/errors"
	"github.com/raoptimus/db-migrator.go/internal/infrastructure/sqlex"
	thelp "github.com/raoptimus/db-migrator.go/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/tarantool/go-tarantool/v2"
)

func TestTarantool_ExecQuery_Successfully(t *testing.T) {
	ctx := context.Background()
	conn := NewMockConnection(t)
	conn.EXPECT().
		ExecContext(ctx, "return box.info()").
		Return(nil, nil)

	repo := NewTarantool(conn, &Options{
		TableName: "migration",
	})
	err := repo.ExecQuery(ctx, "return box.info()")
	require.NoError(t, err)
}

func TestTarantool_ExecQuery_Failure(t *testing.T) {
	ctx := context.Background()
	conn := NewMockConnection(t)
	tErr := tarantool.Error{Code: 32, Msg: "Unknown space"}
	conn.EXPECT().
		ExecContext(ctx, "return box.info()").
		Return(nil, tErr)

	repo := NewTarantool(conn, &Options{
		TableName: "migration",
	})
	err := repo.ExecQuery(ctx, "return box.info()")
	require.Error(t, err)
	var dbErr *DBError
	require.ErrorAs(t, err, &dbErr)
	assert.Equal(t, "32", dbErr.Code)
	assert.Equal(t, "Unknown space", dbErr.Message)
}

func TestTarantool_ExecQueryTransaction_Successfully(t *testing.T) {
	ctx := context.Background()
	conn := NewMockConnection(t)
	txFn := func(ctx context.Context) error {
		return nil
	}
	conn.EXPECT().
		Transaction(ctx, mock.AnythingOfType("func(context.Context) error")).
		Return(nil)

	repo := NewTarantool(conn, &Options{
		TableName: "migration",
	})
	err := repo.ExecQueryTransaction(ctx, txFn)
	require.NoError(t, err)
}

func TestTarantool_ExecQueryTransaction_Failure(t *testing.T) {
	ctx := context.Background()
	conn := NewMockConnection(t)
	expectedErr := errors.New("transaction failed")
	conn.EXPECT().
		Transaction(ctx, mock.AnythingOfType("func(context.Context) error")).
		Return(expectedErr)

	repo := NewTarantool(conn, &Options{
		TableName: "migration",
	})
	err := repo.ExecQueryTransaction(ctx, func(ctx context.Context) error {
		return nil
	})
	require.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

func TestTarantool_CreateMigrationHistoryTable_Successfully(t *testing.T) {
	ctx := context.Background()

	expectedCreateSpace := `box.schema.space.create('migration', {if_not_exists = true})`
	expectedFormat := `box.space.migration:format({{'version',type = 'string',is_nullable = false},{'apply_time', type = 'unsigned', is_nullable = false}})`
	expectedPrimaryIndex := `box.space.migration:create_index('primary', {parts = {'version'}, if_not_exists = true})`
	expectedSecondaryIndex := `box.space.migration:create_index('secondary', {parts = {{'apply_time'}, {'version'}}, if_not_exists = true})`

	conn := NewMockConnection(t)
	conn.EXPECT().
		ExecContext(ctx, mock.MatchedBy(thelp.CompareSQL(expectedCreateSpace))).
		Return(nil, nil).
		Once()
	conn.EXPECT().
		ExecContext(ctx, mock.MatchedBy(thelp.CompareSQL(expectedFormat))).
		Return(nil, nil).
		Once()
	conn.EXPECT().
		ExecContext(ctx, mock.MatchedBy(thelp.CompareSQL(expectedPrimaryIndex))).
		Return(nil, nil).
		Once()
	conn.EXPECT().
		ExecContext(ctx, mock.MatchedBy(thelp.CompareSQL(expectedSecondaryIndex))).
		Return(nil, nil).
		Once()

	repo := NewTarantool(conn, &Options{
		TableName: "migration",
	})
	err := repo.CreateMigrationHistoryTable(ctx)
	require.NoError(t, err)
}

func TestTarantool_CreateMigrationHistoryTable_CreateSpaceFails(t *testing.T) {
	ctx := context.Background()

	conn := NewMockConnection(t)
	tErr := tarantool.Error{Code: 10, Msg: "Space already exists"}
	conn.EXPECT().
		ExecContext(ctx, mock.AnythingOfType("string")).
		Return(nil, tErr).
		Once()

	repo := NewTarantool(conn, &Options{
		TableName: "migration",
	})
	err := repo.CreateMigrationHistoryTable(ctx)
	require.Error(t, err)
	var dbErr *DBError
	require.ErrorAs(t, err, &dbErr)
	assert.Equal(t, "10", dbErr.Code)
}

func TestTarantool_DropMigrationHistoryTable_Successfully(t *testing.T) {
	ctx := context.Background()

	expectedSQL := `box.space.migration:drop()`

	conn := NewMockConnection(t)
	conn.EXPECT().
		ExecContext(ctx, mock.MatchedBy(thelp.CompareSQL(expectedSQL))).
		Return(nil, nil).
		Once()

	repo := NewTarantool(conn, &Options{
		TableName: "migration",
	})
	err := repo.DropMigrationHistoryTable(ctx)
	require.NoError(t, err)
}

func TestTarantool_DropMigrationHistoryTable_Failure(t *testing.T) {
	ctx := context.Background()

	conn := NewMockConnection(t)
	tErr := tarantool.Error{Code: 36, Msg: "Space not found"}
	conn.EXPECT().
		ExecContext(ctx, mock.AnythingOfType("string")).
		Return(nil, tErr).
		Once()

	repo := NewTarantool(conn, &Options{
		TableName: "migration",
	})
	err := repo.DropMigrationHistoryTable(ctx)
	require.Error(t, err)
	var dbErr *DBError
	require.ErrorAs(t, err, &dbErr)
	assert.Equal(t, "36", dbErr.Code)
}

func TestTarantool_InsertMigration_Successfully(t *testing.T) {
	ctx := context.Background()

	conn := NewMockConnection(t)
	conn.EXPECT().
		ExecContext(ctx, mock.AnythingOfType("string"), "210328_221600_test", mock.AnythingOfType("int64")).
		Return(nil, nil).
		Once()

	repo := NewTarantool(conn, &Options{
		TableName: "migration",
	})
	err := repo.InsertMigration(ctx, "210328_221600_test")
	require.NoError(t, err)
}

func TestTarantool_InsertMigration_Failure(t *testing.T) {
	ctx := context.Background()

	conn := NewMockConnection(t)
	tErr := tarantool.Error{Code: 3, Msg: "Duplicate key exists"}
	conn.EXPECT().
		ExecContext(ctx, mock.AnythingOfType("string"), "210328_221600_test", mock.AnythingOfType("int64")).
		Return(nil, tErr).
		Once()

	repo := NewTarantool(conn, &Options{
		TableName: "migration",
	})
	err := repo.InsertMigration(ctx, "210328_221600_test")
	require.Error(t, err)
	var dbErr *DBError
	require.ErrorAs(t, err, &dbErr)
	assert.Equal(t, "3", dbErr.Code)
}

func TestTarantool_RemoveMigration_Successfully(t *testing.T) {
	ctx := context.Background()

	expectedLua := `box.space.migration:delete(...)`

	conn := NewMockConnection(t)
	conn.EXPECT().
		ExecContext(ctx, mock.MatchedBy(thelp.CompareSQL(expectedLua)), "210328_221600_test").
		Return(nil, nil).
		Once()

	repo := NewTarantool(conn, &Options{
		TableName: "migration",
	})
	err := repo.RemoveMigration(ctx, "210328_221600_test")
	require.NoError(t, err)
}

func TestTarantool_RemoveMigration_Failure(t *testing.T) {
	ctx := context.Background()

	conn := NewMockConnection(t)
	tErr := tarantool.Error{Code: 36, Msg: "Space not found"}
	conn.EXPECT().
		ExecContext(ctx, mock.AnythingOfType("string"), "210328_221600_test").
		Return(nil, tErr).
		Once()

	repo := NewTarantool(conn, &Options{
		TableName: "migration",
	})
	err := repo.RemoveMigration(ctx, "210328_221600_test")
	require.Error(t, err)
	var dbErr *DBError
	require.ErrorAs(t, err, &dbErr)
	assert.Equal(t, "36", dbErr.Code)
}

func TestTarantool_Migrations_Failure(t *testing.T) {
	ctx := context.Background()

	expectedLua := `return box.space.migration:select({}, {iterator='LT', limit = 10})`

	conn := NewMockConnection(t)
	tErr := tarantool.Error{Code: 36, Msg: "Space not found"}
	conn.EXPECT().
		QueryContext(ctx, mock.MatchedBy(thelp.CompareSQL(expectedLua))).
		Return(nil, tErr).
		Once()

	repo := NewTarantool(conn, &Options{
		TableName: "migration",
	})
	_, err := repo.Migrations(ctx, 10)
	require.Error(t, err)
}

func TestTarantool_HasMigrationHistoryTable_Failure(t *testing.T) {
	ctx := context.Background()

	expectedLua := `return box.space.migration ~= nil`

	conn := NewMockConnection(t)
	conn.EXPECT().
		QueryContext(ctx, mock.MatchedBy(thelp.CompareSQL(expectedLua))).
		Return(nil, errors.New("connection error")).
		Once()

	repo := NewTarantool(conn, &Options{
		TableName: "migration",
	})
	_, err := repo.HasMigrationHistoryTable(ctx)
	require.Error(t, err)
}

func TestTarantool_MigrationsCount_Failure(t *testing.T) {
	ctx := context.Background()

	expectedLua := `return box.space.migration:len()`

	conn := NewMockConnection(t)
	conn.EXPECT().
		QueryContext(ctx, mock.MatchedBy(thelp.CompareSQL(expectedLua))).
		Return(nil, errors.New("connection error")).
		Once()

	repo := NewTarantool(conn, &Options{
		TableName: "migration",
	})
	_, err := repo.MigrationsCount(ctx)
	require.Error(t, err)
}

func TestTarantool_ExistsMigration_Failure(t *testing.T) {
	ctx := context.Background()

	expectedLua := `box.space.migration:count('210328_221600_test', {iterator='EQ'})`

	conn := NewMockConnection(t)
	conn.EXPECT().
		QueryContext(ctx, mock.MatchedBy(thelp.CompareSQL(expectedLua))).
		Return(nil, errors.New("connection error")).
		Once()

	repo := NewTarantool(conn, &Options{
		TableName: "migration",
	})
	_, err := repo.ExistsMigration(ctx, "210328_221600_test")
	require.Error(t, err)
}

func TestTarantool_TableNameWithSchema(t *testing.T) {
	tests := []struct {
		name      string
		tableName string
		want      string
	}{
		{
			name:      "standard table name",
			tableName: "migration",
			want:      "migration",
		},
		{
			name:      "custom table name",
			tableName: "custom_migrations",
			want:      "custom_migrations",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewTarantool(nil, &Options{
				TableName: tt.tableName,
			})
			got := repo.TableNameWithSchema()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTarantool_dbError_TarantoolError(t *testing.T) {
	repo := NewTarantool(nil, &Options{
		TableName: "migration",
	})

	tErr := tarantool.Error{Code: 3, Msg: "Duplicate key exists"}
	query := "return box.space.migration:len()"

	err := repo.dbError(tErr, query)

	var dbErr *DBError
	require.ErrorAs(t, err, &dbErr)
	assert.Equal(t, "3", dbErr.Code)
	assert.Equal(t, "Duplicate key exists", dbErr.Message)
	assert.Equal(t, query, dbErr.InternalQuery)
	assert.Equal(t, "ERR", dbErr.Severity)
}

func TestTarantool_dbError_NonTarantoolError(t *testing.T) {
	repo := NewTarantool(nil, &Options{
		TableName: "migration",
	})

	regularErr := errors.New("some other error")
	query := "return box.space.migration:len()"

	err := repo.dbError(regularErr, query)

	assert.Equal(t, regularErr, err)

	var dbErr *DBError
	assert.False(t, errors.As(err, &dbErr))
}

func TestTarantool_Migrations_Successfully(t *testing.T) {
	ctx := context.Background()

	expectedLua := `return box.space.migration:select({}, {iterator='LT', limit = 10})`

	rows := sqlex.NewRowsWithSlice([]interface{}{
		[]any{"210328_221600_create_users", int64(1616968560)},
		[]any{"210329_121500_add_index", int64(1617020100)},
	})

	conn := NewMockConnection(t)
	conn.EXPECT().
		QueryContext(ctx, mock.MatchedBy(thelp.CompareSQL(expectedLua))).
		Return(rows, nil).
		Once()

	repo := NewTarantool(conn, &Options{
		TableName: "migration",
	})
	migrations, err := repo.Migrations(ctx, 10)

	require.NoError(t, err)
	require.Len(t, migrations, 2)
	assert.Equal(t, "210328_221600_create_users", migrations[0].Version)
	assert.Equal(t, int64(1616968560), migrations[0].ApplyTime)
}

func TestTarantool_Migrations_EmptyResult_Successfully(t *testing.T) {
	ctx := context.Background()

	rows := sqlex.NewRowsWithSlice([]interface{}{})

	conn := NewMockConnection(t)
	conn.EXPECT().
		QueryContext(ctx, mock.AnythingOfType("string")).
		Return(rows, nil).
		Once()

	repo := NewTarantool(conn, &Options{
		TableName: "migration",
	})
	migrations, err := repo.Migrations(ctx, 10)

	require.NoError(t, err)
	assert.Empty(t, migrations)
}

func TestTarantool_HasMigrationHistoryTable_TableExists_Successfully(t *testing.T) {
	ctx := context.Background()

	expectedLua := `return box.space.migration ~= nil`

	rows := sqlex.NewRowsWithSlice([]interface{}{true})

	conn := NewMockConnection(t)
	conn.EXPECT().
		QueryContext(ctx, mock.MatchedBy(thelp.CompareSQL(expectedLua))).
		Return(rows, nil).
		Once()

	repo := NewTarantool(conn, &Options{
		TableName: "migration",
	})
	exists, err := repo.HasMigrationHistoryTable(ctx)

	require.NoError(t, err)
	assert.True(t, exists)
}

func TestTarantool_HasMigrationHistoryTable_TableNotExists_Successfully(t *testing.T) {
	ctx := context.Background()

	rows := sqlex.NewRowsWithSlice([]interface{}{false})

	conn := NewMockConnection(t)
	conn.EXPECT().
		QueryContext(ctx, mock.AnythingOfType("string")).
		Return(rows, nil).
		Once()

	repo := NewTarantool(conn, &Options{
		TableName: "migration",
	})
	exists, err := repo.HasMigrationHistoryTable(ctx)

	require.NoError(t, err)
	assert.False(t, exists)
}

func TestTarantool_MigrationsCount_Successfully(t *testing.T) {
	ctx := context.Background()

	expectedLua := `return box.space.migration:len()`

	rows := sqlex.NewRowsWithSlice([]interface{}{5})

	conn := NewMockConnection(t)
	conn.EXPECT().
		QueryContext(ctx, mock.MatchedBy(thelp.CompareSQL(expectedLua))).
		Return(rows, nil).
		Once()

	repo := NewTarantool(conn, &Options{
		TableName: "migration",
	})
	count, err := repo.MigrationsCount(ctx)

	require.NoError(t, err)
	assert.Equal(t, 5, count)
}

func TestTarantool_QueryScalar_Successfully(t *testing.T) {
	ctx := context.Background()

	rows := sqlex.NewRowsWithSlice([]interface{}{42})

	conn := NewMockConnection(t)
	conn.EXPECT().
		QueryContext(ctx, "return 42").
		Return(rows, nil).
		Once()

	repo := NewTarantool(conn, &Options{
		TableName: "migration",
	})
	var result int
	err := repo.QueryScalar(ctx, "return 42", &result)

	require.NoError(t, err)
	assert.Equal(t, 42, result)
}

func TestTarantool_QueryScalar_InvalidPtr_Failure(t *testing.T) {
	ctx := context.Background()

	conn := NewMockConnection(t)

	repo := NewTarantool(conn, &Options{
		TableName: "migration",
	})
	var result []int
	err := repo.QueryScalar(ctx, "return 42", &result)

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrPtrValueMustBeAPointerAndScalar)
}

func TestTarantool_ExistsMigration_Exists_Successfully(t *testing.T) {
	ctx := context.Background()

	rows := sqlex.NewRowsWithSlice([]interface{}{true})

	conn := NewMockConnection(t)
	conn.EXPECT().
		QueryContext(ctx, mock.AnythingOfType("string")).
		Return(rows, nil).
		Once()

	repo := NewTarantool(conn, &Options{
		TableName: "migration",
	})
	exists, err := repo.ExistsMigration(ctx, "210328_221600_test")

	require.NoError(t, err)
	assert.True(t, exists)
}

func TestTarantool_ExistsMigration_NotExists_Successfully(t *testing.T) {
	ctx := context.Background()

	rows := sqlex.NewRowsWithSlice([]interface{}{false})

	conn := NewMockConnection(t)
	conn.EXPECT().
		QueryContext(ctx, mock.AnythingOfType("string")).
		Return(rows, nil).
		Once()

	repo := NewTarantool(conn, &Options{
		TableName: "migration",
	})
	exists, err := repo.ExistsMigration(ctx, "210328_221600_test")

	require.NoError(t, err)
	assert.False(t, exists)
}

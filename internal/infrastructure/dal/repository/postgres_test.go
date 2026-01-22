package repository

import (
	"context"
	"testing"

	"github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/raoptimus/db-migrator.go/internal/infrastructure/dal/connection"
	"github.com/raoptimus/db-migrator.go/internal/infrastructure/sqlex"
	thelp "github.com/raoptimus/db-migrator.go/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestPostgres_ExecQuery_Successfully(t *testing.T) {
	ctx := context.Background()
	conn := NewMockConnection(t)
	conn.EXPECT().
		Driver().
		Return(connection.DriverPostgres)
	conn.EXPECT().
		ExecContext(ctx, "SELECT 1").
		Return(nil, nil)

	repo, err := New(conn, &Options{})
	require.NoError(t, err)
	err = repo.ExecQuery(ctx, "SELECT 1")
	require.NoError(t, err)
}

func TestPostgres_ExecQuery_Failure(t *testing.T) {
	ctx := context.Background()
	conn := NewMockConnection(t)
	conn.EXPECT().
		Driver().
		Return(connection.DriverPostgres)
	conn.EXPECT().
		ExecContext(ctx, "SELECT 1").
		Return(nil, &pq.Error{Severity: pq.Efatal})

	repo, err := New(conn, &Options{})
	require.NoError(t, err)
	err = repo.ExecQuery(ctx, "SELECT 1")
	require.Error(t, err)
	var dbErr *DBError
	require.ErrorAs(t, err, &dbErr)
	assert.Equal(t, pq.Efatal, dbErr.Severity)
	assert.Equal(t, "SELECT 1", dbErr.InternalQuery)
}

func TestPostgres_ExecQueryTransaction_Successfully(t *testing.T) {
	ctx := context.Background()

	conn := NewMockConnection(t)
	conn.EXPECT().
		Transaction(ctx, mock.AnythingOfType("func(context.Context) error")).
		Return(nil).
		Once()

	repo := NewPostgres(conn, &Options{
		TableName:  "migration",
		SchemaName: "public",
	})
	err := repo.ExecQueryTransaction(ctx, func(ctx context.Context) error {
		return nil
	})

	require.NoError(t, err)
}

func TestPostgres_ExecQueryTransaction_Failure(t *testing.T) {
	ctx := context.Background()
	expectedErr := errors.New("transaction failed")

	conn := NewMockConnection(t)
	conn.EXPECT().
		Transaction(ctx, mock.AnythingOfType("func(context.Context) error")).
		Return(expectedErr).
		Once()

	repo := NewPostgres(conn, &Options{
		TableName:  "migration",
		SchemaName: "public",
	})
	err := repo.ExecQueryTransaction(ctx, func(ctx context.Context) error {
		return nil
	})

	require.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

func TestPostgres_Migrations_Successfully(t *testing.T) {
	ctx := context.Background()

	expectedSQL := `
		SELECT version, apply_time
		FROM public.migration
		ORDER BY apply_time DESC, version DESC
		LIMIT $1
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

	repo := NewPostgres(conn, &Options{
		TableName:  "migration",
		SchemaName: "public",
	})
	migrations, err := repo.Migrations(ctx, 10)

	require.NoError(t, err)
	require.Len(t, migrations, 2)
	assert.Equal(t, "210328_221600_create_users", migrations[0].Version)
	assert.Equal(t, int64(1616968560), migrations[0].ApplyTime)
}

func TestPostgres_Migrations_Failure(t *testing.T) {
	ctx := context.Background()

	conn := NewMockConnection(t)
	conn.EXPECT().
		QueryContext(ctx, mock.AnythingOfType("string"), 10).
		Return(nil, &pq.Error{Severity: pq.Efatal, Message: "connection lost"}).
		Once()

	repo := NewPostgres(conn, &Options{
		TableName:  "migration",
		SchemaName: "public",
	})
	_, err := repo.Migrations(ctx, 10)

	require.Error(t, err)
}

func TestPostgres_Migrations_EmptyResult_Successfully(t *testing.T) {
	ctx := context.Background()

	rows := sqlex.NewRowsWithSlice([]interface{}{})

	conn := NewMockConnection(t)
	conn.EXPECT().
		QueryContext(ctx, mock.AnythingOfType("string"), 10).
		Return(rows, nil).
		Once()

	repo := NewPostgres(conn, &Options{
		TableName:  "migration",
		SchemaName: "public",
	})
	migrations, err := repo.Migrations(ctx, 10)

	require.NoError(t, err)
	assert.Empty(t, migrations)
}

func TestPostgres_HasMigrationHistoryTable_TableExists_Successfully(t *testing.T) {
	ctx := context.Background()

	rows := sqlex.NewRowsWithSlice([]interface{}{
		[]any{"public", "migration"},
	})

	conn := NewMockConnection(t)
	conn.EXPECT().
		QueryContext(ctx, mock.AnythingOfType("string"), "migration", "public").
		Return(rows, nil).
		Once()

	repo := NewPostgres(conn, &Options{
		TableName:  "migration",
		SchemaName: "public",
	})
	exists, err := repo.HasMigrationHistoryTable(ctx)

	require.NoError(t, err)
	assert.True(t, exists)
}

func TestPostgres_HasMigrationHistoryTable_TableNotExists_Successfully(t *testing.T) {
	ctx := context.Background()

	rows := sqlex.NewRowsWithSlice([]interface{}{})

	conn := NewMockConnection(t)
	conn.EXPECT().
		QueryContext(ctx, mock.AnythingOfType("string"), "migration", "public").
		Return(rows, nil).
		Once()

	repo := NewPostgres(conn, &Options{
		TableName:  "migration",
		SchemaName: "public",
	})
	exists, err := repo.HasMigrationHistoryTable(ctx)

	require.NoError(t, err)
	assert.False(t, exists)
}

func TestPostgres_HasMigrationHistoryTable_Failure(t *testing.T) {
	ctx := context.Background()

	conn := NewMockConnection(t)
	conn.EXPECT().
		QueryContext(ctx, mock.AnythingOfType("string"), "migration", "public").
		Return(nil, errors.New("connection error")).
		Once()

	repo := NewPostgres(conn, &Options{
		TableName:  "migration",
		SchemaName: "public",
	})
	_, err := repo.HasMigrationHistoryTable(ctx)

	require.Error(t, err)
}

func TestPostgres_InsertMigration_Successfully(t *testing.T) {
	ctx := context.Background()

	conn := NewMockConnection(t)
	conn.EXPECT().
		ExecContext(ctx, mock.AnythingOfType("string"), "210328_221600_test", mock.AnythingOfType("uint32")).
		Return(nil, nil).
		Once()

	repo := NewPostgres(conn, &Options{
		TableName:  "migration",
		SchemaName: "public",
	})
	err := repo.InsertMigration(ctx, "210328_221600_test")

	require.NoError(t, err)
}

func TestPostgres_InsertMigration_Failure(t *testing.T) {
	ctx := context.Background()

	conn := NewMockConnection(t)
	conn.EXPECT().
		ExecContext(ctx, mock.AnythingOfType("string"), "210328_221600_test", mock.AnythingOfType("uint32")).
		Return(nil, &pq.Error{Code: "23505", Message: "duplicate key"}).
		Once()

	repo := NewPostgres(conn, &Options{
		TableName:  "migration",
		SchemaName: "public",
	})
	err := repo.InsertMigration(ctx, "210328_221600_test")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "insert migration")
}

func TestPostgres_RemoveMigration_Successfully(t *testing.T) {
	ctx := context.Background()

	expectedSQL := `DELETE FROM public.migration WHERE (version) = ($1)`

	conn := NewMockConnection(t)
	conn.EXPECT().
		ExecContext(ctx, mock.MatchedBy(thelp.CompareSQL(expectedSQL)), "210328_221600_test").
		Return(nil, nil).
		Once()

	repo := NewPostgres(conn, &Options{
		TableName:  "migration",
		SchemaName: "public",
	})
	err := repo.RemoveMigration(ctx, "210328_221600_test")

	require.NoError(t, err)
}

func TestPostgres_RemoveMigration_Failure(t *testing.T) {
	ctx := context.Background()

	conn := NewMockConnection(t)
	conn.EXPECT().
		ExecContext(ctx, mock.AnythingOfType("string"), "210328_221600_test").
		Return(nil, &pq.Error{Code: "42P01", Message: "relation does not exist"}).
		Once()

	repo := NewPostgres(conn, &Options{
		TableName:  "migration",
		SchemaName: "public",
	})
	err := repo.RemoveMigration(ctx, "210328_221600_test")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "remove migration")
}

func TestPostgres_CreateMigrationHistoryTable_Successfully(t *testing.T) {
	ctx := context.Background()

	expectedSQL := `
		CREATE TABLE public.migration (
		  version varchar(180) PRIMARY KEY,
		  apply_time integer
		)
	`

	conn := NewMockConnection(t)
	conn.EXPECT().
		ExecContext(ctx, mock.MatchedBy(thelp.CompareSQL(expectedSQL))).
		Return(nil, nil).
		Once()

	repo := NewPostgres(conn, &Options{
		TableName:  "migration",
		SchemaName: "public",
	})
	err := repo.CreateMigrationHistoryTable(ctx)

	require.NoError(t, err)
}

func TestPostgres_CreateMigrationHistoryTable_Failure(t *testing.T) {
	ctx := context.Background()

	conn := NewMockConnection(t)
	conn.EXPECT().
		ExecContext(ctx, mock.AnythingOfType("string")).
		Return(nil, &pq.Error{Code: "42P07", Message: "relation already exists"}).
		Once()

	repo := NewPostgres(conn, &Options{
		TableName:  "migration",
		SchemaName: "public",
	})
	err := repo.CreateMigrationHistoryTable(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "create migration history table")
}

func TestPostgres_DropMigrationHistoryTable_Successfully(t *testing.T) {
	ctx := context.Background()

	expectedSQL := `DROP TABLE public.migration`

	conn := NewMockConnection(t)
	conn.EXPECT().
		ExecContext(ctx, mock.MatchedBy(thelp.CompareSQL(expectedSQL))).
		Return(nil, nil).
		Once()

	repo := NewPostgres(conn, &Options{
		TableName:  "migration",
		SchemaName: "public",
	})
	err := repo.DropMigrationHistoryTable(ctx)

	require.NoError(t, err)
}

func TestPostgres_DropMigrationHistoryTable_Failure(t *testing.T) {
	ctx := context.Background()

	conn := NewMockConnection(t)
	conn.EXPECT().
		ExecContext(ctx, mock.AnythingOfType("string")).
		Return(nil, &pq.Error{Code: "42P01", Message: "relation does not exist"}).
		Once()

	repo := NewPostgres(conn, &Options{
		TableName:  "migration",
		SchemaName: "public",
	})
	err := repo.DropMigrationHistoryTable(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "drop migration history table")
}

func TestPostgres_MigrationsCount_Successfully(t *testing.T) {
	ctx := context.Background()

	rows := sqlex.NewRowsWithSlice([]interface{}{5})

	conn := NewMockConnection(t)
	conn.EXPECT().
		QueryContext(ctx, mock.AnythingOfType("string")).
		Return(rows, nil).
		Once()

	repo := NewPostgres(conn, &Options{
		TableName:  "migration",
		SchemaName: "public",
	})
	count, err := repo.MigrationsCount(ctx)

	require.NoError(t, err)
	assert.Equal(t, 5, count)
}

func TestPostgres_MigrationsCount_Failure(t *testing.T) {
	ctx := context.Background()

	conn := NewMockConnection(t)
	conn.EXPECT().
		QueryContext(ctx, mock.AnythingOfType("string")).
		Return(nil, errors.New("connection error")).
		Once()

	repo := NewPostgres(conn, &Options{
		TableName:  "migration",
		SchemaName: "public",
	})
	_, err := repo.MigrationsCount(ctx)

	require.Error(t, err)
}

func TestPostgres_QueryScalar_Successfully(t *testing.T) {
	ctx := context.Background()

	rows := sqlex.NewRowsWithSlice([]interface{}{42})

	conn := NewMockConnection(t)
	conn.EXPECT().
		QueryContext(ctx, "SELECT count(*)").
		Return(rows, nil).
		Once()

	repo := NewPostgres(conn, &Options{
		TableName:  "migration",
		SchemaName: "public",
	})
	var result int
	err := repo.QueryScalar(ctx, "SELECT count(*)", &result)

	require.NoError(t, err)
	assert.Equal(t, 42, result)
}

func TestPostgres_QueryScalar_InvalidPtr_Failure(t *testing.T) {
	ctx := context.Background()

	conn := NewMockConnection(t)

	repo := NewPostgres(conn, &Options{
		TableName:  "migration",
		SchemaName: "public",
	})
	var result []int
	err := repo.QueryScalar(ctx, "SELECT count(*)", &result)

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrPtrValueMustBeAPointerAndScalar)
}

func TestPostgres_ExistsMigration_Exists_Successfully(t *testing.T) {
	ctx := context.Background()

	rows := sqlex.NewRowsWithSlice([]interface{}{true})

	conn := NewMockConnection(t)
	conn.EXPECT().
		QueryContext(ctx, mock.AnythingOfType("string"), "210328_221600_test").
		Return(rows, nil).
		Once()

	repo := NewPostgres(conn, &Options{
		TableName:  "migration",
		SchemaName: "public",
	})
	exists, err := repo.ExistsMigration(ctx, "210328_221600_test")

	require.NoError(t, err)
	assert.True(t, exists)
}

func TestPostgres_ExistsMigration_NotExists_Successfully(t *testing.T) {
	ctx := context.Background()

	rows := sqlex.NewRowsWithSlice([]interface{}{false})

	conn := NewMockConnection(t)
	conn.EXPECT().
		QueryContext(ctx, mock.AnythingOfType("string"), "210328_221600_test").
		Return(rows, nil).
		Once()

	repo := NewPostgres(conn, &Options{
		TableName:  "migration",
		SchemaName: "public",
	})
	exists, err := repo.ExistsMigration(ctx, "210328_221600_test")

	require.NoError(t, err)
	assert.False(t, exists)
}

func TestPostgres_ExistsMigration_Failure(t *testing.T) {
	ctx := context.Background()

	conn := NewMockConnection(t)
	conn.EXPECT().
		QueryContext(ctx, mock.AnythingOfType("string"), "210328_221600_test").
		Return(nil, errors.New("connection error")).
		Once()

	repo := NewPostgres(conn, &Options{
		TableName:  "migration",
		SchemaName: "public",
	})
	_, err := repo.ExistsMigration(ctx, "210328_221600_test")

	require.Error(t, err)
}

func TestPostgres_TableNameWithSchema_Successfully(t *testing.T) {
	tests := []struct {
		name       string
		tableName  string
		schemaName string
		want       string
	}{
		{
			name:       "public schema",
			tableName:  "migration",
			schemaName: "public",
			want:       "public.migration",
		},
		{
			name:       "custom schema",
			tableName:  "custom_migrations",
			schemaName: "myschema",
			want:       "myschema.custom_migrations",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewPostgres(nil, &Options{
				TableName:  tt.tableName,
				SchemaName: tt.schemaName,
			})
			got := repo.TableNameWithSchema()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPostgres_dbError_PqError_Successfully(t *testing.T) {
	repo := NewPostgres(nil, &Options{
		TableName:  "migration",
		SchemaName: "public",
	})

	pgErr := &pq.Error{
		Severity: pq.Efatal,
		Code:     "42P01",
		Message:  "relation does not exist",
		Detail:   "Table 'migration' not found",
	}
	query := "SELECT * FROM migration"

	err := repo.dbError(pgErr, query)

	var dbErr *DBError
	require.ErrorAs(t, err, &dbErr)
	assert.Equal(t, "42P01", dbErr.Code)
	assert.Equal(t, pq.Efatal, dbErr.Severity)
	assert.Equal(t, "relation does not exist", dbErr.Message)
	assert.Equal(t, "Table 'migration' not found", dbErr.Details)
	assert.Equal(t, query, dbErr.InternalQuery)
}

func TestPostgres_dbError_PqError_EmptyQuery_Successfully(t *testing.T) {
	repo := NewPostgres(nil, &Options{
		TableName:  "migration",
		SchemaName: "public",
	})

	pgErr := &pq.Error{
		Severity:      pq.Efatal,
		Code:          "42P01",
		Message:       "relation does not exist",
		InternalQuery: "internal query",
	}

	err := repo.dbError(pgErr, "")

	var dbErr *DBError
	require.ErrorAs(t, err, &dbErr)
	assert.Equal(t, "internal query", dbErr.InternalQuery)
}

func TestPostgres_dbError_NonPqError_Successfully(t *testing.T) {
	repo := NewPostgres(nil, &Options{
		TableName:  "migration",
		SchemaName: "public",
	})

	regularErr := errors.New("some other error")
	query := "SELECT * FROM migration"

	err := repo.dbError(regularErr, query)

	assert.Equal(t, regularErr, err)

	var dbErr *DBError
	assert.False(t, errors.As(err, &dbErr))
}

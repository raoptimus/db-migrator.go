package repository

import (
	"context"
	"testing"

	clickhouse "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/pkg/errors"
	"github.com/raoptimus/db-migrator.go/internal/infrastructure/sqlex"
	thelp "github.com/raoptimus/db-migrator.go/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestClickhouse_CreateMigrationHistoryTable_Successfully(t *testing.T) {
	ctx := context.Background()

	expectedSQL := `
		CREATE TABLE default.migrates ON CLUSTER test_cluster (
			version String, 
			date Date DEFAULT toDate(apply_time),
			apply_time UInt32,
			is_deleted UInt8
		) ENGINE = ReplicatedReplacingMergeTree('/clickhouse/tables/{shard}/test_cluster_migrates', '{replica}', apply_time)
		PRIMARY KEY (version)
		PARTITION BY (toYYYYMM(date))
		ORDER BY (version)
		SETTINGS index_granularity=8192
	`
	expectedSQL2 := `
		CREATE TABLE default.d_migrates ON CLUSTER test_cluster AS default.migrates
        ENGINE = Distributed('test_cluster', 'default', migrates, cityHash64(toString(version)))
	`
	conn := NewMockConnection(t)
	conn.EXPECT().
		ExecContext(ctx, mock.MatchedBy(thelp.CompareSQL(expectedSQL))).
		Return(nil, nil).
		Once()
	conn.EXPECT().
		ExecContext(ctx, mock.MatchedBy(thelp.CompareSQL(expectedSQL2))).
		Return(nil, nil).
		Once()

	repo := NewClickhouse(conn, &Options{
		TableName:   "migrates",
		SchemaName:  "default",
		ClusterName: "test_cluster",
		Replicated:  false,
	})
	err := repo.CreateMigrationHistoryTable(ctx)
	require.NoError(t, err)
}

func TestClickhouse_Migrations_Failure(t *testing.T) {
	ctx := context.Background()

	expectedSQL := `
		SELECT version, apply_time 
		FROM default.d_migrates 
		WHERE is_deleted = 0 
		ORDER BY apply_time DESC, version DESC 
		LIMIT ?
	`

	conn := NewMockConnection(t)
	conn.EXPECT().
		QueryContext(ctx, mock.MatchedBy(thelp.CompareSQL(expectedSQL)), 1).
		Return(nil, errors.New("oops")).
		Once()

	repo := NewClickhouse(conn, &Options{
		TableName:   "migrates",
		SchemaName:  "default",
		ClusterName: "test_cluster",
		Replicated:  false,
	})
	_, err := repo.Migrations(ctx, 1)
	assert.Error(t, err)
}

func TestClickhouse_ExistsMigration_Failure(t *testing.T) {
	ctx := context.Background()

	expectedSQL := `
		SELECT 1 FROM default.d_migrates WHERE version = ? AND is_deleted = 0
	`

	conn := NewMockConnection(t)
	conn.EXPECT().
		QueryContext(ctx, mock.MatchedBy(thelp.CompareSQL(expectedSQL)), "20250611_104500_test").
		Return(nil, errors.New("oops")).
		Once()

	repo := NewClickhouse(conn, &Options{
		TableName:   "migrates",
		SchemaName:  "default",
		ClusterName: "test_cluster",
		Replicated:  false,
	})
	_, err := repo.ExistsMigration(ctx, "20250611_104500_test")
	assert.Error(t, err)
}

func TestClickhouse_HasMigrationHistoryTable_Failure(t *testing.T) {
	ctx := context.Background()

	expectedSQL := `
		SELECT database, table 
		FROM system.columns 
		WHERE table = ? AND database = currentDatabase()
	`

	conn := NewMockConnection(t)
	conn.EXPECT().
		QueryContext(ctx, mock.MatchedBy(thelp.CompareSQL(expectedSQL)), "d_migrates").
		Return(nil, errors.New("oops")).
		Once()

	repo := NewClickhouse(conn, &Options{
		TableName:   "migrates",
		SchemaName:  "default",
		ClusterName: "test_cluster",
		Replicated:  false,
	})
	_, err := repo.HasMigrationHistoryTable(ctx)
	assert.Error(t, err)
}

func TestClickhouse_Migrations_Successfully(t *testing.T) {
	ctx := context.Background()

	expectedSQL := `
		SELECT version, apply_time
		FROM default.d_migrates
		WHERE is_deleted = 0
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

	repo := NewClickhouse(conn, &Options{
		TableName:   "migrates",
		SchemaName:  "default",
		ClusterName: "test_cluster",
		Replicated:  false,
	})
	migrations, err := repo.Migrations(ctx, 10)

	require.NoError(t, err)
	require.Len(t, migrations, 2)
	assert.Equal(t, "210328_221600_create_users", migrations[0].Version)
	assert.Equal(t, int64(1616968560), migrations[0].ApplyTime)
	assert.Equal(t, "210329_121500_add_index", migrations[1].Version)
	assert.Equal(t, int64(1617020100), migrations[1].ApplyTime)
}

func TestClickhouse_Migrations_EmptyResult_Successfully(t *testing.T) {
	ctx := context.Background()

	rows := sqlex.NewRowsWithSlice([]interface{}{})

	conn := NewMockConnection(t)
	conn.EXPECT().
		QueryContext(ctx, mock.AnythingOfType("string"), 10).
		Return(rows, nil).
		Once()

	repo := NewClickhouse(conn, &Options{
		TableName:   "migrates",
		SchemaName:  "default",
		ClusterName: "test_cluster",
		Replicated:  false,
	})
	migrations, err := repo.Migrations(ctx, 10)

	require.NoError(t, err)
	assert.Empty(t, migrations)
}

func TestClickhouse_HasMigrationHistoryTable_TableExists_Successfully(t *testing.T) {
	ctx := context.Background()

	expectedSQL := `
		SELECT database, table
		FROM system.columns
		WHERE table = ? AND database = currentDatabase()
	`

	rows := sqlex.NewRowsWithSlice([]interface{}{
		[]any{"default", "d_migrates"},
	})

	conn := NewMockConnection(t)
	conn.EXPECT().
		QueryContext(ctx, mock.MatchedBy(thelp.CompareSQL(expectedSQL)), "d_migrates").
		Return(rows, nil).
		Once()

	repo := NewClickhouse(conn, &Options{
		TableName:   "migrates",
		SchemaName:  "default",
		ClusterName: "test_cluster",
		Replicated:  false,
	})
	exists, err := repo.HasMigrationHistoryTable(ctx)

	require.NoError(t, err)
	assert.True(t, exists)
}

func TestClickhouse_HasMigrationHistoryTable_TableNotExists_Successfully(t *testing.T) {
	ctx := context.Background()

	rows := sqlex.NewRowsWithSlice([]interface{}{})

	conn := NewMockConnection(t)
	conn.EXPECT().
		QueryContext(ctx, mock.AnythingOfType("string"), "d_migrates").
		Return(rows, nil).
		Once()

	repo := NewClickhouse(conn, &Options{
		TableName:   "migrates",
		SchemaName:  "default",
		ClusterName: "test_cluster",
		Replicated:  false,
	})
	exists, err := repo.HasMigrationHistoryTable(ctx)

	require.NoError(t, err)
	assert.False(t, exists)
}

func TestClickhouse_InsertMigration_Successfully(t *testing.T) {
	ctx := context.Background()

	conn := NewMockConnection(t)
	conn.EXPECT().
		Transaction(ctx, mock.AnythingOfType("func(context.Context) error")).
		RunAndReturn(func(ctx context.Context, txFn func(context.Context) error) error {
			return txFn(ctx)
		}).
		Once()
	conn.EXPECT().
		ExecContext(ctx, mock.AnythingOfType("string"), "210328_221600_test", mock.AnythingOfType("uint32"), 0).
		Return(nil, nil).
		Once()
	conn.EXPECT().
		ExecContext(ctx, mock.AnythingOfType("string")).
		Return(nil, nil).
		Once()

	repo := NewClickhouse(conn, &Options{
		TableName:   "migrates",
		SchemaName:  "default",
		ClusterName: "test_cluster",
		Replicated:  false,
	})
	err := repo.InsertMigration(ctx, "210328_221600_test")

	require.NoError(t, err)
}

func TestClickhouse_InsertMigration_TransactionError_Failure(t *testing.T) {
	ctx := context.Background()

	conn := NewMockConnection(t)
	conn.EXPECT().
		Transaction(ctx, mock.AnythingOfType("func(context.Context) error")).
		Return(errors.New("transaction failed")).
		Once()

	repo := NewClickhouse(conn, &Options{
		TableName:   "migrates",
		SchemaName:  "default",
		ClusterName: "test_cluster",
		Replicated:  false,
	})
	err := repo.InsertMigration(ctx, "210328_221600_test")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "insert migration")
}

func TestClickhouse_RemoveMigration_Successfully(t *testing.T) {
	ctx := context.Background()

	conn := NewMockConnection(t)
	conn.EXPECT().
		Transaction(ctx, mock.AnythingOfType("func(context.Context) error")).
		RunAndReturn(func(ctx context.Context, txFn func(context.Context) error) error {
			return txFn(ctx)
		}).
		Once()
	conn.EXPECT().
		ExecContext(ctx, mock.AnythingOfType("string"), "210328_221600_test", mock.AnythingOfType("uint32"), 1).
		Return(nil, nil).
		Once()
	conn.EXPECT().
		ExecContext(ctx, mock.AnythingOfType("string")).
		Return(nil, nil).
		Once()

	repo := NewClickhouse(conn, &Options{
		TableName:   "migrates",
		SchemaName:  "default",
		ClusterName: "test_cluster",
		Replicated:  false,
	})
	err := repo.RemoveMigration(ctx, "210328_221600_test")

	require.NoError(t, err)
}

func TestClickhouse_ExecQuery_Successfully(t *testing.T) {
	ctx := context.Background()

	conn := NewMockConnection(t)
	conn.EXPECT().
		ExecContext(ctx, "SELECT 1").
		Return(nil, nil).
		Once()

	repo := NewClickhouse(conn, &Options{
		TableName:   "migrates",
		SchemaName:  "default",
		ClusterName: "test_cluster",
		Replicated:  false,
	})
	err := repo.ExecQuery(ctx, "SELECT 1")

	require.NoError(t, err)
}

func TestClickhouse_ExecQuery_Failure(t *testing.T) {
	ctx := context.Background()

	clickhouseErr := &clickhouse.Exception{
		Code:       62,
		Message:    "Syntax error",
		StackTrace: "trace",
	}

	conn := NewMockConnection(t)
	conn.EXPECT().
		ExecContext(ctx, "SELECT 1").
		Return(nil, clickhouseErr).
		Once()

	repo := NewClickhouse(conn, &Options{
		TableName:   "migrates",
		SchemaName:  "default",
		ClusterName: "test_cluster",
		Replicated:  false,
	})
	err := repo.ExecQuery(ctx, "SELECT 1")

	require.Error(t, err)
	var dbErr *DBError
	require.ErrorAs(t, err, &dbErr)
}

func TestClickhouse_ExecQueryTransaction_Successfully(t *testing.T) {
	ctx := context.Background()

	conn := NewMockConnection(t)
	conn.EXPECT().
		Transaction(ctx, mock.AnythingOfType("func(context.Context) error")).
		Return(nil).
		Once()

	repo := NewClickhouse(conn, &Options{
		TableName:   "migrates",
		SchemaName:  "default",
		ClusterName: "test_cluster",
		Replicated:  false,
	})
	err := repo.ExecQueryTransaction(ctx, func(ctx context.Context) error {
		return nil
	})

	require.NoError(t, err)
}

func TestClickhouse_MigrationsCount_Successfully(t *testing.T) {
	ctx := context.Background()

	expectedSQL := `SELECT count(*) FROM default.d_migrates WHERE is_deleted = 0`

	rows := sqlex.NewRowsWithSlice([]interface{}{5})

	conn := NewMockConnection(t)
	conn.EXPECT().
		QueryContext(ctx, mock.MatchedBy(thelp.CompareSQL(expectedSQL))).
		Return(rows, nil).
		Once()

	repo := NewClickhouse(conn, &Options{
		TableName:   "migrates",
		SchemaName:  "default",
		ClusterName: "test_cluster",
		Replicated:  false,
	})
	count, err := repo.MigrationsCount(ctx)

	require.NoError(t, err)
	assert.Equal(t, 5, count)
}

func TestClickhouse_MigrationsCount_Failure(t *testing.T) {
	ctx := context.Background()

	conn := NewMockConnection(t)
	conn.EXPECT().
		QueryContext(ctx, mock.AnythingOfType("string")).
		Return(nil, errors.New("connection error")).
		Once()

	repo := NewClickhouse(conn, &Options{
		TableName:   "migrates",
		SchemaName:  "default",
		ClusterName: "test_cluster",
		Replicated:  false,
	})
	_, err := repo.MigrationsCount(ctx)

	require.Error(t, err)
}

func TestClickhouse_QueryScalar_Successfully(t *testing.T) {
	ctx := context.Background()

	rows := sqlex.NewRowsWithSlice([]interface{}{42})

	conn := NewMockConnection(t)
	conn.EXPECT().
		QueryContext(ctx, "SELECT count(*)").
		Return(rows, nil).
		Once()

	repo := NewClickhouse(conn, &Options{
		TableName:   "migrates",
		SchemaName:  "default",
		ClusterName: "test_cluster",
		Replicated:  false,
	})
	var result int
	err := repo.QueryScalar(ctx, "SELECT count(*)", &result)

	require.NoError(t, err)
	assert.Equal(t, 42, result)
}

func TestClickhouse_QueryScalar_InvalidPtr_Failure(t *testing.T) {
	ctx := context.Background()

	conn := NewMockConnection(t)

	repo := NewClickhouse(conn, &Options{
		TableName:   "migrates",
		SchemaName:  "default",
		ClusterName: "test_cluster",
		Replicated:  false,
	})
	var result []int
	err := repo.QueryScalar(ctx, "SELECT count(*)", &result)

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrPtrValueMustBeAPointerAndScalar)
}

func TestClickhouse_ExistsMigration_Exists_Successfully(t *testing.T) {
	ctx := context.Background()

	rows := sqlex.NewRowsWithSlice([]interface{}{1})

	conn := NewMockConnection(t)
	conn.EXPECT().
		QueryContext(ctx, mock.AnythingOfType("string"), "210328_221600_test").
		Return(rows, nil).
		Once()

	repo := NewClickhouse(conn, &Options{
		TableName:   "migrates",
		SchemaName:  "default",
		ClusterName: "test_cluster",
		Replicated:  false,
	})
	exists, err := repo.ExistsMigration(ctx, "210328_221600_test")

	require.NoError(t, err)
	assert.True(t, exists)
}

func TestClickhouse_ExistsMigration_NotExists_Successfully(t *testing.T) {
	ctx := context.Background()

	rows := sqlex.NewRowsWithSlice([]interface{}{})

	conn := NewMockConnection(t)
	conn.EXPECT().
		QueryContext(ctx, mock.AnythingOfType("string"), "210328_221600_test").
		Return(rows, nil).
		Once()

	repo := NewClickhouse(conn, &Options{
		TableName:   "migrates",
		SchemaName:  "default",
		ClusterName: "test_cluster",
		Replicated:  false,
	})
	exists, err := repo.ExistsMigration(ctx, "210328_221600_test")

	require.NoError(t, err)
	assert.False(t, exists)
}

func TestClickhouse_TableNameWithSchema_Successfully(t *testing.T) {
	tests := []struct {
		name       string
		tableName  string
		schemaName string
		want       string
	}{
		{
			name:       "standard names",
			tableName:  "migration",
			schemaName: "default",
			want:       "default.migration",
		},
		{
			name:       "custom names",
			tableName:  "my_migrations",
			schemaName: "mydb",
			want:       "mydb.my_migrations",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewClickhouse(nil, &Options{
				TableName:  tt.tableName,
				SchemaName: tt.schemaName,
			})
			got := repo.TableNameWithSchema()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestClickhouse_DropMigrationHistoryTable_Successfully(t *testing.T) {
	ctx := context.Background()

	conn := NewMockConnection(t)
	conn.EXPECT().
		ExecContext(ctx, mock.MatchedBy(thelp.CompareSQL("DROP TABLE default.migrates ON CLUSTER test_cluster NO DELAY"))).
		Return(nil, nil).
		Once()
	conn.EXPECT().
		ExecContext(ctx, mock.MatchedBy(thelp.CompareSQL("DROP TABLE default.d_migrates ON CLUSTER test_cluster NO DELAY"))).
		Return(nil, nil).
		Once()

	repo := NewClickhouse(conn, &Options{
		TableName:   "migrates",
		SchemaName:  "default",
		ClusterName: "test_cluster",
		Replicated:  false,
	})
	err := repo.DropMigrationHistoryTable(ctx)

	require.NoError(t, err)
}

func TestClickhouse_DropMigrationHistoryTable_NoCluster_Successfully(t *testing.T) {
	ctx := context.Background()

	conn := NewMockConnection(t)
	conn.EXPECT().
		ExecContext(ctx, mock.MatchedBy(thelp.CompareSQL("DROP TABLE default.migrates"))).
		Return(nil, nil).
		Once()

	repo := NewClickhouse(conn, &Options{
		TableName:  "migrates",
		SchemaName: "default",
	})
	err := repo.DropMigrationHistoryTable(ctx)

	require.NoError(t, err)
}

func TestClickhouse_DropMigrationHistoryTable_Failure(t *testing.T) {
	ctx := context.Background()

	conn := NewMockConnection(t)
	conn.EXPECT().
		ExecContext(ctx, mock.AnythingOfType("string")).
		Return(nil, errors.New("drop failed")).
		Once()

	repo := NewClickhouse(conn, &Options{
		TableName:   "migrates",
		SchemaName:  "default",
		ClusterName: "test_cluster",
		Replicated:  false,
	})
	err := repo.DropMigrationHistoryTable(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "drop migration history table")
}

func TestClickhouse_CreateMigrationHistoryTable_NoCluster_Successfully(t *testing.T) {
	ctx := context.Background()

	expectedSQL := `
		CREATE TABLE default.migrates (
			version String,
			date Date DEFAULT toDate(apply_time),
			apply_time UInt32,
			is_deleted UInt8
		) ENGINE = ReplacingMergeTree(apply_time)
		PRIMARY KEY (version)
		PARTITION BY (toYYYYMM(date))
		ORDER BY (version)
		SETTINGS index_granularity=8192
	`

	conn := NewMockConnection(t)
	conn.EXPECT().
		ExecContext(ctx, mock.MatchedBy(thelp.CompareSQL(expectedSQL))).
		Return(nil, nil).
		Once()

	repo := NewClickhouse(conn, &Options{
		TableName:  "migrates",
		SchemaName: "default",
	})
	err := repo.CreateMigrationHistoryTable(ctx)

	require.NoError(t, err)
}

func TestClickhouse_CreateMigrationHistoryTable_Replicated_Successfully(t *testing.T) {
	ctx := context.Background()

	conn := NewMockConnection(t)
	conn.EXPECT().
		ExecContext(ctx, mock.AnythingOfType("string")).
		Return(nil, nil).
		Once()

	repo := NewClickhouse(conn, &Options{
		TableName:   "migrates",
		SchemaName:  "default",
		ClusterName: "test_cluster",
		Replicated:  true,
	})
	err := repo.CreateMigrationHistoryTable(ctx)

	require.NoError(t, err)
}

func TestClickhouse_CreateMigrationHistoryTable_Failure(t *testing.T) {
	ctx := context.Background()

	conn := NewMockConnection(t)
	conn.EXPECT().
		ExecContext(ctx, mock.AnythingOfType("string")).
		Return(nil, errors.New("create failed")).
		Once()

	repo := NewClickhouse(conn, &Options{
		TableName:  "migrates",
		SchemaName: "default",
	})
	err := repo.CreateMigrationHistoryTable(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "create migration history table")
}

func TestClickhouse_dbError_ClickhouseException_Successfully(t *testing.T) {
	repo := NewClickhouse(nil, &Options{
		TableName:  "migrates",
		SchemaName: "default",
	})

	clickhouseErr := &clickhouse.Exception{
		Code:       62,
		Message:    "Syntax error",
		StackTrace: "trace info",
	}
	query := "SELECT * FROM migrates"

	err := repo.dbError(clickhouseErr, query)

	var dbErr *DBError
	require.ErrorAs(t, err, &dbErr)
	assert.Equal(t, "Syntax error", dbErr.Message)
	assert.Equal(t, "trace info", dbErr.Details)
	assert.Equal(t, query, dbErr.InternalQuery)
}

func TestClickhouse_dbError_NonClickhouseException_Successfully(t *testing.T) {
	repo := NewClickhouse(nil, &Options{
		TableName:  "migrates",
		SchemaName: "default",
	})

	regularErr := errors.New("some other error")
	query := "SELECT * FROM migrates"

	err := repo.dbError(regularErr, query)

	assert.Equal(t, regularErr, err)

	var dbErr *DBError
	assert.False(t, errors.As(err, &dbErr))
}

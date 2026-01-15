package repository

import (
	"context"
	"testing"

	"github.com/pkg/errors"
	"github.com/raoptimus/db-migrator.go/internal/testutil"
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

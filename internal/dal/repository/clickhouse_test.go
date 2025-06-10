package repository

import (
	"context"
	"testing"

	"github.com/raoptimus/db-migrator.go/internal/dal/repository/mockrepository"
	"github.com/raoptimus/db-migrator.go/pkg/testhelp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
	conn := mockrepository.NewConnection(t)
	conn.EXPECT().
		ExecContext(ctx, mock.MatchedBy(testhelper.CompareSQL(expectedSQL))).
		Return(nil, nil).
		Once()
	conn.EXPECT().
		ExecContext(ctx, mock.MatchedBy(testhelper.CompareSQL(expectedSQL2))).
		Return(nil, nil).
		Once()

	repo := NewClickhouse(conn, &Options{
		TableName:   "migrates",
		SchemaName:  "default",
		ClusterName: "test_cluster",
		Replicated:  false,
	})
	err := repo.CreateMigrationHistoryTable(ctx)
	assert.NoError(t, err)
}

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
	"github.com/raoptimus/db-migrator.go/internal/infrastructure/dal/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// historyNS is the default history namespace slice used across tests.
var historyNS = []string{"migration"}

// newIcebergRepo creates a default test Iceberg repo with a fresh mock catalog.
func newIcebergRepo(t *testing.T) (*Iceberg, *MockIcebergCatalog) {
	t.Helper()
	cat := NewMockIcebergCatalog(t)
	repo := NewIceberg(cat, &Options{TableName: "migration"})
	return repo, cat
}

// --- CreateMigrationHistoryTable ---

func TestIceberg_CreateMigrationHistoryTable_Successfully(t *testing.T) {
	ctx := context.Background()
	repo, cat := newIcebergRepo(t)

	cat.EXPECT().
		CreateNamespace(ctx, historyNS, (map[string]string)(nil)).
		Return(nil).
		Once()

	err := repo.CreateMigrationHistoryTable(ctx)
	require.NoError(t, err)
}

func TestIceberg_CreateMigrationHistoryTable_Failure(t *testing.T) {
	ctx := context.Background()
	repo, cat := newIcebergRepo(t)

	catalogErr := errors.New("catalog unavailable")
	cat.EXPECT().
		CreateNamespace(ctx, historyNS, (map[string]string)(nil)).
		Return(catalogErr).
		Once()

	err := repo.CreateMigrationHistoryTable(ctx)
	require.Error(t, err)
	assert.ErrorContains(t, err, "create migration history table")
	assert.ErrorContains(t, err, "catalog unavailable")
}

// --- DropMigrationHistoryTable ---

func TestIceberg_DropMigrationHistoryTable_Successfully(t *testing.T) {
	ctx := context.Background()
	repo, cat := newIcebergRepo(t)

	cat.EXPECT().
		DropNamespace(ctx, historyNS).
		Return(nil).
		Once()

	err := repo.DropMigrationHistoryTable(ctx)
	require.NoError(t, err)
}

func TestIceberg_DropMigrationHistoryTable_Failure(t *testing.T) {
	ctx := context.Background()
	repo, cat := newIcebergRepo(t)

	cat.EXPECT().
		DropNamespace(ctx, historyNS).
		Return(errors.New("drop failed")).
		Once()

	err := repo.DropMigrationHistoryTable(ctx)
	require.Error(t, err)
	assert.ErrorContains(t, err, "drop migration history table")
}

// --- HasMigrationHistoryTable ---

func TestIceberg_HasMigrationHistoryTable_Exists(t *testing.T) {
	ctx := context.Background()
	repo, cat := newIcebergRepo(t)

	cat.EXPECT().
		NamespaceExists(ctx, historyNS).
		Return(true, nil).
		Once()

	exists, err := repo.HasMigrationHistoryTable(ctx)
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestIceberg_HasMigrationHistoryTable_NotExists(t *testing.T) {
	ctx := context.Background()
	repo, cat := newIcebergRepo(t)

	cat.EXPECT().
		NamespaceExists(ctx, historyNS).
		Return(false, nil).
		Once()

	exists, err := repo.HasMigrationHistoryTable(ctx)
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestIceberg_HasMigrationHistoryTable_Failure(t *testing.T) {
	ctx := context.Background()
	repo, cat := newIcebergRepo(t)

	cat.EXPECT().
		NamespaceExists(ctx, historyNS).
		Return(false, errors.New("network error")).
		Once()

	_, err := repo.HasMigrationHistoryTable(ctx)
	require.Error(t, err)
	assert.ErrorContains(t, err, "check migration history table")
}

// --- InsertMigrationWithApplyTime ---

func TestIceberg_InsertMigrationWithApplyTime_Successfully(t *testing.T) {
	ctx := context.Background()
	repo, cat := newIcebergRepo(t)

	version := "210328_221600_create_users"
	applyTime := int64(1616968560)
	expectedUpdates := map[string]string{
		"migrate." + version: "1616968560",
	}

	cat.EXPECT().
		UpdateNamespaceProperties(ctx, historyNS, ([]string)(nil), expectedUpdates).
		Return(nil).
		Once()

	err := repo.InsertMigrationWithApplyTime(ctx, version, applyTime)
	require.NoError(t, err)
}

func TestIceberg_InsertMigrationWithApplyTime_Failure(t *testing.T) {
	ctx := context.Background()
	repo, cat := newIcebergRepo(t)

	cat.EXPECT().
		UpdateNamespaceProperties(ctx, historyNS, ([]string)(nil), map[string]string{
			"migrate.210328_221600_create_users": "1616968560",
		}).
		Return(errors.New("write error")).
		Once()

	err := repo.InsertMigrationWithApplyTime(ctx, "210328_221600_create_users", 1616968560)
	require.Error(t, err)
	assert.ErrorContains(t, err, "insert migration")
}

// --- InsertMigration (uses time.Now internally; verify catalog call happens) ---

func TestIceberg_InsertMigration_CallsCatalog(t *testing.T) {
	ctx := context.Background()
	repo, cat := newIcebergRepo(t)

	version := "210328_221600_create_users"

	// We cannot assert the exact apply_time because it uses time.Now().
	// We verify that UpdateNamespaceProperties is called with the correct key.
	cat.EXPECT().
		UpdateNamespaceProperties(
			ctx,
			historyNS,
			([]string)(nil),
			mock.MatchedBy(matchUpdatesContainsKey("migrate."+version)),
		).
		Return(nil).
		Once()

	err := repo.InsertMigration(ctx, version)
	require.NoError(t, err)
}

// --- RemoveMigration ---

func TestIceberg_RemoveMigration_Successfully(t *testing.T) {
	ctx := context.Background()
	repo, cat := newIcebergRepo(t)

	version := "210328_221600_create_users"
	expectedRemovals := []string{"migrate." + version}

	cat.EXPECT().
		UpdateNamespaceProperties(ctx, historyNS, expectedRemovals, (map[string]string)(nil)).
		Return(nil).
		Once()

	err := repo.RemoveMigration(ctx, version)
	require.NoError(t, err)
}

func TestIceberg_RemoveMigration_Failure(t *testing.T) {
	ctx := context.Background()
	repo, cat := newIcebergRepo(t)

	cat.EXPECT().
		UpdateNamespaceProperties(ctx, historyNS,
			[]string{"migrate.210328_221600_create_users"},
			(map[string]string)(nil),
		).
		Return(errors.New("remove error")).
		Once()

	err := repo.RemoveMigration(ctx, "210328_221600_create_users")
	require.Error(t, err)
	assert.ErrorContains(t, err, "remove migration")
}

// --- Migrations ---

func TestIceberg_Migrations_Successfully(t *testing.T) {
	ctx := context.Background()
	repo, cat := newIcebergRepo(t)

	props := map[string]string{
		"migrate.210328_221600_create_users": "1616968560",
		"migrate.210329_121500_add_index":    "1617020100",
		"other.key":                      "ignored",
	}

	cat.EXPECT().
		LoadNamespaceProperties(ctx, historyNS).
		Return(props, nil).
		Once()

	migrations, err := repo.Migrations(ctx, 0)
	require.NoError(t, err)
	// All mig.* entries should be returned.
	require.Len(t, migrations, 2)
	// Returned in DESC order by version (lexicographic).
	assert.Equal(t, "210329_121500_add_index", migrations[0].Version)
	assert.Equal(t, "210328_221600_create_users", migrations[1].Version)
}

func TestIceberg_Migrations_WithLimit(t *testing.T) {
	ctx := context.Background()
	repo, cat := newIcebergRepo(t)

	props := map[string]string{
		"migrate.210328_221600_v1": "1616968560",
		"migrate.210329_120000_v2": "1617020100",
		"migrate.210330_090000_v3": "1617098400",
	}

	cat.EXPECT().
		LoadNamespaceProperties(ctx, historyNS).
		Return(props, nil).
		Once()

	migrations, err := repo.Migrations(ctx, 2)
	require.NoError(t, err)
	require.Len(t, migrations, 2)
	// Highest two versions in DESC order.
	assert.Equal(t, "210330_090000_v3", migrations[0].Version)
	assert.Equal(t, "210329_120000_v2", migrations[1].Version)
}

func TestIceberg_Migrations_EmptyNamespace(t *testing.T) {
	ctx := context.Background()
	repo, cat := newIcebergRepo(t)

	cat.EXPECT().
		LoadNamespaceProperties(ctx, historyNS).
		Return(map[string]string{}, nil).
		Once()

	migrations, err := repo.Migrations(ctx, 0)
	require.NoError(t, err)
	assert.Empty(t, migrations)
}

func TestIceberg_Migrations_NonMigKeyIgnored(t *testing.T) {
	ctx := context.Background()
	repo, cat := newIcebergRepo(t)

	props := map[string]string{
		"location":                    "s3://bucket/path",
		"owner":                       "admin",
		"migrate.210328_221600_my_table":  "1616968560",
	}

	cat.EXPECT().
		LoadNamespaceProperties(ctx, historyNS).
		Return(props, nil).
		Once()

	migrations, err := repo.Migrations(ctx, 0)
	require.NoError(t, err)
	require.Len(t, migrations, 1)
	assert.Equal(t, "210328_221600_my_table", migrations[0].Version)
}

func TestIceberg_Migrations_InvalidApplyTime_Failure(t *testing.T) {
	ctx := context.Background()
	repo, cat := newIcebergRepo(t)

	props := map[string]string{
		"migrate.210328_221600_broken": "not-a-number",
	}

	cat.EXPECT().
		LoadNamespaceProperties(ctx, historyNS).
		Return(props, nil).
		Once()

	_, err := repo.Migrations(ctx, 0)
	require.Error(t, err)
	assert.ErrorContains(t, err, "invalid apply_time value")
	assert.ErrorContains(t, err, "210328_221600_broken")
}

func TestIceberg_Migrations_CatalogFailure(t *testing.T) {
	ctx := context.Background()
	repo, cat := newIcebergRepo(t)

	cat.EXPECT().
		LoadNamespaceProperties(ctx, historyNS).
		Return(nil, errors.New("namespace not found")).
		Once()

	_, err := repo.Migrations(ctx, 0)
	require.Error(t, err)
	assert.ErrorContains(t, err, "get migrations")
}

// --- MigrationsCount ---

func TestIceberg_MigrationsCount_Successfully(t *testing.T) {
	ctx := context.Background()
	repo, cat := newIcebergRepo(t)

	props := map[string]string{
		"migrate.210328_221600_v1": "1616968560",
		"migrate.210329_120000_v2": "1617020100",
		"location":             "s3://bucket",
	}

	cat.EXPECT().
		LoadNamespaceProperties(ctx, historyNS).
		Return(props, nil).
		Once()

	count, err := repo.MigrationsCount(ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, count)
}

func TestIceberg_MigrationsCount_Empty(t *testing.T) {
	ctx := context.Background()
	repo, cat := newIcebergRepo(t)

	cat.EXPECT().
		LoadNamespaceProperties(ctx, historyNS).
		Return(map[string]string{}, nil).
		Once()

	count, err := repo.MigrationsCount(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

// --- ExistsMigration ---

func TestIceberg_ExistsMigration_Exists(t *testing.T) {
	ctx := context.Background()
	repo, cat := newIcebergRepo(t)

	version := "210328_221600_create_users"
	props := map[string]string{
		"migrate." + version: "1616968560",
	}

	cat.EXPECT().
		LoadNamespaceProperties(ctx, historyNS).
		Return(props, nil).
		Once()

	exists, err := repo.ExistsMigration(ctx, version)
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestIceberg_ExistsMigration_NotExists(t *testing.T) {
	ctx := context.Background()
	repo, cat := newIcebergRepo(t)

	cat.EXPECT().
		LoadNamespaceProperties(ctx, historyNS).
		Return(map[string]string{}, nil).
		Once()

	exists, err := repo.ExistsMigration(ctx, "210328_221600_create_users")
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestIceberg_ExistsMigration_Failure(t *testing.T) {
	ctx := context.Background()
	repo, cat := newIcebergRepo(t)

	cat.EXPECT().
		LoadNamespaceProperties(ctx, historyNS).
		Return(nil, errors.New("catalog error")).
		Once()

	_, err := repo.ExistsMigration(ctx, "210328_221600_create_users")
	require.Error(t, err)
	assert.ErrorContains(t, err, "check migration exists")
}

// --- MigrationsByMaxApplyTime ---

func TestIceberg_MigrationsByMaxApplyTime_Successfully(t *testing.T) {
	ctx := context.Background()
	repo, cat := newIcebergRepo(t)

	props := map[string]string{
		"migrate.210328_221600_create_users": "1616968560",
		"migrate.210329_121500_add_index":    "1617020100",
		"migrate.210330_090000_drop_column":  "1617020100", // same apply_time as previous (batch)
	}

	cat.EXPECT().
		LoadNamespaceProperties(ctx, historyNS).
		Return(props, nil).
		Once()

	migrations, err := repo.MigrationsByMaxApplyTime(ctx)
	require.NoError(t, err)
	require.Len(t, migrations, 2)
	// Verify only the batch with maxApplyTime=1617020100 is returned.
	for _, m := range migrations {
		assert.Equal(t, int64(1617020100), m.ApplyTime)
	}
	// Sorted descending by version.
	assert.Equal(t, "210330_090000_drop_column", migrations[0].Version)
	assert.Equal(t, "210329_121500_add_index", migrations[1].Version)
}

func TestIceberg_MigrationsByMaxApplyTime_SingleMigration(t *testing.T) {
	ctx := context.Background()
	repo, cat := newIcebergRepo(t)

	props := map[string]string{
		"migrate.210328_221600_create_users": "1616968560",
	}

	cat.EXPECT().
		LoadNamespaceProperties(ctx, historyNS).
		Return(props, nil).
		Once()

	migrations, err := repo.MigrationsByMaxApplyTime(ctx)
	require.NoError(t, err)
	require.Len(t, migrations, 1)
	assert.Equal(t, "210328_221600_create_users", migrations[0].Version)
}

func TestIceberg_MigrationsByMaxApplyTime_EmptyNamespace(t *testing.T) {
	ctx := context.Background()
	repo, cat := newIcebergRepo(t)

	cat.EXPECT().
		LoadNamespaceProperties(ctx, historyNS).
		Return(map[string]string{}, nil).
		Once()

	migrations, err := repo.MigrationsByMaxApplyTime(ctx)
	require.NoError(t, err)
	assert.Nil(t, migrations)
}

func TestIceberg_MigrationsByMaxApplyTime_AllShareSameApplyTime(t *testing.T) {
	ctx := context.Background()
	repo, cat := newIcebergRepo(t)

	const applyTime = "1617020100"
	props := map[string]string{
		"migrate.210329_121500_v1": applyTime,
		"migrate.210329_121501_v2": applyTime,
		"migrate.210329_121502_v3": applyTime,
	}

	cat.EXPECT().
		LoadNamespaceProperties(ctx, historyNS).
		Return(props, nil).
		Once()

	migrations, err := repo.MigrationsByMaxApplyTime(ctx)
	require.NoError(t, err)
	require.Len(t, migrations, 3)
}

func TestIceberg_MigrationsByMaxApplyTime_Failure(t *testing.T) {
	ctx := context.Background()
	repo, cat := newIcebergRepo(t)

	cat.EXPECT().
		LoadNamespaceProperties(ctx, historyNS).
		Return(nil, errors.New("catalog unavailable")).
		Once()

	_, err := repo.MigrationsByMaxApplyTime(ctx)
	require.Error(t, err)
	assert.ErrorContains(t, err, "get migrations by max apply time")
}

// --- SupportsDDLTransactions / ExecQueryTransaction ---

func TestIceberg_SupportsDDLTransactions(t *testing.T) {
	cat := NewMockIcebergCatalog(t)
	repo := NewIceberg(cat, &Options{TableName: "migration"})
	assert.False(t, repo.SupportsDDLTransactions())
}

func TestIceberg_ExecQueryTransaction_Passthrough(t *testing.T) {
	ctx := context.Background()
	cat := NewMockIcebergCatalog(t)
	repo := NewIceberg(cat, &Options{TableName: "migration"})

	called := false
	err := repo.ExecQueryTransaction(ctx, func(_ context.Context) error {
		called = true
		return nil
	})
	require.NoError(t, err)
	assert.True(t, called)
}

// --- TableNameWithSchema ---

func TestIceberg_TableNameWithSchema(t *testing.T) {
	tests := []struct {
		name      string
		tableName string
		want      string
	}{
		{
			name:      "default table name",
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
			cat := NewMockIcebergCatalog(t)
			repo := NewIceberg(cat, &Options{TableName: tt.tableName})
			assert.Equal(t, tt.want, repo.TableNameWithSchema())
		})
	}
}

// --- ExecQuery / QueryScalar ---

// TestIceberg_ExecQuery_InvalidDDL verifies that an unparseable statement returns an error.
// ExecQuery calls Warehouse() to get the catalog prefix before parsing.
func TestIceberg_ExecQuery_InvalidDDL(t *testing.T) {
	ctx := context.Background()
	repo, cat := newIcebergRepo(t)

	cat.EXPECT().Warehouse().Return("").Once()

	// "CREATE TABLE foo ..." is invalid because "..." is not valid DDL.
	err := repo.ExecQuery(ctx, "CREATE TABLE foo ...")
	require.Error(t, err)
}

func TestIceberg_QueryScalar_NotImplemented(t *testing.T) {
	ctx := context.Background()
	cat := NewMockIcebergCatalog(t)
	repo := NewIceberg(cat, &Options{TableName: "migration"})

	var result int
	err := repo.QueryScalar(ctx, "SELECT 1", &result)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrNotImplemented)
}

// --- helpers ---

// matchUpdatesContainsKey returns a matcher that asserts the updates map has the given key.
func matchUpdatesContainsKey(key string) func(map[string]string) bool {
	return func(m map[string]string) bool {
		_, ok := m[key]
		return ok
	}
}

// Ensure entity.Migrations sorting behaviour (sanity check for the reverseSlice helper).
func TestIceberg_reverseSlice(t *testing.T) {
	s := entity.Migrations{
		{Version: "210328_221600_v1", ApplyTime: 1},
		{Version: "210329_120000_v2", ApplyTime: 2},
		{Version: "210330_090000_v3", ApplyTime: 3},
	}
	s.SortByVersion() // already ascending
	reverseSlice(s)
	require.Equal(t, "210330_090000_v3", s[0].Version)
	require.Equal(t, "210328_221600_v1", s[2].Version)
}

// ─── ExecQuery — dispatch tests ──────────────────────────────────────────────

// TestIceberg_ExecQuery_Dispatch verifies that every DDL operation kind is
// routed to the correct IcebergCatalog method. No network; uses MockIcebergCatalog.
func TestIceberg_ExecQuery_Dispatch(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name  string
		query string
		setup func(cat *MockIcebergCatalog)
	}{
		{
			name:  "CREATE NAMESPACE → CreateNamespace",
			query: "CREATE NAMESPACE analytics",
			setup: func(cat *MockIcebergCatalog) {
				cat.EXPECT().
					Warehouse().Return("").Once()
				cat.EXPECT().
					CreateNamespace(ctx, []string{"analytics"}, (map[string]string)(nil)).
					Return(nil).Once()
			},
		},
		{
			name:  "DROP NAMESPACE → DropNamespace",
			query: "DROP NAMESPACE analytics",
			setup: func(cat *MockIcebergCatalog) {
				cat.EXPECT().Warehouse().Return("").Once()
				cat.EXPECT().
					DropNamespace(ctx, []string{"analytics"}).
					Return(nil).Once()
			},
		},
		{
			name:  "CREATE TABLE → CreateTable",
			query: "CREATE TABLE analytics.events (id long)",
			setup: func(cat *MockIcebergCatalog) {
				cat.EXPECT().Warehouse().Return("").Once()
				cat.EXPECT().
					CreateTable(ctx, mock.AnythingOfType("ddl.Ident"), mock.AnythingOfType("ddl.CreateTableSpec")).
					Return(nil).Once()
			},
		},
		{
			name:  "DROP TABLE → DropTable",
			query: "DROP TABLE analytics.events",
			setup: func(cat *MockIcebergCatalog) {
				cat.EXPECT().Warehouse().Return("").Once()
				cat.EXPECT().
					DropTable(ctx, mock.AnythingOfType("ddl.Ident")).
					Return(nil).Once()
			},
		},
		{
			name:  "RENAME TABLE → RenameTable",
			query: "RENAME TABLE analytics.events TO analytics.events_v2",
			setup: func(cat *MockIcebergCatalog) {
				cat.EXPECT().Warehouse().Return("").Once()
				cat.EXPECT().
					RenameTable(ctx, mock.AnythingOfType("ddl.Ident"), mock.AnythingOfType("ddl.Ident")).
					Return(nil).Once()
			},
		},
		{
			name:  "ADD COLUMN → ApplySchemaChange",
			query: "ALTER TABLE analytics.events ADD COLUMN name string",
			setup: func(cat *MockIcebergCatalog) {
				cat.EXPECT().Warehouse().Return("").Once()
				cat.EXPECT().
					ApplySchemaChange(ctx, mock.AnythingOfType("ddl.Operation")).
					Return(nil).Once()
			},
		},
		{
			name:  "DROP COLUMN → ApplySchemaChange",
			query: "ALTER TABLE analytics.events DROP COLUMN name",
			setup: func(cat *MockIcebergCatalog) {
				cat.EXPECT().Warehouse().Return("").Once()
				cat.EXPECT().
					ApplySchemaChange(ctx, mock.AnythingOfType("ddl.Operation")).
					Return(nil).Once()
			},
		},
		{
			name:  "RENAME COLUMN → ApplySchemaChange",
			query: "ALTER TABLE analytics.events RENAME COLUMN name TO title",
			setup: func(cat *MockIcebergCatalog) {
				cat.EXPECT().Warehouse().Return("").Once()
				cat.EXPECT().
					ApplySchemaChange(ctx, mock.AnythingOfType("ddl.Operation")).
					Return(nil).Once()
			},
		},
		{
			name:  "ALTER COLUMN TYPE → ApplySchemaChange",
			query: "ALTER TABLE analytics.events ALTER COLUMN id TYPE long",
			setup: func(cat *MockIcebergCatalog) {
				cat.EXPECT().Warehouse().Return("").Once()
				cat.EXPECT().
					ApplySchemaChange(ctx, mock.AnythingOfType("ddl.Operation")).
					Return(nil).Once()
			},
		},
		{
			name:  "ADD PARTITION FIELD → ApplySpecChange",
			query: "ALTER TABLE analytics.events ADD PARTITION FIELD days(ts)",
			setup: func(cat *MockIcebergCatalog) {
				cat.EXPECT().Warehouse().Return("").Once()
				cat.EXPECT().
					ApplySpecChange(ctx, mock.AnythingOfType("ddl.Operation")).
					Return(nil).Once()
			},
		},
		{
			name:  "DROP PARTITION FIELD → ApplySpecChange",
			query: "ALTER TABLE analytics.events DROP PARTITION FIELD days(ts)",
			setup: func(cat *MockIcebergCatalog) {
				cat.EXPECT().Warehouse().Return("").Once()
				cat.EXPECT().
					ApplySpecChange(ctx, mock.AnythingOfType("ddl.Operation")).
					Return(nil).Once()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, cat := newIcebergRepo(t)
			tt.setup(cat)
			err := repo.ExecQuery(ctx, tt.query)
			require.NoError(t, err)
		})
	}
}

// TestIceberg_ExecQuery_ParseError verifies that a parse error is propagated directly.
func TestIceberg_ExecQuery_ParseError(t *testing.T) {
	ctx := context.Background()
	repo, cat := newIcebergRepo(t)

	cat.EXPECT().Warehouse().Return("").Once()

	err := repo.ExecQuery(ctx, "NOT VALID DDL AT ALL")
	require.Error(t, err)
}

// TestIceberg_ExecQuery_CatalogError verifies that a catalog error is propagated.
func TestIceberg_ExecQuery_CatalogError(t *testing.T) {
	ctx := context.Background()
	repo, cat := newIcebergRepo(t)

	cat.EXPECT().Warehouse().Return("").Once()
	cat.EXPECT().
		DropTable(ctx, mock.AnythingOfType("ddl.Ident")).
		Return(errors.New("catalog unreachable")).Once()

	err := repo.ExecQuery(ctx, "DROP TABLE analytics.events")
	require.Error(t, err)
	assert.ErrorContains(t, err, "catalog unreachable")
}

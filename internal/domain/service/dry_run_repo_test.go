/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package service

import (
	"context"
	"testing"

	"github.com/raoptimus/db-migrator.go/internal/infrastructure/dal/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Before CreateMigrationHistoryTable, the dry-run repository must delegate read
// operations to the wrapped repository.
func TestDryRunRepository_Delegates_BeforeVirtualTable_Successfully(t *testing.T) {
	ctx := t.Context()
	repo := NewMockRepository(t)
	sut := NewDryRunRepository(repo)

	repo.EXPECT().ExistsMigration(ctx, "230101_120000_a").Return(true, nil).Once()
	repo.EXPECT().Migrations(ctx, 10).Return(entity.Migrations{{Version: "230101_120000_a"}}, nil).Once()
	repo.EXPECT().HasMigrationHistoryTable(ctx).Return(true, nil).Once()
	repo.EXPECT().MigrationsCount(ctx).Return(3, nil).Once()
	repo.EXPECT().MigrationsByMaxApplyTime(ctx).Return(entity.Migrations{{Version: "230101_120000_a"}}, nil).Once()
	repo.EXPECT().TableNameWithSchema().Return("public.migration").Once()

	exists, err := sut.ExistsMigration(ctx, "230101_120000_a")
	require.NoError(t, err)
	assert.True(t, exists)

	migrations, err := sut.Migrations(ctx, 10)
	require.NoError(t, err)
	assert.Len(t, migrations, 1)

	has, err := sut.HasMigrationHistoryTable(ctx)
	require.NoError(t, err)
	assert.True(t, has)

	count, err := sut.MigrationsCount(ctx)
	require.NoError(t, err)
	assert.Equal(t, 3, count)

	batch, err := sut.MigrationsByMaxApplyTime(ctx)
	require.NoError(t, err)
	assert.Len(t, batch, 1)

	assert.Equal(t, "public.migration", sut.TableNameWithSchema())
}

// After CreateMigrationHistoryTable, the virtual table masks the real one: reads
// return empty results and must NOT touch the wrapped repository. A mock with no
// expectations fails the test if any delegation happens.
func TestDryRunRepository_VirtualTable_MasksReads_Successfully(t *testing.T) {
	ctx := t.Context()
	repo := NewMockRepository(t)
	sut := NewDryRunRepository(repo)

	require.NoError(t, sut.CreateMigrationHistoryTable(ctx))

	exists, err := sut.ExistsMigration(ctx, "230101_120000_a")
	require.NoError(t, err)
	assert.False(t, exists)

	migrations, err := sut.Migrations(ctx, 10)
	require.NoError(t, err)
	assert.Nil(t, migrations)

	has, err := sut.HasMigrationHistoryTable(ctx)
	require.NoError(t, err)
	assert.True(t, has)

	count, err := sut.MigrationsCount(ctx)
	require.NoError(t, err)
	assert.Zero(t, count)

	batch, err := sut.MigrationsByMaxApplyTime(ctx)
	require.NoError(t, err)
	assert.Nil(t, batch)

	assert.Empty(t, sut.TableNameWithSchema())
}

// Write operations are no-ops in dry-run mode and never reach the wrapped repository.
func TestDryRunRepository_WriteOperations_AreNoOps_Successfully(t *testing.T) {
	ctx := t.Context()
	repo := NewMockRepository(t)
	sut := NewDryRunRepository(repo)

	assert.False(t, sut.SupportsDDLTransactions())
	require.NoError(t, sut.InsertMigration(ctx, "230101_120000_a"))
	require.NoError(t, sut.RemoveMigration(ctx, "230101_120000_a"))
	require.NoError(t, sut.ExecQuery(ctx, "CREATE TABLE t (id INT)"))
	require.NoError(t, sut.DropMigrationHistoryTable(ctx))
	require.NoError(t, sut.InsertMigrationWithApplyTime(ctx, "230101_120000_a", 1_700_000_000))

	// ExecQueryTransaction must not invoke the callback in dry-run mode.
	called := false
	require.NoError(t, sut.ExecQueryTransaction(ctx, func(context.Context) error {
		called = true
		return nil
	}))
	assert.False(t, called, "transaction callback must not run in dry-run mode")
}

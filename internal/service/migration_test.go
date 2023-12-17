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
	"io"
	"strings"
	"testing"

	"github.com/raoptimus/db-migrator.go/internal/action/mockaction"
	"github.com/raoptimus/db-migrator.go/internal/dal/entity"
	"github.com/raoptimus/db-migrator.go/internal/service/mockservice"
	"github.com/raoptimus/db-migrator.go/pkg/console"
	"github.com/stretchr/testify/assert"
)

func TestMigration_BeginCommand(t *testing.T) {
	repo := mockservice.NewRepository(t)
	file := mockservice.NewFile(t)
	c := mockaction.NewConsole(t)
	c.EXPECT().
		Infof("    > execute SQL: %s ...\n", "select 1").
		Once()
	serv := NewMigration(&Options{}, c, file, repo)
	serv.BeginCommand("select 1")
}

func TestMigration_ApplyFile_MultiSTMT_Successfully(t *testing.T) {
	t.Skip("Skip")
	ctx := context.Background()
	fileName := "000000_000000_test.up.sql"

	sqlReader := strings.NewReader("select 1; select 2;")
	sqlReaderCloser := io.NopCloser(sqlReader)

	file := mockservice.NewFile(t)
	file.EXPECT().Exists(fileName).Return(true, nil)
	file.EXPECT().Open(fileName).Return(sqlReaderCloser, nil)

	repo := mockservice.NewRepository(t)
	repo.EXPECT().
		ExecQuery(ctx, "select 1").
		RunAndReturn(func(ctx context.Context, s string, i ...interface{}) error {
			return nil
		}).
		Once()
	repo.EXPECT().
		ExecQuery(ctx, "select 2").
		RunAndReturn(func(ctx context.Context, s string, i ...interface{}) error {
			return nil
		}).
		Once()
	repo.EXPECT().
		InsertMigration(ctx, "000000_000000_test").
		Return(nil)

	serv := NewMigration(&Options{}, console.NewDummy(true), file, repo)
	err := serv.ApplyFile(ctx,
		&entity.Migration{Version: "000000_000000_test"},
		fileName,
		false,
	)
	assert.NoError(t, err)
}

func TestMigration_ApplyFile_SimpleSTMT_Successfully(t *testing.T) {
	t.Skip("Skip")
	ctx := context.Background()
	fileName := "000000_000000_test.up.sql"

	file := mockservice.NewFile(t)
	file.EXPECT().Exists(fileName).Return(true, nil)
	file.EXPECT().ReadAll(fileName).Return([]byte("select 1;"), nil)

	repo := mockservice.NewRepository(t)
	repo.EXPECT().
		ExecQuery(ctx, "select 1;").
		Return(nil)
	repo.EXPECT().
		InsertMigration(ctx, "000000_000000_test").
		Return(nil)

	serv := NewMigration(&Options{}, console.NewDummy(true), file, repo)

	err := serv.ApplyFile(ctx,
		&entity.Migration{Version: "000000_000000_test"},
		fileName,
		false,
	)
	assert.NoError(t, err)
}

/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package migrator

import (
	"github.com/raoptimus/db-migrator.go/internal/action"
	"github.com/raoptimus/db-migrator.go/internal/builder"
	"github.com/raoptimus/db-migrator.go/internal/dal/connection"
	"github.com/raoptimus/db-migrator.go/internal/dal/repository"
	"github.com/raoptimus/db-migrator.go/internal/service"
	"github.com/raoptimus/db-migrator.go/pkg/console"
	"github.com/raoptimus/db-migrator.go/pkg/iohelp"
	"github.com/raoptimus/db-migrator.go/pkg/timex"
)

type (
	DBService struct {
		options              *Options
		fileNameBuilder      FileNameBuilder
		migrationServiceFunc func() (*service.Migration, error)

		conn *connection.Connection
		repo *repository.Repository
	}
	Options struct {
		DSN                string
		Directory          string
		TableName          string
		ClusterName        string
		Compact            bool
		Interactive        bool
		MaxSQLOutputLength int
	}
)

func New(options *Options) *DBService {
	fb := builder.NewFileName(iohelp.StdFile, options.Directory)
	dbs := &DBService{
		options:         options,
		fileNameBuilder: fb,
	}
	dbs.migrationServiceFunc = dbs.migrationService
	return dbs
}

func (s *DBService) Create() *action.Create {
	return action.NewCreate(
		timex.StdTime,
		iohelp.StdFile,
		console.Std,
		s.fileNameBuilder,
		s.options.Directory,
	)
}

func (s *DBService) Upgrade() (*action.Upgrade, error) {
	serv, err := s.migrationServiceFunc()
	if err != nil {
		return nil, err
	}

	return action.NewUpgrade(
		console.Std,
		serv,
		s.fileNameBuilder,
		s.options.Interactive,
	), nil
}

func (s *DBService) Downgrade() (*action.Downgrade, error) {
	serv, err := s.migrationServiceFunc()
	if err != nil {
		return nil, err
	}
	return action.NewDowngrade(serv, s.fileNameBuilder, s.options.Interactive), nil
}

func (s *DBService) To() (*action.To, error) {
	serv, err := s.migrationServiceFunc()
	if err != nil {
		return nil, err
	}
	return action.NewTo(serv, s.fileNameBuilder, s.options.Interactive), nil
}

func (s *DBService) History() (*action.History, error) {
	serv, err := s.migrationServiceFunc()
	if err != nil {
		return nil, err
	}
	return action.NewHistory(serv), nil
}

func (s *DBService) HistoryNew() (*action.HistoryNew, error) {
	serv, err := s.migrationServiceFunc()
	if err != nil {
		return nil, err
	}
	return action.NewHistoryNew(serv), nil
}

func (s *DBService) Redo() (*action.Redo, error) {
	serv, err := s.migrationServiceFunc()
	if err != nil {
		return nil, err
	}
	return action.NewRedo(serv, s.fileNameBuilder, s.options.Interactive), nil
}

func (s *DBService) migrationService() (*service.Migration, error) {
	if s.conn == nil {
		conn, err := connection.New(s.options.DSN)
		if err != nil {
			return nil, err
		}
		s.conn = conn
	}

	if s.repo == nil {
		repo, err := repository.New(
			s.conn,
			&repository.Options{
				TableName:   s.options.TableName,
				ClusterName: s.options.ClusterName,
			},
		)
		if err != nil {
			return nil, err
		}
		s.repo = repo
	}

	return service.NewMigration(
		&service.Options{
			MaxSQLOutputLength: s.options.MaxSQLOutputLength,
			Directory:          s.options.Directory,
			Compact:            s.options.Compact,
		},
		console.Std,
		iohelp.StdFile,
		s.repo,
	), nil
}

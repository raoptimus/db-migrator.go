/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package migrator

import (
	"net/url"
	"strings"

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
		options         *Options
		fileNameBuilder FileNameBuilder

		conn    *connection.Connection
		repo    *repository.Repository
		service *service.Migration
	}
	Options struct {
		DSN                string
		Directory          string
		TableName          string
		ClusterName        string
		Replicated         bool
		Compact            bool
		Interactive        bool
		MaxSQLOutputLength int
	}
)

func New(options *Options) *DBService {
	fb := builder.NewFileName(iohelp.StdFile, options.Directory)
	return &DBService{
		options:         options,
		fileNameBuilder: fb,
	}
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
	serv, err := s.MigrationService()
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
	serv, err := s.MigrationService()
	if err != nil {
		return nil, err
	}

	return action.NewDowngrade(serv, s.fileNameBuilder, s.options.Interactive), nil
}

func (s *DBService) To() (*action.To, error) {
	serv, err := s.MigrationService()
	if err != nil {
		return nil, err
	}

	return action.NewTo(serv, s.fileNameBuilder, s.options.Interactive), nil
}

func (s *DBService) History() (*action.History, error) {
	serv, err := s.MigrationService()
	if err != nil {
		return nil, err
	}

	return action.NewHistory(serv), nil
}

func (s *DBService) HistoryNew() (*action.HistoryNew, error) {
	serv, err := s.MigrationService()
	if err != nil {
		return nil, err
	}

	return action.NewHistoryNew(serv), nil
}

func (s *DBService) Redo() (*action.Redo, error) {
	serv, err := s.MigrationService()
	if err != nil {
		return nil, err
	}

	return action.NewRedo(serv, s.fileNameBuilder, s.options.Interactive), nil
}

func (s *DBService) MigrationService() (*service.Migration, error) {
	if s.service != nil {
		return s.service, nil
	}

	var err error

	if s.conn == nil {
		s.conn, err = connection.New(s.options.DSN)
		if err != nil {
			return nil, err
		}
	}

	udsn, _, _ := strings.Cut(s.options.DSN, "@")
	dsn, err := url.Parse(udsn + "@")
	if err != nil {
		return nil, err
	}

	if s.repo == nil {
		s.repo, err = repository.New(
			s.conn,
			&repository.Options{
				TableName:   s.options.TableName,
				ClusterName: s.options.ClusterName,
				Replicated:  s.options.Replicated,
			},
		)
		if err != nil {
			return nil, err
		}
	}

	pass, _ := dsn.User.Password()

	s.service = service.NewMigration(
		&service.Options{
			MaxSQLOutputLength: s.options.MaxSQLOutputLength,
			Directory:          s.options.Directory,
			Compact:            s.options.Compact,

			Username: dsn.User.Username(),
			Password: pass,
		},
		console.Std,
		iohelp.StdFile,
		s.repo,
	)

	return s.service, nil
}

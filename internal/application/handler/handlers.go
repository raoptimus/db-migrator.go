/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package handler

import (
	"github.com/raoptimus/db-migrator.go/internal/application/presenter"
	"github.com/raoptimus/db-migrator.go/internal/domain/builder"
	iohelp "github.com/raoptimus/db-migrator.go/internal/helper/io"
)

type Handlers struct {
	Create     Handler
	Upgrade    Handler
	Downgrade  Handler
	Redo       Handler
	To         Handler
	History    Handler
	HistoryNew Handler
}

func NewHandlers(options *Options, logger Logger) *Handlers {
	fileNameBuilder := builder.NewFileName(iohelp.StdFile, options.Directory)
	migrationPresenter := presenter.NewMigrationPresenter(logger)

	return &Handlers{
		Create:     NewCreate(options, logger, iohelp.StdFile, fileNameBuilder),
		Upgrade:    NewServiceWrapHandler(options, logger, NewUpgrade(options, migrationPresenter, fileNameBuilder)),
		Downgrade:  NewServiceWrapHandler(options, logger, NewDowngrade(options, migrationPresenter, fileNameBuilder)),
		Redo:       NewServiceWrapHandler(options, logger, NewRedo(options, migrationPresenter, fileNameBuilder)),
		To:         NewServiceWrapHandler(options, logger, NewTo(options, logger, fileNameBuilder)),
		History:    NewServiceWrapHandler(options, logger, NewHistory(options, migrationPresenter)),
		HistoryNew: NewServiceWrapHandler(options, logger, NewHistoryNew(options, migrationPresenter)),
	}
}

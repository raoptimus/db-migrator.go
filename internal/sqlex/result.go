/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package sqlex

import (
	"database/sql"

	"github.com/pkg/errors"
)

var ErrIsNotSupportedByThisDriver = errors.New("is not supported by this driver")

type Result interface {
	sql.Result
}

// Done implements [Result] for an INSERT or UPDATE operation
// which mutates a number of rows.
type Done bool

var _ Result = Done(true)

func (Done) LastInsertId() (int64, error) {
	return 0, errors.WithMessage(ErrIsNotSupportedByThisDriver, "LastInsertId")
}

func (v Done) RowsAffected() (int64, error) {
	return 0, errors.WithMessage(ErrIsNotSupportedByThisDriver, "RowsAffected")
}

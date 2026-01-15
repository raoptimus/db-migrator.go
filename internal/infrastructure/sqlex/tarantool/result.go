/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package tarantool

import (
	"github.com/pkg/errors"
	"github.com/raoptimus/db-migrator.go/internal/infrastructure/sqlex"
)

// ErrIsNotSupportedByThisDriver indicates that the requested operation is not supported by the Tarantool driver.
var ErrIsNotSupportedByThisDriver = errors.New("is not supported by this driver")

// Done implements sqlex.Result for Tarantool operations.
// It represents the successful completion of an INSERT or UPDATE operation.
// Note that Tarantool does not provide LastInsertId or RowsAffected, so these methods return errors.
type Done bool

var _ sqlex.Result = Done(true)

// LastInsertId returns an error as this operation is not supported by Tarantool.
func (Done) LastInsertId() (int64, error) {
	return 0, errors.WithMessage(ErrIsNotSupportedByThisDriver, "LastInsertId")
}

// RowsAffected returns an error as this operation is not supported by Tarantool.
func (v Done) RowsAffected() (int64, error) {
	return 0, errors.WithMessage(ErrIsNotSupportedByThisDriver, "RowsAffected")
}

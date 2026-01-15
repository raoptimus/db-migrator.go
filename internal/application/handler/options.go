/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package handler

import (
	"github.com/pkg/errors"

	"github.com/raoptimus/db-migrator.go/internal/domain/validator"
)

const maxConnAttempts = 100

// Options contains configuration parameters for database migration operations.
type Options struct {
	PlaceholderCustom  string
	DSN                string
	MaxConnAttempts    int
	Directory          string
	TableName          string
	ClusterName        string
	Replicated         bool
	Compact            bool
	Interactive        bool
	MaxSQLOutputLength int
}

func (o *Options) Validate() error {
	if err := validator.ValidateIdentifier(o.PlaceholderCustom); err != nil {
		return errors.WithMessage(err, "placeholderCustom")
	}
	if err := validator.ValidateStringLen(1, maxConnAttempts, o.MaxConnAttempts); err != nil {
		return errors.WithMessage(err, "maxConnAttempts")
	}
	if err := validator.ValidateIdentifier(o.TableName); err != nil {
		return errors.WithMessage(err, "tableName")
	}
	if err := validator.ValidateIdentifier(o.ClusterName); err != nil {
		return errors.WithMessage(err, "clusterName")
	}

	return nil
}

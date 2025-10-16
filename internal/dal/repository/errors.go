/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package repository

import (
	"strings"

	"github.com/pkg/errors"
)

var (
	ErrPtrValueMustBeAPointerAndScalar = errors.New("ptr value must be a pointer and scalar")
)

type DBError struct {
	Code          string
	Severity      string
	Message       string
	Details       string
	InternalQuery string
	Cause         error
}

func (d *DBError) Error() string {
	var sb strings.Builder
	sb.WriteString("SQLSTATE[")
	sb.WriteString(d.Code)
	sb.WriteString("]: ")

	if d.Severity != "" {
		sb.WriteString(d.Severity)
		sb.WriteString(": ")
	}

	sb.WriteString(d.Message)
	sb.WriteString("\n")

	if d.InternalQuery != "" {
		sb.WriteString("The SQL being executed was: ")
		sb.WriteString(d.InternalQuery)
		sb.WriteString("\n")
	}

	if d.Details != "" {
		sb.WriteString("Details: ")
		sb.WriteString(d.Details)
		sb.WriteString("\n")
	}

	return strings.TrimRight(sb.String(), "\n")
}

func (d *DBError) Unwrap() error {
	return d.Cause
}

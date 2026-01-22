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
	// ErrPtrValueMustBeAPointerAndScalar is returned when a value is not a pointer to a scalar type.
	ErrPtrValueMustBeAPointerAndScalar = errors.New("ptr value must be a pointer and scalar")
)

// DBError represents a database error with additional metadata.
// It provides structured information about database errors including SQL state code, severity, and query details.
type DBError struct {
	Code          string
	Severity      string
	Message       string
	Details       string
	InternalQuery string
	Cause         error
}

// Error returns the formatted error message including SQL state, severity, message, query, and details.
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

// Unwrap returns the underlying cause of the database error for error chain unwrapping.
func (d *DBError) Unwrap() error {
	return d.Cause
}

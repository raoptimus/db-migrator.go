/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package repository

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDBError_Error_AllFieldsPopulated_Successfully(t *testing.T) {
	dbErr := &DBError{
		Code:          "42P01",
		Severity:      "ERROR",
		Message:       "relation does not exist",
		Details:       "Table 'migration' not found",
		InternalQuery: "SELECT * FROM migration",
		Cause:         errors.New("original error"),
	}

	errStr := dbErr.Error()

	assert.Contains(t, errStr, "SQLSTATE[42P01]")
	assert.Contains(t, errStr, "ERROR:")
	assert.Contains(t, errStr, "relation does not exist")
	assert.Contains(t, errStr, "The SQL being executed was: SELECT * FROM migration")
	assert.Contains(t, errStr, "Details: Table 'migration' not found")
}

func TestDBError_Error_MinimalFields_Successfully(t *testing.T) {
	dbErr := &DBError{
		Code:    "1064",
		Message: "You have an error in your SQL syntax",
	}

	errStr := dbErr.Error()

	assert.Contains(t, errStr, "SQLSTATE[1064]")
	assert.Contains(t, errStr, "You have an error in your SQL syntax")
	assert.NotContains(t, errStr, "The SQL being executed was:")
	assert.NotContains(t, errStr, "Details:")
}

func TestDBError_Error_OnlySeverity_Successfully(t *testing.T) {
	dbErr := &DBError{
		Code:     "23505",
		Severity: "FATAL",
		Message:  "duplicate key value violates unique constraint",
	}

	errStr := dbErr.Error()

	assert.Contains(t, errStr, "SQLSTATE[23505]")
	assert.Contains(t, errStr, "FATAL:")
	assert.Contains(t, errStr, "duplicate key value violates unique constraint")
}

func TestDBError_Error_OnlyInternalQuery_Successfully(t *testing.T) {
	dbErr := &DBError{
		Code:          "1146",
		Message:       "Table does not exist",
		InternalQuery: "DROP TABLE users",
	}

	errStr := dbErr.Error()

	assert.Contains(t, errStr, "SQLSTATE[1146]")
	assert.Contains(t, errStr, "Table does not exist")
	assert.Contains(t, errStr, "The SQL being executed was: DROP TABLE users")
}

func TestDBError_Error_OnlyDetails_Successfully(t *testing.T) {
	dbErr := &DBError{
		Code:    "23503",
		Message: "foreign key violation",
		Details: "Key (id)=(1) is still referenced from table 'orders'",
	}

	errStr := dbErr.Error()

	assert.Contains(t, errStr, "SQLSTATE[23503]")
	assert.Contains(t, errStr, "foreign key violation")
	assert.Contains(t, errStr, "Details: Key (id)=(1) is still referenced from table 'orders'")
}

func TestDBError_Error_EmptySeverityNoColon_Successfully(t *testing.T) {
	dbErr := &DBError{
		Code:    "00000",
		Message: "success",
	}

	errStr := dbErr.Error()

	// Without severity, there should be no colon before message
	assert.Equal(t, "SQLSTATE[00000]: success", errStr)
}

func TestDBError_Unwrap_ReturnsCause_Successfully(t *testing.T) {
	originalErr := errors.New("original database error")
	dbErr := &DBError{
		Code:    "42000",
		Message: "syntax error",
		Cause:   originalErr,
	}

	unwrapped := dbErr.Unwrap()

	require.Equal(t, originalErr, unwrapped)
}

func TestDBError_Unwrap_NilCause_Successfully(t *testing.T) {
	dbErr := &DBError{
		Code:    "42000",
		Message: "syntax error",
		Cause:   nil,
	}

	unwrapped := dbErr.Unwrap()

	require.Nil(t, unwrapped)
}

func TestDBError_ErrorsAs_Successfully(t *testing.T) {
	originalErr := errors.New("original error")
	dbErr := &DBError{
		Code:    "42P01",
		Message: "relation does not exist",
		Cause:   originalErr,
	}

	var target *DBError
	ok := errors.As(dbErr, &target)

	require.True(t, ok)
	assert.Equal(t, "42P01", target.Code)
	assert.Equal(t, "relation does not exist", target.Message)
}

func TestDBError_ErrorsIs_WithCause_Successfully(t *testing.T) {
	targetErr := errors.New("target error")
	dbErr := &DBError{
		Code:    "42P01",
		Message: "relation does not exist",
		Cause:   targetErr,
	}

	// errors.Is should follow Unwrap chain
	ok := errors.Is(dbErr, targetErr)
	require.True(t, ok)
}

func TestDBError_ImplementsError_Successfully(t *testing.T) {
	var err error = &DBError{
		Code:    "42000",
		Message: "test error",
	}

	// Verify it implements error interface
	require.NotNil(t, err.Error())
}

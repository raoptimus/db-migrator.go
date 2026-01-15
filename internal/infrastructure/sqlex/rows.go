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
	"fmt"
	"io"
	"reflect"

	"github.com/pkg/errors"
)

var (
	// ErrNilPtrValue occurs when a nil pointer is passed to Scan.
	ErrNilPtrValue = errors.New("nil ptr value")
	// ErrPtrValueMustBeAPointer occurs when a non-pointer value is passed to Scan.
	ErrPtrValueMustBeAPointer = errors.New("ptr value must be a pointer")
)

// Rows abstracts the interface for iterating over query result rows
// and scanning values into destination variables.
type Rows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
	Close() error
}

// RowsWithSlice implements the Rows interface using an in-memory slice of data.
// This is primarily used for testing and mock data scenarios.
type RowsWithSlice struct {
	i    int
	rows []interface{}
}

// RowsWithSQLRows wraps the standard database/sql.Rows to implement the custom Rows interface.
type RowsWithSQLRows struct {
	*sql.Rows
}

// NewRowsWithSQLRows creates a new RowsWithSQLRows wrapper around the provided sql.Rows.
func NewRowsWithSQLRows(rows *sql.Rows) *RowsWithSQLRows {
	return &RowsWithSQLRows{Rows: rows}
}

// NewRowsWithSlice creates a new RowsWithSlice from the provided slice of row data.
func NewRowsWithSlice(rows []interface{}) *RowsWithSlice {
	return &RowsWithSlice{
		rows: rows,
	}
}

// Next advances to the next row in the result set and returns true if a row is available.
func (v *RowsWithSlice) Next() bool {
	return v.i < len(v.rows)
}

// Scan copies the columns from the current row into the values pointed at by dest.
// It performs type conversion and validation for in-memory slice data.
func (v *RowsWithSlice) Scan(dest ...any) error {
	if v.i >= len(v.rows) {
		return errors.WithStack(io.EOF)
	}

	for k := range dest {
		destValPtr := dest[k]
		if destValPtr == nil {
			return errors.WithStack(ErrNilPtrValue)
		}
		destType := reflect.TypeOf(destValPtr)
		if destType.Kind() != reflect.Pointer {
			return errors.WithStack(ErrPtrValueMustBeAPointer)
		}
		destTypeElem := destType.Elem()

		row := v.rows[v.i]
		rowVals, ok := row.([]any)
		if !ok {
			rowVals = []any{row}
		}

		if len(dest) != len(rowVals) {
			return errors.WithStack(
				fmt.Errorf("sql: expected %d destination arguments in Scan, not %d",
					len(rowVals),
					len(dest),
				))
		}

		rowVal := rowVals[k]
		if rowVal == nil {
			continue
		}
		rowType := reflect.TypeOf(rowVal)
		rowTypeElem := rowType
		if rowType.Kind() == reflect.Pointer {
			rowTypeElem = rowType.Elem()
		}

		rowValOf := reflect.ValueOf(rowVal)
		if rowTypeElem.Kind() == destTypeElem.Kind() {
			reflect.ValueOf(destValPtr).Elem().Set(rowValOf)
			continue
		}

		if rowValOf.CanConvert(destTypeElem) {
			reflect.ValueOf(destValPtr).Elem().Set(rowValOf.Convert(destTypeElem))
			continue
		}

		return errors.WithStack(
			errors.Errorf("variable of type %T cannot be set to variable of type %T",
				rowVal,
				destValPtr,
			))
	}

	v.i++

	return nil
}

// Err returns the error, if any, that was encountered during iteration.
// For RowsWithSlice, this always returns nil.
func (v *RowsWithSlice) Err() error {
	return nil
}

// Close closes the Rows, preventing further enumeration.
// For RowsWithSlice, this is a no-op that always returns nil.
func (v *RowsWithSlice) Close() error {
	return nil
}

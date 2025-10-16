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
	ErrNilPtrValue            = errors.New("nil ptr value")
	ErrPtrValueMustBeAPointer = errors.New("ptr value must be a pointer")
)

type Rows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
	Close() error
}

type RowsWithSlice struct {
	i    int
	rows []interface{}
}

type RowsWithSQLRows struct {
	*sql.Rows
}

func NewRowsWithSQLRows(rows *sql.Rows) *RowsWithSQLRows {
	return &RowsWithSQLRows{Rows: rows}
}

func NewRowsWithSlice(rows []interface{}) *RowsWithSlice {
	return &RowsWithSlice{
		rows: rows,
	}
}

func (v *RowsWithSlice) Next() bool {
	return v.i < len(v.rows)
}

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

func (v *RowsWithSlice) Err() error {
	return nil
}

func (v *RowsWithSlice) Close() error {
	return nil
}

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
	"io"
	"reflect"

	"github.com/pkg/errors"
)

var ErrNilPtrValue = errors.New("nil ptr value")
var ErrPtrValueMustBeAPointer = errors.New("ptr value must be a pointer")

type Rows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
}

type RowsWithSlice struct {
	next func() bool
	scan func(dest ...any) error
	err  func() error
}

func NewRowsWithSlice(next func() bool, scan func(dest ...any) error, err func() error) *RowsWithSlice {
	return &RowsWithSlice{
		next: next,
		scan: scan,
		err:  err,
	}
}

func (v *RowsWithSlice) Next() bool {
	return v.next()
}

func (v *RowsWithSlice) Scan(dest ...any) error {
	return v.scan(dest...)
}

func (v *RowsWithSlice) Err() error {
	return v.err()
}

func NewRowsByData(rows []interface{}) *RowsWithSlice {
	var i = -1
	return NewRowsWithSlice(
		func() bool {
			i++

			return i < len(rows)
		},
		func(dest ...any) error {
			if i >= len(rows) {
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

				row := rows[i]
				rowVals, ok := row.([]any)
				if !ok {
					rowVals = []any{row}
				}

				if k >= len(rowVals) {
					return errors.New("")
				}

				rowVal := rowVals[k]
				rowType := reflect.TypeOf(rowVal)
				rowTypeElem := rowType
				if rowType.Kind() == reflect.Pointer {
					if rowVal == nil {
						continue
					}
					rowTypeElem = rowType.Elem()
				}

				if rowTypeElem.Kind() == destTypeElem.Kind() {
					reflect.ValueOf(destValPtr).Elem().Set(reflect.ValueOf(rowVal))
				}
			}

			return nil
		},
		func() error {
			return nil
		},
	)
}

func NewRowsBySQLRows(rows *sql.Rows) *RowsWithSlice {
	return NewRowsWithSlice(
		func() bool {
			return rows.Next()
		},
		func(dest ...any) error {
			return rows.Scan(dest...)
		},
		func() error {
			return rows.Err()
		},
	)
}

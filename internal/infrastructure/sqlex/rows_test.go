/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package sqlex

import (
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

// --- NewRowsWithSlice Tests ---

func TestNewRowsWithSlice_CreatesRowsWithData_Successfully(t *testing.T) {
	rows := []interface{}{
		[]any{1, "test"},
		[]any{2, "test2"},
	}

	result := NewRowsWithSlice(rows)

	require.NotNil(t, result)
	require.Equal(t, rows, result.rows)
	require.Equal(t, 0, result.i)
}

func TestNewRowsWithSlice_CreatesRowsWithEmptySlice_Successfully(t *testing.T) {
	rows := []interface{}{}

	result := NewRowsWithSlice(rows)

	require.NotNil(t, result)
	require.Empty(t, result.rows)
	require.Equal(t, 0, result.i)
}

func TestNewRowsWithSlice_CreatesRowsWithNilSlice_Successfully(t *testing.T) {
	var rows []interface{}

	result := NewRowsWithSlice(rows)

	require.NotNil(t, result)
	require.Nil(t, result.rows)
	require.Equal(t, 0, result.i)
}

// --- Next Tests ---

func TestRowsWithSlice_Next_ReturnsTrue_WhenDataAvailable(t *testing.T) {
	rows := NewRowsWithSlice([]interface{}{
		[]any{1, "test"},
		[]any{2, "test2"},
	})

	require.True(t, rows.Next())
}

func TestRowsWithSlice_Next_ReturnsFalse_WhenNoDataAvailable(t *testing.T) {
	rows := NewRowsWithSlice([]interface{}{})

	require.False(t, rows.Next())
}

func TestRowsWithSlice_Next_IteratesThroughAllRows_Successfully(t *testing.T) {
	rows := NewRowsWithSlice([]interface{}{
		[]any{1},
		[]any{2},
		[]any{3},
	})

	// First row
	require.True(t, rows.Next())
	var val1 int
	err := rows.Scan(&val1)
	require.NoError(t, err)
	require.Equal(t, 1, val1)

	// Second row
	require.True(t, rows.Next())
	var val2 int
	err = rows.Scan(&val2)
	require.NoError(t, err)
	require.Equal(t, 2, val2)

	// Third row
	require.True(t, rows.Next())
	var val3 int
	err = rows.Scan(&val3)
	require.NoError(t, err)
	require.Equal(t, 3, val3)

	// No more rows
	require.False(t, rows.Next())
}

// --- Scan Tests ---

func TestRowsWithSlice_Scan_ScansIntValue_Successfully(t *testing.T) {
	rows := NewRowsWithSlice([]interface{}{
		42,
	})

	var result int
	err := rows.Scan(&result)

	require.NoError(t, err)
	require.Equal(t, 42, result)
}

func TestRowsWithSlice_Scan_ScansStringValue_Successfully(t *testing.T) {
	rows := NewRowsWithSlice([]interface{}{
		"test_string",
	})

	var result string
	err := rows.Scan(&result)

	require.NoError(t, err)
	require.Equal(t, "test_string", result)
}

func TestRowsWithSlice_Scan_ScansInt64Value_Successfully(t *testing.T) {
	rows := NewRowsWithSlice([]interface{}{
		int64(9223372036854775807),
	})

	var result int64
	err := rows.Scan(&result)

	require.NoError(t, err)
	require.Equal(t, int64(9223372036854775807), result)
}

func TestRowsWithSlice_Scan_ScansMultipleColumns_Successfully(t *testing.T) {
	rows := NewRowsWithSlice([]interface{}{
		[]any{1, "name", int64(100)},
	})

	var id int
	var name string
	var count int64
	err := rows.Scan(&id, &name, &count)

	require.NoError(t, err)
	require.Equal(t, 1, id)
	require.Equal(t, "name", name)
	require.Equal(t, int64(100), count)
}

func TestRowsWithSlice_Scan_ScansWithTypeConversion_Successfully(t *testing.T) {
	tests := []struct {
		name          string
		rowValue      interface{}
		expectedValue int64
	}{
		{
			name:          "converts int to int64",
			rowValue:      int(42),
			expectedValue: int64(42),
		},
		{
			name:          "converts int32 to int64",
			rowValue:      int32(42),
			expectedValue: int64(42),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows := NewRowsWithSlice([]interface{}{tt.rowValue})

			var result int64
			err := rows.Scan(&result)

			require.NoError(t, err)
			require.Equal(t, tt.expectedValue, result)
		})
	}
}

func TestRowsWithSlice_Scan_SkipsNilValue_Successfully(t *testing.T) {
	rows := NewRowsWithSlice([]interface{}{
		[]any{nil, "test"},
	})

	var id int
	var name string
	err := rows.Scan(&id, &name)

	require.NoError(t, err)
	require.Equal(t, 0, id) // Zero value since nil was skipped
	require.Equal(t, "test", name)
}

func TestRowsWithSlice_Scan_DestCountMismatch_Failure(t *testing.T) {
	tests := []struct {
		name     string
		rowData  interface{}
		destArgs int
	}{
		{
			name:     "more dest than columns",
			rowData:  []any{1, "test"},
			destArgs: 3,
		},
		{
			name:     "fewer dest than columns",
			rowData:  []any{1, "test", 3},
			destArgs: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows := NewRowsWithSlice([]interface{}{tt.rowData})

			var a, b, c int
			var err error
			switch tt.destArgs {
			case 2:
				err = rows.Scan(&a, &b)
			case 3:
				err = rows.Scan(&a, &b, &c)
			}

			require.Error(t, err)
			require.Contains(t, err.Error(), "expected")
			require.Contains(t, err.Error(), "destination arguments")
		})
	}
}

func TestRowsWithSlice_Scan_NilDestPointer_Failure(t *testing.T) {
	rows := NewRowsWithSlice([]interface{}{
		42,
	})

	err := rows.Scan(nil)

	require.Error(t, err)
	require.ErrorIs(t, err, ErrNilPtrValue)
}

func TestRowsWithSlice_Scan_NonPointerDest_Failure(t *testing.T) {
	rows := NewRowsWithSlice([]interface{}{
		42,
	})

	var result int
	err := rows.Scan(result) // Passing value instead of pointer

	require.Error(t, err)
	require.ErrorIs(t, err, ErrPtrValueMustBeAPointer)
}

func TestRowsWithSlice_Scan_IncompatibleTypes_Failure(t *testing.T) {
	rows := NewRowsWithSlice([]interface{}{
		"not_convertible_to_int",
	})

	var result int
	err := rows.Scan(&result)

	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot be set to variable")
}

func TestRowsWithSlice_Scan_EOFWhenNoMoreRows_Failure(t *testing.T) {
	rows := NewRowsWithSlice([]interface{}{
		42,
	})

	var result int
	// Scan the first row
	err := rows.Scan(&result)
	require.NoError(t, err)

	// Try to scan when no more rows
	err = rows.Scan(&result)

	require.Error(t, err)
	require.ErrorIs(t, err, io.EOF)
}

func TestRowsWithSlice_Scan_EmptySlice_Failure(t *testing.T) {
	rows := NewRowsWithSlice([]interface{}{})

	var result int
	err := rows.Scan(&result)

	require.Error(t, err)
	require.ErrorIs(t, err, io.EOF)
}

// --- Scan Boundary Value Tests ---

func TestRowsWithSlice_Scan_BoundaryValues_Successfully(t *testing.T) {
	tests := []struct {
		name          string
		rowValue      interface{}
		expectedValue interface{}
	}{
		{
			name:          "int zero value",
			rowValue:      0,
			expectedValue: 0,
		},
		{
			name:          "int max value",
			rowValue:      2147483647,
			expectedValue: 2147483647,
		},
		{
			name:          "int min value",
			rowValue:      -2147483648,
			expectedValue: -2147483648,
		},
		{
			name:          "int64 max value",
			rowValue:      int64(9223372036854775807),
			expectedValue: int64(9223372036854775807),
		},
		{
			name:          "int64 min value",
			rowValue:      int64(-9223372036854775808),
			expectedValue: int64(-9223372036854775808),
		},
		{
			name:          "empty string",
			rowValue:      "",
			expectedValue: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows := NewRowsWithSlice([]interface{}{tt.rowValue})

			switch expected := tt.expectedValue.(type) {
			case int:
				var result int
				err := rows.Scan(&result)
				require.NoError(t, err)
				require.Equal(t, expected, result)
			case int64:
				var result int64
				err := rows.Scan(&result)
				require.NoError(t, err)
				require.Equal(t, expected, result)
			case string:
				var result string
				err := rows.Scan(&result)
				require.NoError(t, err)
				require.Equal(t, expected, result)
			}
		})
	}
}

// --- Err Tests ---

func TestRowsWithSlice_Err_ReturnsNil_Successfully(t *testing.T) {
	rows := NewRowsWithSlice([]interface{}{
		42,
	})

	err := rows.Err()

	require.NoError(t, err)
}

func TestRowsWithSlice_Err_ReturnsNilAfterIteration_Successfully(t *testing.T) {
	rows := NewRowsWithSlice([]interface{}{
		[]any{1},
		[]any{2},
	})

	// Iterate through all rows
	for rows.Next() {
		var val int
		_ = rows.Scan(&val)
	}

	err := rows.Err()

	require.NoError(t, err)
}

func TestRowsWithSlice_Err_ReturnsNilWithEmptySlice_Successfully(t *testing.T) {
	rows := NewRowsWithSlice([]interface{}{})

	err := rows.Err()

	require.NoError(t, err)
}

// --- Close Tests ---

func TestRowsWithSlice_Close_ReturnsNil_Successfully(t *testing.T) {
	rows := NewRowsWithSlice([]interface{}{
		42,
	})

	err := rows.Close()

	require.NoError(t, err)
}

func TestRowsWithSlice_Close_ReturnsNilAfterIteration_Successfully(t *testing.T) {
	rows := NewRowsWithSlice([]interface{}{
		[]any{1},
		[]any{2},
	})

	// Iterate through all rows
	for rows.Next() {
		var val int
		_ = rows.Scan(&val)
	}

	err := rows.Close()

	require.NoError(t, err)
}

func TestRowsWithSlice_Close_ReturnsNilWithEmptySlice_Successfully(t *testing.T) {
	rows := NewRowsWithSlice([]interface{}{})

	err := rows.Close()

	require.NoError(t, err)
}

func TestRowsWithSlice_Close_CanBeCalledMultipleTimes_Successfully(t *testing.T) {
	rows := NewRowsWithSlice([]interface{}{42})

	err1 := rows.Close()
	err2 := rows.Close()
	err3 := rows.Close()

	require.NoError(t, err1)
	require.NoError(t, err2)
	require.NoError(t, err3)
}

// --- Rows Interface Compliance Test ---

func TestRowsWithSlice_ImplementsRowsInterface(t *testing.T) {
	var _ Rows = (*RowsWithSlice)(nil)
}

// --- NewRowsWithSQLRows Tests ---

func TestNewRowsWithSQLRows_CreatesWrapper_Successfully(t *testing.T) {
	// Note: We cannot test with a real sql.Rows without a database connection
	// This test verifies that the constructor does not panic with nil input
	// In production, sql.Rows would never be nil when obtained from a query
	result := NewRowsWithSQLRows(nil)

	require.NotNil(t, result)
	require.Nil(t, result.Rows)
}

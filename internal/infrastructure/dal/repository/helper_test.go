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
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckArgIsPtrAndScalar_ValidPointerTypes_Successfully(t *testing.T) {
	tests := []struct {
		name string
		ptr  any
	}{
		{
			name: "pointer to int",
			ptr:  new(int),
		},
		{
			name: "pointer to int8",
			ptr:  new(int8),
		},
		{
			name: "pointer to int16",
			ptr:  new(int16),
		},
		{
			name: "pointer to int32",
			ptr:  new(int32),
		},
		{
			name: "pointer to int64",
			ptr:  new(int64),
		},
		{
			name: "pointer to uint",
			ptr:  new(uint),
		},
		{
			name: "pointer to uint8",
			ptr:  new(uint8),
		},
		{
			name: "pointer to uint16",
			ptr:  new(uint16),
		},
		{
			name: "pointer to uint32",
			ptr:  new(uint32),
		},
		{
			name: "pointer to uint64",
			ptr:  new(uint64),
		},
		{
			name: "pointer to float32",
			ptr:  new(float32),
		},
		{
			name: "pointer to float64",
			ptr:  new(float64),
		},
		{
			name: "pointer to bool",
			ptr:  new(bool),
		},
		{
			name: "pointer to string",
			ptr:  new(string),
		},
		{
			name: "pointer to time.Time",
			ptr:  new(time.Time),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkArgIsPtrAndScalar(tt.ptr)
			require.NoError(t, err)
		})
	}
}

func TestCheckArgIsPtrAndScalar_InvalidTypes_Failure(t *testing.T) {
	tests := []struct {
		name string
		ptr  any
	}{
		{
			name: "non-pointer int",
			ptr:  42,
		},
		{
			name: "non-pointer string",
			ptr:  "hello",
		},
		{
			name: "pointer to slice",
			ptr:  new([]int),
		},
		{
			name: "pointer to map",
			ptr:  new(map[string]int),
		},
		{
			name: "pointer to struct",
			ptr:  new(struct{ Field int }),
		},
		{
			name: "nil value",
			ptr:  nil,
		},
		{
			name: "pointer to pointer to int",
			ptr:  new(*int),
		},
		{
			name: "slice value",
			ptr:  []int{1, 2, 3},
		},
		{
			name: "map value",
			ptr:  map[string]int{"key": 1},
		},
		{
			name: "struct value",
			ptr:  struct{ Field int }{Field: 1},
		},
		{
			name: "pointer to interface",
			ptr:  new(interface{}),
		},
		{
			name: "pointer to channel",
			ptr:  new(chan int),
		},
		{
			name: "pointer to func",
			ptr:  new(func()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkArgIsPtrAndScalar(tt.ptr)
			require.Error(t, err)
			assert.ErrorIs(t, err, ErrPtrValueMustBeAPointerAndScalar)
		})
	}
}

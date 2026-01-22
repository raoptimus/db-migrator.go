/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package dbmigrator

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSentinelErrors_AreDistinct(t *testing.T) {
	require.NotEqual(t, ErrMigrationAlreadyExists, ErrAppliedMigrationNotFound)
}

func TestSentinelErrors_ErrorMessages(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		wantMessage string
	}{
		{
			name:        "ErrMigrationAlreadyExists has correct message",
			err:         ErrMigrationAlreadyExists,
			wantMessage: "migration already exists",
		},
		{
			name:        "ErrAppliedMigrationNotFound has correct message",
			err:         ErrAppliedMigrationNotFound,
			wantMessage: "applied migration not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantMessage, tt.err.Error())
		})
	}
}

func TestSentinelErrors_CanBeWrapped(t *testing.T) {
	tests := []struct {
		name    string
		baseErr error
	}{
		{
			name:    "ErrMigrationAlreadyExists can be wrapped",
			baseErr: ErrMigrationAlreadyExists,
		},
		{
			name:    "ErrAppliedMigrationNotFound can be wrapped",
			baseErr: ErrAppliedMigrationNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wrapped := errors.WithStack(tt.baseErr)
			assert.True(t, errors.Is(wrapped, tt.baseErr))

			wrappedWithMessage := errors.WithMessage(tt.baseErr, "context")
			assert.True(t, errors.Is(wrappedWithMessage, tt.baseErr))
		})
	}
}

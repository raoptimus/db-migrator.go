/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateVersion(t *testing.T) {
	tests := []struct {
		name    string
		version string
		wantErr error
	}{
		{
			name:    "valid version with underscore name",
			version: "200905_192800_create_test",
			wantErr: nil,
		},
		{
			name:    "valid version with underscore in table name",
			version: "210328_221600_add_user_table",
			wantErr: nil,
		},
		{
			name:    "valid version with dashes in name",
			version: "210328_221600_add-user-table",
			wantErr: nil,
		},
		{
			name:    "valid version with mixed underscore and dashes",
			version: "210328_221600_add-user_table",
			wantErr: nil,
		},
		{
			name:    "valid version with numbers in name",
			version: "210328_221600_add_column2",
			wantErr: nil,
		},
		{
			name:    "valid version minimal name",
			version: "200101_000000_ab",
			wantErr: nil,
		},
		{
			name:    "empty string",
			version: "",
			wantErr: ErrVersionIsNotValid,
		},
		{
			name:    "short version string",
			version: "210328_22",
			wantErr: ErrVersionIsNotValid,
		},
		{
			name:    "missing name part",
			version: "210328_221600",
			wantErr: ErrVersionIsNotValid,
		},
		{
			name:    "missing underscore between date and time",
			version: "210328221600_test",
			wantErr: ErrVersionIsNotValid,
		},
		{
			name:    "missing underscore before name",
			version: "210328_221600test",
			wantErr: ErrVersionIsNotValid,
		},
		{
			name:    "invalid month 13",
			version: "211328_221600_test",
			wantErr: ErrVersionIsNotValid,
		},
		{
			name:    "invalid month 00",
			version: "210028_221600_test",
			wantErr: ErrVersionIsNotValid,
		},
		{
			name:    "invalid day 32",
			version: "210132_221600_test",
			wantErr: ErrVersionIsNotValid,
		},
		{
			name:    "invalid day 00",
			version: "210100_221600_test",
			wantErr: ErrVersionIsNotValid,
		},
		{
			name:    "invalid hour 25",
			version: "210328_251600_test",
			wantErr: ErrVersionIsNotValid,
		},
		{
			name:    "invalid minute 61",
			version: "210328_226100_test",
			wantErr: ErrVersionIsNotValid,
		},
		{
			name:    "invalid second 61",
			version: "210328_221661_test",
			wantErr: ErrVersionIsNotValid,
		},
		{
			name:    "future date",
			version: "350328_221600_test",
			wantErr: ErrVersionIsNotValid,
		},
		{
			name:    "name starts with uppercase",
			version: "210328_221600_Test",
			wantErr: ErrVersionIsNotValid,
		},
		{
			name:    "name starts with digit",
			version: "210328_221600_1test",
			wantErr: ErrVersionIsNotValid,
		},
		{
			name:    "name with special characters",
			version: "210328_221600_test@name",
			wantErr: ErrVersionIsNotValid,
		},
		{
			name:    "name with spaces",
			version: "210328_221600_test name",
			wantErr: ErrVersionIsNotValid,
		},
		{
			name:    "name with dots",
			version: "210328_221600_test.name",
			wantErr: ErrVersionIsNotValid,
		},
		{
			name:    "only numbers",
			version: "123456789012",
			wantErr: ErrVersionIsNotValid,
		},
		{
			name:    "letters in date part",
			version: "21ab28_221600_test",
			wantErr: ErrVersionIsNotValid,
		},
		{
			name:    "letters in time part",
			version: "210328_22ab00_test",
			wantErr: ErrVersionIsNotValid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateVersion(tt.version)
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

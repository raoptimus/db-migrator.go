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

func TestValidateFileName(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		wantErr  error
	}{
		{
			name:     "valid up file",
			fileName: "200905_192800_create_test.up.sql",
			wantErr:  nil,
		},
		{
			name:     "valid down file",
			fileName: "200905_192800_create_test.down.sql",
			wantErr:  nil,
		},
		{
			name:     "valid safe up file",
			fileName: "200905_192800_create_test.safe.up.sql",
			wantErr:  nil,
		},
		{
			name:     "valid safe down file",
			fileName: "200905_192800_create_test.safe.down.sql",
			wantErr:  nil,
		},
		{
			name:     "valid file with dashes in name",
			fileName: "210328_221600_add-user-table.up.sql",
			wantErr:  nil,
		},
		{
			name:     "valid file with underscores in name",
			fileName: "210328_221600_add_user_table.up.sql",
			wantErr:  nil,
		},
		{
			name:     "valid file with numbers in name",
			fileName: "210328_221600_add_column2.up.sql",
			wantErr:  nil,
		},
		{
			name:     "valid safe file with complex name",
			fileName: "210328_221600_add-user_table2.safe.down.sql",
			wantErr:  nil,
		},
		{
			name:     "empty string",
			fileName: "",
			wantErr:  ErrVersionIsNotValid,
		},
		{
			name:     "missing sql extension",
			fileName: "200905_192800_test.up",
			wantErr:  ErrFileNameIsNotValid,
		},
		{
			name:     "wrong extension txt",
			fileName: "200905_192800_test.up.txt",
			wantErr:  ErrFileNameIsNotValid,
		},
		{
			name:     "missing action part",
			fileName: "200905_192800_test.sql",
			wantErr:  ErrFileNameIsNotValid,
		},
		{
			name:     "invalid action migrate",
			fileName: "200905_192800_test.migrate.sql",
			wantErr:  ErrFileNameIsNotValid,
		},
		{
			name:     "safe in wrong position",
			fileName: "200905_192800_test.up.safe.sql",
			wantErr:  ErrFileNameIsNotValid,
		},
		{
			name:     "uppercase UP action",
			fileName: "200905_192800_test.UP.sql",
			wantErr:  ErrFileNameIsNotValid,
		},
		{
			name:     "uppercase DOWN action",
			fileName: "200905_192800_test.DOWN.sql",
			wantErr:  ErrFileNameIsNotValid,
		},
		{
			name:     "uppercase SAFE modifier",
			fileName: "200905_192800_test.SAFE.up.sql",
			wantErr:  ErrFileNameIsNotValid,
		},
		{
			name:     "name starts with uppercase",
			fileName: "210328_221600_Test.up.sql",
			wantErr:  ErrFileNameIsNotValid,
		},
		{
			name:     "name starts with digit",
			fileName: "210328_221600_1test.up.sql",
			wantErr:  ErrFileNameIsNotValid,
		},
		{
			name:     "invalid month 13",
			fileName: "211328_221600_test.up.sql",
			wantErr:  ErrFileNameIsNotValid,
		},
		{
			name:     "invalid day 32",
			fileName: "210132_221600_test.up.sql",
			wantErr:  ErrFileNameIsNotValid,
		},
		{
			name:     "invalid hour 25",
			fileName: "210328_251600_test.up.sql",
			wantErr:  ErrFileNameIsNotValid,
		},
		{
			name:     "future date",
			fileName: "350328_221600_test.up.sql",
			wantErr:  ErrFileNameIsNotValid,
		},
		{
			name:     "double dot before extension",
			fileName: "200905_192800_test..up.sql",
			wantErr:  ErrFileNameIsNotValid,
		},
		{
			name:     "missing name part",
			fileName: "200905_192800_.up.sql",
			wantErr:  ErrFileNameIsNotValid,
		},
		{
			name:     "special characters in name",
			fileName: "200905_192800_test@name.up.sql",
			wantErr:  ErrFileNameIsNotValid,
		},
		{
			name:     "spaces in name",
			fileName: "200905_192800_test name.up.sql",
			wantErr:  ErrFileNameIsNotValid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFileName(tt.fileName)
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package entity

import (
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigration_ApplyTimeFormat_ReturnsFormattedTime_Successfully(t *testing.T) {
	t.Parallel()

	// Expected format pattern: YYYY-MM-DD HH:MM:SS
	formatPattern := regexp.MustCompile(`^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}$`)

	tests := []struct {
		name      string
		applyTime int64
	}{
		{
			name:      "zero timestamp returns valid format",
			applyTime: 0,
		},
		{
			name:      "positive timestamp returns valid format",
			applyTime: 1599332880,
		},
		{
			name:      "large timestamp returns valid format",
			applyTime: 1672531200,
		},
		{
			name:      "negative timestamp returns valid format",
			applyTime: -86400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			migration := Migration{
				Version:   "200905_192800",
				ApplyTime: tt.applyTime,
			}

			result := migration.ApplyTimeFormat()

			assert.True(t, formatPattern.MatchString(result),
				"result %q should match format YYYY-MM-DD HH:MM:SS", result)

			// Verify the result matches what time.Unix produces with local timezone
			expected := time.Unix(tt.applyTime, 0).Format("2006-01-02 15:04:05")
			assert.Equal(t, expected, result)
		})
	}
}

func TestMigration_ApplyTimeFormat_ReturnsExpectedLocalTime_Successfully(t *testing.T) {
	t.Parallel()

	// This test verifies that ApplyTimeFormat correctly converts Unix timestamp
	// to local time format. We use time.Unix to compute expected values dynamically
	// since the output depends on local timezone.

	tests := []struct {
		name      string
		applyTime int64
	}{
		{
			name:      "epoch timestamp",
			applyTime: 0,
		},
		{
			name:      "typical migration timestamp",
			applyTime: 1599332880,
		},
		{
			name:      "year boundary timestamp",
			applyTime: 1672531200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			migration := Migration{
				Version:   "test_version",
				ApplyTime: tt.applyTime,
			}

			result := migration.ApplyTimeFormat()

			// Parse the result back and verify it represents the same point in time
			parsed, err := time.ParseInLocation("2006-01-02 15:04:05", result, time.Local)
			require.NoError(t, err)

			assert.Equal(t, tt.applyTime, parsed.Unix())
		})
	}
}

func TestMigrations_Len_ReturnsCorrectLength_Successfully(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		migrations Migrations
		expected   int
	}{
		{
			name:       "empty slice returns zero",
			migrations: Migrations{},
			expected:   0,
		},
		{
			name: "single element returns one",
			migrations: Migrations{
				{Version: "200905_192800", ApplyTime: 1599332880},
			},
			expected: 1,
		},
		{
			name: "multiple elements returns correct count",
			migrations: Migrations{
				{Version: "200905_192800", ApplyTime: 1599332880},
				{Version: "200906_100000", ApplyTime: 1599386400},
				{Version: "200907_120000", ApplyTime: 1599480000},
			},
			expected: 3,
		},
		{
			name:       "nil slice returns zero",
			migrations: nil,
			expected:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.migrations.Len()

			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMigrations_Less_ComparesVersionsCorrectly_Successfully(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		migrations Migrations
		i          int
		j          int
		expected   bool
	}{
		{
			name: "earlier version is less than later version",
			migrations: Migrations{
				{Version: "200905_192800", ApplyTime: 1599332880},
				{Version: "200906_100000", ApplyTime: 1599386400},
			},
			i:        0,
			j:        1,
			expected: true,
		},
		{
			name: "later version is not less than earlier version",
			migrations: Migrations{
				{Version: "200905_192800", ApplyTime: 1599332880},
				{Version: "200906_100000", ApplyTime: 1599386400},
			},
			i:        1,
			j:        0,
			expected: false,
		},
		{
			name: "same version is not less than itself",
			migrations: Migrations{
				{Version: "200905_192800", ApplyTime: 1599332880},
				{Version: "200905_192800", ApplyTime: 1599332880},
			},
			i:        0,
			j:        1,
			expected: false,
		},
		{
			name: "lexicographic comparison for different year prefixes",
			migrations: Migrations{
				{Version: "190101_000000", ApplyTime: 1546300800},
				{Version: "200101_000000", ApplyTime: 1577836800},
			},
			i:        0,
			j:        1,
			expected: true,
		},
		{
			name: "lexicographic comparison with different time parts",
			migrations: Migrations{
				{Version: "200905_192800", ApplyTime: 1599332880},
				{Version: "200905_192801", ApplyTime: 1599332881},
			},
			i:        0,
			j:        1,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.migrations.Less(tt.i, tt.j)

			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMigrations_Swap_SwapsElementsCorrectly_Successfully(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		migrations         Migrations
		i                  int
		j                  int
		expectedVersionAtI string
		expectedVersionAtJ string
	}{
		{
			name: "swap first and second elements",
			migrations: Migrations{
				{Version: "200905_192800", ApplyTime: 1599332880},
				{Version: "200906_100000", ApplyTime: 1599386400},
			},
			i:                  0,
			j:                  1,
			expectedVersionAtI: "200906_100000",
			expectedVersionAtJ: "200905_192800",
		},
		{
			name: "swap same index does nothing",
			migrations: Migrations{
				{Version: "200905_192800", ApplyTime: 1599332880},
				{Version: "200906_100000", ApplyTime: 1599386400},
			},
			i:                  0,
			j:                  0,
			expectedVersionAtI: "200905_192800",
			expectedVersionAtJ: "200905_192800",
		},
		{
			name: "swap first and last elements in larger slice",
			migrations: Migrations{
				{Version: "200905_192800", ApplyTime: 1599332880},
				{Version: "200906_100000", ApplyTime: 1599386400},
				{Version: "200907_120000", ApplyTime: 1599480000},
			},
			i:                  0,
			j:                  2,
			expectedVersionAtI: "200907_120000",
			expectedVersionAtJ: "200905_192800",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create a copy to avoid affecting other tests
			migrations := make(Migrations, len(tt.migrations))
			copy(migrations, tt.migrations)

			migrations.Swap(tt.i, tt.j)

			assert.Equal(t, tt.expectedVersionAtI, migrations[tt.i].Version)
			assert.Equal(t, tt.expectedVersionAtJ, migrations[tt.j].Version)
		})
	}
}

func TestMigrations_SortByVersion_SortsCorrectly_Successfully(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		migrations       Migrations
		expectedVersions []string
	}{
		{
			name:             "empty slice remains empty",
			migrations:       Migrations{},
			expectedVersions: []string{},
		},
		{
			name: "single element remains unchanged",
			migrations: Migrations{
				{Version: "200905_192800", ApplyTime: 1599332880},
			},
			expectedVersions: []string{"200905_192800"},
		},
		{
			name: "already sorted slice remains sorted",
			migrations: Migrations{
				{Version: "200905_192800", ApplyTime: 1599332880},
				{Version: "200906_100000", ApplyTime: 1599386400},
				{Version: "200907_120000", ApplyTime: 1599480000},
			},
			expectedVersions: []string{"200905_192800", "200906_100000", "200907_120000"},
		},
		{
			name: "reverse sorted slice becomes sorted",
			migrations: Migrations{
				{Version: "200907_120000", ApplyTime: 1599480000},
				{Version: "200906_100000", ApplyTime: 1599386400},
				{Version: "200905_192800", ApplyTime: 1599332880},
			},
			expectedVersions: []string{"200905_192800", "200906_100000", "200907_120000"},
		},
		{
			name: "unsorted slice becomes sorted",
			migrations: Migrations{
				{Version: "200906_100000", ApplyTime: 1599386400},
				{Version: "200905_192800", ApplyTime: 1599332880},
				{Version: "200908_150000", ApplyTime: 1599573600},
				{Version: "200907_120000", ApplyTime: 1599480000},
			},
			expectedVersions: []string{"200905_192800", "200906_100000", "200907_120000", "200908_150000"},
		},
		{
			name: "slice with duplicate versions preserves duplicates",
			migrations: Migrations{
				{Version: "200906_100000", ApplyTime: 1599386401},
				{Version: "200905_192800", ApplyTime: 1599332880},
				{Version: "200906_100000", ApplyTime: 1599386400},
			},
			expectedVersions: []string{"200905_192800", "200906_100000", "200906_100000"},
		},
		{
			name: "slice with different year prefixes sorts correctly",
			migrations: Migrations{
				{Version: "210101_000000", ApplyTime: 1609459200},
				{Version: "190601_120000", ApplyTime: 1559390400},
				{Version: "200101_000000", ApplyTime: 1577836800},
			},
			expectedVersions: []string{"190601_120000", "200101_000000", "210101_000000"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create a copy to avoid affecting other tests
			migrations := make(Migrations, len(tt.migrations))
			copy(migrations, tt.migrations)

			migrations.SortByVersion()

			require.Equal(t, len(tt.expectedVersions), len(migrations))

			for i, expectedVersion := range tt.expectedVersions {
				assert.Equal(t, expectedVersion, migrations[i].Version,
					"version at index %d should be %s", i, expectedVersion)
			}
		})
	}
}

func TestMigration_StructFields_HaveCorrectValues_Successfully(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		version           string
		applyTime         int64
		expectedVersion   string
		expectedApplyTime int64
	}{
		{
			name:              "migration with standard version format",
			version:           "200905_192800",
			applyTime:         1599332880,
			expectedVersion:   "200905_192800",
			expectedApplyTime: 1599332880,
		},
		{
			name:              "migration with empty version",
			version:           "",
			applyTime:         0,
			expectedVersion:   "",
			expectedApplyTime: 0,
		},
		{
			name:              "migration with long version string",
			version:           "200905_192800_create_users_table",
			applyTime:         1599332880,
			expectedVersion:   "200905_192800_create_users_table",
			expectedApplyTime: 1599332880,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			migration := Migration{
				Version:   tt.version,
				ApplyTime: tt.applyTime,
			}

			assert.Equal(t, tt.expectedVersion, migration.Version)
			assert.Equal(t, tt.expectedApplyTime, migration.ApplyTime)
		})
	}
}

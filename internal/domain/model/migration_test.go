/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMigration_ApplyTimeFormat(t *testing.T) {
	tests := []struct {
		name      string
		applyTime int64
	}{
		{
			name:      "unix timestamp formats correctly",
			applyTime: 1616961360,
		},
		{
			name:      "zero timestamp formats correctly",
			applyTime: 0,
		},
		{
			name:      "far future date formats correctly",
			applyTime: 4102444800,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Migration{
				Version:   "test",
				ApplyTime: tt.applyTime,
			}
			got := m.ApplyTimeFormat()
			// Check format pattern: YYYY-MM-DD HH:MM:SS
			assert.Regexp(t, `^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}$`, got)
		})
	}
}

func TestMigration_ApplyTimeFormat_ConsistentWithInput(t *testing.T) {
	m := Migration{
		Version:   "test",
		ApplyTime: 1616961360,
	}

	got1 := m.ApplyTimeFormat()
	got2 := m.ApplyTimeFormat()

	// Same input should produce same output
	assert.Equal(t, got1, got2)
}

func TestMigrations_Len(t *testing.T) {
	tests := []struct {
		name       string
		migrations Migrations
		want       int
	}{
		{
			name:       "empty migrations",
			migrations: Migrations{},
			want:       0,
		},
		{
			name: "single migration",
			migrations: Migrations{
				{Version: "210328_221600_test"},
			},
			want: 1,
		},
		{
			name: "multiple migrations",
			migrations: Migrations{
				{Version: "210328_221600_first"},
				{Version: "210328_221700_second"},
				{Version: "210328_221800_third"},
			},
			want: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.migrations.Len()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMigrations_Less(t *testing.T) {
	migrations := Migrations{
		{Version: "210328_221700_second"},
		{Version: "210328_221600_first"},
		{Version: "210328_221800_third"},
	}

	// first < second by version
	assert.True(t, migrations.Less(1, 0))  // 210328_221600 < 210328_221700
	assert.False(t, migrations.Less(0, 1)) // 210328_221700 > 210328_221600
	assert.True(t, migrations.Less(0, 2))  // 210328_221700 < 210328_221800
}

func TestMigrations_Swap(t *testing.T) {
	migrations := Migrations{
		{Version: "210328_221600_first"},
		{Version: "210328_221700_second"},
	}

	migrations.Swap(0, 1)

	assert.Equal(t, "210328_221700_second", migrations[0].Version)
	assert.Equal(t, "210328_221600_first", migrations[1].Version)
}

func TestMigrations_SortByVersion(t *testing.T) {
	tests := []struct {
		name       string
		migrations Migrations
		want       []string
	}{
		{
			name:       "empty migrations",
			migrations: Migrations{},
			want:       []string{},
		},
		{
			name: "already sorted",
			migrations: Migrations{
				{Version: "210328_221600_first"},
				{Version: "210328_221700_second"},
				{Version: "210328_221800_third"},
			},
			want: []string{
				"210328_221600_first",
				"210328_221700_second",
				"210328_221800_third",
			},
		},
		{
			name: "reverse order",
			migrations: Migrations{
				{Version: "210328_221800_third"},
				{Version: "210328_221700_second"},
				{Version: "210328_221600_first"},
			},
			want: []string{
				"210328_221600_first",
				"210328_221700_second",
				"210328_221800_third",
			},
		},
		{
			name: "mixed order",
			migrations: Migrations{
				{Version: "210328_221700_second"},
				{Version: "210328_221800_third"},
				{Version: "210328_221600_first"},
			},
			want: []string{
				"210328_221600_first",
				"210328_221700_second",
				"210328_221800_third",
			},
		},
		{
			name: "same date different time",
			migrations: Migrations{
				{Version: "210101_120000_midday"},
				{Version: "210101_000000_midnight"},
				{Version: "210101_235959_end"},
			},
			want: []string{
				"210101_000000_midnight",
				"210101_120000_midday",
				"210101_235959_end",
			},
		},
		{
			name: "different dates",
			migrations: Migrations{
				{Version: "220101_000000_next_year"},
				{Version: "200101_000000_old"},
				{Version: "210615_000000_mid_year"},
			},
			want: []string{
				"200101_000000_old",
				"210615_000000_mid_year",
				"220101_000000_next_year",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.migrations.SortByVersion()

			got := make([]string, len(tt.migrations))
			for i, m := range tt.migrations {
				got[i] = m.Version
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMigration_AllFields(t *testing.T) {
	m := Migration{
		Version:     "210328_221600_test",
		ApplyTime:   1616961360,
		BodySQL:     "CREATE TABLE test (id INT)",
		ExecutedSQL: "CREATE TABLE test (id INT)",
		Release:     "v1.0.0",
	}

	assert.Equal(t, "210328_221600_test", m.Version)
	assert.Equal(t, int64(1616961360), m.ApplyTime)
	assert.Equal(t, "CREATE TABLE test (id INT)", m.BodySQL)
	assert.Equal(t, "CREATE TABLE test (id INT)", m.ExecutedSQL)
	assert.Equal(t, "v1.0.0", m.Release)
}

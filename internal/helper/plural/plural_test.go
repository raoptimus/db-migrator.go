/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package plural

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNumberPlural_CountIsOne_ReturnsSingularForm(t *testing.T) {
	tests := []struct {
		name  string
		count int
		one   string
		many  string
		want  string
	}{
		{
			name:  "count is exactly 1",
			count: 1,
			one:   "item",
			many:  "items",
			want:  "item",
		},
		{
			name:  "count is 0",
			count: 0,
			one:   "item",
			many:  "items",
			want:  "item",
		},
		{
			name:  "count is negative",
			count: -1,
			one:   "item",
			many:  "items",
			want:  "item",
		},
		{
			name:  "count is large negative",
			count: -100,
			one:   "record",
			many:  "records",
			want:  "record",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NumberPlural(tt.count, tt.one, tt.many)

			require.Equal(t, tt.want, got)
		})
	}
}

func TestNumberPlural_CountIsGreaterThanOne_ReturnsPluralForm(t *testing.T) {
	tests := []struct {
		name  string
		count int
		one   string
		many  string
		want  string
	}{
		{
			name:  "count is 2 (boundary)",
			count: 2,
			one:   "item",
			many:  "items",
			want:  "items",
		},
		{
			name:  "count is 10",
			count: 10,
			one:   "item",
			many:  "items",
			want:  "items",
		},
		{
			name:  "count is large number",
			count: 1000000,
			one:   "record",
			many:  "records",
			want:  "records",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NumberPlural(tt.count, tt.one, tt.many)

			require.Equal(t, tt.want, got)
		})
	}
}

func TestNumberPlural_EmptyStrings_ReturnsEmptyString(t *testing.T) {
	tests := []struct {
		name  string
		count int
		one   string
		many  string
		want  string
	}{
		{
			name:  "empty one with count 1",
			count: 1,
			one:   "",
			many:  "items",
			want:  "",
		},
		{
			name:  "empty many with count 2",
			count: 2,
			one:   "item",
			many:  "",
			want:  "",
		},
		{
			name:  "both empty with count 1",
			count: 1,
			one:   "",
			many:  "",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NumberPlural(tt.count, tt.one, tt.many)

			require.Equal(t, tt.want, got)
		})
	}
}

func TestMigration_ReturnsSingularForm(t *testing.T) {
	tests := []struct {
		name  string
		count int
		want  string
	}{
		{
			name:  "count is 1",
			count: 1,
			want:  "migration",
		},
		{
			name:  "count is 0",
			count: 0,
			want:  "migration",
		},
		{
			name:  "count is negative",
			count: -5,
			want:  "migration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Migration(tt.count)

			require.Equal(t, tt.want, got)
		})
	}
}

func TestMigration_ReturnsPluralForm(t *testing.T) {
	tests := []struct {
		name  string
		count int
		want  string
	}{
		{
			name:  "count is 2",
			count: 2,
			want:  "migrations",
		},
		{
			name:  "count is 5",
			count: 5,
			want:  "migrations",
		},
		{
			name:  "count is 100",
			count: 100,
			want:  "migrations",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Migration(tt.count)

			require.Equal(t, tt.want, got)
		})
	}
}

func TestMigrationWas_ReturnsSingularForm(t *testing.T) {
	tests := []struct {
		name  string
		count int
		want  string
	}{
		{
			name:  "count is 1",
			count: 1,
			want:  "migration was",
		},
		{
			name:  "count is 0",
			count: 0,
			want:  "migration was",
		},
		{
			name:  "count is negative",
			count: -10,
			want:  "migration was",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MigrationWas(tt.count)

			require.Equal(t, tt.want, got)
		})
	}
}

func TestMigrationWas_ReturnsPluralForm(t *testing.T) {
	tests := []struct {
		name  string
		count int
		want  string
	}{
		{
			name:  "count is 2",
			count: 2,
			want:  "migrations were",
		},
		{
			name:  "count is 3",
			count: 3,
			want:  "migrations were",
		},
		{
			name:  "count is 50",
			count: 50,
			want:  "migrations were",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MigrationWas(tt.count)

			require.Equal(t, tt.want, got)
		})
	}
}

func TestMigrationHas_ReturnsSingularForm(t *testing.T) {
	tests := []struct {
		name  string
		count int
		want  string
	}{
		{
			name:  "count is 1",
			count: 1,
			want:  "migration has",
		},
		{
			name:  "count is 0",
			count: 0,
			want:  "migration has",
		},
		{
			name:  "count is negative",
			count: -1,
			want:  "migration has",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MigrationHas(tt.count)

			require.Equal(t, tt.want, got)
		})
	}
}

func TestMigrationHas_ReturnsPluralForm(t *testing.T) {
	tests := []struct {
		name  string
		count int
		want  string
	}{
		{
			name:  "count is 2",
			count: 2,
			want:  "migrations have",
		},
		{
			name:  "count is 7",
			count: 7,
			want:  "migrations have",
		},
		{
			name:  "count is 999",
			count: 999,
			want:  "migrations have",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MigrationHas(tt.count)

			require.Equal(t, tt.want, got)
		})
	}
}

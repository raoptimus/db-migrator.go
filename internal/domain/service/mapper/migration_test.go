/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package mapper

import (
	"testing"

	"github.com/raoptimus/db-migrator.go/internal/domain/model"
	"github.com/raoptimus/db-migrator.go/internal/infrastructure/dal/entity"
	"github.com/stretchr/testify/assert"
)

func TestEntityToDomain(t *testing.T) {
	tests := []struct {
		name   string
		entity entity.Migration
		want   model.Migration
	}{
		{
			name: "converts entity with all fields",
			entity: entity.Migration{
				Version:   "210328_221600_test",
				ApplyTime: 1616961360,
			},
			want: model.Migration{
				Version:   "210328_221600_test",
				ApplyTime: 1616961360,
			},
		},
		{
			name: "converts entity with zero apply time",
			entity: entity.Migration{
				Version:   "200101_000000_init",
				ApplyTime: 0,
			},
			want: model.Migration{
				Version:   "200101_000000_init",
				ApplyTime: 0,
			},
		},
		{
			name:   "converts empty entity",
			entity: entity.Migration{},
			want:   model.Migration{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EntityToDomain(tt.entity)
			assert.Equal(t, tt.want.Version, got.Version)
			assert.Equal(t, tt.want.ApplyTime, got.ApplyTime)
		})
	}
}

func TestDomainToEntity(t *testing.T) {
	tests := []struct {
		name      string
		migration model.Migration
		want      entity.Migration
	}{
		{
			name: "converts migration with all fields",
			migration: model.Migration{
				Version:   "210328_221600_test",
				ApplyTime: 1616961360,
			},
			want: entity.Migration{
				Version:   "210328_221600_test",
				ApplyTime: 1616961360,
			},
		},
		{
			name: "converts migration with extra fields (they are ignored)",
			migration: model.Migration{
				Version:     "210328_221600_test",
				ApplyTime:   1616961360,
				BodySQL:     "CREATE TABLE test",
				ExecutedSQL: "CREATE TABLE test",
				Release:     "v1.0.0",
			},
			want: entity.Migration{
				Version:   "210328_221600_test",
				ApplyTime: 1616961360,
			},
		},
		{
			name:      "converts empty migration",
			migration: model.Migration{},
			want:      entity.Migration{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DomainToEntity(tt.migration)
			assert.Equal(t, tt.want.Version, got.Version)
			assert.Equal(t, tt.want.ApplyTime, got.ApplyTime)
		})
	}
}

func TestEntitiesToDomain(t *testing.T) {
	tests := []struct {
		name     string
		entities entity.Migrations
		want     model.Migrations
	}{
		{
			name: "converts multiple entities",
			entities: entity.Migrations{
				{Version: "210328_221600_first", ApplyTime: 1616961360},
				{Version: "210328_221700_second", ApplyTime: 1616961420},
				{Version: "210328_221800_third", ApplyTime: 1616961480},
			},
			want: model.Migrations{
				{Version: "210328_221600_first", ApplyTime: 1616961360},
				{Version: "210328_221700_second", ApplyTime: 1616961420},
				{Version: "210328_221800_third", ApplyTime: 1616961480},
			},
		},
		{
			name:     "converts empty slice",
			entities: entity.Migrations{},
			want:     model.Migrations{},
		},
		{
			name:     "converts nil slice",
			entities: nil,
			want:     model.Migrations{},
		},
		{
			name: "converts single entity",
			entities: entity.Migrations{
				{Version: "210328_221600_single", ApplyTime: 1616961360},
			},
			want: model.Migrations{
				{Version: "210328_221600_single", ApplyTime: 1616961360},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EntitiesToDomain(tt.entities)
			assert.Equal(t, len(tt.want), len(got))
			for i := range tt.want {
				assert.Equal(t, tt.want[i].Version, got[i].Version)
				assert.Equal(t, tt.want[i].ApplyTime, got[i].ApplyTime)
			}
		})
	}
}

func TestDomainsToEntities(t *testing.T) {
	tests := []struct {
		name       string
		migrations model.Migrations
		want       entity.Migrations
	}{
		{
			name: "converts multiple migrations",
			migrations: model.Migrations{
				{Version: "210328_221600_first", ApplyTime: 1616961360},
				{Version: "210328_221700_second", ApplyTime: 1616961420},
				{Version: "210328_221800_third", ApplyTime: 1616961480},
			},
			want: entity.Migrations{
				{Version: "210328_221600_first", ApplyTime: 1616961360},
				{Version: "210328_221700_second", ApplyTime: 1616961420},
				{Version: "210328_221800_third", ApplyTime: 1616961480},
			},
		},
		{
			name:       "converts empty slice",
			migrations: model.Migrations{},
			want:       entity.Migrations{},
		},
		{
			name:       "converts nil slice",
			migrations: nil,
			want:       entity.Migrations{},
		},
		{
			name: "converts migrations with extra fields",
			migrations: model.Migrations{
				{
					Version:     "210328_221600_test",
					ApplyTime:   1616961360,
					BodySQL:     "CREATE TABLE test",
					ExecutedSQL: "CREATE TABLE test",
					Release:     "v1.0.0",
				},
			},
			want: entity.Migrations{
				{Version: "210328_221600_test", ApplyTime: 1616961360},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DomainsToEntities(tt.migrations)
			assert.Equal(t, len(tt.want), len(got))
			for i := range tt.want {
				assert.Equal(t, tt.want[i].Version, got[i].Version)
				assert.Equal(t, tt.want[i].ApplyTime, got[i].ApplyTime)
			}
		})
	}
}

func TestRoundTripConversion(t *testing.T) {
	original := entity.Migration{
		Version:   "210328_221600_roundtrip",
		ApplyTime: 1616961360,
	}

	domain := EntityToDomain(original)
	converted := DomainToEntity(domain)

	assert.Equal(t, original.Version, converted.Version)
	assert.Equal(t, original.ApplyTime, converted.ApplyTime)
}

func TestRoundTripConversionSlices(t *testing.T) {
	original := entity.Migrations{
		{Version: "210328_221600_first", ApplyTime: 1616961360},
		{Version: "210328_221700_second", ApplyTime: 1616961420},
	}

	domain := EntitiesToDomain(original)
	converted := DomainsToEntities(domain)

	assert.Equal(t, len(original), len(converted))
	for i := range original {
		assert.Equal(t, original[i].Version, converted[i].Version)
		assert.Equal(t, original[i].ApplyTime, converted[i].ApplyTime)
	}
}

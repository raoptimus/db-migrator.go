/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package mapper

import (
	"github.com/raoptimus/db-migrator.go/internal/domain/model"
	"github.com/raoptimus/db-migrator.go/internal/infrastructure/dal/entity"
)

// EntityToDomain converts a DAL entity.Migration to a domain model.Migration.
func EntityToDomain(e entity.Migration) model.Migration {
	return model.Migration{
		Version:   e.Version,
		ApplyTime: e.ApplyTime,
	}
}

// DomainToEntity converts a domain model.Migration to a DAL entity.Migration.
func DomainToEntity(m model.Migration) entity.Migration {
	return entity.Migration{
		Version:   m.Version,
		ApplyTime: m.ApplyTime,
	}
}

// EntitiesToDomain converts a slice of DAL entity.Migration to domain model.Migrations.
func EntitiesToDomain(entities entity.Migrations) model.Migrations {
	result := make(model.Migrations, len(entities))
	for i, e := range entities {
		result[i] = EntityToDomain(e)
	}
	return result
}

// DomainsToEntities converts a slice of domain model.Migration to DAL entity.Migrations.
func DomainsToEntities(migrations model.Migrations) entity.Migrations {
	result := make(entity.Migrations, len(migrations))
	for i, m := range migrations {
		result[i] = DomainToEntity(m)
	}
	return result
}

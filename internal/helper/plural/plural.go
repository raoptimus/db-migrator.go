/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package plural

const (
	migration      = "migration"
	migrations     = "migrations"
	migrationWas   = "migration was"
	migrationsWere = "migrations were"
	migrationHas   = "migration has"
	migrationsHave = "migrations have"
)

// NumberPlural returns the appropriate singular or plural form based on the count.
// If count is greater than 1, it returns the many form; otherwise, it returns the one form.
func NumberPlural(c int, one, many string) string {
	if c > 1 {
		return many
	}

	return one
}

// Migration returns "migration" for count 1 or "migrations" for count greater than 1.
func Migration(c int) string {
	return NumberPlural(c, migration, migrations)
}

// MigrationWas returns "migration was" for count 1 or "migrations were" for count greater than 1.
func MigrationWas(c int) string {
	return NumberPlural(c, migrationWas, migrationsWere)
}

// MigrationHas returns "migration has" for count 1 or "migrations have" for count greater than 1.
func MigrationHas(c int) string {
	return NumberPlural(c, migrationHas, migrationsHave)
}

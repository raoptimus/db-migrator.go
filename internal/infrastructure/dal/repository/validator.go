/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package repository

import "github.com/pkg/errors"

const maxIdentifierLen = 65000

var ErrInvalidIdentifier = errors.New("invalid SQL identifier")

// ValidateIdentifier validates SQL identifiers (table names, schema names, cluster names).
// Allowed characters: a-z, A-Z, 0-9, underscore (_)
// Empty string is allowed for optional fields.
// Returns ErrInvalidIdentifier if the identifier contains invalid characters or is too long.
func ValidateIdentifier(name string) error {
	if name == "" {
		return nil // empty is allowed for optional fields
	}

	if len(name) > maxIdentifierLen {
		return errors.Wrapf(ErrInvalidIdentifier, "identifier too long: %d chars", len(name))
	}

	for _, r := range name {
		if !isValidIdentifierChar(r) {
			return errors.Wrapf(ErrInvalidIdentifier, "invalid character '%c' in: %s", r, name)
		}
	}

	return nil
}

func isValidIdentifierChar(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
		(r >= '0' && r <= '9') || r == '_'
}

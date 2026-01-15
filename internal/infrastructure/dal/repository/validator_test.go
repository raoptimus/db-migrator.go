/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package repository

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateIdentifier_ValidNames(t *testing.T) {
	validNames := []string{
		"migration",
		"my_table",
		"Table123",
		"UPPERCASE_TABLE",
		"table_2024_01_15",
		"_underscore_prefix",
		"table_with_many_underscores_123",
		"a",
		"A1",
		"_",
		"", // empty is valid for optional fields
	}

	for _, name := range validNames {
		t.Run(name, func(t *testing.T) {
			err := ValidateIdentifier(name)
			assert.NoError(t, err, "Expected %s to be valid", name)
		})
	}
}

func TestValidateIdentifier_InvalidNames(t *testing.T) {
	invalidNames := map[string]string{
		"table-name":       "contains hyphen",
		"table.name":       "contains dot",
		"table name":       "contains space",
		"table;DROP TABLE": "SQL injection attempt semicolon",
		"table/*comment*/": "SQL injection with comment",
		"table'OR'1'='1":   "SQL injection with quotes",
		"table`name":       "contains backtick",
		"table\"name":      "contains quote",
		"table$name":       "contains dollar",
		"table@name":       "contains at sign",
		"table#name":       "contains hash",
		"table%name":       "contains percent",
		"table&name":       "contains ampersand",
		"table*name":       "contains asterisk",
		"table(name)":      "contains parentheses",
		"table[name]":      "contains brackets",
		"table{name}":      "contains braces",
		"table+name":       "contains plus",
		"table=name":       "contains equals",
		"table,name":       "contains comma",
		"table<name>":      "contains angle brackets",
		"table|name":       "contains pipe",
		"table\\name":      "contains backslash",
		"table/name":       "contains slash",
		"table:name":       "contains colon",
		"table?name":       "contains question mark",
		"table!name":       "contains exclamation",
		"table~name":       "contains tilde",
		"таблица":          "contains cyrillic",
		"表":                "contains chinese characters",
		"table\nname":      "contains newline",
		"table\tname":      "contains tab",
		"table\rname":      "contains carriage return",
	}

	for name, reason := range invalidNames {
		t.Run(reason, func(t *testing.T) {
			err := ValidateIdentifier(name)
			assert.Error(t, err, "Expected '%s' to be invalid: %s", name, reason)
			assert.ErrorIs(t, err, ErrInvalidIdentifier)
			assert.Contains(t, err.Error(), "invalid character", "Error message should mention invalid character")
		})
	}
}

func TestValidateIdentifier_TooLong(t *testing.T) {
	longName := strings.Repeat("a", 70000)
	err := ValidateIdentifier(longName)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidIdentifier)
	assert.Contains(t, err.Error(), "too long", "Error should mention that identifier is too long")
}

func TestValidateIdentifier_ExactlyMaxLength(t *testing.T) {
	// maxIdentifierLen is 65000, test exactly at boundary
	exactlyMaxName := strings.Repeat("a", 65000)
	err := ValidateIdentifier(exactlyMaxName)
	assert.NoError(t, err, "Identifier exactly at max length should be valid")
}

func TestValidateIdentifier_EmptyString(t *testing.T) {
	err := ValidateIdentifier("")
	assert.NoError(t, err, "Empty string should be valid for optional fields")
}

func TestValidateIdentifier_SQLInjectionAttempts(t *testing.T) {
	injectionAttempts := []string{
		"'; DROP TABLE users; --",
		"1' OR '1'='1",
		"admin'--",
		"' OR 1=1--",
		"' UNION SELECT * FROM users--",
		"'; DELETE FROM migration; --",
		"table` OR `1`=`1",
		"table\"; DROP TABLE users; --",
	}

	for _, attempt := range injectionAttempts {
		t.Run(attempt, func(t *testing.T) {
			err := ValidateIdentifier(attempt)
			assert.Error(t, err, "SQL injection attempt should be rejected: %s", attempt)
			assert.ErrorIs(t, err, ErrInvalidIdentifier)
		})
	}
}

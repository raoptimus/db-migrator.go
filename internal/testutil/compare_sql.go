/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package thelp

import (
	"regexp"
	"strings"
)

var (
	//nolint:gocritic // Ок
	patternWhitespace      = regexp.MustCompile(`[\s\r\n\t]+`)
	patternSpaceCommaSpace = regexp.MustCompile(`\s*,\s+`)
)

// CompareSQL returns a function that compares an actual SQL string against the expected SQL string.
// Both strings are cleaned (whitespace normalized, etc.) before comparison.
func CompareSQL(expected string) func(actual string) bool {
	return func(actual string) bool {
		expected = CleanSQL(expected)
		actual = CleanSQL(actual)
		result := expected == actual

		return result
	}
}

// CleanSQL normalizes SQL text by removing extra whitespace, normalizing spacing around parentheses and commas.
// This allows for consistent comparison of SQL statements that may have different formatting.
func CleanSQL(sql string) string {
	sql = strings.TrimSpace(sql)

	sql = patternWhitespace.ReplaceAllString(sql, " ")
	sql = patternSpaceCommaSpace.ReplaceAllString(sql, ",")

	sql = strings.ReplaceAll(sql, " )", ")")
	sql = strings.ReplaceAll(sql, "( ", "(")

	return sql
}

package testhelper

import (
	"regexp"
	"strings"
)

var (
	//nolint:gocritic // Ок
	patternWhitespace      = regexp.MustCompile(`[\s\r\n\t]+`)
	patternSpaceCommaSpace = regexp.MustCompile(`\s*,\s+`)
)

func CompareSQL(expected string) func(actual string) bool {
	return func(actual string) bool {
		expected = CleanSQL(expected)
		actual = CleanSQL(actual)
		return expected == actual
	}
}

func CleanSQL(sql string) string {
	sql = strings.TrimSpace(sql)

	sql = patternWhitespace.ReplaceAllString(sql, " ")
	sql = patternSpaceCommaSpace.ReplaceAllString(sql, ",")

	sql = strings.ReplaceAll(sql, " )", ")")
	sql = strings.ReplaceAll(sql, "( ", "(")

	return sql
}

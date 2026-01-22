package validator

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateIdentifier_ValidIdentifiers_Success(t *testing.T) {
	tests := []struct {
		name string
		id   string
	}{
		{
			name: "empty string is valid",
			id:   "",
		},
		{
			name: "lowercase letters",
			id:   "tablename",
		},
		{
			name: "uppercase letters",
			id:   "TABLENAME",
		},
		{
			name: "mixed case letters",
			id:   "TableName",
		},
		{
			name: "with numbers",
			id:   "table123",
		},
		{
			name: "with underscore",
			id:   "table_name",
		},
		{
			name: "underscore prefix",
			id:   "_tablename",
		},
		{
			name: "numbers only",
			id:   "123456",
		},
		{
			name: "single character",
			id:   "a",
		},
		{
			name: "single underscore",
			id:   "_",
		},
		{
			name: "complex valid identifier",
			id:   "MyTable_123_name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateIdentifier(tt.id)
			require.NoError(t, err)
		})
	}
}

func TestValidateIdentifier_InvalidIdentifiers_Failure(t *testing.T) {
	tests := []struct {
		name string
		id   string
	}{
		{
			name: "contains space",
			id:   "table name",
		},
		{
			name: "contains semicolon",
			id:   "table;name",
		},
		{
			name: "contains single quote",
			id:   "table'name",
		},
		{
			name: "contains double quote",
			id:   `table"name`,
		},
		{
			name: "contains dash",
			id:   "table-name",
		},
		{
			name: "contains dot",
			id:   "table.name",
		},
		{
			name: "contains special chars",
			id:   "table@name",
		},
		{
			name: "sql injection attempt",
			id:   "table'; DROP TABLE users;--",
		},
		{
			name: "contains newline",
			id:   "table\nname",
		},
		{
			name: "contains tab",
			id:   "table\tname",
		},
		{
			name: "contains unicode",
			id:   "таблица",
		},
		{
			name: "contains emoji",
			id:   "table😀name",
		},
		{
			name: "exceeds max length",
			id:   strings.Repeat("a", 257),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateIdentifier(tt.id)
			require.Error(t, err)
			assert.ErrorIs(t, err, ErrIdentifierIsNotValid)
		})
	}
}

func TestValidateIdentifier_MaxLengthBoundary(t *testing.T) {
	// Test at max length (256) - should be valid
	maxLenID := strings.Repeat("a", 256)
	err := ValidateIdentifier(maxLenID)
	require.NoError(t, err)

	// Test at max length + 1 (257) - should be invalid
	overMaxLenID := strings.Repeat("a", 257)
	err = ValidateIdentifier(overMaxLenID)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrIdentifierIsNotValid)
}

func TestIsValidIdentifierChar_ValidChars(t *testing.T) {
	validChars := []rune{
		'a', 'b', 'z', // lowercase
		'A', 'B', 'Z', // uppercase
		'0', '5', '9', // digits
		'_',           // underscore
	}

	for _, r := range validChars {
		t.Run(string(r), func(t *testing.T) {
			assert.True(t, isValidIdentifierChar(r))
		})
	}
}

func TestIsValidIdentifierChar_InvalidChars(t *testing.T) {
	invalidChars := []rune{
		' ',  // space
		'-',  // dash
		'.',  // dot
		';',  // semicolon
		'\'', // single quote
		'"',  // double quote
		'@',  // at
		'#',  // hash
		'$',  // dollar
		'%',  // percent
		'&',  // ampersand
		'*',  // asterisk
		'(',  // parentheses
		')',
		'[',
		']',
		'{',
		'}',
		'\n', // newline
		'\t', // tab
		'я',  // cyrillic
	}

	for _, r := range invalidChars {
		t.Run(string(r), func(t *testing.T) {
			assert.False(t, isValidIdentifierChar(r))
		})
	}
}

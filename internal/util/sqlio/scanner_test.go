/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package sqlio

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	testCases := []struct {
		name        string
		multiStmt   string
		delimiter   string
		expected    []string
		expectedErr error
	}{
		{
			name:        "single statement, no delimiter",
			multiStmt:   "single statement, no delimiter",
			expected:    []string{"single statement, no delimiter"},
			expectedErr: nil,
		},
		{
			name:        "single statement, one delimiter",
			multiStmt:   "single statement, one delimiter",
			expected:    []string{"single statement, one delimiter"},
			expectedErr: nil,
		},
		{
			name:        "two statements, no trailing delimiter",
			multiStmt:   "statement one; statement two",
			expected:    []string{"statement one", "statement two"},
			expectedErr: nil,
		},
		{
			name:        "two statements, with trailing delimiter",
			multiStmt:   "statement one; statement two;",
			expected:    []string{"statement one", "statement two"},
			expectedErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			stmts := make([]string, 0, len(tc.expected))
			scanner := NewScanner(strings.NewReader(tc.multiStmt))

			for scanner.Scan() {
				stmts = append(stmts, scanner.SQL())
			}
			err := scanner.Err()
			assert.Equal(t, tc.expectedErr, err)
			assert.Equal(t, tc.expected, stmts)
		})
	}
}

func TestParseDiscontinue(t *testing.T) {
	multiStmt := "statement one; statement two"
	expected := []string{"statement one", "statement two"}
	stmts := make([]string, 0, len(expected))
	scanner := NewScanner(strings.NewReader(multiStmt))
	for scanner.Scan() {
		stmts = append(stmts, scanner.SQL())
	}
	err := scanner.Err()
	assert.Nil(t, err)
	assert.Equal(t, expected, stmts)
}

func TestParsePostgresFunctions(t *testing.T) {
	expected := []string{`CREATE test`,
		`CREATE OR REPLACE FUNCTION test_index_update() RETURNS trigger AS $$
BEGIN
    	something;
    ELSEIF(TG_OP = 'UPDATE') THEN
		something;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql`, `CREATE TRIGGER test_index_update_trigger`}
	multiStmt := strings.Join(expected, "; ")
	stmts := make([]string, 0, len(expected))
	scanner := NewScanner(strings.NewReader(multiStmt))
	for scanner.Scan() {
		stmts = append(stmts, scanner.SQL())
	}
	err := scanner.Err()
	assert.Nil(t, err)
	assert.Equal(t, expected, stmts)
}

func TestParseFailure(t *testing.T) {
	multiStmt := `CREATE test;
CREATE OR REPLACE FUNCTION test_index_update() RETURNS trigger AS $$
BEGIN
    	something;
    ELSEIF(TG_OP = 'UPDATE') THEN
		something;
    END IF;

    RETURN NEW;
END;
 LANGUAGE plpgsql
;
CREATE TRIGGER test_index_update_trigger`

	scanner := NewScanner(strings.NewReader(multiStmt))
	for scanner.Scan() {
		scanner.SQL()
	}
	err := scanner.Err()
	assert.Error(t, err)
}

func TestBufferReadAll(t *testing.T) {
	multiStmt := `CREATE test;
CREATE OR REPLACE FUNCTION test_index_update() RETURNS trigger AS $$
BEGIN
    	something;
    ELSEIF(TG_OP = 'UPDATE') THEN
		something;
    END IF;

    RETURN NEW;
END;
$$
 LANGUAGE plpgsql
;
CREATE TRIGGER test_index_update_trigger`

	StartBufSize = 100

	scanner := NewScanner(strings.NewReader(multiStmt))
	for scanner.Scan() {
		scanner.SQL()
	}
	err := scanner.Err()
	assert.NoError(t, err)
}

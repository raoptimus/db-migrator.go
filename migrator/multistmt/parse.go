package multistmt

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"strings"
)

const maxMigrationSize = 10 * 1 << 20

var multiStmtDelimiter = []byte(";")

// StartBufSize is the default starting size of the buffer used to scan and parse multi-statement migrations
var StartBufSize = 4096

// Handler handles a single migration parsed from a multi-statement migration.
// It's given the single migration to handle and returns whether or not further statements
// from the multi-statement migration should be parsed and handled.
type Handler func(sqlQuery string) error

func splitWithDelimiter() func(d []byte, atEOF bool) (int, []byte, error) {
	return func(d []byte, atEOF bool) (int, []byte, error) {
		// SplitFunc inspired by bufio.ScanLines() implementation
		if atEOF {
			if len(d) == 0 {
				return 0, nil, nil
			}
			return len(d), d, nil
		}
		if i := bytes.Index(d, multiStmtDelimiter); i >= 0 {
			return i + len(multiStmtDelimiter), d[:i+len(multiStmtDelimiter)], nil
		}
		return 0, nil, nil
	}
}

// Parse parses the given multi-statement migration
func Parse(reader io.Reader, callback Handler) error {
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 0, StartBufSize), maxMigrationSize)
	scanner.Split(splitWithDelimiter())

	var sqlQuery string
	for scanner.Scan() {
		sqlQuery = string(scanner.Bytes())
		sqlQuery = strings.Trim(sqlQuery, " \n")
		if sqlQuery == "" {
			continue
		}
		if err := callback(sqlQuery); err != nil {
			return err
		}
	}

	return scanner.Err()
}

func ParseSQLFile(filename string, callback Handler) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}

	return Parse(f, callback)
}

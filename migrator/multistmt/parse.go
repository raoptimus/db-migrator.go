package multistmt

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

const maxMigrationSize = 10 * 1 << 20

var (
	multiStmtDelimiter  = []byte(";")
	psqlPLFuncDelimiter = []byte("$$")
	skip                = 0
)

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
		}

		openPi, pLen := bytes.Index(d, psqlPLFuncDelimiter), len(psqlPLFuncDelimiter)
		delI, delLen := bytes.Index(d, multiStmtDelimiter), len(multiStmtDelimiter)

		switch {
		case openPi > delI:
			if len(d[:delI+delLen]) == 0 {
				return 0, nil, nil
			}
			return delI + delLen, d[:delI+delLen], nil

		case openPi >= 0 && openPi < delI:
			closePi := bytes.Index(d[openPi+pLen:], psqlPLFuncDelimiter)
			if closePi < 0 {
				var err error
				if atEOF {
					err = fmt.Errorf("closed tag %s not found", psqlPLFuncDelimiter)
				}
				return 0, nil, err
			}
			offset := closePi + openPi + pLen
			delI = bytes.Index(d[offset:], multiStmtDelimiter)
			offset = offset + delI + delLen
			if delI < 0 {
				var err error
				if atEOF {
					err = fmt.Errorf("closed tag %s not found", multiStmtDelimiter)
				}
				return 0, nil, err
			}
			return offset, d[:offset], nil

		case delI >= 0:
			return delI + delLen, d[:delI+delLen], nil

		default:
			if atEOF {
				return len(d), d, nil
			}
			return 0, nil, nil
		}
	}
}

// Parse parses the given multi-statement migration
func Parse(reader io.Reader, callback Handler) error {
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 0, StartBufSize), maxMigrationSize)
	scanner.Split(splitWithDelimiter())

	var (
		sqlQuery string
	)
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

func ReadSQLFile(filename string, callback Handler) error {
	var (
		sqlBytes []byte
		err      error
	)
	sqlBytes, err = ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	return callback(string(sqlBytes))
}

func ReadOrParseSQLFile(filename string, multiSTMT bool, callback Handler) error {
	if multiSTMT {
		return ReadSQLFile(filename, callback)
	}

	return ParseSQLFile(filename, callback)
}

/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package sqlio

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
)

const (
	maxMigrationSize = 10 * 1 << 20
)

var (
	multiStmtDelimiter  = []byte(";")
	psqlPLFuncDelimiter = []byte("$$")
)

// StartBufSize is the default starting size of the buffer used to scan and parse multi-statement migrations.
var StartBufSize = 4096

type Scanner struct {
	scanner *bufio.Scanner
	sql     string
	err     error
	done    bool
}

func NewScanner(r io.Reader) *Scanner {
	s := bufio.NewScanner(r)
	s.Buffer(make([]byte, 0, StartBufSize), maxMigrationSize)
	s.Split(splitWithDelimiter())

	return &Scanner{
		scanner: s,
	}
}

func (s *Scanner) SQL() string {
	return s.sql
}

func (s *Scanner) Err() error {
	return s.err
}

func (s *Scanner) Scan() bool {
	if s.done {
		return false
	}
	for s.scanner.Scan() {
		s.sql = strings.TrimSpace(s.scanner.Text())
		s.sql = strings.Trim(s.sql, ";")
		if s.sql == "" {
			continue
		}
		return true
	}

	s.err = s.scanner.Err()
	return false
}

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

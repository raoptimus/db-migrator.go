/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package validator

import (
	"regexp"
	"time"

	"github.com/pkg/errors"
)

const patternFileName = `^(?P<Year>\d{2})(?P<Month>\d{2})(?P<Day>\d{2})\_(?P<Hour>\d{2})(?P<Minute>\d{2})(?P<Second>\d{2})\_[a-z][a-z0-9\_\-]+(\.safe)?\.(up|down)\.sql$`

var (
	ErrFileNameIsNotValid = errors.New("file name is not valid. File name must be eq pattern: YYMMDD_hhmmss_[a-z][a-z0-9\\_\\-]+(\\.safe)?\\.(up|down)\\.sql")

	groupsLenFileName = len(regexpFileName.SubexpNames())
	regexpFileName    = regexp.MustCompile(patternFileName)
)

// ValidateFileName validates that a migration file name follows the required naming convention.
// The name must match the pattern: YYMMDD_hhmmss_name[.safe].(up|down).sql
// The timestamp must be valid and not in the future (accounting for timezone differences).
func ValidateFileName(name string) error {
	if len(name) == 0 {
		return ErrVersionIsNotValid
	}

	groups := regexpFileName.FindStringSubmatch(name)
	if len(groups) < groupsLenFileName {
		return ErrFileNameIsNotValid
	}

	yy, mm, dd, h, m, s := groups[1], groups[2], groups[3], groups[4], groups[5], groups[6]

	dt, err := time.Parse(time.DateTime, "20"+yy+"-"+mm+"-"+dd+" "+h+":"+m+":"+s)
	if err != nil {
		return ErrFileNameIsNotValid
	}

	if dt.After(time.Now().Add(maxTZ)) {
		return ErrFileNameIsNotValid
	}

	return nil
}

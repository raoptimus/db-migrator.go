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

const maxTZ = 8 * time.Hour
const patternVersion = `^(?P<Year>\d{2})(?P<Month>\d{2})(?P<Day>\d{2})\_(?P<Hour>\d{2})(?P<Minute>\d{2})(?P<Second>\d{2})\_[a-z][a-z0-9\_\-]+$`

var (
	ErrVersionIsNotValid = errors.New("Version is not valid. Version must be eq pattern: YYMMDD_hhmmss_[a-z0-9_]+")

	regexpVersion    = regexp.MustCompile(patternVersion)
	groupsLenVersion = len(regexpVersion.SubexpNames())
)

// ValidateVersion validates that a migration version string follows the required format.
// The version must match the pattern: YYMMDD_hhmmss_name
// The timestamp must be valid and not in the future (accounting for timezone differences).
func ValidateVersion(version string) error {
	if len(version) == 0 {
		return ErrVersionIsNotValid
	}

	groups := regexpVersion.FindStringSubmatch(version)
	if len(groups) < groupsLenVersion {
		return ErrVersionIsNotValid
	}

	yy, mm, dd, h, m, s := groups[1], groups[2], groups[3], groups[4], groups[5], groups[6]

	dt, err := time.Parse(time.DateTime, "20"+yy+"-"+mm+"-"+dd+" "+h+":"+m+":"+s)
	if err != nil {
		return ErrVersionIsNotValid
	}

	if dt.After(time.Now().Add(maxTZ)) {
		return ErrVersionIsNotValid
	}

	return nil
}

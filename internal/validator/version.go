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
const patternVersion = `^(?P<Year>\d{2})(?P<Month>\d{2})(?P<Day>\d{2})\_(?P<Hour>\d{2})(?P<Minute>\d{2})(?P<Second>\d{2})\_[a-zA-Z][a-zA-Z0-9\_\-]+$`

var (
	regexpVersion        = regexp.MustCompile(patternVersion)
	groupsLenVersion     = len(regexpVersion.SubexpNames())
	errVersionIsNotValid = errors.New("Version is not valid. Version must be eq pattern: YYMMDD_hhmmss_[a-z0-9_]+")
)

func ValidateVersion(version string) error {
	if len(version) == 0 {
		return errVersionIsNotValid
	}

	groups := regexpVersion.FindStringSubmatch(version)
	if len(groups) < groupsLenVersion {
		return errVersionIsNotValid
	}

	yy, mm, dd, h, m, s := groups[1], groups[2], groups[3], groups[4], groups[5], groups[6]

	dt, err := time.Parse(time.DateTime, "20"+yy+"-"+mm+"-"+dd+" "+h+":"+m+":"+s)
	if err != nil {
		return errVersionIsNotValid
	}

	if dt.After(time.Now().Add(maxTZ)) {
		return errVersionIsNotValid
	}

	return nil
}

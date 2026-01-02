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

const patternFileName = `^(?P<Year>\d{2})(?P<Month>\d{2})(?P<Day>\d{2})\_(?P<Hour>\d{2})(?P<Minute>\d{2})(?P<Second>\d{2})\_[a-zA-Z][a-zA-Z0-9\_\-]+(\.safe)?\.(up|down)\.sql$`

var (
	regexpFileName        = regexp.MustCompile(patternFileName)
	groupsLenFileName     = len(regexpFileName.SubexpNames())
	errFileNameIsNotValid = errors.New("file name is not valid. File name must be eq pattern: YYMMDD_hhmmss_[a-z][a-z0-9\\_\\-]+(\\.safe)?\\.(up|down)\\.sql")
)

func ValidateFileName(name string) error {
	if len(name) == 0 {
		return errVersionIsNotValid
	}

	groups := regexpFileName.FindStringSubmatch(name)
	if len(groups) < groupsLenFileName {
		return errFileNameIsNotValid
	}

	yy, mm, dd, h, m, s := groups[1], groups[2], groups[3], groups[4], groups[5], groups[6]

	dt, err := time.Parse(time.DateTime, "20"+yy+"-"+mm+"-"+dd+" "+h+":"+m+":"+s)
	if err != nil {
		return errFileNameIsNotValid
	}

	if dt.After(time.Now().Add(maxTZ)) {
		return errFileNameIsNotValid
	}

	return nil
}

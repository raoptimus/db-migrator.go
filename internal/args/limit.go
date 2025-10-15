/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package args

import (
	"fmt"
	"strconv"

	"github.com/pkg/errors"
)

const (
	empty = ""
	all   = "all"
)

var ErrArgumentMustBeGreaterThanZero = errors.New("the step argument must be greater than 0")

func ParseStepStringOrDefault(value string, defaults int) (int, error) {
	switch value {
	case empty:
		return defaults, nil
	case all:
		return 0, nil
	default:
		i, err := strconv.Atoi(value)
		if err != nil {
			return -1, fmt.Errorf("the step argument %s is not valid", value)
		}

		if i < 1 {
			return -1, ErrArgumentMustBeGreaterThanZero
		}

		return i, nil
	}
}

package ddl

import (
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// transformArgCount is the number of arguments expected for bucket and truncate transforms.
const transformArgCount = 2

// parseTransform parses a partition transform expression such as:
//
//	identity(col)
//	years(col)  months(col)  days(col)  hours(col)
//	bucket(16, col)
//	truncate(8, col)
//
// Unknown transform names return ErrUnknownTransform.
func parseTransform(expr string) (PartitionField, error) {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return PartitionField{}, errors.Wrapf(ErrParse, "empty partition transform expression")
	}

	// Find the opening parenthesis.
	parenIdx := strings.Index(expr, "(")
	if parenIdx < 0 {
		return PartitionField{}, errors.Wrapf(ErrParse, "invalid partition transform %q: missing parentheses", expr)
	}
	if !strings.HasSuffix(expr, ")") {
		return PartitionField{}, errors.Wrapf(ErrParse, "invalid partition transform %q: missing closing parenthesis", expr)
	}

	funcName := strings.ToLower(strings.TrimSpace(expr[:parenIdx]))
	argsStr := strings.TrimSpace(expr[parenIdx+1 : len(expr)-1])

	switch funcName {
	case "identity":
		col, err := singleArg(argsStr, expr)
		if err != nil {
			return PartitionField{}, err
		}
		return PartitionField{Transform: Identity, SourceCol: col}, nil

	case "years":
		col, err := singleArg(argsStr, expr)
		if err != nil {
			return PartitionField{}, err
		}
		return PartitionField{Transform: Years, SourceCol: col}, nil

	case "months":
		col, err := singleArg(argsStr, expr)
		if err != nil {
			return PartitionField{}, err
		}
		return PartitionField{Transform: Months, SourceCol: col}, nil

	case "days":
		col, err := singleArg(argsStr, expr)
		if err != nil {
			return PartitionField{}, err
		}
		return PartitionField{Transform: Days, SourceCol: col}, nil

	case "hours":
		col, err := singleArg(argsStr, expr)
		if err != nil {
			return PartitionField{}, err
		}
		return PartitionField{Transform: Hours, SourceCol: col}, nil

	case "bucket":
		n, col, err := paramAndCol(argsStr, expr)
		if err != nil {
			return PartitionField{}, err
		}
		return PartitionField{Transform: Bucket, Param: n, SourceCol: col}, nil

	case "truncate":
		n, col, err := paramAndCol(argsStr, expr)
		if err != nil {
			return PartitionField{}, err
		}
		return PartitionField{Transform: Truncate, Param: n, SourceCol: col}, nil

	default:
		return PartitionField{}, errors.Wrapf(ErrUnknownTransform, "transform %q is not supported", funcName)
	}
}

// singleArg extracts a single column argument from the argument string.
func singleArg(args, expr string) (string, error) {
	col := strings.TrimSpace(args)
	if col == "" {
		return "", errors.Wrapf(ErrParse, "partition transform %q: expected column name argument", expr)
	}
	return col, nil
}

// paramAndCol parses "N, col" arguments for bucket/truncate transforms.
func paramAndCol(args, expr string) (n int, col string, err error) {
	parts := strings.SplitN(args, ",", transformArgCount)
	if len(parts) != transformArgCount {
		return 0, "", errors.Wrapf(ErrParse, "partition transform %q: expected (N, col) arguments", expr)
	}
	n, err = strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return 0, "", errors.Wrapf(ErrParse, "partition transform %q: invalid N argument %q", expr, parts[0])
	}
	col = strings.TrimSpace(parts[1])
	if col == "" {
		return 0, "", errors.Wrapf(ErrParse, "partition transform %q: missing column name", expr)
	}
	return n, col, nil
}

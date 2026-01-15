package repository

import (
	"time"
)

// checkArgIsPtrAndScalar validates that the provided argument is a pointer to a scalar type.
// It returns ErrPtrValueMustBeAPointerAndScalar if the argument is not a pointer to an integer,
// unsigned integer, float, boolean, string, or time.Time.
func checkArgIsPtrAndScalar(ptr any) error {
	switch ptr.(type) {
	case *int, *int8, *int16, *int32, *int64:
		return nil
	case *uint, *uint8, *uint16, *uint32, *uint64:
		return nil
	case *float32, *float64:
		return nil
	case *bool:
		return nil
	case *string:
		return nil
	case *time.Time:
		return nil
	default:
		return ErrPtrValueMustBeAPointerAndScalar
	}
}

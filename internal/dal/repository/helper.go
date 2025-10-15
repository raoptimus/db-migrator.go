package repository

import (
	"time"
)

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

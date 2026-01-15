package validator

import (
	"strconv"
)

type TooShortError struct {
	message string
}

func NewTooShortError(minLen int) *TooShortError {
	return &TooShortError{
		message: "This value should contain at least " + strconv.Itoa(minLen) + ".",
	}
}

func (e *TooShortError) Error() string {
	return e.message
}

type TooLongError struct {
	message string
}

func NewTooLongError(maxLen int) *TooLongError {
	return &TooLongError{
		message: "This value should contain at most " + strconv.Itoa(maxLen) + ".",
	}
}

func (e *TooLongError) Error() string {
	return e.message
}

func ValidateStringLen(minLen, maxLen, value int) error {
	if value < minLen {
		return NewTooShortError(minLen)
	}
	if value > maxLen {
		return NewTooLongError(maxLen)
	}

	return nil
}

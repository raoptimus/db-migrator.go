package validator

import "github.com/pkg/errors"

const maxIdentifierLen = 256

var ErrIdentifierIsNotValid = errors.New("identifier is not valid")

func ValidateIdentifier(id string) error {
	if !isValidIdentifier(id) {
		return ErrIdentifierIsNotValid
	}

	return nil
}

func isValidIdentifier(name string) bool {
	if len(name) > maxIdentifierLen {
		return false
	}
	for _, r := range name {
		if !isValidIdentifierChar(r) {
			return false
		}
	}

	return true
}

func isValidIdentifierChar(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
		(r >= '0' && r <= '9') || r == '_'
}

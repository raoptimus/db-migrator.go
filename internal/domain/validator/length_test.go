package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateStringLen_ValidLength_Success(t *testing.T) {
	tests := []struct {
		name   string
		minLen int
		maxLen int
		value  int
	}{
		{
			name:   "value equals min",
			minLen: 1,
			maxLen: 10,
			value:  1,
		},
		{
			name:   "value equals max",
			minLen: 1,
			maxLen: 10,
			value:  10,
		},
		{
			name:   "value in middle",
			minLen: 1,
			maxLen: 10,
			value:  5,
		},
		{
			name:   "min equals max equals value",
			minLen: 5,
			maxLen: 5,
			value:  5,
		},
		{
			name:   "zero min allowed",
			minLen: 0,
			maxLen: 10,
			value:  0,
		},
		{
			name:   "large values",
			minLen: 1000,
			maxLen: 10000,
			value:  5000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStringLen(tt.minLen, tt.maxLen, tt.value)
			require.NoError(t, err)
		})
	}
}

func TestValidateStringLen_TooShort_Failure(t *testing.T) {
	tests := []struct {
		name   string
		minLen int
		maxLen int
		value  int
	}{
		{
			name:   "value less than min by 1",
			minLen: 5,
			maxLen: 10,
			value:  4,
		},
		{
			name:   "value is zero when min is 1",
			minLen: 1,
			maxLen: 10,
			value:  0,
		},
		{
			name:   "negative value",
			minLen: 0,
			maxLen: 10,
			value:  -1,
		},
		{
			name:   "significantly less than min",
			minLen: 100,
			maxLen: 200,
			value:  10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStringLen(tt.minLen, tt.maxLen, tt.value)
			require.Error(t, err)

			var tooShortErr *TooShortError
			assert.ErrorAs(t, err, &tooShortErr)
		})
	}
}

func TestValidateStringLen_TooLong_Failure(t *testing.T) {
	tests := []struct {
		name   string
		minLen int
		maxLen int
		value  int
	}{
		{
			name:   "value more than max by 1",
			minLen: 1,
			maxLen: 10,
			value:  11,
		},
		{
			name:   "significantly more than max",
			minLen: 1,
			maxLen: 10,
			value:  100,
		},
		{
			name:   "large excess",
			minLen: 1,
			maxLen: 100,
			value:  1000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStringLen(tt.minLen, tt.maxLen, tt.value)
			require.Error(t, err)

			var tooLongErr *TooLongError
			assert.ErrorAs(t, err, &tooLongErr)
		})
	}
}

func TestNewTooShortError_CreatesCorrectMessage(t *testing.T) {
	tests := []struct {
		name            string
		minLen          int
		expectedMessage string
	}{
		{
			name:            "min 1",
			minLen:          1,
			expectedMessage: "This value should contain at least 1.",
		},
		{
			name:            "min 10",
			minLen:          10,
			expectedMessage: "This value should contain at least 10.",
		},
		{
			name:            "min 100",
			minLen:          100,
			expectedMessage: "This value should contain at least 100.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewTooShortError(tt.minLen)
			assert.Equal(t, tt.expectedMessage, err.Error())
		})
	}
}

func TestNewTooLongError_CreatesCorrectMessage(t *testing.T) {
	tests := []struct {
		name            string
		maxLen          int
		expectedMessage string
	}{
		{
			name:            "max 1",
			maxLen:          1,
			expectedMessage: "This value should contain at most 1.",
		},
		{
			name:            "max 10",
			maxLen:          10,
			expectedMessage: "This value should contain at most 10.",
		},
		{
			name:            "max 100",
			maxLen:          100,
			expectedMessage: "This value should contain at most 100.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewTooLongError(tt.maxLen)
			assert.Equal(t, tt.expectedMessage, err.Error())
		})
	}
}

func TestTooShortError_ImplementsError(t *testing.T) {
	err := NewTooShortError(5)
	var _ error = err // Compile-time check that TooShortError implements error
	assert.NotEmpty(t, err.Error())
}

func TestTooLongError_ImplementsError(t *testing.T) {
	err := NewTooLongError(5)
	var _ error = err // Compile-time check that TooLongError implements error
	assert.NotEmpty(t, err.Error())
}

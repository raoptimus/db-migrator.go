package ddl

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParseTransform_Errors exercises the malformed-input branches of parseTransform
// and its argument helpers (singleArg, paramAndCol).
func TestParseTransform_Errors(t *testing.T) {
	tests := []struct {
		name string
		in   string
	}{
		{name: "empty expression", in: "   "},
		{name: "missing parentheses", in: "identity"},
		{name: "missing closing parenthesis", in: "identity(col"},
		{name: "single-arg transform without column", in: "days()"},
		{name: "bucket with one argument", in: "bucket(16)"},
		{name: "bucket with non-numeric N", in: "bucket(x, col)"},
		{name: "truncate with empty column", in: "truncate(8, )"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseTransform(tt.in)
			require.Error(t, err)
			assert.ErrorIs(t, err, ErrParse)
		})
	}
}

// TestParseTransform_AllKinds covers each supported transform's happy path.
func TestParseTransform_AllKinds(t *testing.T) {
	tests := []struct {
		in   string
		want TransformKind
	}{
		{in: "identity(id)", want: Identity},
		{in: "years(ts)", want: Years},
		{in: "months(ts)", want: Months},
		{in: "days(ts)", want: Days},
		{in: "hours(ts)", want: Hours},
		{in: "bucket(16, user_id)", want: Bucket},
		{in: "truncate(8, name)", want: Truncate},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, err := parseTransform(tt.in)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got.Transform)
		})
	}
}

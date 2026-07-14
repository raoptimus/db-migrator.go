package ddl

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParseType_CompositeErrors exercises the error branches of the composite and
// numeric type parsers (decimal, struct, array, list, map) and the bracket helpers.
func TestParseType_CompositeErrors(t *testing.T) {
	tests := []struct {
		name string
		in   string
	}{
		// decimal
		{name: "decimal missing closing paren", in: "DECIMAL("},
		{name: "decimal non-numeric precision", in: "DECIMAL(x)"},
		{name: "decimal non-numeric scale", in: "DECIMAL(10,y)"},
		// struct
		{name: "struct field without colon", in: "STRUCT<foo>"},
		{name: "struct field unknown type", in: "STRUCT<a:bogus>"},
		{name: "struct unmatched angle bracket", in: "STRUCT<a:int"},
		{name: "struct unmatched closing bracket in field", in: "STRUCT<a:int)>"},
		// array
		{name: "array unknown element type", in: "ARRAY<bogus>"},
		{name: "array unmatched angle bracket", in: "ARRAY<int"},
		// list
		{name: "list unknown element type", in: "LIST<bogus>"},
		{name: "list unmatched angle bracket", in: "LIST<int"},
		// map
		{name: "map single argument", in: "MAP<int>"},
		{name: "map unknown key type", in: "MAP<bogus,int>"},
		{name: "map unknown value type", in: "MAP<int,bogus>"},
		{name: "map unmatched angle bracket", in: "MAP<int,int"},
		// primitive fallthrough
		{name: "unknown primitive", in: "totally_unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseType(tt.in)
			require.Error(t, err)
			assert.ErrorIs(t, err, ErrParse)
		})
	}
}

// TestParseType_CompositeSuccess covers the nested composite happy paths so the
// recursive descent through struct/array/map is exercised end to end.
func TestParseType_NestedCompositeSuccess(t *testing.T) {
	tests := []struct {
		name string
		in   string
		kind TypeKind
	}{
		{name: "decimal default precision", in: "DECIMAL", kind: Decimal},
		{name: "decimal precision only", in: "DECIMAL(20)", kind: Decimal},
		{name: "nested struct in map value", in: "MAP<string,STRUCT<a:int,b:long>>", kind: Map},
		{name: "array of struct", in: "ARRAY<STRUCT<id:string>>", kind: List},
		{name: "list of map", in: "LIST<MAP<string,int>>", kind: List},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseType(tt.in)
			require.NoError(t, err)
			assert.Equal(t, tt.kind, got.Kind)
		})
	}
}

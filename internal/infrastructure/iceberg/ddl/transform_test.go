package ddl

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParseTransform covers all 7 supported Iceberg partition transforms.
func TestParseTransform(t *testing.T) {
	t.Parallel()

	tests := []struct {
		expr          string
		wantTransform TransformKind
		wantParam     int
		wantSourceCol string
	}{
		{
			expr:          "identity(id)",
			wantTransform: Identity,
			wantSourceCol: "id",
		},
		{
			expr:          "years(ts)",
			wantTransform: Years,
			wantSourceCol: "ts",
		},
		{
			expr:          "months(ts)",
			wantTransform: Months,
			wantSourceCol: "ts",
		},
		{
			expr:          "days(ts)",
			wantTransform: Days,
			wantSourceCol: "ts",
		},
		{
			expr:          "hours(ts)",
			wantTransform: Hours,
			wantSourceCol: "ts",
		},
		{
			// bucket(16, id)
			expr:          "bucket(16, id)",
			wantTransform: Bucket,
			wantParam:     16,
			wantSourceCol: "id",
		},
		{
			// truncate(8, name)
			expr:          "truncate(8, name)",
			wantTransform: Truncate,
			wantParam:     8,
			wantSourceCol: "name",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.expr, func(t *testing.T) {
			t.Parallel()
			pf, err := parseTransform(tt.expr)
			require.NoError(t, err)
			assert.Equal(t, tt.wantTransform, pf.Transform)
			assert.Equal(t, tt.wantSourceCol, pf.SourceCol)
			if tt.wantParam != 0 {
				assert.Equal(t, tt.wantParam, pf.Param)
			}
		})
	}
}

// TestParseTransform_UnknownTransform verifies that unsupported transforms return ErrUnknownTransform.
func TestParseTransform_UnknownTransform(t *testing.T) {
	t.Parallel()

	// weeks(ts) is not a supported Iceberg transform → ErrUnknownTransform
	_, err := parseTransform("weeks(ts)")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrUnknownTransform),
		"expected errors.Is(ErrUnknownTransform) but got: %v", err)
}

// TestParse_Transforms_InCreateTable exercises transforms through the full Parse path.
func TestParse_Transforms_InCreateTable(t *testing.T) {
	t.Parallel()

	// "CREATE TABLE raw.t (id long, ts timestamp, name string) USING iceberg PARTITIONED BY (<transform>)"
	tests := []struct {
		transform     string
		wantTransform TransformKind
		wantParam     int
		wantSourceCol string
	}{
		{"identity(id)", Identity, 0, "id"},
		{"years(ts)", Years, 0, "ts"},
		{"months(ts)", Months, 0, "ts"},
		{"days(ts)", Days, 0, "ts"},
		{"hours(ts)", Hours, 0, "ts"},
		{"bucket(16, id)", Bucket, 16, "id"},
		{"truncate(8, name)", Truncate, 8, "name"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.transform, func(t *testing.T) {
			t.Parallel()
			stmt := "CREATE TABLE raw.t (id long, ts timestamp, name string) USING iceberg PARTITIONED BY (" + tt.transform + ")"
			op, err := Parse("", stmt)
			require.NoError(t, err)
			require.Len(t, op.Create.Partition, 1)
			pf := op.Create.Partition[0]
			assert.Equal(t, tt.wantTransform, pf.Transform)
			assert.Equal(t, tt.wantSourceCol, pf.SourceCol)
			if tt.wantParam != 0 {
				assert.Equal(t, tt.wantParam, pf.Param)
			}
		})
	}
}

// TestParse_UnknownTransform_InCreateTable verifies ErrUnknownTransform propagates from Parse.
func TestParse_UnknownTransform_InCreateTable(t *testing.T) {
	t.Parallel()

	// weeks(ts) is not supported — must propagate ErrUnknownTransform.
	stmt := "CREATE TABLE raw.t (ts timestamp) USING iceberg PARTITIONED BY (weeks(ts))"
	_, err := Parse("", stmt)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrUnknownTransform),
		"expected errors.Is(ErrUnknownTransform) but got: %v", err)
}

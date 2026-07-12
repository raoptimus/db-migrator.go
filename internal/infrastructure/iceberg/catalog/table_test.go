/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package catalog

// Unit tests for the IR → iceberg-go mapping helpers.
// These tests do NOT make network calls; they only verify that buildSchema,
// buildPartitionSpec and toIcebergType produce correct iceberg-go objects.

import (
	"testing"

	iceberg "github.com/apache/iceberg-go"
	"github.com/raoptimus/db-migrator.go/internal/infrastructure/iceberg/ddl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// collectPartitionFields collects all PartitionField values from the spec's
// iter.Seq2 iterator into a slice for easy assertion in tests.
func collectPartitionFields(spec *iceberg.PartitionSpec) []iceberg.PartitionField {
	var result []iceberg.PartitionField
	for _, f := range spec.Fields() {
		result = append(result, f)
	}
	return result
}

// ─── toIcebergType ──────────────────────────────────────────────────────────

func TestToIcebergType_Primitives(t *testing.T) {
	ctr := &idCounter{}
	tests := []struct {
		kind     ddl.TypeKind
		wantType iceberg.Type
	}{
		{ddl.Boolean, iceberg.BooleanType{}},
		{ddl.Int, iceberg.Int32Type{}},
		{ddl.Long, iceberg.Int64Type{}},
		{ddl.Float, iceberg.Float32Type{}},
		{ddl.Double, iceberg.Float64Type{}},
		{ddl.Date, iceberg.DateType{}},
		{ddl.Time, iceberg.TimeType{}},
		{ddl.TimestampTz, iceberg.TimestampTzType{}},
		{ddl.Timestamp, iceberg.TimestampType{}},
		{ddl.String, iceberg.StringType{}},
		{ddl.UUID, iceberg.UUIDType{}},
		{ddl.Binary, iceberg.BinaryType{}},
	}
	for _, tt := range tests {
		t.Run(tt.wantType.Type(), func(t *testing.T) {
			got, err := toIcebergType(ddl.IcebergType{Kind: tt.kind}, ctr)
			require.NoError(t, err)
			assert.Equal(t, tt.wantType, got)
		})
	}
}

func TestToIcebergType_Decimal(t *testing.T) {
	ctr := &idCounter{}
	got, err := toIcebergType(ddl.IcebergType{Kind: ddl.Decimal, Prec: 18, Scale: 4}, ctr)
	require.NoError(t, err)
	dt, ok := got.(iceberg.DecimalType)
	require.True(t, ok, "expected DecimalType, got %T", got)
	assert.Equal(t, 18, dt.Precision())
	assert.Equal(t, 4, dt.Scale())
}

func TestToIcebergType_Struct(t *testing.T) {
	ctr := &idCounter{}
	structType := ddl.IcebergType{
		Kind: ddl.Struct,
		Fields: []ddl.Field{
			{Name: "x", Type: ddl.IcebergType{Kind: ddl.Int}, Required: true, Doc: "x coord"},
			{Name: "y", Type: ddl.IcebergType{Kind: ddl.Int}},
		},
	}
	got, err := toIcebergType(structType, ctr)
	require.NoError(t, err)
	st, ok := got.(*iceberg.StructType)
	require.True(t, ok)
	require.Len(t, st.FieldList, 2)
	assert.Equal(t, "x", st.FieldList[0].Name)
	assert.True(t, st.FieldList[0].Required)
	assert.Equal(t, "x coord", st.FieldList[0].Doc)
	assert.Equal(t, "y", st.FieldList[1].Name)
	assert.False(t, st.FieldList[1].Required)
}

func TestToIcebergType_List(t *testing.T) {
	ctr := &idCounter{}
	elem := ddl.IcebergType{Kind: ddl.String}
	listType := ddl.IcebergType{Kind: ddl.List, Elem: &elem}
	got, err := toIcebergType(listType, ctr)
	require.NoError(t, err)
	lt, ok := got.(*iceberg.ListType)
	require.True(t, ok)
	assert.Equal(t, iceberg.StringType{}, lt.Element)
	assert.Greater(t, lt.ElementID, 0)
}

func TestToIcebergType_Map(t *testing.T) {
	ctr := &idCounter{}
	keyT := ddl.IcebergType{Kind: ddl.String}
	valT := ddl.IcebergType{Kind: ddl.Long}
	mapType := ddl.IcebergType{Kind: ddl.Map, Key: &keyT, Val: &valT}
	got, err := toIcebergType(mapType, ctr)
	require.NoError(t, err)
	mt, ok := got.(*iceberg.MapType)
	require.True(t, ok)
	assert.Equal(t, iceberg.StringType{}, mt.KeyType)
	assert.Equal(t, iceberg.Int64Type{}, mt.ValueType)
	assert.Greater(t, mt.KeyID, 0)
	assert.Greater(t, mt.ValueID, 0)
	assert.NotEqual(t, mt.KeyID, mt.ValueID)
}

func TestToIcebergType_ListMissingElem(t *testing.T) {
	ctr := &idCounter{}
	_, err := toIcebergType(ddl.IcebergType{Kind: ddl.List}, ctr)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no element type")
}

func TestToIcebergType_MapMissingKey(t *testing.T) {
	ctr := &idCounter{}
	val := ddl.IcebergType{Kind: ddl.String}
	_, err := toIcebergType(ddl.IcebergType{Kind: ddl.Map, Val: &val}, ctr)
	require.Error(t, err)
}

// ─── buildSchema ────────────────────────────────────────────────────────────

func TestBuildSchema_Basic(t *testing.T) {
	fields := []ddl.Field{
		{Name: "id", Type: ddl.IcebergType{Kind: ddl.Long}, Required: true},
		{Name: "event_time", Type: ddl.IcebergType{Kind: ddl.TimestampTz}, Doc: "when it happened"},
		{Name: "payload", Type: ddl.IcebergType{Kind: ddl.String}},
	}
	schema, fieldIDs, err := buildSchema(fields)
	require.NoError(t, err)
	require.NotNil(t, schema)
	require.NotNil(t, fieldIDs)
	// IDs are sequential starting from 1
	require.Len(t, schema.Fields(), 3)
	assert.Equal(t, 1, schema.Fields()[0].ID)
	assert.Equal(t, "id", schema.Fields()[0].Name)
	assert.True(t, schema.Fields()[0].Required)
	assert.Equal(t, 2, schema.Fields()[1].ID)
	assert.Equal(t, "event_time", schema.Fields()[1].Name)
	assert.Equal(t, "when it happened", schema.Fields()[1].Doc)
	assert.Equal(t, iceberg.TimestampTzType{}, schema.Fields()[1].Type)
	assert.Equal(t, 3, schema.Fields()[2].ID)
	// fieldIDs map must match schema field IDs
	assert.Equal(t, 1, fieldIDs["id"])
	assert.Equal(t, 2, fieldIDs["event_time"])
	assert.Equal(t, 3, fieldIDs["payload"])
}

// ─── buildPartitionSpec ──────────────────────────────────────────────────────

func TestBuildPartitionSpec_Days(t *testing.T) {
	schemaFields := []ddl.Field{
		{Name: "id", Type: ddl.IcebergType{Kind: ddl.Long}},
		{Name: "event_time", Type: ddl.IcebergType{Kind: ddl.TimestampTz}},
	}
	_, fieldIDs, err := buildSchema(schemaFields)
	require.NoError(t, err)
	partitions := []ddl.PartitionField{
		{Transform: ddl.Days, SourceCol: "event_time", Name: "event_day"},
	}
	spec, err := buildPartitionSpec(fieldIDs, partitions)
	require.NoError(t, err)
	require.NotNil(t, spec)
	pfields := collectPartitionFields(spec)
	require.Len(t, pfields, 1)
	assert.Equal(t, "event_day", pfields[0].Name)
	assert.Equal(t, iceberg.DayTransform{}, pfields[0].Transform)
	// SourceID == 2 (second field in schema, 1-based)
	assert.Equal(t, 2, pfields[0].SourceID())
}

func TestBuildPartitionSpec_Bucket(t *testing.T) {
	schemaFields := []ddl.Field{
		{Name: "user_id", Type: ddl.IcebergType{Kind: ddl.Long}},
	}
	_, fieldIDs, err := buildSchema(schemaFields)
	require.NoError(t, err)
	partitions := []ddl.PartitionField{
		{Transform: ddl.Bucket, Param: 16, SourceCol: "user_id"},
	}
	spec, err := buildPartitionSpec(fieldIDs, partitions)
	require.NoError(t, err)
	pfields := collectPartitionFields(spec)
	require.Len(t, pfields, 1)
	bt, ok := pfields[0].Transform.(iceberg.BucketTransform)
	require.True(t, ok)
	assert.Equal(t, 16, bt.NumBuckets)
}

func TestBuildPartitionSpec_Truncate(t *testing.T) {
	schemaFields := []ddl.Field{
		{Name: "name", Type: ddl.IcebergType{Kind: ddl.String}},
	}
	_, fieldIDs, err := buildSchema(schemaFields)
	require.NoError(t, err)
	partitions := []ddl.PartitionField{
		{Transform: ddl.Truncate, Param: 10, SourceCol: "name"},
	}
	spec, err := buildPartitionSpec(fieldIDs, partitions)
	require.NoError(t, err)
	pfields := collectPartitionFields(spec)
	require.Len(t, pfields, 1)
	tt, ok := pfields[0].Transform.(iceberg.TruncateTransform)
	require.True(t, ok)
	assert.Equal(t, 10, tt.Width)
}

func TestBuildPartitionSpec_UnknownColumn(t *testing.T) {
	schemaFields := []ddl.Field{
		{Name: "id", Type: ddl.IcebergType{Kind: ddl.Long}},
	}
	_, fieldIDs, err := buildSchema(schemaFields)
	require.NoError(t, err)
	partitions := []ddl.PartitionField{
		{Transform: ddl.Days, SourceCol: "nonexistent"},
	}
	_, err = buildPartitionSpec(fieldIDs, partitions)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nonexistent")
}

func TestBuildPartitionSpec_AllTransforms(t *testing.T) {
	schemaFields := []ddl.Field{
		{Name: "ts", Type: ddl.IcebergType{Kind: ddl.TimestampTz}},
	}
	_, fieldIDs, err := buildSchema(schemaFields)
	require.NoError(t, err)
	transforms := []struct {
		kind ddl.TransformKind
		want iceberg.Transform
	}{
		{ddl.Identity, iceberg.IdentityTransform{}},
		{ddl.Years, iceberg.YearTransform{}},
		{ddl.Months, iceberg.MonthTransform{}},
		{ddl.Days, iceberg.DayTransform{}},
		{ddl.Hours, iceberg.HourTransform{}},
	}
	for _, tt := range transforms {
		t.Run(tt.want.String(), func(t *testing.T) {
			spec, err := buildPartitionSpec(fieldIDs, []ddl.PartitionField{
				{Transform: tt.kind, SourceCol: "ts"},
			})
			require.NoError(t, err)
			pfields := collectPartitionFields(spec)
			require.Len(t, pfields, 1)
			assert.Equal(t, tt.want, pfields[0].Transform)
		})
	}
}

// TestBuildPartitionSpec_NestedBeforeScalar verifies that when a nested type
// (struct, list, map) precedes a scalar partition column, the SourceID in the
// partition spec matches the real field ID assigned by buildSchema — NOT a
// positional i+1 value.
//
// Schema: a struct<x:int>, b timestamp
// buildSchema IDs:  x=1, a=2, b=3
// Partition: days(b) → SourceID must be 3, not 2 (which would be wrong positional value).
func TestBuildPartitionSpec_NestedBeforeScalar(t *testing.T) {
	schemaFields := []ddl.Field{
		{
			Name: "a",
			Type: ddl.IcebergType{
				Kind: ddl.Struct,
				Fields: []ddl.Field{
					{Name: "x", Type: ddl.IcebergType{Kind: ddl.Int}},
				},
			},
		},
		{Name: "b", Type: ddl.IcebergType{Kind: ddl.Timestamp}},
	}

	schema, fieldIDs, err := buildSchema(schemaFields)
	require.NoError(t, err)

	// Verify schema field IDs: x=1, a=2, b=3
	schFields := schema.Fields()
	require.Len(t, schFields, 2)
	aField := schFields[0]
	bField := schFields[1]
	assert.Equal(t, "a", aField.Name)
	assert.Equal(t, "b", bField.Name)
	// b's ID must be 3 (not 2), because the nested field x consumed ID 1.
	assert.Equal(t, 3, bField.ID, "b must have field ID 3 (x=1, a=2, b=3)")

	// fieldIDs map must reflect actual IDs
	assert.Equal(t, 2, fieldIDs["a"])
	assert.Equal(t, 3, fieldIDs["b"])

	// Partition by days(b): SourceID must equal b's real field ID (3).
	spec, err := buildPartitionSpec(fieldIDs, []ddl.PartitionField{
		{Transform: ddl.Days, SourceCol: "b", Name: "b_day"},
	})
	require.NoError(t, err)
	pfields := collectPartitionFields(spec)
	require.Len(t, pfields, 1)
	assert.Equal(t, "b_day", pfields[0].Name)
	assert.Equal(t, iceberg.DayTransform{}, pfields[0].Transform)
	assert.Equal(t, bField.ID, pfields[0].SourceID(),
		"SourceID must match b's actual schema field ID, not positional i+1")
}

// TestBuildPartitionSpec_MultipleNestedBeforeScalar verifies correct SourceID
// when multiple nested types precede the partitioned column.
//
// Schema: m map<string,long>, l list<int>, ts timestamp
// m occupies IDs for key and value, l for element, then top-level fields get subsequent IDs.
func TestBuildPartitionSpec_MultipleNestedBeforeScalar(t *testing.T) {
	keyT := ddl.IcebergType{Kind: ddl.String}
	valT := ddl.IcebergType{Kind: ddl.Long}
	elemT := ddl.IcebergType{Kind: ddl.Int}

	schemaFields := []ddl.Field{
		{Name: "m", Type: ddl.IcebergType{Kind: ddl.Map, Key: &keyT, Val: &valT}},
		{Name: "l", Type: ddl.IcebergType{Kind: ddl.List, Elem: &elemT}},
		{Name: "ts", Type: ddl.IcebergType{Kind: ddl.TimestampTz}},
	}

	schema, fieldIDs, err := buildSchema(schemaFields)
	require.NoError(t, err)

	schFields := schema.Fields()
	require.Len(t, schFields, 3)

	// Find ts field and its real ID
	var tsFieldID int
	for _, f := range schFields {
		if f.Name == "ts" {
			tsFieldID = f.ID
		}
	}
	require.Greater(t, tsFieldID, 3,
		"ts field ID must be > 3 because map key/value and list element consume IDs before it")

	// fieldIDs map must return the real ID for ts
	assert.Equal(t, tsFieldID, fieldIDs["ts"])

	// Partition by hours(ts): SourceID must equal ts's real field ID.
	spec, err := buildPartitionSpec(fieldIDs, []ddl.PartitionField{
		{Transform: ddl.Hours, SourceCol: "ts", Name: "ts_hour"},
	})
	require.NoError(t, err)
	pfields := collectPartitionFields(spec)
	require.Len(t, pfields, 1)
	assert.Equal(t, tsFieldID, pfields[0].SourceID(),
		"SourceID must match ts's actual schema field ID")
}

// ─── ident helper ───────────────────────────────────────────────────────────

func TestIdent_NSAndTable(t *testing.T) {
	id := ddl.Ident{Namespace: []string{"analytics", "sub"}, Table: "events"}
	got := ident(id)
	assert.Equal(t, []string{"analytics", "sub", "events"}, []string(got))
}

func TestIdent_NSOnly(t *testing.T) {
	id := ddl.Ident{Namespace: []string{"analytics"}, Table: ""}
	got := ident(id)
	assert.Equal(t, []string{"analytics"}, []string(got))
}

// ─── findPartitionFieldName ──────────────────────────────────────────────────

// buildTestSpec is a helper that constructs an iceberg.PartitionSpec from
// raw iceberg.PartitionField values without schema validation.
func buildTestSpec(fields ...iceberg.PartitionField) iceberg.PartitionSpec {
	return iceberg.NewPartitionSpec(fields...)
}

// buildTestSchema creates an *iceberg.Schema with the given top-level fields
// (using sequential IDs from 1).
func buildTestSchema(t *testing.T, ddlFields []ddl.Field) *iceberg.Schema {
	t.Helper()
	schema, _, err := buildSchema(ddlFields)
	require.NoError(t, err)
	return schema
}

// TestFindPartitionFieldName_BucketByConvention verifies that the function
// resolves the partition field name for bucket(16, gid) when the partition
// field was created with the ADD PARTITION FIELD convention
// (name = "<col>_bucket[16]", e.g. "gid_bucket[16]") and op.Name is empty.
//
// This is the exact scenario that caused the reported bug:
// ADD: us.AddField("gid", BucketTransform{16}, "gid_bucket[16]")
// DROP: us.RemoveField("gid")  ← wrong; must be "gid_bucket[16]"
func TestFindPartitionFieldName_BucketByConvention(t *testing.T) {
	schema := buildTestSchema(t, []ddl.Field{
		{Name: "id", Type: ddl.IcebergType{Kind: ddl.Long}, Required: true},
		{Name: "gid", Type: ddl.IcebergType{Kind: ddl.String}},
	})

	// gid has field ID 2 (id=1, gid=2).
	gidField, ok := schema.FindFieldByName("gid")
	require.True(t, ok)

	spec := buildTestSpec(iceberg.PartitionField{
		SourceIDs: []int{gidField.ID},
		FieldID:   partitionFieldIDBase,
		Name:      "gid_bucket[16]",
		Transform: iceberg.BucketTransform{NumBuckets: 16},
	})

	op := &ddl.PartitionField{
		Transform: ddl.Bucket,
		Param:     16,
		SourceCol: "gid",
		Name:      "", // no explicit name — must be resolved from spec
	}

	got, err := findPartitionFieldName(schema, spec, op)
	require.NoError(t, err)
	assert.Equal(t, "gid_bucket[16]", got,
		"must return the actual partition field name, not the source column name")
}

// TestFindPartitionFieldName_IdentityByConvention verifies the identity
// transform case where the partition field name typically equals the source
// column name (e.g. created via ADD IDENTITY / CREATE TABLE PARTITIONED BY (col)).
func TestFindPartitionFieldName_IdentityByConvention(t *testing.T) {
	schema := buildTestSchema(t, []ddl.Field{
		{Name: "region", Type: ddl.IcebergType{Kind: ddl.String}},
	})

	regionField, ok := schema.FindFieldByName("region")
	require.True(t, ok)

	spec := buildTestSpec(iceberg.PartitionField{
		SourceIDs: []int{regionField.ID},
		FieldID:   partitionFieldIDBase,
		Name:      "region",
		Transform: iceberg.IdentityTransform{},
	})

	op := &ddl.PartitionField{
		Transform: ddl.Identity,
		SourceCol: "region",
		Name:      "",
	}

	got, err := findPartitionFieldName(schema, spec, op)
	require.NoError(t, err)
	assert.Equal(t, "region", got)
}

// TestFindPartitionFieldName_DaysTransform verifies days(event_time) resolution.
func TestFindPartitionFieldName_DaysTransform(t *testing.T) {
	schema := buildTestSchema(t, []ddl.Field{
		{Name: "event_time", Type: ddl.IcebergType{Kind: ddl.TimestampTz}},
	})

	etField, ok := schema.FindFieldByName("event_time")
	require.True(t, ok)

	spec := buildTestSpec(iceberg.PartitionField{
		SourceIDs: []int{etField.ID},
		FieldID:   partitionFieldIDBase,
		Name:      "event_time_day",
		Transform: iceberg.DayTransform{},
	})

	op := &ddl.PartitionField{
		Transform: ddl.Days,
		SourceCol: "event_time",
		Name:      "",
	}

	got, err := findPartitionFieldName(schema, spec, op)
	require.NoError(t, err)
	assert.Equal(t, "event_time_day", got)
}

// TestFindPartitionFieldName_ExplicitNameOverride verifies that when op.Name
// is set, the function returns it directly without looking up the spec.
func TestFindPartitionFieldName_ExplicitNameOverride(t *testing.T) {
	schema := buildTestSchema(t, []ddl.Field{
		{Name: "col", Type: ddl.IcebergType{Kind: ddl.Long}},
	})
	// spec is empty — the function should not need to consult it.
	spec := buildTestSpec()

	op := &ddl.PartitionField{
		Transform: ddl.Bucket,
		Param:     8,
		SourceCol: "col",
		Name:      "my_custom_partition_name",
	}

	got, err := findPartitionFieldName(schema, spec, op)
	require.NoError(t, err)
	assert.Equal(t, "my_custom_partition_name", got)
}

// TestFindPartitionFieldName_SourceColNotInSchema verifies a clear error when
// the source column does not exist in the schema.
func TestFindPartitionFieldName_SourceColNotInSchema(t *testing.T) {
	schema := buildTestSchema(t, []ddl.Field{
		{Name: "id", Type: ddl.IcebergType{Kind: ddl.Long}},
	})
	spec := buildTestSpec()

	op := &ddl.PartitionField{
		Transform: ddl.Bucket,
		Param:     16,
		SourceCol: "nonexistent_col",
		Name:      "",
	}

	_, err := findPartitionFieldName(schema, spec, op)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nonexistent_col")
}

// TestFindPartitionFieldName_PartitionFieldNotInSpec verifies a clear error when
// the source column exists in the schema but no matching partition field is
// found in the current spec (e.g. wrong transform or field not yet added).
func TestFindPartitionFieldName_PartitionFieldNotInSpec(t *testing.T) {
	schema := buildTestSchema(t, []ddl.Field{
		{Name: "ts", Type: ddl.IcebergType{Kind: ddl.TimestampTz}},
	})

	tsField, ok := schema.FindFieldByName("ts")
	require.True(t, ok)

	// Spec has days(ts), but op asks for hours(ts).
	spec := buildTestSpec(iceberg.PartitionField{
		SourceIDs: []int{tsField.ID},
		FieldID:   partitionFieldIDBase,
		Name:      "ts_day",
		Transform: iceberg.DayTransform{},
	})

	op := &ddl.PartitionField{
		Transform: ddl.Hours,
		SourceCol: "ts",
		Name:      "",
	}

	_, err := findPartitionFieldName(schema, spec, op)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found in current spec")
}

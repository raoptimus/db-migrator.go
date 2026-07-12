package ddl

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParseType covers all 13 types from @ФТ-3 Примеры table.
func TestParseType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input     string
		wantKind  TypeKind
		wantPrec  int
		wantScale int
		wantElem  *TypeKind
		wantKey   *TypeKind
		wantVal   *TypeKind
		wantFields []Field // for struct
	}{
		{input: "boolean", wantKind: Boolean},
		{input: "BOOLEAN", wantKind: Boolean},
		{input: "int", wantKind: Int},
		{input: "INT", wantKind: Int},
		{input: "long", wantKind: Long},
		{input: "LONG", wantKind: Long},
		{input: "bigint", wantKind: Long},
		{input: "BIGINT", wantKind: Long},
		{input: "float", wantKind: Float},
		{input: "FLOAT", wantKind: Float},
		{input: "double", wantKind: Double},
		{input: "DOUBLE", wantKind: Double},
		{
			input: "decimal(10,2)", wantKind: Decimal,
			wantPrec: 10, wantScale: 2,
		},
		{
			input: "DECIMAL(10,2)", wantKind: Decimal,
			wantPrec: 10, wantScale: 2,
		},
		{input: "date", wantKind: Date},
		{input: "DATE", wantKind: Date},
		{input: "time", wantKind: Time},
		// TIMESTAMP → TimestampTz (with zone) — critical ФТ-12 mapping
		{input: "timestamp", wantKind: TimestampTz},
		{input: "TIMESTAMP", wantKind: TimestampTz},
		// TIMESTAMP_NTZ → Timestamp (without zone) — critical ФТ-12 mapping
		{input: "timestamp_ntz", wantKind: Timestamp},
		{input: "TIMESTAMP_NTZ", wantKind: Timestamp},
		{input: "string", wantKind: String},
		{input: "STRING", wantKind: String},
		{input: "String", wantKind: String},
		{input: "uuid", wantKind: UUID},
		{input: "UUID", wantKind: UUID},
		{input: "binary", wantKind: Binary},
		{input: "BINARY", wantKind: Binary},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			got, err := parseType(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.wantKind, got.Kind)
			if tt.wantKind == Decimal {
				assert.Equal(t, tt.wantPrec, got.Prec)
				assert.Equal(t, tt.wantScale, got.Scale)
			}
		})
	}
}

// TestParseType_Composite covers struct, array/list, map with recursive types.
func TestParseType_Composite(t *testing.T) {
	t.Parallel()

	t.Run("struct<a:int,b:string>", func(t *testing.T) {
		t.Parallel()
		got, err := parseType("struct<a:int,b:string>")
		require.NoError(t, err)
		assert.Equal(t, Struct, got.Kind)
		require.Len(t, got.Fields, 2)
		assert.Equal(t, "a", got.Fields[0].Name)
		assert.Equal(t, Int, got.Fields[0].Type.Kind)
		assert.Equal(t, "b", got.Fields[1].Name)
		assert.Equal(t, String, got.Fields[1].Type.Kind)
	})

	t.Run("STRUCT<a:INT,b:STRING>", func(t *testing.T) {
		t.Parallel()
		got, err := parseType("STRUCT<a:INT,b:STRING>")
		require.NoError(t, err)
		assert.Equal(t, Struct, got.Kind)
		require.Len(t, got.Fields, 2)
	})

	t.Run("array<string>", func(t *testing.T) {
		t.Parallel()
		got, err := parseType("array<string>")
		require.NoError(t, err)
		assert.Equal(t, List, got.Kind)
		require.NotNil(t, got.Elem)
		assert.Equal(t, String, got.Elem.Kind)
	})

	// BDD @ФТ-3 type table row: "list<string>" — list<T> is a synonym for array<T>.
	t.Run("list<string>", func(t *testing.T) {
		t.Parallel()
		got, err := parseType("list<string>")
		require.NoError(t, err)
		assert.Equal(t, List, got.Kind)
		require.NotNil(t, got.Elem)
		assert.Equal(t, String, got.Elem.Kind)
	})

	t.Run("list<map<string,long>> (nested)", func(t *testing.T) {
		t.Parallel()
		got, err := parseType("list<map<string,long>>")
		require.NoError(t, err)
		assert.Equal(t, List, got.Kind)
		require.NotNil(t, got.Elem)
		assert.Equal(t, Map, got.Elem.Kind)
		require.NotNil(t, got.Elem.Key)
		assert.Equal(t, String, got.Elem.Key.Kind)
		require.NotNil(t, got.Elem.Val)
		assert.Equal(t, Long, got.Elem.Val.Kind)
	})

	t.Run("map<string,long>", func(t *testing.T) {
		t.Parallel()
		got, err := parseType("map<string,long>")
		require.NoError(t, err)
		assert.Equal(t, Map, got.Kind)
		require.NotNil(t, got.Key)
		require.NotNil(t, got.Val)
		assert.Equal(t, String, got.Key.Kind)
		assert.Equal(t, Long, got.Val.Kind)
	})

	t.Run("MAP<STRING,LONG>", func(t *testing.T) {
		t.Parallel()
		got, err := parseType("MAP<STRING,LONG>")
		require.NoError(t, err)
		assert.Equal(t, Map, got.Kind)
		assert.Equal(t, String, got.Key.Kind)
		assert.Equal(t, Long, got.Val.Kind)
	})

	t.Run("nested: array<struct<a:int>>", func(t *testing.T) {
		t.Parallel()
		got, err := parseType("array<struct<a:int>>")
		require.NoError(t, err)
		assert.Equal(t, List, got.Kind)
		require.NotNil(t, got.Elem)
		assert.Equal(t, Struct, got.Elem.Kind)
		require.Len(t, got.Elem.Fields, 1)
		assert.Equal(t, Int, got.Elem.Fields[0].Type.Kind)
	})
}

// TestParseType_CreateTable_ColumnTypes exercises types through the full Parse path (BDD examples).
func TestParseType_CreateTable_ColumnTypes(t *testing.T) {
	t.Parallel()

	// From BDD @ФТ-3: "Поддерживаемые типы колонок" table (13 rows).
	// Each row is "CREATE TABLE analytics.t (c <type>)" and the type must map correctly.
	tests := []struct {
		typeExpr string
		wantKind TypeKind
	}{
		{"boolean", Boolean},
		{"int", Int},
		{"long", Long},
		{"double", Double},
		{"decimal(10,2)", Decimal},
		{"date", Date},
		{"timestamp", TimestampTz}, // BDD row says "timestamp" → timestamptz (with zone)
		{"string", String},
		{"uuid", UUID},
		{"binary", Binary},
		{"map<string,long>", Map},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.typeExpr, func(t *testing.T) {
			t.Parallel()
			stmt := "CREATE TABLE analytics.t (c " + tt.typeExpr + ")"
			op, err := Parse("", stmt)
			require.NoError(t, err)
			require.Len(t, op.Create.Schema, 1)
			assert.Equal(t, tt.wantKind, op.Create.Schema[0].Type.Kind)
		})
	}

	// Struct and list need special treatment due to tokenizer handling angle brackets.
	t.Run("struct<a:int,b:string>", func(t *testing.T) {
		t.Parallel()
		op, err := Parse("", "CREATE TABLE analytics.t (c struct<a:int,b:string>)")
		require.NoError(t, err)
		require.Len(t, op.Create.Schema, 1)
		assert.Equal(t, Struct, op.Create.Schema[0].Type.Kind)
		require.Len(t, op.Create.Schema[0].Type.Fields, 2)
	})

	t.Run("array<string>", func(t *testing.T) {
		t.Parallel()
		op, err := Parse("", "CREATE TABLE analytics.t (c array<string>)")
		require.NoError(t, err)
		require.Len(t, op.Create.Schema, 1)
		assert.Equal(t, List, op.Create.Schema[0].Type.Kind)
	})

	// BDD @ФТ-3: literal "list<string>" in type column → must parse to List kind.
	t.Run("list<string>", func(t *testing.T) {
		t.Parallel()
		op, err := Parse("", "CREATE TABLE analytics.t (c list<string>)")
		require.NoError(t, err)
		require.Len(t, op.Create.Schema, 1)
		assert.Equal(t, List, op.Create.Schema[0].Type.Kind)
	})
}

// TestParseType_Errors covers unknown type.
func TestParseType_Errors(t *testing.T) {
	t.Parallel()

	_, err := parseType("NVARCHAR(100)")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrParse))
}

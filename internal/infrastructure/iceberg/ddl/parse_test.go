package ddl

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ─── DDL operations ────────────────────────────────────────────────────────────

func TestParse_DDLOperations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		catalog  string
		stmt     string
		wantKind OpKind
		check    func(t *testing.T, op Operation)
	}{
		{
			name:     "CREATE NAMESPACE",
			stmt:     "CREATE NAMESPACE analytics",
			wantKind: CreateNamespace,
			check: func(t *testing.T, op Operation) {
				assert.Equal(t, []string{"analytics"}, op.Table.Namespace)
				assert.Empty(t, op.Table.Table)
			},
		},
		{
			name:     "DROP NAMESPACE",
			stmt:     "DROP NAMESPACE analytics",
			wantKind: DropNamespace,
			check: func(t *testing.T, op Operation) {
				assert.Equal(t, []string{"analytics"}, op.Table.Namespace)
			},
		},
		{
			name:     "CREATE TABLE",
			stmt:     "CREATE TABLE analytics.events (id long)",
			wantKind: CreateTable,
			check: func(t *testing.T, op Operation) {
				assert.Equal(t, []string{"analytics"}, op.Table.Namespace)
				assert.Equal(t, "events", op.Table.Table)
				require.NotNil(t, op.Create)
				require.Len(t, op.Create.Schema, 1)
				assert.Equal(t, "id", op.Create.Schema[0].Name)
				assert.Equal(t, Long, op.Create.Schema[0].Type.Kind)
			},
		},
		{
			name:     "DROP TABLE",
			stmt:     "DROP TABLE analytics.events",
			wantKind: DropTable,
			check: func(t *testing.T, op Operation) {
				assert.Equal(t, []string{"analytics"}, op.Table.Namespace)
				assert.Equal(t, "events", op.Table.Table)
			},
		},
		{
			name:     "RENAME TABLE",
			stmt:     "RENAME TABLE analytics.events TO analytics.events_v2",
			wantKind: RenameTable,
			check: func(t *testing.T, op Operation) {
				assert.Equal(t, []string{"analytics"}, op.Table.Namespace)
				assert.Equal(t, "events", op.Table.Table)
				require.NotNil(t, op.RenameTo)
				assert.Equal(t, []string{"analytics"}, op.RenameTo.Namespace)
				assert.Equal(t, "events_v2", op.RenameTo.Table)
			},
		},
		{
			name:     "ALTER TABLE ADD COLUMN",
			stmt:     "ALTER TABLE analytics.events ADD COLUMN name string",
			wantKind: AddColumn,
			check: func(t *testing.T, op Operation) {
				assert.Equal(t, []string{"analytics"}, op.Table.Namespace)
				require.NotNil(t, op.Column)
				assert.Equal(t, "name", op.Column.Name)
				assert.Equal(t, String, op.Column.Type.Kind)
			},
		},
		{
			name:     "ALTER TABLE DROP COLUMN",
			stmt:     "ALTER TABLE analytics.events DROP COLUMN name",
			wantKind: DropColumn,
			check: func(t *testing.T, op Operation) {
				require.NotNil(t, op.Column)
				assert.Equal(t, "name", op.Column.Name)
			},
		},
		{
			name:     "ALTER TABLE RENAME COLUMN",
			stmt:     "ALTER TABLE analytics.events RENAME COLUMN name TO title",
			wantKind: RenameColumn,
			check: func(t *testing.T, op Operation) {
				require.NotNil(t, op.Column)
				assert.Equal(t, "name", op.Column.Name)
				assert.Equal(t, "title", op.NewName)
			},
		},
		{
			name:     "ALTER TABLE ALTER COLUMN TYPE",
			stmt:     "ALTER TABLE analytics.events ALTER COLUMN id TYPE long",
			wantKind: AlterColumnType,
			check: func(t *testing.T, op Operation) {
				require.NotNil(t, op.Column)
				assert.Equal(t, "id", op.Column.Name)
				assert.Equal(t, Long, op.Column.Type.Kind)
			},
		},
		{
			name:     "ALTER TABLE ADD PARTITION FIELD",
			stmt:     "ALTER TABLE analytics.events ADD PARTITION FIELD days(ts)",
			wantKind: AddPartitionField,
			check: func(t *testing.T, op Operation) {
				require.NotNil(t, op.Partition)
				assert.Equal(t, Days, op.Partition.Transform)
				assert.Equal(t, "ts", op.Partition.SourceCol)
			},
		},
		{
			name:     "ALTER TABLE DROP PARTITION FIELD",
			stmt:     "ALTER TABLE analytics.events DROP PARTITION FIELD days(ts)",
			wantKind: DropPartitionField,
			check: func(t *testing.T, op Operation) {
				require.NotNil(t, op.Partition)
				assert.Equal(t, Days, op.Partition.Transform)
				assert.Equal(t, "ts", op.Partition.SourceCol)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			op, err := Parse(tt.catalog, tt.stmt)
			require.NoError(t, err)
			assert.Equal(t, tt.wantKind, op.Kind)
			if tt.check != nil {
				tt.check(t, op)
			}
		})
	}
}

// ─── Referential CREATE TABLE ──────────────────────────────────────────────────

func TestParse_ReferentialCreateTable(t *testing.T) {
	t.Parallel()

	stmt := `CREATE TABLE iceberg.analytics.events (
		event_id   STRING    COMMENT 'unique event identifier',
		user_id    STRING    COMMENT 'external user identifier',
		event_type STRING    COMMENT 'event category',
		event_time TIMESTAMP COMMENT 'canonical event time (UTC)',
		created_at TIMESTAMP COMMENT 'record ingestion time',
		source     STRING    COMMENT 'ingestion source label'
	)
	USING iceberg
	PARTITIONED BY (days(event_time))
	COMMENT 'analytics events (bronze layer)'
	TBLPROPERTIES (
		'format-version' = '2',
		'write.format.default' = 'parquet',
		'write.parquet.compression-codec' = 'zstd'
	)`

	op, err := Parse("iceberg", stmt)
	require.NoError(t, err)

	// Kind
	assert.Equal(t, CreateTable, op.Kind)

	// Identifier: leading "iceberg" stripped (catalog=iceberg), namespace=["analytics"], table="events"
	assert.Equal(t, []string{"analytics"}, op.Table.Namespace)
	assert.Equal(t, "events", op.Table.Table)

	require.NotNil(t, op.Create)

	// Schema: 6 columns
	require.Len(t, op.Create.Schema, 6)

	// event_id
	assert.Equal(t, "event_id", op.Create.Schema[0].Name)
	assert.Equal(t, String, op.Create.Schema[0].Type.Kind)
	assert.Equal(t, "unique event identifier", op.Create.Schema[0].Doc)

	// user_id
	assert.Equal(t, "user_id", op.Create.Schema[1].Name)
	assert.Equal(t, String, op.Create.Schema[1].Type.Kind)
	assert.Equal(t, "external user identifier", op.Create.Schema[1].Doc)

	// event_type
	assert.Equal(t, "event_type", op.Create.Schema[2].Name)
	assert.Equal(t, String, op.Create.Schema[2].Type.Kind)
	assert.Equal(t, "event category", op.Create.Schema[2].Doc)

	// event_time: TIMESTAMP → TimestampTz (mapped to timestamptz with UTC timezone)
	assert.Equal(t, "event_time", op.Create.Schema[3].Name)
	assert.Equal(t, TimestampTz, op.Create.Schema[3].Type.Kind)
	assert.Equal(t, "canonical event time (UTC)", op.Create.Schema[3].Doc)

	// created_at: TIMESTAMP → TimestampTz
	assert.Equal(t, "created_at", op.Create.Schema[4].Name)
	assert.Equal(t, TimestampTz, op.Create.Schema[4].Type.Kind)
	assert.Equal(t, "record ingestion time", op.Create.Schema[4].Doc)

	// source
	assert.Equal(t, "source", op.Create.Schema[5].Name)
	assert.Equal(t, String, op.Create.Schema[5].Type.Kind)
	assert.Equal(t, "ingestion source label", op.Create.Schema[5].Doc)

	// Partition: days(event_time)
	require.Len(t, op.Create.Partition, 1)
	assert.Equal(t, Days, op.Create.Partition[0].Transform)
	assert.Equal(t, "event_time", op.Create.Partition[0].SourceCol)

	// TBLPROPERTIES
	assert.Equal(t, "2", op.Create.Props["format-version"])
	assert.Equal(t, "parquet", op.Create.Props["write.format.default"])
	assert.Equal(t, "zstd", op.Create.Props["write.parquet.compression-codec"])

	// Table COMMENT
	assert.Equal(t, "analytics events (bronze layer)", op.Create.Comment)
}

// ─── Timestamp types ───────────────────────────────────────────────────────────

func TestParse_TimestampTypes(t *testing.T) {
	t.Parallel()

	t.Run("TIMESTAMP → TimestampTz (with zone)", func(t *testing.T) {
		t.Parallel()
		op, err := Parse("", "CREATE TABLE raw.t (ts TIMESTAMP)")
		require.NoError(t, err)
		require.Len(t, op.Create.Schema, 1)
		assert.Equal(t, TimestampTz, op.Create.Schema[0].Type.Kind)
	})

	t.Run("TIMESTAMP_NTZ → Timestamp (without zone)", func(t *testing.T) {
		t.Parallel()
		op, err := Parse("", "CREATE TABLE raw.t (ts TIMESTAMP_NTZ)")
		require.NoError(t, err)
		require.Len(t, op.Create.Schema, 1)
		assert.Equal(t, Timestamp, op.Create.Schema[0].Type.Kind)
	})
}

// ─── Identifier resolution ─────────────────────────────────────────────────────

func TestParse_IdentResolution(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		catalog   string
		stmt      string
		wantNS    []string
		wantTable string
		wantErrIs error
	}{
		{
			name:      "catalog stripped: iceberg.raw.events → ns[raw], table=events",
			catalog:   "iceberg",
			stmt:      "CREATE TABLE iceberg.raw.events (id long)",
			wantNS:    []string{"raw"},
			wantTable: "events",
		},
		{
			name:      "multi-level namespace: bronze.raw.events (catalog=iceberg) → ns[bronze,raw], table=events",
			catalog:   "iceberg",
			stmt:      "CREATE TABLE bronze.raw.events (id long)",
			wantNS:    []string{"bronze", "raw"},
			wantTable: "events",
		},
		{
			name:      "no catalog prefix: analytics.events",
			catalog:   "iceberg",
			stmt:      "CREATE TABLE analytics.events (id long)",
			wantNS:    []string{"analytics"},
			wantTable: "events",
		},
		{
			name:      "namespace required: single segment without namespace",
			catalog:   "iceberg",
			stmt:      "CREATE TABLE events (id long)",
			wantErrIs: ErrNamespaceRequired,
		},
		{
			name:      "namespace required: no catalog, single segment",
			catalog:   "",
			stmt:      "CREATE TABLE events (id long)",
			wantErrIs: ErrNamespaceRequired,
		},
		// Backtick-quoted identifiers (FIX 1)
		{
			name:      "backtick: `raw`.`t` → ns[raw], table=t",
			catalog:   "",
			stmt:      "CREATE TABLE `raw`.`t` (id long)",
			wantNS:    []string{"raw"},
			wantTable: "t",
		},
		{
			name:      "backtick: `analytics`.`events` → ns[analytics], table=events",
			catalog:   "",
			stmt:      "DROP TABLE `analytics`.`events`",
			wantNS:    []string{"analytics"},
			wantTable: "events",
		},
		{
			name:      "backtick: multi-level `bronze`.`raw`.`events` → ns[bronze,raw], table=events",
			catalog:   "",
			stmt:      "DROP TABLE `bronze`.`raw`.`events`",
			wantNS:    []string{"bronze", "raw"},
			wantTable: "events",
		},
		{
			name:      "backtick: catalog-stripping `iceberg`.`raw`.`events` (catalog=iceberg) → ns[raw], table=events",
			catalog:   "iceberg",
			stmt:      "CREATE TABLE `iceberg`.`raw`.`events` (id long)",
			wantNS:    []string{"raw"},
			wantTable: "events",
		},
		{
			name:      "backtick: single segment → ErrNamespaceRequired",
			catalog:   "",
			stmt:      "CREATE TABLE `events` (id long)",
			wantErrIs: ErrNamespaceRequired,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			op, err := Parse(tt.catalog, tt.stmt)
			if tt.wantErrIs != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErrIs), "expected errors.Is(%v) but got: %v", tt.wantErrIs, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantNS, op.Table.Namespace)
			assert.Equal(t, tt.wantTable, op.Table.Table)
		})
	}
}

// ─── Negative / error cases ────────────────────────────────────────────────────

func TestParse_Negatives(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		stmt      string
		wantErrIs error
	}{
		{
			name:      "unsupported DDL: TRUNCATE TABLE",
			stmt:      "TRUNCATE TABLE analytics.events",
			wantErrIs: ErrUnsupportedDDL,
		},
		{
			name:      "parse error: malformed CREATE TABLE",
			stmt:      "CREATE TABLE (( )",
			wantErrIs: ErrParse,
		},
		{
			name:      "parse error: empty statement",
			stmt:      "",
			wantErrIs: ErrParse,
		},
		{
			name:      "unsupported DDL: INSERT INTO",
			stmt:      "INSERT INTO raw.t VALUES (1)",
			wantErrIs: ErrUnsupportedDDL,
		},
		{
			name:      "unsupported DDL: CREATE INDEX",
			stmt:      "CREATE INDEX idx ON raw.t (id)",
			wantErrIs: ErrUnsupportedDDL,
		},
		{
			name:      "unsupported DDL: DROP INDEX",
			stmt:      "DROP INDEX idx ON raw.t",
			wantErrIs: ErrUnsupportedDDL,
		},
		{
			name:      "unsupported DDL: RENAME unknown",
			stmt:      "RENAME VIEW raw.v TO raw.v2",
			wantErrIs: ErrUnsupportedDDL,
		},
		{
			name:      "unsupported DDL: ALTER unknown",
			stmt:      "ALTER VIEW raw.v AS SELECT 1",
			wantErrIs: ErrUnsupportedDDL,
		},
		{
			name:      "parse error: incomplete CREATE",
			stmt:      "CREATE",
			wantErrIs: ErrParse,
		},
		{
			name:      "parse error: incomplete DROP",
			stmt:      "DROP",
			wantErrIs: ErrParse,
		},
		{
			name:      "parse error: incomplete RENAME",
			stmt:      "RENAME",
			wantErrIs: ErrParse,
		},
		{
			name:      "parse error: incomplete ALTER",
			stmt:      "ALTER",
			wantErrIs: ErrParse,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := Parse("", tt.stmt)
			require.Error(t, err)
			assert.True(t, errors.Is(err, tt.wantErrIs),
				"expected errors.Is(%v) but got: %v", tt.wantErrIs, err)
		})
	}
}

// TestParse_DropTableIfExists exercises the IF EXISTS path.
func TestParse_DropTableIfExists(t *testing.T) {
	t.Parallel()
	op, err := Parse("", "DROP TABLE IF EXISTS analytics.events")
	require.NoError(t, err)
	assert.Equal(t, DropTable, op.Kind)
	assert.Equal(t, "events", op.Table.Table)
}

// TestParse_CreateTable_MultipleClauses exercises additional CREATE TABLE code paths.
func TestParse_CreateTable_MultipleClauses(t *testing.T) {
	t.Parallel()

	t.Run("multiple transforms", func(t *testing.T) {
		t.Parallel()
		op, err := Parse("", "CREATE TABLE raw.t (id long, ts timestamp) USING iceberg PARTITIONED BY (identity(id), days(ts))")
		require.NoError(t, err)
		require.Len(t, op.Create.Partition, 2)
		assert.Equal(t, Identity, op.Create.Partition[0].Transform)
		assert.Equal(t, Days, op.Create.Partition[1].Transform)
	})

	t.Run("TBLPROPERTIES without COMMENT", func(t *testing.T) {
		t.Parallel()
		op, err := Parse("", "CREATE TABLE raw.t (id long) TBLPROPERTIES ('k' = 'v')")
		require.NoError(t, err)
		assert.Equal(t, "v", op.Create.Props["k"])
		assert.Empty(t, op.Create.Comment)
	})
}

// TestParse_AlterTable_AddColumnWithComment verifies COMMENT is captured in ADD COLUMN.
func TestParse_AlterTable_AddColumnWithComment(t *testing.T) {
	t.Parallel()
	op, err := Parse("", "ALTER TABLE raw.t ADD COLUMN info string COMMENT 'some info'")
	require.NoError(t, err)
	require.NotNil(t, op.Column)
	assert.Equal(t, "info", op.Column.Name)
	assert.Equal(t, "some info", op.Column.Doc)
}

// TestParse_DecimalTypeParsing covers DECIMAL type via parseType directly.
func TestParse_DecimalTypeParsing(t *testing.T) {
	t.Parallel()

	t.Run("decimal no params uses defaults", func(t *testing.T) {
		t.Parallel()
		tp, err := parseType("decimal")
		require.NoError(t, err)
		assert.Equal(t, Decimal, tp.Kind)
		assert.Equal(t, defaultDecimalPrecision, tp.Prec)
	})

	t.Run("invalid decimal format", func(t *testing.T) {
		t.Parallel()
		_, err := parseType("decimal[10,2]")
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrParse))
	})
}

// TestParse_Tokenizer_SpecialCases exercises edge cases in the tokenizer.
func TestParse_Tokenizer_SpecialCases(t *testing.T) {
	t.Parallel()

	t.Run("escaped single quote in comment", func(t *testing.T) {
		t.Parallel()
		// Single-quoted string with escaped single quote — tokenizer shouldn't crash
		op, err := Parse("", "CREATE TABLE raw.t (id long) COMMENT 'it''s a test'")
		require.NoError(t, err)
		// The comment should include the content (escaped quotes normalized by SQL, not by our parser)
		assert.NotEmpty(t, op.Create.Comment)
	})

	t.Run("semicolon stripped", func(t *testing.T) {
		t.Parallel()
		op, err := Parse("", "DROP TABLE raw.t;")
		require.NoError(t, err)
		assert.Equal(t, DropTable, op.Kind)
	})
}

// TestParse_CreateNamespace_WithCatalog exercises catalog stripping for namespace operations.
func TestParse_CreateNamespace_WithCatalog(t *testing.T) {
	t.Parallel()

	t.Run("catalog stripped from namespace", func(t *testing.T) {
		t.Parallel()
		op, err := Parse("iceberg", "CREATE NAMESPACE iceberg.raw")
		require.NoError(t, err)
		assert.Equal(t, CreateNamespace, op.Kind)
		assert.Equal(t, []string{"raw"}, op.Table.Namespace)
	})

	t.Run("multi-level namespace", func(t *testing.T) {
		t.Parallel()
		op, err := Parse("", "DROP NAMESPACE bronze.raw")
		require.NoError(t, err)
		assert.Equal(t, DropNamespace, op.Kind)
		assert.Equal(t, []string{"bronze", "raw"}, op.Table.Namespace)
	})

	// Backtick-quoted namespace identifiers (FIX 1).
	t.Run("backtick: `iceberg`.`raw` with catalog stripping", func(t *testing.T) {
		t.Parallel()
		op, err := Parse("iceberg", "CREATE NAMESPACE `iceberg`.`raw`")
		require.NoError(t, err)
		assert.Equal(t, CreateNamespace, op.Kind)
		assert.Equal(t, []string{"raw"}, op.Table.Namespace)
	})

	t.Run("backtick: `analytics` single-segment namespace", func(t *testing.T) {
		t.Parallel()
		op, err := Parse("", "DROP NAMESPACE `analytics`")
		require.NoError(t, err)
		assert.Equal(t, DropNamespace, op.Kind)
		assert.Equal(t, []string{"analytics"}, op.Table.Namespace)
	})
}

// TestParse_CreateNamespace_IfNotExists verifies that the optional IF NOT EXISTS clause is
// parsed and recorded in the IR, enabling idempotent namespace creation.
func TestParse_CreateNamespace_IfNotExists(t *testing.T) {
	t.Parallel()

	t.Run("with IF NOT EXISTS", func(t *testing.T) {
		t.Parallel()
		op, err := Parse("iceberg", "CREATE NAMESPACE IF NOT EXISTS iceberg.raw")
		require.NoError(t, err)
		assert.Equal(t, CreateNamespace, op.Kind)
		assert.Equal(t, []string{"raw"}, op.Table.Namespace)
		assert.True(t, op.IfNotExists)
	})

	t.Run("without IF NOT EXISTS", func(t *testing.T) {
		t.Parallel()
		op, err := Parse("", "CREATE NAMESPACE analytics")
		require.NoError(t, err)
		assert.Equal(t, CreateNamespace, op.Kind)
		assert.Equal(t, []string{"analytics"}, op.Table.Namespace)
		assert.False(t, op.IfNotExists)
	})
}

// TestParse_CreateTable_IfNotExists verifies that the optional IF NOT EXISTS clause is parsed and
// recorded in the IR, enabling idempotent table creation (regression: the clause was previously
// consumed but discarded, so CREATE TABLE IF NOT EXISTS still failed with AlreadyExistsException).
func TestParse_CreateTable_IfNotExists(t *testing.T) {
	t.Parallel()

	t.Run("with IF NOT EXISTS", func(t *testing.T) {
		t.Parallel()
		op, err := Parse("iceberg", "CREATE TABLE IF NOT EXISTS iceberg.analytics.events (id long)")
		require.NoError(t, err)
		assert.Equal(t, CreateTable, op.Kind)
		assert.Equal(t, []string{"analytics"}, op.Table.Namespace)
		assert.Equal(t, "events", op.Table.Table)
		assert.True(t, op.IfNotExists)
	})

	t.Run("without IF NOT EXISTS", func(t *testing.T) {
		t.Parallel()
		op, err := Parse("iceberg", "CREATE TABLE iceberg.analytics.events (id long)")
		require.NoError(t, err)
		assert.Equal(t, CreateTable, op.Kind)
		assert.False(t, op.IfNotExists)
	})
}

// TestParse_DropTable_IfExists verifies that the optional IF EXISTS clause is parsed and recorded
// in the IR, enabling idempotent DROP TABLE (regression: the clause was previously consumed but
// discarded).
func TestParse_DropTable_IfExists(t *testing.T) {
	t.Parallel()

	t.Run("with IF EXISTS", func(t *testing.T) {
		t.Parallel()
		op, err := Parse("iceberg", "DROP TABLE IF EXISTS iceberg.analytics.events")
		require.NoError(t, err)
		assert.Equal(t, DropTable, op.Kind)
		assert.Equal(t, []string{"analytics"}, op.Table.Namespace)
		assert.Equal(t, "events", op.Table.Table)
		assert.True(t, op.IfExists)
	})

	t.Run("without IF EXISTS", func(t *testing.T) {
		t.Parallel()
		op, err := Parse("iceberg", "DROP TABLE iceberg.analytics.events")
		require.NoError(t, err)
		assert.Equal(t, DropTable, op.Kind)
		assert.False(t, op.IfExists)
	})
}

// TestParse_DropNamespace_IfExists verifies that the optional IF EXISTS clause is parsed and
// recorded in the IR, enabling idempotent DROP NAMESPACE (regression: the parser did not understand
// IF EXISTS for namespaces at all).
func TestParse_DropNamespace_IfExists(t *testing.T) {
	t.Parallel()

	t.Run("with IF EXISTS", func(t *testing.T) {
		t.Parallel()
		op, err := Parse("iceberg", "DROP NAMESPACE IF EXISTS iceberg.raw")
		require.NoError(t, err)
		assert.Equal(t, DropNamespace, op.Kind)
		assert.Equal(t, []string{"raw"}, op.Table.Namespace)
		assert.True(t, op.IfExists)
	})

	t.Run("without IF EXISTS", func(t *testing.T) {
		t.Parallel()
		op, err := Parse("", "DROP NAMESPACE analytics")
		require.NoError(t, err)
		assert.Equal(t, DropNamespace, op.Kind)
		assert.Equal(t, []string{"analytics"}, op.Table.Namespace)
		assert.False(t, op.IfExists)
	})
}

// TestParse_WriteOrderedBy verifies parsing of ALTER TABLE … WRITE ORDERED BY, including
// direction, null ordering, transforms, defaults, and the WRITE UNORDERED reset form.
func TestParse_WriteOrderedBy(t *testing.T) {
	t.Parallel()

	t.Run("plain columns with direction and defaults", func(t *testing.T) {
		t.Parallel()
		op, err := Parse("iceberg", "ALTER TABLE iceberg.analytics.events WRITE ORDERED BY event_time DESC, user_id")
		require.NoError(t, err)
		assert.Equal(t, SetSortOrder, op.Kind)
		assert.Equal(t, "events", op.Table.Table)
		require.NotNil(t, op.Sort)
		assert.False(t, op.Sort.Unordered)
		require.Len(t, op.Sort.Fields, 2)

		// event_time DESC → direction DESC, default null order NULLS LAST.
		assert.Equal(t, Identity, op.Sort.Fields[0].Transform)
		assert.Equal(t, "event_time", op.Sort.Fields[0].SourceCol)
		assert.Equal(t, SortDesc, op.Sort.Fields[0].Direction)
		assert.Equal(t, NullsLast, op.Sort.Fields[0].NullOrder)

		// user_id (no direction) → default ASC, default null order NULLS FIRST.
		assert.Equal(t, "user_id", op.Sort.Fields[1].SourceCol)
		assert.Equal(t, SortAsc, op.Sort.Fields[1].Direction)
		assert.Equal(t, NullsFirst, op.Sort.Fields[1].NullOrder)
	})

	t.Run("explicit NULLS FIRST/LAST overrides default", func(t *testing.T) {
		t.Parallel()
		op, err := Parse("", "ALTER TABLE analytics.events WRITE ORDERED BY a ASC NULLS LAST, b DESC NULLS FIRST")
		require.NoError(t, err)
		require.NotNil(t, op.Sort)
		require.Len(t, op.Sort.Fields, 2)
		assert.Equal(t, SortAsc, op.Sort.Fields[0].Direction)
		assert.Equal(t, NullsLast, op.Sort.Fields[0].NullOrder)
		assert.Equal(t, SortDesc, op.Sort.Fields[1].Direction)
		assert.Equal(t, NullsFirst, op.Sort.Fields[1].NullOrder)
	})

	t.Run("transform sort column", func(t *testing.T) {
		t.Parallel()
		op, err := Parse("", "ALTER TABLE analytics.events WRITE ORDERED BY bucket(16, user_id) DESC, days(event_time)")
		require.NoError(t, err)
		require.NotNil(t, op.Sort)
		require.Len(t, op.Sort.Fields, 2)
		assert.Equal(t, Bucket, op.Sort.Fields[0].Transform)
		assert.Equal(t, 16, op.Sort.Fields[0].Param)
		assert.Equal(t, "user_id", op.Sort.Fields[0].SourceCol)
		assert.Equal(t, SortDesc, op.Sort.Fields[0].Direction)
		assert.Equal(t, Days, op.Sort.Fields[1].Transform)
		assert.Equal(t, "event_time", op.Sort.Fields[1].SourceCol)
	})

	t.Run("optional surrounding parentheses accepted", func(t *testing.T) {
		t.Parallel()
		op, err := Parse("", "ALTER TABLE analytics.events WRITE ORDERED BY (a, b DESC)")
		require.NoError(t, err)
		require.NotNil(t, op.Sort)
		require.Len(t, op.Sort.Fields, 2)
	})

	t.Run("WRITE UNORDERED clears sort order", func(t *testing.T) {
		t.Parallel()
		op, err := Parse("", "ALTER TABLE analytics.events WRITE UNORDERED")
		require.NoError(t, err)
		assert.Equal(t, SetSortOrder, op.Kind)
		require.NotNil(t, op.Sort)
		assert.True(t, op.Sort.Unordered)
		assert.Empty(t, op.Sort.Fields)
	})

	t.Run("unsupported WRITE variant returns error", func(t *testing.T) {
		t.Parallel()
		_, err := Parse("", "ALTER TABLE analytics.events WRITE DISTRIBUTED BY PARTITION")
		require.Error(t, err)
	})

	t.Run("NULLS without FIRST/LAST is a parse error", func(t *testing.T) {
		t.Parallel()
		_, err := Parse("", "ALTER TABLE analytics.events WRITE ORDERED BY a NULLS")
		require.Error(t, err)
	})
}

// TestParse_NotNull_OutsideSubset verifies that NOT NULL (outside subset v1) returns ErrParse
// and does not panic. Field.Required is not supported in subset v1.
func TestParse_NotNull_OutsideSubset(t *testing.T) {
	t.Parallel()
	_, err := Parse("", "CREATE TABLE raw.t (id long NOT NULL)")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrParse),
		"NOT NULL should produce ErrParse (outside subset v1), got: %v", err)
}

// TestParse_Comments verifies that SQL comments are stripped before parsing
// while string-literal content is preserved intact.
func TestParse_Comments(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		catalog  string
		stmt     string
		wantKind OpKind
		check    func(t *testing.T, op Operation)
	}{
		{
			// Reproduces the live-smoke bug: leading line comment above statement.
			name:     "leading line comment above CREATE NAMESPACE",
			stmt:     "-- Create the raw namespace (bronze layer)\nCREATE NAMESPACE raw",
			wantKind: CreateNamespace,
			check: func(t *testing.T, op Operation) {
				assert.Equal(t, []string{"raw"}, op.Table.Namespace)
			},
		},
		{
			name:     "trailing line comment after statement",
			stmt:     "CREATE NAMESPACE raw -- this is the bronze layer",
			wantKind: CreateNamespace,
			check: func(t *testing.T, op Operation) {
				assert.Equal(t, []string{"raw"}, op.Table.Namespace)
			},
		},
		{
			name:     "inline block comment between clauses",
			stmt:     "DROP TABLE /* temporary cleanup */ analytics.events",
			wantKind: DropTable,
			check: func(t *testing.T, op Operation) {
				assert.Equal(t, "events", op.Table.Table)
			},
		},
		{
			name:     "comment containing SQL keywords is ignored",
			stmt:     "-- CREATE TABLE foo\nCREATE NAMESPACE real_ns",
			wantKind: CreateNamespace,
			check: func(t *testing.T, op Operation) {
				assert.Equal(t, []string{"real_ns"}, op.Table.Namespace)
			},
		},
		{
			// String literals must NOT be stripped even when they contain -- or /* */.
			name:    "string literal with dashes and block comment markers preserved",
			catalog: "iceberg",
			stmt: `CREATE TABLE raw.t (
				c string COMMENT 'has -- dashes and /* stars */ inside'
			)`,
			wantKind: CreateTable,
			check: func(t *testing.T, op Operation) {
				require.NotNil(t, op.Create)
				require.Len(t, op.Create.Schema, 1)
				assert.Equal(t, "c", op.Create.Schema[0].Name)
				assert.Equal(t, "has -- dashes and /* stars */ inside", op.Create.Schema[0].Doc)
			},
		},
		{
			// Multi-line block comment spanning the entire preceding line.
			name:     "multi-line block comment between clauses",
			stmt:     "CREATE TABLE raw.t (id long)\n/* configure storage */\nUSING iceberg",
			wantKind: CreateTable,
			check: func(t *testing.T, op Operation) {
				require.NotNil(t, op.Create)
				require.Len(t, op.Create.Schema, 1)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			op, err := Parse(tt.catalog, tt.stmt)
			require.NoError(t, err)
			assert.Equal(t, tt.wantKind, op.Kind)
			if tt.check != nil {
				tt.check(t, op)
			}
		})
	}
}

// TestParse_AlterTable_ErrorPaths exercises ALTER TABLE error paths.
func TestParse_AlterTable_ErrorPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		stmt    string
		wantErr error
	}{
		{
			name:    "ALTER TABLE unsupported sub-command",
			stmt:    "ALTER TABLE raw.t SET TBLPROPERTIES ('k'='v')",
			wantErr: ErrUnsupportedDDL,
		},
		{
			name:    "ALTER TABLE ADD unsupported",
			stmt:    "ALTER TABLE raw.t ADD CONSTRAINT pk PRIMARY KEY (id)",
			wantErr: ErrUnsupportedDDL,
		},
		{
			name:    "ALTER TABLE DROP unsupported",
			stmt:    "ALTER TABLE raw.t DROP CONSTRAINT pk",
			wantErr: ErrUnsupportedDDL,
		},
		{
			name:    "ALTER TABLE RENAME unsupported",
			stmt:    "ALTER TABLE raw.t RENAME TO raw.t2",
			wantErr: ErrUnsupportedDDL,
		},
		{
			name:    "ALTER TABLE ALTER unsupported",
			stmt:    "ALTER TABLE raw.t ALTER TYPE int",
			wantErr: ErrUnsupportedDDL,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := Parse("", tt.stmt)
			require.Error(t, err)
			assert.True(t, errors.Is(err, tt.wantErr),
				"expected errors.Is(%v) but got: %v", tt.wantErr, err)
		})
	}
}

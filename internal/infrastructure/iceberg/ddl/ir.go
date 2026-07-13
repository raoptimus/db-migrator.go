package ddl

// OpKind identifies the kind of DDL operation.
type OpKind int

const (
	CreateNamespace    OpKind = iota // CREATE NAMESPACE
	DropNamespace                    // DROP NAMESPACE
	CreateTable                      // CREATE TABLE
	DropTable                        // DROP TABLE
	RenameTable                      // RENAME TABLE
	AddColumn                        // ALTER TABLE … ADD COLUMN
	DropColumn                       // ALTER TABLE … DROP COLUMN
	RenameColumn                     // ALTER TABLE … RENAME COLUMN
	AlterColumnType                  // ALTER TABLE … ALTER COLUMN … TYPE
	AddPartitionField                // ALTER TABLE … ADD PARTITION FIELD
	DropPartitionField               // ALTER TABLE … DROP PARTITION FIELD
)

// Ident is a fully-qualified table or namespace identifier with the catalog prefix already stripped.
type Ident struct {
	// Namespace holds every identifier segment except the last (table name).
	// For namespace-only operations (CREATE/DROP NAMESPACE) Table is empty.
	Namespace []string
	Table     string
}

// Operation is the IR produced by the parser for a single Spark-SQL DDL statement.
type Operation struct {
	Kind      OpKind
	Table     Ident
	RenameTo  *Ident            // RenameTable: destination identifier
	Column    *Field            // AddColumn / DropColumn / AlterColumnType / RenameColumn (source column)
	NewName   string            // RenameColumn: new column name
	Partition *PartitionField   // AddPartitionField / DropPartitionField
	Create    *CreateTableSpec  // CreateTable: full table specification
	Props     map[string]string // CreateNamespace: optional properties
}

// CreateTableSpec holds the full specification of a CREATE TABLE statement.
type CreateTableSpec struct {
	Schema    []Field
	Partition []PartitionField
	Props     map[string]string // TBLPROPERTIES
	Comment   string            // table-level COMMENT
}

// Field describes a single table column or struct member.
type Field struct {
	Name     string
	Type     IcebergType
	Doc      string // COMMENT clause → doc of the field
	Required bool
}

// TypeKind enumerates all supported Iceberg type kinds.
type TypeKind int

const (
	Boolean     TypeKind = iota
	Int                  // 32-bit integer
	Long                 // 64-bit integer (also BIGINT)
	Float                // 32-bit float
	Double               // 64-bit float
	Decimal              // fixed-precision decimal with Prec and Scale
	Date                 // calendar date
	Time                 // time of day
	TimestampTz          // TIMESTAMP with timezone (UTC) — Spark TIMESTAMP
	Timestamp            // TIMESTAMP without timezone — Spark TIMESTAMP_NTZ
	String               // variable-length string
	UUID                 // UUID
	Binary               // binary data
	Struct               // nested struct
	List                 // ordered list (Spark ARRAY<T>)
	Map                  // key/value map
)

// IcebergType is a recursive type descriptor.
type IcebergType struct {
	Kind        TypeKind
	Prec, Scale int          // Decimal: precision and scale
	Fields      []Field      // Struct: nested fields
	Elem        *IcebergType // List: element type
	Key, Val    *IcebergType // Map: key and value types
}

// TransformKind identifies a partition transform function.
type TransformKind int

const (
	Identity TransformKind = iota
	Years
	Months
	Days
	Hours
	Bucket   // bucket(N, col)
	Truncate // truncate(N, col)
)

// PartitionField describes a single partition transform applied to a source column.
type PartitionField struct {
	Transform TransformKind
	Param     int    // Bucket / Truncate: N
	SourceCol string // source column name
	Name      string // optional explicit partition field name
}

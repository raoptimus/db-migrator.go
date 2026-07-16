/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package catalog

import (
	"context"
	"fmt"

	iceberg "github.com/apache/iceberg-go"
	icebergcatalog "github.com/apache/iceberg-go/catalog"
	"github.com/apache/iceberg-go/catalog/rest"
	"github.com/apache/iceberg-go/table"
	"github.com/pkg/errors"
	"github.com/raoptimus/db-migrator.go/internal/infrastructure/iceberg/ddl"
)

// partitionFieldIDBase is the base field-ID used for partition spec fields.
// Iceberg convention: partition field IDs start at 1000 to avoid conflicts with schema field IDs.
const partitionFieldIDBase = 1000

// ident converts a ddl.Ident to a table.Identifier ([]string).
// It appends the table name after the namespace segments.
func ident(id ddl.Ident) table.Identifier {
	parts := make(table.Identifier, 0, len(id.Namespace)+1)
	parts = append(parts, id.Namespace...)
	if id.Table != "" {
		parts = append(parts, id.Table)
	}
	return parts
}

// CreateTable creates an Iceberg table from the given IR specification.
// It builds the iceberg-go Schema (with nested types and doc strings),
// the PartitionSpec, and merges table properties and COMMENT.
func (c *Client) CreateTable(ctx context.Context, id ddl.Ident, spec ddl.CreateTableSpec) error {
	schema, fieldIDs, err := buildSchema(spec.Schema)
	if err != nil {
		return errors.WithMessage(err, "build schema")
	}

	partSpec, err := buildPartitionSpec(fieldIDs, spec.Partition)
	if err != nil {
		return errors.WithMessage(err, "build partition spec")
	}

	props := make(iceberg.Properties, len(spec.Props)+1)
	for k, v := range spec.Props {
		props[k] = v
	}
	if spec.Comment != "" {
		props["comment"] = spec.Comment
	}

	const maxCreateTableOpts = 2
	opts := make([]icebergcatalog.CreateTableOpt, 0, maxCreateTableOpts)
	opts = append(opts, icebergcatalog.WithPartitionSpec(partSpec))
	if len(props) > 0 {
		opts = append(opts, icebergcatalog.WithProperties(props))
	}

	_, err = c.cat.CreateTable(ctx, ident(id), schema, opts...)
	if err != nil {
		return errors.WithMessage(err, "create table")
	}
	return nil
}

// TableExists reports whether the given table exists in the catalog.
//
// It prefers a single lightweight HEAD probe (CheckTableExists), which touches neither
// S3 nor table metadata. Older REST servers (JdbcCatalog builds) do not implement HEAD
// and reject it with 400; on that definitive signal we permanently switch to the
// GET-based ListTables path for the rest of the process (same reasoning as the namespace
// GET-over-HEAD fix). Transient HEAD errors trigger a one-off fallback without disabling
// HEAD for later probes.
func (c *Client) TableExists(ctx context.Context, id ddl.Ident) (bool, error) {
	if !c.headUnsupported {
		exists, err := c.cat.CheckTableExists(ctx, ident(id))
		if err == nil {
			return exists, nil
		}
		if errors.Is(err, rest.ErrBadRequest) {
			// HEAD unsupported by this server — stop probing with HEAD for the rest of the run.
			c.headUnsupported = true
		}
		// Fall through to the GET-based path for this call regardless of error kind.
	}

	// GET fallback: list tables in the namespace (no S3, no metadata read) and match by name.
	// SetPageSize(0) suppresses the unconditional pageSize query param that trips a
	// NumberFormatException on the reference REST server (see Ping).
	ctx = c.cat.SetPageSize(ctx, 0)
	for tbl, err := range c.cat.ListTables(ctx, id.Namespace) {
		if err != nil {
			return false, errors.WithMessage(err, "check table exists")
		}
		if len(tbl) > 0 && tbl[len(tbl)-1] == id.Table {
			return true, nil
		}
	}
	return false, nil
}

// DropTable drops an Iceberg table.
func (c *Client) DropTable(ctx context.Context, id ddl.Ident) error {
	if err := c.cat.DropTable(ctx, ident(id)); err != nil {
		return errors.WithMessage(err, "drop table")
	}
	return nil
}

// RenameTable renames an Iceberg table from from to to.
func (c *Client) RenameTable(ctx context.Context, from, to ddl.Ident) error {
	if _, err := c.cat.RenameTable(ctx, ident(from), ident(to)); err != nil {
		return errors.WithMessage(err, "rename table")
	}
	return nil
}

// ApplySchemaChange applies a schema-level DDL operation to an Iceberg table.
// Supported operations: AddColumn, DropColumn, RenameColumn, AlterColumnType.
// Uses iceberg-go transaction with allowIncompatibleChanges=false, which means
// narrowing type changes (e.g. long→int) are rejected by the library itself (Р8 fail-fast).
//
//nolint:gocritic // hugeParam: op is passed by value to match the IcebergCatalog interface contract
func (c *Client) ApplySchemaChange(ctx context.Context, op ddl.Operation) error {
	tbl, err := c.cat.LoadTable(ctx, ident(op.Table))
	if err != nil {
		return errors.WithMessage(err, "load table for schema change")
	}

	txn := tbl.NewTransaction()
	us := txn.UpdateSchema(true /* caseSensitive */, false /* allowIncompatibleChanges */)

	switch op.Kind {
	case ddl.AddColumn:
		if op.Column == nil {
			return errors.New("AddColumn: column spec is nil")
		}
		colType, err := toIcebergType(op.Column.Type, &idCounter{n: tbl.Schema().HighestFieldID()})
		if err != nil {
			return errors.WithMessagef(err, "AddColumn: map type for %s", op.Column.Name)
		}
		us.AddColumn([]string{op.Column.Name}, colType, op.Column.Doc, op.Column.Required, nil)

	case ddl.DropColumn:
		if op.Column == nil {
			return errors.New("DropColumn: column spec is nil")
		}
		us.DeleteColumn([]string{op.Column.Name})

	case ddl.RenameColumn:
		if op.Column == nil {
			return errors.New("RenameColumn: column spec is nil")
		}
		us.RenameColumn([]string{op.Column.Name}, op.NewName)

	case ddl.AlterColumnType:
		if op.Column == nil {
			return errors.New("AlterColumnType: column spec is nil")
		}
		colType, err := toIcebergType(op.Column.Type, &idCounter{n: tbl.Schema().HighestFieldID()})
		if err != nil {
			return errors.WithMessagef(err, "AlterColumnType: map type for %s", op.Column.Name)
		}
		us.UpdateColumn([]string{op.Column.Name}, table.ColumnUpdate{
			FieldType: iceberg.Optional[iceberg.Type]{Val: colType, Valid: true},
		})

	default:
		return errors.Errorf("unsupported schema change kind: %d", op.Kind)
	}

	if err := us.Commit(); err != nil {
		return errors.WithMessage(err, "commit schema change")
	}
	if _, err := txn.Commit(ctx); err != nil {
		return errors.WithMessage(err, "persist schema change")
	}
	return nil
}

// ApplySpecChange applies a partition-spec DDL operation to an Iceberg table.
// Supported operations: AddPartitionField, DropPartitionField.
//
//nolint:gocritic // hugeParam: op is passed by value to match the IcebergCatalog interface contract
func (c *Client) ApplySpecChange(ctx context.Context, op ddl.Operation) error {
	tbl, err := c.cat.LoadTable(ctx, ident(op.Table))
	if err != nil {
		return errors.WithMessage(err, "load table for spec change")
	}

	txn := tbl.NewTransaction()
	us := txn.UpdateSpec(true /* caseSensitive */)

	switch op.Kind {
	case ddl.AddPartitionField:
		if op.Partition == nil {
			return errors.New("AddPartitionField: partition spec is nil")
		}
		pf := op.Partition
		transform, err := toIcebergTransform(pf.Transform, pf.Param)
		if err != nil {
			return errors.WithMessagef(err, "AddPartitionField: map transform for %s", pf.SourceCol)
		}
		if pf.Transform == ddl.Identity {
			us.AddIdentity(pf.SourceCol)
		} else {
			name := pf.Name
			if name == "" {
				name = fmt.Sprintf("%s_%s", pf.SourceCol, transform.String())
			}
			us.AddField(pf.SourceCol, transform, name)
		}

	case ddl.DropPartitionField:
		if op.Partition == nil {
			return errors.New("DropPartitionField: partition spec is nil")
		}
		fieldName, err := findPartitionFieldName(tbl.Schema(), tbl.Spec(), op.Partition)
		if err != nil {
			return errors.WithMessage(err, "DropPartitionField")
		}
		us.RemoveField(fieldName)

	default:
		return errors.Errorf("unsupported spec change kind: %d", op.Kind)
	}

	if err := us.Commit(); err != nil {
		return errors.WithMessage(err, "commit spec change")
	}
	if _, err := txn.Commit(ctx); err != nil {
		return errors.WithMessage(err, "persist spec change")
	}
	return nil
}

// ApplySortOrderChange sets (or clears) a table's write sort order.
//
// iceberg-go v0.6.0 has no high-level sort-order transaction builder (unlike UpdateSpec /
// UpdateSchema), so the change is applied through the exported catalog-level CommitTable with
// AddSortOrder + SetDefaultSortOrder updates, guarded by an AssertTableUUID requirement for
// optimistic concurrency. WRITE UNORDERED resets the default to the always-present unsorted
// order (id 0).
//
//nolint:gocritic // hugeParam: op is passed by value to match the IcebergCatalog interface contract
func (c *Client) ApplySortOrderChange(ctx context.Context, op ddl.Operation) error {
	if op.Sort == nil {
		return errors.New("SetSortOrder: sort spec is nil")
	}
	tbl, err := c.cat.LoadTable(ctx, ident(op.Table))
	if err != nil {
		return errors.WithMessage(err, "load table for sort order change")
	}

	var updates []table.Update
	if op.Sort.Unordered {
		updates = []table.Update{table.NewSetDefaultSortOrderUpdate(table.UnsortedSortOrderID)}
	} else {
		schema := tbl.Schema()
		fields := make([]table.SortField, 0, len(op.Sort.Fields))
		for _, sf := range op.Sort.Fields {
			col, ok := schema.FindFieldByName(sf.SourceCol)
			if !ok {
				return errors.Errorf("sort column %q not found in table schema", sf.SourceCol)
			}
			transform, err := toIcebergTransform(sf.Transform, sf.Param)
			if err != nil {
				return errors.WithMessagef(err, "map transform for sort column %s", sf.SourceCol)
			}
			fields = append(fields, table.SortField{
				SourceIDs: []int{col.ID},
				Transform: transform,
				Direction: toIcebergSortDirection(sf.Direction),
				NullOrder: toIcebergNullOrder(sf.NullOrder),
			})
		}
		so, err := table.NewSortOrder(table.InitialSortOrderID, fields)
		if err != nil {
			return errors.WithMessage(err, "build sort order")
		}
		// SetDefaultSortOrder(-1) points the default at the just-added order (its final id is
		// assigned server-side / by the metadata builder).
		updates = []table.Update{
			table.NewAddSortOrderUpdate(&so),
			table.NewSetDefaultSortOrderUpdate(-1),
		}
	}

	reqs := []table.Requirement{table.AssertTableUUID(tbl.Metadata().TableUUID())}
	if _, _, err := c.cat.CommitTable(ctx, ident(op.Table), reqs, updates); err != nil {
		return errors.WithMessage(err, "commit sort order")
	}
	return nil
}

// toIcebergSortDirection maps a ddl.SortDirection to iceberg-go's table.SortDirection.
func toIcebergSortDirection(d ddl.SortDirection) table.SortDirection {
	if d == ddl.SortDesc {
		return table.SortDESC
	}
	return table.SortASC
}

// toIcebergNullOrder maps a ddl.NullOrder to iceberg-go's table.NullOrder.
func toIcebergNullOrder(n ddl.NullOrder) table.NullOrder {
	if n == ddl.NullsLast {
		return table.NullsLast
	}
	return table.NullsFirst
}

// ─── schema building ────────────────────────────────────────────────────────

// idCounter tracks the next field ID to assign during schema construction.
type idCounter struct {
	n int
}

func (c *idCounter) next() int {
	c.n++
	return c.n
}

// buildSchema converts a slice of ddl.Field into an *iceberg.Schema.
// Field IDs are assigned sequentially starting from 1.
// It also returns a map of top-level column name → assigned field ID so that
// buildPartitionSpec can resolve SourceIDs correctly even when nested types
// consume IDs before a scalar column (e.g. struct<x:int> before timestamp).
func buildSchema(fields []ddl.Field) (*iceberg.Schema, map[string]int, error) {
	ctr := &idCounter{}
	nested, err := buildNestedFields(fields, ctr)
	if err != nil {
		return nil, nil, err
	}
	// Build name→fieldID map from the constructed top-level fields.
	fieldIDs := make(map[string]int, len(nested))
	for _, f := range nested {
		fieldIDs[f.Name] = f.ID
	}
	return iceberg.NewSchema(0, nested...), fieldIDs, nil
}

// buildNestedFields converts []ddl.Field to []iceberg.NestedField using the shared counter.
func buildNestedFields(fields []ddl.Field, ctr *idCounter) ([]iceberg.NestedField, error) {
	result := make([]iceberg.NestedField, 0, len(fields))
	for _, f := range fields {
		iType, err := toIcebergType(f.Type, ctr)
		if err != nil {
			return nil, errors.WithMessagef(err, "field %s", f.Name)
		}
		result = append(result, iceberg.NestedField{
			ID:       ctr.next(),
			Name:     f.Name,
			Type:     iType,
			Required: f.Required,
			Doc:      f.Doc,
		})
	}
	return result, nil
}

// toIcebergType maps a ddl.IcebergType to an iceberg.Type recursively.
// ctr is used to assign IDs to nested struct/list/map type fields.
//
//nolint:ireturn // returns iceberg.Type interface by design; iceberg-go API requires interface values
func toIcebergType(t ddl.IcebergType, ctr *idCounter) (iceberg.Type, error) {
	switch t.Kind {
	case ddl.Boolean:
		return iceberg.BooleanType{}, nil
	case ddl.Int:
		return iceberg.Int32Type{}, nil
	case ddl.Long:
		return iceberg.Int64Type{}, nil
	case ddl.Float:
		return iceberg.Float32Type{}, nil
	case ddl.Double:
		return iceberg.Float64Type{}, nil
	case ddl.Decimal:
		return iceberg.DecimalTypeOf(t.Prec, t.Scale), nil
	case ddl.Date:
		return iceberg.DateType{}, nil
	case ddl.Time:
		return iceberg.TimeType{}, nil
	case ddl.TimestampTz:
		return iceberg.TimestampTzType{}, nil
	case ddl.Timestamp:
		return iceberg.TimestampType{}, nil
	case ddl.String:
		return iceberg.StringType{}, nil
	case ddl.UUID:
		return iceberg.UUIDType{}, nil
	case ddl.Binary:
		return iceberg.BinaryType{}, nil

	case ddl.Struct:
		nested, err := buildNestedFields(t.Fields, ctr)
		if err != nil {
			return nil, err
		}
		return &iceberg.StructType{FieldList: nested}, nil

	case ddl.List:
		if t.Elem == nil {
			return nil, errors.New("list type has no element type")
		}
		elemType, err := toIcebergType(*t.Elem, ctr)
		if err != nil {
			return nil, errors.WithMessage(err, "list element")
		}
		return &iceberg.ListType{
			ElementID:       ctr.next(),
			Element:         elemType,
			ElementRequired: false,
		}, nil

	case ddl.Map:
		if t.Key == nil || t.Val == nil {
			return nil, errors.New("map type missing key or value type")
		}
		keyType, err := toIcebergType(*t.Key, ctr)
		if err != nil {
			return nil, errors.WithMessage(err, "map key")
		}
		valType, err := toIcebergType(*t.Val, ctr)
		if err != nil {
			return nil, errors.WithMessage(err, "map value")
		}
		return &iceberg.MapType{
			KeyID:     ctr.next(),
			KeyType:   keyType,
			ValueID:   ctr.next(),
			ValueType: valType,
		}, nil

	default:
		return nil, errors.Errorf("unsupported type kind: %d", t.Kind)
	}
}

// ─── partition spec building ─────────────────────────────────────────────────

// buildPartitionSpec converts a []ddl.PartitionField and a name→fieldID map
// (produced by buildSchema) into an *iceberg.PartitionSpec.
// Using the actual field IDs from buildSchema ensures correctness when nested
// types (struct/list/map) precede the partitioned column and consume IDs.
func buildPartitionSpec(fieldIDs map[string]int, partitions []ddl.PartitionField) (*iceberg.PartitionSpec, error) {
	fields := make([]iceberg.PartitionField, 0, len(partitions))
	for idx, p := range partitions {
		transform, err := toIcebergTransform(p.Transform, p.Param)
		if err != nil {
			return nil, errors.WithMessagef(err, "partition field %d (%s)", idx, p.SourceCol)
		}

		sourceID, ok := fieldIDs[p.SourceCol]
		if !ok {
			return nil, errors.Errorf("partition source column %q not found in schema", p.SourceCol)
		}

		name := p.Name
		if name == "" {
			name = fmt.Sprintf("%s_%s", p.SourceCol, transform.String())
		}

		fields = append(fields, iceberg.PartitionField{
			SourceIDs: []int{sourceID},
			FieldID:   partitionFieldIDBase + idx, // standard Iceberg partition field ID base
			Name:      name,
			Transform: transform,
		})
	}

	spec := iceberg.NewPartitionSpec(fields...)
	return &spec, nil
}

// findPartitionFieldName resolves the actual partition field name from the current
// PartitionSpec by matching (source column ID + transform). This is required for
// DropPartitionField because RemoveField accepts the partition field's Name, not
// the source column name. ADD PARTITION FIELD uses the convention
// "<sourceCol>_<transform.String()>" (e.g. "gid_bucket[16]"), but tables created
// via CREATE TABLE PARTITIONED BY may use a different naming convention, so we
// always look up the real name from the live spec rather than reconstructing it.
//
// Lookup priority:
//  1. If op.Name is set, match by Name directly (caller overrides).
//  2. Otherwise, find a field in spec.FieldsBySourceID(srcFieldID) whose
//     Transform.String() matches the desired transform.
//
// Returns an error if the source column is not found in the schema, or if no
// matching partition field is found in the current spec.
func findPartitionFieldName(schema *iceberg.Schema, spec iceberg.PartitionSpec, op *ddl.PartitionField) (string, error) {
	// Priority 1: caller provided an explicit partition field name.
	if op.Name != "" {
		return op.Name, nil
	}

	// Resolve source column ID from schema.
	srcField, ok := schema.FindFieldByName(op.SourceCol)
	if !ok {
		return "", errors.Errorf("source column %q not found in table schema", op.SourceCol)
	}
	srcFieldID := srcField.ID

	// Build the desired transform for comparison.
	wantTransform, err := toIcebergTransform(op.Transform, op.Param)
	if err != nil {
		return "", errors.WithMessagef(err, "resolve transform for %s", op.SourceCol)
	}
	wantStr := wantTransform.String()

	// Search partition fields with matching source column ID and transform.
	for _, pf := range spec.FieldsBySourceID(srcFieldID) {
		if pf.Transform.String() == wantStr {
			return pf.Name, nil
		}
	}

	return "", errors.Errorf("partition field for %s(%s) not found in current spec", wantStr, op.SourceCol)
}

// toIcebergTransform maps a ddl.TransformKind (and optional numeric param) to an iceberg.Transform.
//
//nolint:ireturn // returns iceberg.Transform interface by design; iceberg-go API requires interface values
func toIcebergTransform(kind ddl.TransformKind, param int) (iceberg.Transform, error) {
	switch kind {
	case ddl.Identity:
		return iceberg.IdentityTransform{}, nil
	case ddl.Years:
		return iceberg.YearTransform{}, nil
	case ddl.Months:
		return iceberg.MonthTransform{}, nil
	case ddl.Days:
		return iceberg.DayTransform{}, nil
	case ddl.Hours:
		return iceberg.HourTransform{}, nil
	case ddl.Bucket:
		return iceberg.BucketTransform{NumBuckets: param}, nil
	case ddl.Truncate:
		return iceberg.TruncateTransform{Width: param}, nil
	default:
		return nil, errors.Errorf("unsupported transform kind: %d", kind)
	}
}

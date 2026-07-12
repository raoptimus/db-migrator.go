package ddl

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// defaultDecimalPrecision is the Iceberg default precision for DECIMAL with no arguments.
const defaultDecimalPrecision = 38

// parseType parses a Spark-SQL type expression into an IcebergType.
// Supported primitives (case-insensitive): boolean, int, long, bigint, float, double, decimal(p,s),
// date, time, timestamp, timestamp_ntz, string, uuid, binary.
// Supported composite types (recursive): struct<name:type,...>, array<type>, list<type>, map<key,val>.
//
// Subset v1 constraints — the following are intentionally NOT supported:
//   - NOT NULL / nullability constraints: Field.Required stays zero-value (false); writing
//     "col long NOT NULL" in a column definition will return ErrParse.
//   - CHAR(n) / NVARCHAR(n) / VARCHAR(n) and other SQL-Server/MySQL type aliases.
func parseType(s string) (IcebergType, error) {
	s = strings.TrimSpace(s)
	upper := strings.ToUpper(s)

	switch {
	case upper == "BOOLEAN":
		return IcebergType{Kind: Boolean}, nil
	case upper == "INT" || upper == "INTEGER":
		return IcebergType{Kind: Int}, nil
	case upper == "LONG" || upper == "BIGINT":
		return IcebergType{Kind: Long}, nil
	case upper == "FLOAT":
		return IcebergType{Kind: Float}, nil
	case upper == "DOUBLE":
		return IcebergType{Kind: Double}, nil
	case upper == "DATE":
		return IcebergType{Kind: Date}, nil
	case upper == "TIME":
		return IcebergType{Kind: Time}, nil
	case upper == "TIMESTAMP":
		// Spark TIMESTAMP → Iceberg timestamptz (with timezone, UTC) — ФТ-12
		return IcebergType{Kind: TimestampTz}, nil
	case upper == "TIMESTAMP_NTZ":
		// Spark TIMESTAMP_NTZ → Iceberg timestamp (without timezone) — ФТ-12
		return IcebergType{Kind: Timestamp}, nil
	case upper == "STRING":
		return IcebergType{Kind: String}, nil
	case upper == "UUID":
		return IcebergType{Kind: UUID}, nil
	case upper == "BINARY":
		return IcebergType{Kind: Binary}, nil
	case strings.HasPrefix(upper, "DECIMAL"):
		return parseDecimal(s)
	case strings.HasPrefix(upper, "STRUCT<"):
		return parseStructType(s)
	case strings.HasPrefix(upper, "ARRAY<"):
		return parseArrayType(s)
	case strings.HasPrefix(upper, "LIST<"):
		// list<T> is a synonym for array<T> — both map to IcebergType{Kind: List}.
		return parseListType(s)
	case strings.HasPrefix(upper, "MAP<"):
		return parseMapType(s)
	default:
		return IcebergType{}, errors.Wrapf(ErrParse, "unknown type %q", s)
	}
}

// parseDecimal parses DECIMAL(p,s) or DECIMAL(p).
func parseDecimal(s string) (IcebergType, error) {
	upper := strings.ToUpper(s)
	if !strings.HasPrefix(upper, "DECIMAL") {
		return IcebergType{}, errors.Wrapf(ErrParse, "invalid decimal type %q", s)
	}
	inner := strings.TrimSpace(s[len("DECIMAL"):])
	if inner == "" {
		// DECIMAL without parameters — use defaults
		return IcebergType{Kind: Decimal, Prec: defaultDecimalPrecision, Scale: 0}, nil
	}
	if !strings.HasPrefix(inner, "(") || !strings.HasSuffix(inner, ")") {
		return IcebergType{}, errors.Wrapf(ErrParse, "invalid decimal type %q: expected DECIMAL(p,s)", s)
	}
	inner = inner[1 : len(inner)-1]
	const maxParts = 2
	parts := strings.SplitN(inner, ",", maxParts)
	prec, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return IcebergType{}, errors.Wrapf(ErrParse, "invalid decimal precision in %q", s)
	}
	scale := 0
	if len(parts) == maxParts {
		scale, err = strconv.Atoi(strings.TrimSpace(parts[1]))
		if err != nil {
			return IcebergType{}, errors.Wrapf(ErrParse, "invalid decimal scale in %q", s)
		}
	}
	return IcebergType{Kind: Decimal, Prec: prec, Scale: scale}, nil
}

// parseStructType parses STRUCT<name:type, name:type, ...>.
func parseStructType(s string) (IcebergType, error) {
	upper := strings.ToUpper(s)
	if !strings.HasPrefix(upper, "STRUCT<") {
		return IcebergType{}, errors.Wrapf(ErrParse, "invalid struct type %q", s)
	}
	// Find the matching closing >
	inner, err := extractAngleBracket(s[len("STRUCT"):])
	if err != nil {
		return IcebergType{}, errors.Wrapf(ErrParse, "invalid struct type %q: %v", s, err)
	}
	parts, err := splitTopLevel(inner, ',')
	if err != nil {
		return IcebergType{}, errors.Wrapf(ErrParse, "invalid struct type %q: %v", s, err)
	}
	fields := make([]Field, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		colonIdx := strings.Index(part, ":")
		if colonIdx < 0 {
			return IcebergType{}, errors.Wrapf(ErrParse, "invalid struct field %q in %q", part, s)
		}
		name := strings.TrimSpace(part[:colonIdx])
		typStr := strings.TrimSpace(part[colonIdx+1:])
		ft, err := parseType(typStr)
		if err != nil {
			return IcebergType{}, errors.Wrapf(err, "in struct field %q", name)
		}
		fields = append(fields, Field{Name: name, Type: ft})
	}
	return IcebergType{Kind: Struct, Fields: fields}, nil
}

// parseArrayType parses ARRAY<element_type>.
func parseArrayType(s string) (IcebergType, error) {
	upper := strings.ToUpper(s)
	if !strings.HasPrefix(upper, "ARRAY<") {
		return IcebergType{}, errors.Wrapf(ErrParse, "invalid array type %q", s)
	}
	inner, err := extractAngleBracket(s[len("ARRAY"):])
	if err != nil {
		return IcebergType{}, errors.Wrapf(ErrParse, "invalid array type %q: %v", s, err)
	}
	elem, err := parseType(inner)
	if err != nil {
		return IcebergType{}, errors.Wrapf(err, "in array element type")
	}
	return IcebergType{Kind: List, Elem: &elem}, nil
}

// parseListType parses LIST<element_type>. list<T> is a synonym for array<T> in Spark SQL
// and maps to IcebergType{Kind: List} — the same as ARRAY<T>.
func parseListType(s string) (IcebergType, error) {
	upper := strings.ToUpper(s)
	if !strings.HasPrefix(upper, "LIST<") {
		return IcebergType{}, errors.Wrapf(ErrParse, "invalid list type %q", s)
	}
	inner, err := extractAngleBracket(s[len("LIST"):])
	if err != nil {
		return IcebergType{}, errors.Wrapf(ErrParse, "invalid list type %q: %v", s, err)
	}
	elem, err := parseType(inner)
	if err != nil {
		return IcebergType{}, errors.Wrapf(err, "in list element type")
	}
	return IcebergType{Kind: List, Elem: &elem}, nil
}

// parseMapType parses MAP<key_type, value_type>.
func parseMapType(s string) (IcebergType, error) {
	upper := strings.ToUpper(s)
	if !strings.HasPrefix(upper, "MAP<") {
		return IcebergType{}, errors.Wrapf(ErrParse, "invalid map type %q", s)
	}
	inner, err := extractAngleBracket(s[len("MAP"):])
	if err != nil {
		return IcebergType{}, errors.Wrapf(ErrParse, "invalid map type %q: %v", s, err)
	}
	const mapTypeArgs = 2
	parts, err := splitTopLevel(inner, ',')
	if err != nil {
		return IcebergType{}, errors.Wrapf(ErrParse, "invalid map type %q: %v", s, err)
	}
	if len(parts) != mapTypeArgs {
		return IcebergType{}, errors.Wrapf(
			ErrParse, "map type requires exactly 2 type arguments, got %d in %q", len(parts), s,
		)
	}
	keyT, err := parseType(strings.TrimSpace(parts[0]))
	if err != nil {
		return IcebergType{}, errors.Wrapf(err, "in map key type")
	}
	valT, err := parseType(strings.TrimSpace(parts[1]))
	if err != nil {
		return IcebergType{}, errors.Wrapf(err, "in map value type")
	}
	return IcebergType{Kind: Map, Key: &keyT, Val: &valT}, nil
}

// extractAngleBracket extracts the content inside the first pair of < > from the start of s.
// s must start with '<'. Returns the content (without the brackets).
func extractAngleBracket(s string) (string, error) {
	s = strings.TrimSpace(s)
	if len(s) == 0 || s[0] != '<' {
		return "", fmt.Errorf("expected '<', got %q", s)
	}
	depth := 0
	for i, ch := range s {
		switch ch {
		case '<':
			depth++
		case '>':
			depth--
			if depth == 0 {
				return s[1:i], nil
			}
		}
	}
	return "", fmt.Errorf("unmatched '<' in %q", s)
}

// splitTopLevel splits s by sep, but only at the top level (not inside < > brackets or ( ) parens).
func splitTopLevel(s string, sep rune) ([]string, error) {
	var parts []string
	depth := 0
	start := 0
	for i, ch := range s {
		switch ch {
		case '<', '(':
			depth++
		case '>', ')':
			depth--
			if depth < 0 {
				return nil, fmt.Errorf("unmatched closing bracket in %q", s)
			}
		case sep:
			if depth == 0 {
				parts = append(parts, s[start:i])
				start = i + 1
			}
		}
	}
	parts = append(parts, s[start:])
	return parts, nil
}

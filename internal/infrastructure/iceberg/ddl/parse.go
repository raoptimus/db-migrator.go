// Package ddl provides a hand-written parser for the Spark-SQL (Iceberg) DDL subset v1.
// It is intentionally pure: no network, no catalog, no filesystem, no random — only string → IR.
package ddl

import (
	"strings"

	"github.com/pkg/errors"
)

// SQL keyword constants used in multiple places across the parser.
const (
	kwDROP      = "DROP"
	kwRENAME    = "RENAME"
	kwALTER     = "ALTER"
	kwNAMESPACE = "NAMESPACE"
	kwTABLE     = "TABLE"
	kwCOMMENT   = "COMMENT"
	kwCOLUMN    = "COLUMN"
	kwPARTITION = "PARTITION"
)

// Parse parses a single Spark-SQL (Iceberg) DDL statement into an Operation.
//
// catalog is the warehouse name from the DSN (e.g. "iceberg") used to strip the leading
// catalog segment from qualified identifiers. Pass an empty string to disable stripping.
//
// The statement must be a single DDL statement; the caller (sqlio.Scanner) is responsible
// for splitting multi-statement files on ";".
func Parse(catalog, statement string) (Operation, error) {
	stmt := strings.TrimSpace(statement)
	// Strip SQL comments before any further processing.
	// This must happen before TrimRight(";") so that a trailing comment like
	// "CREATE NAMESPACE raw -- note;" does not leave a spurious semicolon.
	stmt = stripComments(stmt)
	// Strip trailing semicolon, if any.
	stmt = strings.TrimRight(stmt, ";")
	stmt = strings.TrimSpace(stmt)

	if stmt == "" {
		return Operation{}, errors.Wrapf(ErrParse, "empty statement")
	}

	// Tokenise at the word level; the parser is top-down and uses a token cursor.
	p := newParser(catalog, stmt)
	return p.parse()
}

// stripComments removes SQL line comments (-- … <EOL>) and block comments (/* … */) from s,
// while preserving the content of single-quoted string literals ('…') and backtick-quoted
// identifiers (`…`).  The function is intentionally conservative: it only skips what it
// recognises as a comment delimiter that is NOT inside a string or backtick literal.
func stripComments(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	i := 0
	for i < len(s) {
		// Single-quoted string literal — copy verbatim (includes any '--' or '/*' inside).
		if s[i] == '\'' {
			end, tok := scanStringLiteral(s, i)
			b.WriteString(tok)
			i = end
			continue
		}

		// Backtick-quoted identifier — copy verbatim.
		if s[i] == '`' {
			j := i + 1
			for j < len(s) && s[j] != '`' {
				j++
			}
			if j < len(s) {
				j++ // include closing backtick
			}
			b.WriteString(s[i:j])
			i = j
			continue
		}

		// Line comment: -- … <newline>
		if i+1 < len(s) && s[i] == '-' && s[i+1] == '-' {
			// Skip until end of line (keep the newline so multi-line statements stay intact).
			for i < len(s) && s[i] != '\n' {
				i++
			}
			continue
		}

		// Block comment: /* … */
		if i+1 < len(s) && s[i] == '/' && s[i+1] == '*' {
			i += 2 // skip '/*'
			for i < len(s) {
				if i+1 < len(s) && s[i] == '*' && s[i+1] == '/' {
					i += 2 // skip '*/'
					break
				}
				i++
			}
			continue
		}

		b.WriteByte(s[i])
		i++
	}
	return b.String()
}

// ─��─ parser ────────────────────────────────────────────────────────────────────

type parser struct {
	catalog string // warehouse name for catalog-prefix stripping
	stmt    string // original statement (for error messages)
	tokens  []string
	pos     int
}

func newParser(catalog, stmt string) *parser {
	return &parser{
		catalog: strings.ToLower(strings.TrimSpace(catalog)),
		stmt:    stmt,
		tokens:  tokenize(stmt),
		pos:     0,
	}
}

// parse dispatches on the first token.
func (p *parser) parse() (Operation, error) {
	kw, ok := p.peek()
	if !ok {
		return Operation{}, errors.Wrapf(ErrParse, "empty statement")
	}
	switch strings.ToUpper(kw) {
	case "CREATE":
		return p.parseCreate()
	case kwDROP:
		return p.parseDrop()
	case kwRENAME:
		return p.parseRename()
	case kwALTER:
		return p.parseAlter()
	default:
		return Operation{}, errors.Wrapf(ErrUnsupportedDDL, "statement starts with %q: %s", kw, p.stmt)
	}
}

// ─── CREATE ────────────────────────────────────────────────────────────────────

func (p *parser) parseCreate() (Operation, error) {
	p.mustConsume("CREATE")
	next, ok := p.peek()
	if !ok {
		return Operation{}, errors.Wrapf(ErrParse, "unexpected end after CREATE")
	}
	switch strings.ToUpper(next) {
	case kwNAMESPACE:
		return p.parseCreateNamespace()
	case kwTABLE:
		return p.parseCreateTable()
	default:
		return Operation{}, errors.Wrapf(ErrUnsupportedDDL, "CREATE %s is not supported: %s", next, p.stmt)
	}
}

// consumeIfNotExists consumes an optional "IF NOT EXISTS" clause and reports whether it was present.
func (p *parser) consumeIfNotExists() (bool, error) {
	if !p.peekUpperIs("IF") {
		return false, nil
	}
	p.consume()
	if err := p.expectConsume("NOT"); err != nil {
		return false, err
	}
	if err := p.expectConsume("EXISTS"); err != nil {
		return false, err
	}
	return true, nil
}

// consumeIfExists consumes an optional "IF EXISTS" clause and reports whether it was present.
func (p *parser) consumeIfExists() (bool, error) {
	if !p.peekUpperIs("IF") {
		return false, nil
	}
	p.consume()
	if err := p.expectConsume("EXISTS"); err != nil {
		return false, err
	}
	return true, nil
}

func (p *parser) parseCreateNamespace() (Operation, error) {
	p.mustConsume(kwNAMESPACE)
	ifNotExists, err := p.consumeIfNotExists()
	if err != nil {
		return Operation{}, err
	}
	ns, err := p.parseNamespaceIdent()
	if err != nil {
		return Operation{}, err
	}
	op := Operation{
		Kind:        CreateNamespace,
		Table:       Ident{Namespace: ns},
		IfNotExists: ifNotExists,
	}
	// Remaining tokens may be PROPERTIES (ignore for now, not in subset v1).
	return op, nil
}

func (p *parser) parseCreateTable() (Operation, error) {
	p.mustConsume(kwTABLE)
	ifNotExists, err := p.consumeIfNotExists()
	if err != nil {
		return Operation{}, err
	}
	id, err := p.parseIdent()
	if err != nil {
		return Operation{}, err
	}

	spec := &CreateTableSpec{
		Props: make(map[string]string),
	}

	// Expect '(' column_list ')'
	if err := p.expectConsume("("); err != nil {
		return Operation{}, errors.Wrapf(ErrParse, "CREATE TABLE: expected '(' after table name: %s", p.stmt)
	}
	schema, err := p.parseColumnList()
	if err != nil {
		return Operation{}, err
	}
	spec.Schema = schema

	// Parse optional clauses: USING, PARTITIONED BY, COMMENT, TBLPROPERTIES
	for {
		kw, ok := p.peek()
		if !ok {
			break
		}
		switch strings.ToUpper(kw) {
		case "USING":
			p.consume() // USING
			p.consume() // format (e.g. iceberg) — accepted and ignored
		case "PARTITIONED":
			p.consume() // PARTITIONED
			if err := p.expectConsume("BY"); err != nil {
				return Operation{}, err
			}
			if err := p.expectConsume("("); err != nil {
				return Operation{}, errors.Wrapf(ErrParse, "PARTITIONED BY: expected '(': %s", p.stmt)
			}
			transforms, err := p.parseTransformList()
			if err != nil {
				return Operation{}, err
			}
			spec.Partition = transforms
		case kwCOMMENT:
			p.consume() // COMMENT
			comment, err := p.parseStringLiteral()
			if err != nil {
				return Operation{}, err
			}
			spec.Comment = comment
		case "TBLPROPERTIES":
			p.consume() // TBLPROPERTIES
			props, err := p.parseProperties()
			if err != nil {
				return Operation{}, err
			}
			spec.Props = props
		default:
			// Unknown trailing token — stop (don't error; forward-compatible)
			goto done
		}
	}
done:
	return Operation{
		Kind:        CreateTable,
		Table:       id,
		Create:      spec,
		IfNotExists: ifNotExists,
	}, nil
}

// parseColumnList parses "colname TYPE [COMMENT '…'], …)" stopping at the closing ')'.
//
// Subset v1 constraints:
//   - NOT NULL constraints and Field.Required are outside subset v1; Required stays zero-value (false).
//     Writing "id long NOT NULL" will produce a parse error because NOT NULL is not consumed
//     as part of the type expression.
//   - Empty column lists (no fields) are also outside subset v1.
func (p *parser) parseColumnList() ([]Field, error) {
	var fields []Field
	for {
		// Check for closing paren
		if p.peekIs(")") {
			p.consume()
			break
		}
		name, ok := p.consume()
		if !ok {
			return nil, errors.Wrapf(ErrParse, "CREATE TABLE: unexpected end while parsing column list: %s", p.stmt)
		}
		// Strip backtick quoting if present
		name = unquote(name)

		// Collect type tokens up to COMMENT, ',', or ')'
		typeStr, err := p.collectTypeString()
		if err != nil {
			return nil, err
		}

		ft, err := parseType(typeStr)
		if err != nil {
			return nil, errors.Wrapf(err, "column %q", name)
		}

		doc := ""
		if p.peekUpperIs(kwCOMMENT) {
			p.consume() // COMMENT
			doc, err = p.parseStringLiteral()
			if err != nil {
				return nil, err
			}
		}

		fields = append(fields, Field{Name: name, Type: ft, Doc: doc})

		// Skip optional comma
		if p.peekIs(",") {
			p.consume()
		}
	}
	return fields, nil
}

// collectTypeString reads tokens that form a type expression, handling nested <...> and (...).
// It stops at COMMENT, ',', ')' at depth 0.
func (p *parser) collectTypeString() (string, error) {
	var parts []string
	depth := 0 // tracks angle/paren nesting
	for {
		tok, ok := p.peek()
		if !ok {
			break
		}

		// At depth 0, stop on structural tokens
		if depth == 0 && (tok == "," || tok == ")" || strings.EqualFold(tok, kwCOMMENT)) {
			break
		}

		p.consume()
		parts = append(parts, tok)

		// Track nesting depth (using simplified char-level scan on the token itself)
		for _, ch := range tok {
			switch ch {
			case '<', '(':
				depth++
			case '>', ')':
				if depth > 0 {
					depth--
				}
			}
		}
	}
	if len(parts) == 0 {
		return "", errors.Wrapf(ErrParse, "expected type expression: %s", p.stmt)
	}
	return strings.Join(parts, ""), nil
}

// parseTransformList parses "transform1, transform2, …)" stopping at ')'.
func (p *parser) parseTransformList() ([]PartitionField, error) {
	var result []PartitionField
	for {
		if p.peekIs(")") {
			p.consume()
			break
		}
		// Collect a single transform expression (up to matching ')' for the transform, then ',' or ')')
		expr := p.collectTransformExpr()
		pf, err := parseTransform(expr)
		if err != nil {
			return nil, err
		}
		result = append(result, pf)
		if p.peekIs(",") {
			p.consume()
		}
	}
	return result, nil
}

// collectTransformExpr reads tokens for a single partition transform expression.
// It reads until it hits a ',' or ')' at depth 0, collecting the inner '(...)'.
func (p *parser) collectTransformExpr() string {
	var parts []string
	depth := 0
	for {
		tok, ok := p.peek()
		if !ok {
			break
		}
		if depth == 0 && (tok == "," || tok == ")") {
			break
		}
		p.consume()
		parts = append(parts, tok)
		for _, ch := range tok {
			switch ch {
			case '(':
				depth++
			case ')':
				depth--
			}
		}
		if depth == 0 && len(parts) > 0 && strings.HasSuffix(tok, ")") {
			// We've finished one complete transform expression like "days(ts)"
			break
		}
	}
	return strings.Join(parts, "")
}

// parseProperties parses ( 'key' = 'value', … )
func (p *parser) parseProperties() (map[string]string, error) {
	if err := p.expectConsume("("); err != nil {
		return nil, errors.Wrapf(ErrParse, "TBLPROPERTIES: expected '(': %s", p.stmt)
	}
	props := make(map[string]string)
	for {
		if p.peekIs(")") {
			p.consume()
			break
		}
		key, err := p.parseStringLiteral()
		if err != nil {
			return nil, errors.Wrapf(err, "TBLPROPERTIES key")
		}
		if err := p.expectConsume("="); err != nil {
			return nil, errors.Wrapf(ErrParse, "TBLPROPERTIES: expected '=' after key: %s", p.stmt)
		}
		val, err := p.parseStringLiteral()
		if err != nil {
			return nil, errors.Wrapf(err, "TBLPROPERTIES value")
		}
		props[key] = val
		if p.peekIs(",") {
			p.consume()
		}
	}
	return props, nil
}

// ─── DROP ──────────────────────────────────────────────────────────────────────

func (p *parser) parseDrop() (Operation, error) {
	p.mustConsume(kwDROP)
	next, ok := p.peek()
	if !ok {
		return Operation{}, errors.Wrapf(ErrParse, "unexpected end after DROP")
	}
	switch strings.ToUpper(next) {
	case kwNAMESPACE:
		return p.parseDropNamespace()
	case kwTABLE:
		return p.parseDropTable()
	default:
		return Operation{}, errors.Wrapf(ErrUnsupportedDDL, "DROP %s is not supported: %s", next, p.stmt)
	}
}

func (p *parser) parseDropNamespace() (Operation, error) {
	p.mustConsume(kwNAMESPACE)
	ifExists, err := p.consumeIfExists()
	if err != nil {
		return Operation{}, err
	}
	ns, err := p.parseNamespaceIdent()
	if err != nil {
		return Operation{}, err
	}
	return Operation{
		Kind:     DropNamespace,
		Table:    Ident{Namespace: ns},
		IfExists: ifExists,
	}, nil
}

func (p *parser) parseDropTable() (Operation, error) {
	p.mustConsume(kwTABLE)
	ifExists, err := p.consumeIfExists()
	if err != nil {
		return Operation{}, err
	}
	id, err := p.parseIdent()
	if err != nil {
		return Operation{}, err
	}
	return Operation{Kind: DropTable, Table: id, IfExists: ifExists}, nil
}

// ─── RENAME ────────────────────────────────────────────────────────────────────

func (p *parser) parseRename() (Operation, error) {
	p.mustConsume(kwRENAME)
	next, ok := p.peek()
	if !ok {
		return Operation{}, errors.Wrapf(ErrParse, "unexpected end after RENAME")
	}
	if !strings.EqualFold(next, kwTABLE) {
		return Operation{}, errors.Wrapf(ErrUnsupportedDDL, "RENAME %s is not supported: %s", next, p.stmt)
	}
	p.mustConsume(kwTABLE)
	src, err := p.parseIdent()
	if err != nil {
		return Operation{}, err
	}
	if err := p.expectConsume("TO"); err != nil {
		return Operation{}, errors.Wrapf(ErrParse, "RENAME TABLE: expected TO: %s", p.stmt)
	}
	dst, err := p.parseIdent()
	if err != nil {
		return Operation{}, err
	}
	return Operation{Kind: RenameTable, Table: src, RenameTo: &dst}, nil
}

// ─── ALTER ─────────────────────────────────────────────────────────────────────

func (p *parser) parseAlter() (Operation, error) {
	p.mustConsume(kwALTER)
	next, ok := p.peek()
	if !ok {
		return Operation{}, errors.Wrapf(ErrParse, "unexpected end after ALTER")
	}
	if !strings.EqualFold(next, kwTABLE) {
		return Operation{}, errors.Wrapf(ErrUnsupportedDDL, "ALTER %s is not supported: %s", next, p.stmt)
	}
	p.mustConsume(kwTABLE)

	id, err := p.parseIdent()
	if err != nil {
		return Operation{}, err
	}

	sub, ok := p.peek()
	if !ok {
		return Operation{}, errors.Wrapf(ErrParse, "ALTER TABLE: unexpected end after table name: %s", p.stmt)
	}
	switch strings.ToUpper(sub) {
	case "ADD":
		return p.parseAlterAdd(id)
	case kwDROP:
		return p.parseAlterDrop(id)
	case kwRENAME:
		return p.parseAlterRename(id)
	case kwALTER:
		return p.parseAlterColumn(id)
	case "WRITE":
		return p.parseAlterWrite(id)
	default:
		return Operation{}, errors.Wrapf(ErrUnsupportedDDL, "ALTER TABLE … %s is not supported: %s", sub, p.stmt)
	}
}

// parseAlterWrite handles WRITE ORDERED BY … and WRITE UNORDERED (table write sort order).
func (p *parser) parseAlterWrite(id Ident) (Operation, error) {
	p.mustConsume("WRITE")
	next, ok := p.peek()
	if !ok {
		return Operation{}, errors.Wrapf(ErrParse, "ALTER TABLE … WRITE: unexpected end: %s", p.stmt)
	}
	switch strings.ToUpper(next) {
	case "UNORDERED":
		p.consume()
		return Operation{Kind: SetSortOrder, Table: id, Sort: &SortSpec{Unordered: true}}, nil
	case "ORDERED":
		p.consume()
		if err := p.expectConsume("BY"); err != nil {
			return Operation{}, err
		}
		fields, err := p.parseSortFieldList()
		if err != nil {
			return Operation{}, err
		}
		return Operation{Kind: SetSortOrder, Table: id, Sort: &SortSpec{Fields: fields}}, nil
	default:
		return Operation{}, errors.Wrapf(ErrUnsupportedDDL, "ALTER TABLE … WRITE %s is not supported: %s", next, p.stmt)
	}
}

// parseSortFieldList parses a comma-separated list of sort columns following WRITE ORDERED BY.
// Spark writes the list without surrounding parentheses (WRITE ORDERED BY a, b DESC), but an
// optional wrapping "(...)" is also accepted for forgiveness.
func (p *parser) parseSortFieldList() ([]SortField, error) {
	wrapped := false
	if p.peekIs("(") {
		p.consume()
		wrapped = true
	}

	fields := make([]SortField, 0, 1)
	for {
		sf, err := p.parseSortField()
		if err != nil {
			return nil, err
		}
		fields = append(fields, sf)
		if p.peekIs(",") {
			p.consume()
			continue
		}
		break
	}

	if wrapped {
		if err := p.expectConsume(")"); err != nil {
			return nil, err
		}
	}
	if len(fields) == 0 {
		return nil, errors.Wrapf(ErrParse, "WRITE ORDERED BY: expected at least one sort column: %s", p.stmt)
	}
	return fields, nil
}

// parseSortField parses a single sort column: "<col-or-transform> [ASC|DESC] [NULLS FIRST|LAST]".
// A bare identifier maps to an Identity transform; "func(args)" reuses the partition transform parser.
// Defaults follow Iceberg: direction ASC, and null ordering NULLS FIRST for ASC / NULLS LAST for DESC.
func (p *parser) parseSortField() (SortField, error) {
	first, ok := p.peek()
	if !ok {
		return SortField{}, errors.Wrapf(ErrParse, "WRITE ORDERED BY: expected sort column: %s", p.stmt)
	}

	var pf PartitionField
	// A transform is "<name>(...)": detect it by a '(' immediately following the first token.
	if p.pos+1 < len(p.tokens) && p.tokens[p.pos+1] == "(" {
		expr, err := p.collectTransformExprFull()
		if err != nil {
			return SortField{}, err
		}
		pf, err = parseTransform(expr)
		if err != nil {
			return SortField{}, err
		}
	} else {
		p.consume() // plain column
		pf = PartitionField{Transform: Identity, SourceCol: first}
	}

	direction := SortAsc
	switch {
	case p.peekUpperIs("ASC"):
		p.consume()
	case p.peekUpperIs("DESC"):
		p.consume()
		direction = SortDesc
	}

	// Iceberg default null ordering depends on the direction.
	nullOrder := NullsFirst
	if direction == SortDesc {
		nullOrder = NullsLast
	}
	if p.peekUpperIs("NULLS") {
		p.consume()
		switch {
		case p.peekUpperIs("FIRST"):
			p.consume()
			nullOrder = NullsFirst
		case p.peekUpperIs("LAST"):
			p.consume()
			nullOrder = NullsLast
		default:
			return SortField{}, errors.Wrapf(ErrParse, "WRITE ORDERED BY: expected FIRST or LAST after NULLS: %s", p.stmt)
		}
	}

	return SortField{
		Transform: pf.Transform,
		Param:     pf.Param,
		SourceCol: pf.SourceCol,
		Direction: direction,
		NullOrder: nullOrder,
	}, nil
}

// parseAlterAdd handles ADD COLUMN … and ADD PARTITION FIELD …
func (p *parser) parseAlterAdd(id Ident) (Operation, error) {
	p.mustConsume("ADD")
	next, ok := p.peek()
	if !ok {
		return Operation{}, errors.Wrapf(ErrParse, "ALTER TABLE ADD: unexpected end: %s", p.stmt)
	}
	switch strings.ToUpper(next) {
	case kwCOLUMN:
		p.consume()
		return p.parseAddColumn(id)
	case kwPARTITION:
		p.consume()
		if err := p.expectConsume("FIELD"); err != nil {
			return Operation{}, err
		}
		return p.parseAddPartitionField(id)
	default:
		return Operation{}, errors.Wrapf(ErrUnsupportedDDL, "ALTER TABLE ADD %s is not supported: %s", next, p.stmt)
	}
}

func (p *parser) parseAddColumn(id Ident) (Operation, error) {
	name, ok := p.consume()
	if !ok {
		return Operation{}, errors.Wrapf(ErrParse, "ADD COLUMN: expected column name: %s", p.stmt)
	}
	name = unquote(name)
	typeStr, err := p.collectTypeString()
	if err != nil {
		return Operation{}, err
	}
	ft, err := parseType(typeStr)
	if err != nil {
		return Operation{}, errors.Wrapf(err, "ADD COLUMN %q", name)
	}
	doc := ""
	if p.peekUpperIs(kwCOMMENT) {
		p.consume()
		doc, err = p.parseStringLiteral()
		if err != nil {
			return Operation{}, err
		}
	}
	return Operation{
		Kind:   AddColumn,
		Table:  id,
		Column: &Field{Name: name, Type: ft, Doc: doc},
	}, nil
}

func (p *parser) parseAddPartitionField(id Ident) (Operation, error) {
	expr, err := p.collectTransformExprFull()
	if err != nil {
		return Operation{}, err
	}
	pf, err := parseTransform(expr)
	if err != nil {
		return Operation{}, err
	}
	return Operation{Kind: AddPartitionField, Table: id, Partition: &pf}, nil
}

// parseAlterDrop handles DROP COLUMN … and DROP PARTITION FIELD …
func (p *parser) parseAlterDrop(id Ident) (Operation, error) {
	p.mustConsume(kwDROP)
	next, ok := p.peek()
	if !ok {
		return Operation{}, errors.Wrapf(ErrParse, "ALTER TABLE DROP: unexpected end: %s", p.stmt)
	}
	switch strings.ToUpper(next) {
	case kwCOLUMN:
		p.consume()
		return p.parseDropColumn(id)
	case kwPARTITION:
		p.consume()
		if err := p.expectConsume("FIELD"); err != nil {
			return Operation{}, err
		}
		return p.parseDropPartitionField(id)
	default:
		return Operation{}, errors.Wrapf(ErrUnsupportedDDL, "ALTER TABLE DROP %s is not supported: %s", next, p.stmt)
	}
}

func (p *parser) parseDropColumn(id Ident) (Operation, error) {
	name, ok := p.consume()
	if !ok {
		return Operation{}, errors.Wrapf(ErrParse, "DROP COLUMN: expected column name: %s", p.stmt)
	}
	name = unquote(name)
	return Operation{
		Kind:   DropColumn,
		Table:  id,
		Column: &Field{Name: name},
	}, nil
}

func (p *parser) parseDropPartitionField(id Ident) (Operation, error) {
	expr, err := p.collectTransformExprFull()
	if err != nil {
		return Operation{}, err
	}
	pf, err := parseTransform(expr)
	if err != nil {
		return Operation{}, err
	}
	return Operation{Kind: DropPartitionField, Table: id, Partition: &pf}, nil
}

// parseAlterRename handles RENAME COLUMN a TO b
func (p *parser) parseAlterRename(id Ident) (Operation, error) {
	p.mustConsume(kwRENAME)
	next, ok := p.peek()
	if !ok {
		return Operation{}, errors.Wrapf(ErrParse, "ALTER TABLE RENAME: unexpected end: %s", p.stmt)
	}
	if !strings.EqualFold(next, kwCOLUMN) {
		return Operation{}, errors.Wrapf(ErrUnsupportedDDL, "ALTER TABLE RENAME %s is not supported: %s", next, p.stmt)
	}
	p.consume()
	oldName, ok := p.consume()
	if !ok {
		return Operation{}, errors.Wrapf(ErrParse, "RENAME COLUMN: expected old column name: %s", p.stmt)
	}
	if err := p.expectConsume("TO"); err != nil {
		return Operation{}, errors.Wrapf(ErrParse, "RENAME COLUMN: expected TO: %s", p.stmt)
	}
	newName, ok := p.consume()
	if !ok {
		return Operation{}, errors.Wrapf(ErrParse, "RENAME COLUMN: expected new column name: %s", p.stmt)
	}
	return Operation{
		Kind:    RenameColumn,
		Table:   id,
		Column:  &Field{Name: unquote(oldName)},
		NewName: unquote(newName),
	}, nil
}

// parseAlterColumn handles ALTER COLUMN name TYPE newtype
func (p *parser) parseAlterColumn(id Ident) (Operation, error) {
	p.mustConsume(kwALTER)
	next, ok := p.peek()
	if !ok {
		return Operation{}, errors.Wrapf(ErrParse, "ALTER TABLE ALTER: unexpected end: %s", p.stmt)
	}
	if !strings.EqualFold(next, kwCOLUMN) {
		return Operation{}, errors.Wrapf(ErrUnsupportedDDL, "ALTER TABLE ALTER %s is not supported: %s", next, p.stmt)
	}
	p.consume()
	colName, ok := p.consume()
	if !ok {
		return Operation{}, errors.Wrapf(ErrParse, "ALTER COLUMN: expected column name: %s", p.stmt)
	}
	colName = unquote(colName)
	if err := p.expectConsume("TYPE"); err != nil {
		return Operation{}, errors.Wrapf(ErrParse, "ALTER COLUMN: expected TYPE keyword: %s", p.stmt)
	}
	typeStr, err := p.collectTypeString()
	if err != nil {
		return Operation{}, err
	}
	ft, err := parseType(typeStr)
	if err != nil {
		return Operation{}, errors.Wrapf(err, "ALTER COLUMN %q TYPE", colName)
	}
	return Operation{
		Kind:   AlterColumnType,
		Table:  id,
		Column: &Field{Name: colName, Type: ft},
	}, nil
}

// ─── Identifier parsing ────────────────────────────────────────────────────────

// parseIdent reads a potentially dot-separated qualified identifier and strips the catalog prefix.
// The last segment becomes Table; all preceding segments (after catalog stripping) become Namespace.
// A single segment (no namespace) returns ErrNamespaceRequired.
//
// For namespace-only operations (CREATE/DROP NAMESPACE) the caller promotes id.Table into the namespace.
//
// Each dot-separated segment is unquoted individually so that backtick-quoted identifiers like
// `raw`.`t` produce Namespace=["raw"], Table="t" (not the garbled result of unquoting the join).
func (p *parser) parseIdent() (Ident, error) {
	// The next token(s) form a dotted identifier. The tokenizer may produce them as
	// separate tokens "a", ".", "b", ".", "c" or as a single token "a.b.c".
	raw, ok := p.consume()
	if !ok {
		return Ident{}, errors.Wrapf(ErrParse, "expected identifier: %s", p.stmt)
	}
	// Collect additional ".segment" tokens if the tokenizer split them.
	// Unquote each segment individually before re-joining so that backtick-quoted
	// identifiers (e.g. `raw`.`t`) are handled correctly.
	segments := []string{unquote(raw)}
	for p.peekIs(".") {
		p.consume() // consume '.'
		seg, ok := p.consume()
		if !ok {
			break
		}
		segments = append(segments, unquote(seg))
	}
	joined := strings.Join(segments, ".")
	return parseQualifiedIdent(p.catalog, joined)
}

// parseNamespaceIdent reads a namespace identifier (no table component required).
// All segments after optional catalog stripping form the namespace.
// A single bare segment like "analytics" returns Namespace=["analytics"].
//
// Each dot-separated segment is unquoted individually so that backtick-quoted identifiers
// like `iceberg`.`raw` produce Namespace=["iceberg","raw"] (not garbled text).
func (p *parser) parseNamespaceIdent() ([]string, error) {
	raw, ok := p.consume()
	if !ok {
		return nil, errors.Wrapf(ErrParse, "expected namespace identifier: %s", p.stmt)
	}
	// Collect additional ".segment" tokens if the tokenizer split them.
	// Unquote each segment individually before re-joining so that backtick-quoted
	// identifiers (e.g. `iceberg`.`raw`) are handled correctly.
	segments := []string{unquote(raw)}
	for p.peekIs(".") {
		p.consume() // consume '.'
		seg, ok := p.consume()
		if !ok {
			break
		}
		segments = append(segments, unquote(seg))
	}
	joined := strings.Join(segments, ".")
	return parseNamespaceSegments(p.catalog, joined)
}

// parseNamespaceSegments splits a dotted identifier and strips the catalog prefix.
// Returns all segments as the namespace (no table extraction).
func parseNamespaceSegments(catalog, raw string) ([]string, error) {
	segments := strings.Split(raw, ".")
	filtered := make([]string, 0, len(segments))
	for _, s := range segments {
		if s = strings.TrimSpace(s); s != "" {
			filtered = append(filtered, s)
		}
	}
	if len(filtered) == 0 {
		return nil, errors.Wrapf(ErrParse, "empty namespace identifier %q", raw)
	}
	// Strip leading catalog segment (case-insensitive)
	if catalog != "" && strings.EqualFold(filtered[0], catalog) {
		filtered = filtered[1:]
	}
	if len(filtered) == 0 {
		return nil, errors.Wrapf(ErrParse, "namespace identifier %q stripped to empty after removing catalog", raw)
	}
	return filtered, nil
}

// parseQualifiedIdent splits a dotted identifier string and applies catalog stripping.
func parseQualifiedIdent(catalog, raw string) (Ident, error) {
	// Validate that the raw identifier looks like it could be an identifier (not a punctuation char).
	if raw != "" && !isIdentStart(raw[0]) {
		return Ident{}, errors.Wrapf(ErrParse, "invalid identifier %q: not a valid name", raw)
	}
	segments := strings.Split(raw, ".")
	// Strip empty segments (can happen with leading dots)
	filtered := make([]string, 0, len(segments))
	for _, s := range segments {
		if s = strings.TrimSpace(s); s != "" {
			filtered = append(filtered, s)
		}
	}
	if len(filtered) == 0 {
		return Ident{}, errors.Wrapf(ErrParse, "empty identifier %q", raw)
	}
	// Strip leading catalog segment (case-insensitive) when it matches the warehouse name.
	if catalog != "" && strings.EqualFold(filtered[0], catalog) {
		filtered = filtered[1:]
	}
	if len(filtered) == 0 {
		return Ident{}, errors.Wrapf(ErrNamespaceRequired, "identifier %q stripped to empty after removing catalog", raw)
	}
	// Last segment is the table name; everything before is namespace.
	table := filtered[len(filtered)-1]
	ns := filtered[:len(filtered)-1]
	if len(ns) == 0 {
		return Ident{}, errors.Wrapf(ErrNamespaceRequired, "identifier %q has no namespace", raw)
	}
	return Ident{Namespace: ns, Table: table}, nil
}

// ─── Literal and token helpers ─────────────────────────────────────────────────

// parseStringLiteral consumes and returns the next token, stripping surrounding single quotes.
func (p *parser) parseStringLiteral() (string, error) {
	tok, ok := p.consume()
	if !ok {
		return "", errors.Wrapf(ErrParse, "expected string literal: %s", p.stmt)
	}
	if strings.HasPrefix(tok, "'") && strings.HasSuffix(tok, "'") && len(tok) >= 2 {
		return tok[1 : len(tok)-1], nil
	}
	return tok, nil
}

// collectTransformExprFull reads a complete "func(args)" expression.
func (p *parser) collectTransformExprFull() (string, error) {
	funcName, ok := p.consume()
	if !ok {
		return "", errors.Wrapf(ErrParse, "expected partition transform name: %s", p.stmt)
	}
	if !p.peekIs("(") {
		return "", errors.Wrapf(ErrParse, "expected '(' after transform name %q: %s", funcName, p.stmt)
	}
	// Collect everything up to and including the matching ')'
	depth := 0
	var parts []string
	parts = append(parts, funcName)
	for {
		tok, ok := p.consume()
		if !ok {
			return "", errors.Wrapf(ErrParse, "unexpected end in transform expression: %s", p.stmt)
		}
		parts = append(parts, tok)
		for _, ch := range tok {
			switch ch {
			case '(':
				depth++
			case ')':
				depth--
			}
		}
		if depth == 0 {
			break
		}
	}
	return strings.Join(parts, ""), nil
}

// ─── Token cursor ──────────────────────────────────────────────────────────────

func (p *parser) peek() (string, bool) {
	if p.pos >= len(p.tokens) {
		return "", false
	}
	return p.tokens[p.pos], true
}

func (p *parser) peekIs(s string) bool {
	tok, ok := p.peek()
	return ok && tok == s
}

func (p *parser) peekUpperIs(s string) bool {
	tok, ok := p.peek()
	return ok && strings.EqualFold(tok, s)
}

func (p *parser) consume() (string, bool) {
	if p.pos >= len(p.tokens) {
		return "", false
	}
	tok := p.tokens[p.pos]
	p.pos++
	return tok, true
}

// mustConsume consumes a token unconditionally (used when we already peeked).
func (p *parser) mustConsume(_ string) {
	_, _ = p.consume()
}

// expectConsume consumes the next token and returns an error if it doesn't match (case-insensitive).
func (p *parser) expectConsume(expected string) error {
	tok, ok := p.consume()
	if !ok {
		return errors.Wrapf(ErrParse, "expected %q but reached end of statement: %s", expected, p.stmt)
	}
	if !strings.EqualFold(tok, expected) {
		return errors.Wrapf(ErrParse, "expected %q but got %q: %s", expected, tok, p.stmt)
	}
	return nil
}

// ─── Tokenizer ─────────────────────────────────────────────────────────────────

// tokenize splits a DDL statement into tokens:
//   - identifiers (alphanumeric + underscore)
//   - single-quoted string literals (preserved with quotes)
//   - single-character punctuation: ( ) , = ;
//   - composite type expressions (STRUCT/ARRAY/MAP with angle brackets) as a single token
//
// The tokenizer is not a full SQL lexer; it covers the DDL subset needed by this parser.
func tokenize(s string) []string {
	var tokens []string
	i := 0
	for i < len(s) {
		ch := s[i]

		// Skip whitespace
		if ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' {
			i++
			continue
		}

		// Single-quoted string literal
		if ch == '\'' {
			j, tok := scanStringLiteral(s, i)
			tokens = append(tokens, tok)
			i = j
			continue
		}

		// Backtick-quoted identifier
		if ch == '`' {
			j := i + 1
			for j < len(s) && s[j] != '`' {
				j++
			}
			tokens = append(tokens, s[i:j+1])
			i = j + 1
			continue
		}

		// Punctuation (standalone single characters)
		if ch == '(' || ch == ')' || ch == ',' || ch == '=' || ch == ';' || ch == '.' {
			tokens = append(tokens, string(ch))
			i++
			continue
		}

		// Numeric literal (e.g. 10, 16, 8 in DECIMAL(p,s) or bucket/truncate params)
		if ch >= '0' && ch <= '9' {
			j := i
			for j < len(s) && s[j] >= '0' && s[j] <= '9' {
				j++
			}
			tokens = append(tokens, s[i:j])
			i = j
			continue
		}

		// Identifier / keyword / composite type expression (including angle-bracket nesting)
		if isIdentStart(ch) {
			j, tok := scanIdentOrType(s, i)
			tokens = append(tokens, tok)
			i = j
			continue
		}

		// Unknown character — skip
		i++
	}
	return tokens
}

// scanStringLiteral scans a single-quoted string literal starting at i.
// Returns the position after the closing quote and the literal token (including quotes).
func scanStringLiteral(s string, i int) (end int, tok string) {
	j := i + 1
	for j < len(s) {
		if s[j] == '\'' {
			if j+1 < len(s) && s[j+1] == '\'' {
				j += 2 // escaped single quote
				continue
			}
			break
		}
		j++
	}
	return j + 1, s[i : j+1]
}

// scanIdentOrType scans an identifier, keyword, or composite type expression starting at i.
// Composite types (STRUCT<...>, ARRAY<...>, MAP<...>) are captured as a single token
// by tracking angle-bracket depth.
func scanIdentOrType(s string, i int) (end int, tok string) {
	j := i
	depth := 0
	for j < len(s) {
		c := s[j]
		if c == '<' {
			depth++
			j++
			continue
		}
		if c == '>' {
			depth--
			j++
			if depth <= 0 {
				break
			}
			continue
		}
		if depth > 0 {
			// Inside angle brackets: consume everything including commas, colons, spaces
			j++
			continue
		}
		if isIdentChar(c) {
			j++
			continue
		}
		break
	}
	return j, s[i:j]
}

func isIdentStart(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_'
}

func isIdentChar(ch byte) bool {
	return isIdentStart(ch) || (ch >= '0' && ch <= '9')
}

// unquote strips surrounding backtick or single-quote characters from an identifier.
func unquote(s string) string {
	if len(s) >= 2 && s[0] == '`' && s[len(s)-1] == '`' {
		return s[1 : len(s)-1]
	}
	if len(s) >= 2 && s[0] == '\'' && s[len(s)-1] == '\'' {
		return s[1 : len(s)-1]
	}
	return s
}

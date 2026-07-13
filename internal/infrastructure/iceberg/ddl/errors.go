package ddl

import "github.com/pkg/errors"

// Sentinel errors for the DDL parser.
var (
	// ErrUnsupportedDDL is returned when the statement is syntactically valid but outside the supported subset v1.
	ErrUnsupportedDDL = errors.New("unsupported DDL statement")

	// ErrParse is returned when the statement cannot be parsed (syntax error).
	ErrParse = errors.New("DDL parse error")

	// ErrUnknownTransform is returned when a partition transform function is not recognized.
	ErrUnknownTransform = errors.New("unknown partition transform")

	// ErrNamespaceRequired is returned when a table identifier has no namespace component.
	ErrNamespaceRequired = errors.New("namespace is required in table identifier")
)

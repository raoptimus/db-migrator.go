/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package connection

import (
	"context"

	"github.com/pkg/errors"
	"github.com/raoptimus/db-migrator.go/internal/infrastructure/sqlex"
)

var (
	// ErrNoTransaction indicates that no transaction was found in the context.
	ErrNoTransaction = errors.New("no transaction")
)

// contextKey is a private type used for storing values in context to avoid collisions.
type contextKey int

const (
	// contextKeyTX is the context key for storing transaction instances.
	contextKeyTX contextKey = iota
)

// ContextWithTx returns a new context with the transaction stored in it.
func ContextWithTx(parent context.Context, v sqlex.Tx) context.Context {
	return context.WithValue(parent, contextKeyTX, v)
}

// TxFromContext retrieves a transaction from the context.
// It returns ErrNoTransaction if no transaction is found.
//
//nolint:ireturn,nolintlint // its ok
func TxFromContext(parent context.Context) (sqlex.Tx, error) {
	tx, ok := parent.Value(contextKeyTX).(sqlex.Tx)
	if !ok {
		return nil, ErrNoTransaction
	}

	return tx, nil
}

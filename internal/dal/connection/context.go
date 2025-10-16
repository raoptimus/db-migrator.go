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
	"github.com/raoptimus/db-migrator.go/internal/sqlex"
)

var (
	ErrNoTransaction = errors.New("no transaction")
)

type contextKey int

const (
	contextKeyTX contextKey = iota
)

func ContextWithTx(parent context.Context, v sqlex.Tx) context.Context {
	return context.WithValue(parent, contextKeyTX, v)
}

//nolint:ireturn,nolintlint // its ok
func TxFromContext(parent context.Context) (sqlex.Tx, error) {
	tx, ok := parent.Value(contextKeyTX).(sqlex.Tx)
	if !ok {
		return nil, ErrNoTransaction
	}

	return tx, nil
}

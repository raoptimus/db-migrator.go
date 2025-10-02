package sqlex

import "database/sql"

type Tx interface {
	Rollback() error
	Commit() error
}

type sqlTx struct {
	*sql.Tx
}

//nolint:ireturn,nolintlint // its ok
func NewTx(tx *sql.Tx) Tx {
	return &sqlTx{Tx: tx}
}

package tarantool

import (
	"github.com/raoptimus/db-migrator.go/internal/sqlex"
	"github.com/tarantool/go-tarantool/v2"
)

type tx struct {
	stream *tarantool.Stream
}

//nolint:ireturn,nolintlint // its ok
func NewTx(stream *tarantool.Stream) sqlex.Tx {
	return &tx{stream: stream}
}

func (tx *tx) Commit() error {
	if _, err := tx.stream.Do(tarantool.NewCommitRequest()).Get(); err != nil {
		return err
	}

	return nil
}

func (tx *tx) Rollback() error {
	if _, err := tx.stream.Do(tarantool.NewRollbackRequest()).Get(); err != nil {
		return err
	}

	return nil
}

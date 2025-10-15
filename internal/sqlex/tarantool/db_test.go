package tarantool

import (
	"bytes"
	"context"
	"testing"

	"github.com/raoptimus/db-migrator.go/internal/sqlex/tarantool/mocktarantool"
	"github.com/stretchr/testify/assert"
	"github.com/tarantool/go-tarantool/v2"
)

func TestDB_QueryContext(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		args    []any
		want    []interface{}
		wantErr bool
	}{
		{
			name:  "simple query",
			query: "return box.space.test:select()",
			args:  nil,
			want:  []interface{}{[]any{"v1", int64(123)}},
		},
		{
			name:  "query with args",
			query: "return box.space.test:select({...})",
			args:  []any{"key1"},
			want:  []interface{}{[]any{"key1", int64(456)}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			req := tarantool.NewEvalRequest(tt.query).Context(ctx)
			if len(tt.args) > 0 {
				req = req.Args(tt.args)
			}
			fut := tarantool.NewFuture(req)
			fut.SetResponse(tarantool.Header{}, bytes.NewBuffer([]byte("")))

			conn := mocktarantool.NewConnection(t)
			conn.EXPECT().
				Do(req).Return(fut)

			db := DB{conn: conn}
			_, err := db.QueryContext(ctx, tt.query, tt.args...)
			assert.NoError(t, err)
		})
	}
}

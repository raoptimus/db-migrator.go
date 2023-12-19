module github.com/raoptimus/db-migrator.go

go 1.21

require (
	github.com/ClickHouse/clickhouse-go v1.5.4
	github.com/go-sql-driver/mysql v1.7.1
	github.com/lib/pq v1.10.9
	github.com/pkg/errors v0.9.1
	github.com/raoptimus/db-migrator.go/pkg/console v0.0.0-00010101000000-000000000000
	github.com/raoptimus/db-migrator.go/pkg/iohelp v0.0.0-00010101000000-000000000000
	github.com/raoptimus/db-migrator.go/pkg/sqlio v0.0.0-00010101000000-000000000000
	github.com/raoptimus/db-migrator.go/pkg/timex v0.0.0-00010101000000-000000000000
	github.com/stretchr/testify v1.8.4
	github.com/urfave/cli/v2 v2.25.7
)

require (
	github.com/cloudflare/golz4 v0.0.0-20150217214814-ef862a3cdc58 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/stretchr/objx v0.5.1 // indirect
	github.com/xrash/smetrics v0.0.0-20201216005158-039620a65673 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	github.com/raoptimus/db-migrator.go/pkg/console => ./pkg/console
	github.com/raoptimus/db-migrator.go/pkg/iohelp => ./pkg/iohelp
	github.com/raoptimus/db-migrator.go/pkg/sqlio => ./pkg/sqlio
	github.com/raoptimus/db-migrator.go/pkg/timex => ./pkg/timex
)

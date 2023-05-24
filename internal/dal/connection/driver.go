package connection

type Driver string

const (
	DriverClickhouse Driver = "clickhouse"
	DriverMySQL      Driver = "mysql"
	DriverPostgres   Driver = "postgres"
)

func (d Driver) String() string {
	return string(d)
}

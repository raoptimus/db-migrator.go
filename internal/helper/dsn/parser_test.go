/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package dsn

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name     string
		raw      string
		expected *DSN
		wantErr  bool
	}{
		{
			name: "simple postgres DSN",
			raw:  "postgres://user:pass@localhost:5432/mydb?sslmode=disable",
			expected: &DSN{
				Driver:   "postgres",
				Hosts:    []Host{{Address: "localhost:5432", Host: "localhost", Port: "5432"}},
				Database: "mydb",
				Username: "user",
				Password: "pass",
			},
		},
		{
			name: "clickhouse cluster DSN with multiple hosts",
			raw:  "clickhouse://user:pass@host1:9000,host2:9000,host3:9000/default?compress=true",
			expected: &DSN{
				Driver: "clickhouse",
				Hosts: []Host{
					{Address: "host1:9000", Host: "host1", Port: "9000"},
					{Address: "host2:9000", Host: "host2", Port: "9000"},
					{Address: "host3:9000", Host: "host3", Port: "9000"},
				},
				Database: "default",
				Username: "user",
				Password: "pass",
			},
		},
		{
			name: "DSN without credentials",
			raw:  "tarantool://localhost:3301/testdb",
			expected: &DSN{
				Driver:   "tarantool",
				Hosts:    []Host{{Address: "localhost:3301", Host: "localhost", Port: "3301"}},
				Database: "testdb",
				Username: "",
				Password: "",
			},
		},
		{
			name: "DSN with empty password",
			raw:  "clickhouse://default:@localhost:9000/default",
			expected: &DSN{
				Driver:   "clickhouse",
				Hosts:    []Host{{Address: "localhost:9000", Host: "localhost", Port: "9000"}},
				Database: "default",
				Username: "default",
				Password: "",
			},
		},
		{
			name: "mysql DSN",
			raw:  "mysql://user:pass@tcp(localhost:3306)/mydb",
			expected: &DSN{
				Driver:   "mysql",
				Hosts:    []Host{{Address: "tcp(localhost:3306)", Host: "tcp(localhost", Port: "3306)"}},
				Database: "mydb",
				Username: "user",
				Password: "pass",
			},
		},
		{
			name: "DSN with multiple options",
			raw:  "postgres://user:pass@localhost:5432/mydb?sslmode=disable&connect_timeout=10",
			expected: &DSN{
				Driver:   "postgres",
				Hosts:    []Host{{Address: "localhost:5432", Host: "localhost", Port: "5432"}},
				Database: "mydb",
				Username: "user",
				Password: "pass",
			},
		},
		{
			name: "DSN without database",
			raw:  "postgres://user:pass@localhost:5432/",
			expected: &DSN{
				Driver:   "postgres",
				Hosts:    []Host{{Address: "localhost:5432", Host: "localhost", Port: "5432"}},
				Database: "",
				Username: "user",
				Password: "pass",
			},
		},
		{
			name:    "empty DSN",
			raw:     "",
			wantErr: true,
		},
		{
			name:    "missing driver prefix",
			raw:     "localhost:5432/mydb",
			wantErr: true,
		},
		{
			name:    "no hosts specified",
			raw:     "postgres://user:pass@/mydb",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.raw)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expected.Driver, got.Driver)
			assert.Equal(t, tt.expected.Username, got.Username)
			assert.Equal(t, tt.expected.Password, got.Password)
			assert.Equal(t, tt.expected.Database, got.Database)
			assert.Len(t, got.Hosts, len(tt.expected.Hosts))

			for i, h := range tt.expected.Hosts {
				assert.Equal(t, h.Address, got.Hosts[i].Address)
				assert.Equal(t, h.Host, got.Hosts[i].Host)
				assert.Equal(t, h.Port, got.Hosts[i].Port)
			}

			assert.Equal(t, tt.raw, got.Raw)
		})
	}
}

func TestDSN_Primary(t *testing.T) {
	tests := []struct {
		name     string
		dsn      *DSN
		expected Host
	}{
		{
			name: "single host",
			dsn: &DSN{
				Hosts: []Host{{Address: "localhost:5432", Host: "localhost", Port: "5432"}},
			},
			expected: Host{Address: "localhost:5432", Host: "localhost", Port: "5432"},
		},
		{
			name: "multiple hosts returns first",
			dsn: &DSN{
				Hosts: []Host{
					{Address: "host1:9000", Host: "host1", Port: "9000"},
					{Address: "host2:9000", Host: "host2", Port: "9000"},
				},
			},
			expected: Host{Address: "host1:9000", Host: "host1", Port: "9000"},
		},
		{
			name:     "no hosts returns empty",
			dsn:      &DSN{Hosts: nil},
			expected: Host{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.dsn.Primary()
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestDSN_PrimaryAddress(t *testing.T) {
	dsn := &DSN{
		Hosts: []Host{{Address: "localhost:5432", Host: "localhost", Port: "5432"}},
	}
	assert.Equal(t, "localhost:5432", dsn.PrimaryAddress())

	emptyDSN := &DSN{}
	assert.Equal(t, "", emptyDSN.PrimaryAddress())
}

func TestDSN_IsCluster(t *testing.T) {
	single, err := Parse("postgres://localhost:5432/db")
	require.NoError(t, err)
	assert.False(t, single.IsCluster())

	cluster, err := Parse("clickhouse://host1:9000,host2:9000/db")
	require.NoError(t, err)
	assert.True(t, cluster.IsCluster())
}

func TestDSN_HasCredentials(t *testing.T) {
	tests := []struct {
		name     string
		dsn      *DSN
		expected bool
	}{
		{
			name:     "has both username and password",
			dsn:      &DSN{Username: "user", Password: "pass"},
			expected: true,
		},
		{
			name:     "has only username",
			dsn:      &DSN{Username: "user"},
			expected: true,
		},
		{
			name:     "has only password",
			dsn:      &DSN{Password: "pass"},
			expected: true,
		},
		{
			name:     "no credentials",
			dsn:      &DSN{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.dsn.HasCredentials())
		})
	}
}

func TestMustParse(t *testing.T) {
	t.Run("valid DSN does not panic", func(t *testing.T) {
		assert.NotPanics(t, func() {
			dsn := MustParse("postgres://user:pass@localhost:5432/db")
			assert.Equal(t, "postgres", dsn.Driver)
		})
	})

	t.Run("invalid DSN panics", func(t *testing.T) {
		assert.Panics(t, func() {
			MustParse("")
		})
	})
}

func TestParse_Options(t *testing.T) {
	dsn, err := Parse("postgres://user:pass@localhost:5432/db?sslmode=disable&connect_timeout=10")
	require.NoError(t, err)

	assert.Equal(t, "disable", dsn.Options.Get("sslmode"))
	assert.Equal(t, "10", dsn.Options.Get("connect_timeout"))
}

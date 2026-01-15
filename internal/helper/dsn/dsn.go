/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package dsn

import "net/url"

// Host represents a single host endpoint in a DSN.
type Host struct {
	Address string // host:port (full address)
	Host    string // host only
	Port    string // port only
}

// DSN represents a parsed Data Source Name with support for multiple hosts (cluster mode).
type DSN struct {
	Driver   string     // database driver: clickhouse, postgres, mysql, tarantool
	Hosts    []Host     // list of hosts (supports cluster mode with multiple hosts)
	Database string     // database name
	Username string     // authentication username
	Password string     // authentication password
	Options  url.Values // query parameters
	Raw      string     // original DSN string
}

// Primary returns the primary (first) host.
// Returns empty Host if no hosts are configured.
func (d *DSN) Primary() Host {
	if len(d.Hosts) > 0 {
		return d.Hosts[0]
	}

	return Host{}
}

// PrimaryAddress returns the primary host address (host:port).
// Returns empty string if no hosts are configured.
func (d *DSN) PrimaryAddress() string {
	return d.Primary().Address
}

// IsCluster returns true if DSN contains multiple hosts.
func (d *DSN) IsCluster() bool {
	return len(d.Hosts) > 1
}

// HasCredentials returns true if DSN contains username or password.
func (d *DSN) HasCredentials() bool {
	return d.Username != "" || d.Password != ""
}

/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package dsn

import (
	"net/url"
	"strings"

	"github.com/pkg/errors"
)

// ErrInvalidDSN indicates the DSN format is invalid.
var ErrInvalidDSN = errors.New("invalid DSN format")

// Parse parses a DSN string and returns a DSN structure.
// Supports cluster format with multiple hosts: driver://user:pass@host1:port,host2:port/database?options
func Parse(raw string) (*DSN, error) {
	if raw == "" {
		return nil, ErrInvalidDSN
	}

	// Extract driver from prefix (e.g., "clickhouse" from "clickhouse://...")
	driver, rest, found := strings.Cut(raw, "://")
	if !found {
		return nil, errors.Wrap(ErrInvalidDSN, "missing driver prefix")
	}

	dsn := &DSN{
		Driver: driver,
		Raw:    raw,
	}

	// Split at '@' to separate credentials from host part
	credentialsPart, hostPart, hasAt := strings.Cut(rest, "@")
	if !hasAt {
		// No credentials, credentialsPart is actually hostPart
		hostPart = credentialsPart
		credentialsPart = ""
	}

	// Parse credentials (username:password)
	if credentialsPart != "" {
		username, password, _ := strings.Cut(credentialsPart, ":")
		dsn.Username = username
		dsn.Password = password
	}

	// Split host part from database and options
	hostAndDB, optionsStr, _ := strings.Cut(hostPart, "?")

	// Parse options (query parameters)
	if optionsStr != "" {
		var err error
		dsn.Options, err = url.ParseQuery(optionsStr)
		if err != nil {
			return nil, errors.Wrap(err, "parsing options")
		}
	} else {
		dsn.Options = make(url.Values)
	}

	// Split hosts from database path
	hostsStr, database, _ := strings.Cut(hostAndDB, "/")
	dsn.Database = database

	// Parse multiple hosts (cluster support)
	dsn.Hosts = parseHosts(hostsStr)
	if len(dsn.Hosts) == 0 {
		return nil, errors.Wrap(ErrInvalidDSN, "no hosts specified")
	}

	return dsn, nil
}

// parseHosts parses comma-separated hosts into Host structures.
func parseHosts(hostsStr string) []Host {
	if hostsStr == "" {
		return nil
	}

	hostStrs := strings.Split(hostsStr, ",")
	hosts := make([]Host, 0, len(hostStrs))

	for _, h := range hostStrs {
		h = strings.TrimSpace(h)
		if h == "" {
			continue
		}

		host, port, _ := strings.Cut(h, ":")
		hosts = append(hosts, Host{
			Address: h,
			Host:    host,
			Port:    port,
		})
	}

	return hosts
}

// MustParse parses DSN and panics on error. Useful for tests and initialization.
func MustParse(raw string) *DSN {
	dsn, err := Parse(raw)
	if err != nil {
		panic(err)
	}

	return dsn
}

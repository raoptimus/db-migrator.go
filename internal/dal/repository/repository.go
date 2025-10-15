/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package repository

// New creates repository by connection
func New(conn Connection, options *Options) (*Repository, error) {
	r, err := create(conn, options)
	if err != nil {
		return nil, err
	}
	return &Repository{
		adapter: r,
	}, nil
}

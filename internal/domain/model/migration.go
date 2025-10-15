/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package model

type Migration struct {
	Version     string
	ApplyTime   int
	BodySQL     string
	ExecutedSQL string
	Release     string
}
type Migrations []Migration

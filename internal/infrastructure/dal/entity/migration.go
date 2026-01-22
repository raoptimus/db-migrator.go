/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package entity

import (
	"sort"
	"time"
)

// Migration represents a database migration record stored in the migration history table.
// It contains the version identifier and the timestamp when the migration was applied.
type Migration struct {
	Version   string `db:"version"`
	ApplyTime int64  `db:"apply_time"`
	// BodySQL     string `db:"body_sql"`
	// ExecutedSQL string `db:"executed_sql"`
	// Release     string `db:"release"`
}

// Migrations is a collection of Migration records that implements sort.Interface.
type Migrations []Migration

// Len returns the number of migrations in the collection.
// This method is required by sort.Interface.
func (s Migrations) Len() int {
	return len(s)
}

// Less reports whether the migration at index i should sort before the migration at index j.
// Migrations are sorted by their version string in ascending order.
// This method is required by sort.Interface.
func (s Migrations) Less(i, j int) bool {
	return s[i].Version < s[j].Version
}

// Swap swaps the migrations at indexes i and j.
// This method is required by sort.Interface.
func (s Migrations) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// SortByVersion sorts the migrations by their version string in ascending order.
func (s Migrations) SortByVersion() {
	sort.Sort(s)
}

// ApplyTimeFormat returns the formatted apply time as a string in "YYYY-MM-DD HH:MM:SS" format.
func (s Migration) ApplyTimeFormat() string {
	return time.Unix(s.ApplyTime, 0).Format("2006-01-02 15:04:05")
}

/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package model

import (
	"sort"
	"time"
)

// Migration represents a domain migration record.
type Migration struct {
	Version     string
	ApplyTime   int64
	BodySQL     string
	ExecutedSQL string
	Release     string
}

// ApplyTimeFormat returns the formatted apply time as a string in "YYYY-MM-DD HH:MM:SS" format.
func (m Migration) ApplyTimeFormat() string {
	return time.Unix(m.ApplyTime, 0).Format("2006-01-02 15:04:05")
}

// Migrations is a collection of Migration records that implements sort.Interface.
type Migrations []Migration

// Len returns the number of migrations in the collection.
func (s Migrations) Len() int {
	return len(s)
}

// Less reports whether the migration at index i should sort before the migration at index j.
func (s Migrations) Less(i, j int) bool {
	return s[i].Version < s[j].Version
}

// Swap swaps the migrations at indexes i and j.
func (s Migrations) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// SortByVersion sorts the migrations by their version string in ascending order.
func (s Migrations) SortByVersion() {
	sort.Sort(s)
}

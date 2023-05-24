package entity

import (
	"sort"
	"time"
)

type Migration struct {
	Version   string `db:"version"`
	ApplyTime int    `db:"apply_time"`
	// BodySQL     string `db:"body_sql"`
	// ExecutedSQL string `db:"executed_sql"`
	// Release     string `db:"release"`
}
type Migrations []Migration

func (s Migrations) Len() int {
	return len(s)
}

func (s Migrations) Less(i, j int) bool {
	return s[i].Version < s[j].Version
}

func (s Migrations) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s Migrations) SortByVersion() {
	sort.Sort(s)
}

func (s Migration) ApplyTimeFormat() string {
	return time.Unix(int64(s.ApplyTime), 0).Format("2006-01-02 15:04:05")
}

/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev <resmus@gmail.com>
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */
package db

import (
	"sort"
	"time"
)

type (
	HistoryItem struct {
		Version   string
		ApplyTime int
		Safely    bool
	}
	HistoryItems []HistoryItem
)

func (s HistoryItems) Len() int {
	return len(s)
}

func (s HistoryItems) Less(i, j int) bool {
	return s[i].Version < s[j].Version
}

func (s HistoryItems) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s HistoryItems) SortByVersion() {
	sort.Sort(s)
}

func (s HistoryItem) GetUpFileName() string {
	if s.Safely {
		return s.Version + ".safe.up.sql"
	}
	return s.Version + ".up.sql"
}

func (s HistoryItem) GetDownFileName() string {
	if s.Safely {
		return s.Version + ".safe.down.sql"
	}
	return s.Version + ".down.sql"
}

func (s HistoryItem) ApplyTimeFormat() string {
	return time.Unix(int64(s.ApplyTime), 0).Format("2006-01-02 15:04:05")
}

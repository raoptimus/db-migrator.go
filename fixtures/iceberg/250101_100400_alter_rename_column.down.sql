-- Rename event_kind back to event_type (reverse of 250101_100400)
ALTER TABLE iceberg.analytics.events RENAME COLUMN event_kind TO event_type;

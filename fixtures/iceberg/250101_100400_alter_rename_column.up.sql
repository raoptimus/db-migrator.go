-- Rename event_type to event_kind for clarity
ALTER TABLE iceberg.analytics.events RENAME COLUMN event_type TO event_kind;

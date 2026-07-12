-- Add a new column to track session identifiers
ALTER TABLE iceberg.analytics.events ADD COLUMN session_id STRING;

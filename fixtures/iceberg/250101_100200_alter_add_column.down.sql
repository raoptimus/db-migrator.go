-- Remove the session_id column added in 250101_100200
ALTER TABLE iceberg.analytics.events DROP COLUMN session_id;

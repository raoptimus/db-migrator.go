-- Re-create the table that was dropped in 250101_100700.
-- This is the full definition after all preceding migrations were applied:
-- rename event_type->event_kind, drop source, add retry_count (long), add session_id,
-- add partition bucket(16, user_id).
CREATE TABLE iceberg.analytics.events (
  event_id   STRING    COMMENT 'unique event identifier',
  user_id    STRING    COMMENT 'external user identifier',
  event_kind STRING    COMMENT 'event category',
  event_time TIMESTAMP COMMENT 'canonical event time (UTC)',
  created_at TIMESTAMP COMMENT 'record ingestion time',
  session_id STRING,
  retry_count long
)
USING iceberg
PARTITIONED BY (days(event_time), bucket(16, user_id))
COMMENT 'analytics events (bronze layer)'
TBLPROPERTIES (
  'format-version' = '2',
  'write.format.default' = 'parquet',
  'write.parquet.compression-codec' = 'zstd'
);

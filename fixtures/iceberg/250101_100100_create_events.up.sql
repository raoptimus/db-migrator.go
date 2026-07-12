-- Reference CREATE TABLE (analytics namespace, bronze layer).
-- DSN warehouse = iceberg, so the leading "iceberg" segment is stripped by the driver:
-- iceberg.analytics.events => namespace=[analytics], table=events
CREATE TABLE iceberg.analytics.events (
  event_id   STRING    COMMENT 'unique event identifier',
  user_id    STRING    COMMENT 'external user identifier',
  event_type STRING    COMMENT 'event category',
  event_time TIMESTAMP COMMENT 'canonical event time (UTC)',
  created_at TIMESTAMP COMMENT 'record ingestion time',
  source     STRING    COMMENT 'ingestion source label'
)
USING iceberg
PARTITIONED BY (days(event_time))
COMMENT 'analytics events (bronze layer)'
TBLPROPERTIES (
  'format-version' = '2',
  'write.format.default' = 'parquet',
  'write.parquet.compression-codec' = 'zstd'
);

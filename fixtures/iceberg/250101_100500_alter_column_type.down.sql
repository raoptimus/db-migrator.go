-- Reverse of "ADD COLUMN retry_count + widen to long": drop the column.
ALTER TABLE iceberg.analytics.events DROP COLUMN retry_count;

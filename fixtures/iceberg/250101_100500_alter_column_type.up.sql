-- Add an integer column and widen it to long to demonstrate ALTER COLUMN TYPE with int->long.
-- Iceberg allows widening promotions (e.g. int -> long).
ALTER TABLE iceberg.analytics.events ADD COLUMN retry_count INT;
ALTER TABLE iceberg.analytics.events ALTER COLUMN retry_count TYPE long;

-- Second migration fails: TRUNCATE is not supported DDL in Iceberg.
-- Intentionally broken to verify that a partial release leaves successfully applied migrations recorded (best-effort).
TRUNCATE TABLE raw_release_fail.nonexistent;

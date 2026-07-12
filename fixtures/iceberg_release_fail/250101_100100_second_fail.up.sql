-- Second migration fails: TRUNCATE is not supported DDL in Iceberg.
-- This is intentionally broken to test best-effort release failure (ФТ-7 @negative).
TRUNCATE TABLE raw_release_fail.nonexistent;

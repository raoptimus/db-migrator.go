-- FT-13 @negative: widen id column int->long (valid Iceberg promotion).
-- The paired down.sql attempts the reverse (long->int narrowing) which the catalog rejects.
ALTER TABLE iceberg.raw_ft13.irreversible_demo ALTER COLUMN id TYPE long;

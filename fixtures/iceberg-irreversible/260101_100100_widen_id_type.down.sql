-- FT-13 @negative: narrowing long->int is NOT supported by Iceberg.
-- Executing this down migration is expected to FAIL with a catalog error about
-- incompatible type change. The migration record must remain in history (applied).
ALTER TABLE iceberg.raw_ft13.irreversible_demo ALTER COLUMN id TYPE int;

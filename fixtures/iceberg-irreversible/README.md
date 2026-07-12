Negative test fixtures for BDD ФТ-13 (@negative): demonstrates an irreversible `down` migration.
`down` of `260101_100100_widen_id_type` attempts a type narrowing (long→int) which Iceberg rejects;
the migration must remain marked as applied in history. Used by task 06 integration tests.

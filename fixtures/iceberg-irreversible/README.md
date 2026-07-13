Negative test fixtures demonstrating an irreversible `down` migration.
`down` of `260101_100100_widen_id_type` attempts a type narrowing (long→int) which the Iceberg catalog rejects;
the migration record must remain marked as applied in history.

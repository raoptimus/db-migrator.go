-- Create a table in the release namespace for release/rollback integration tests.
CREATE TABLE raw_release.events (
  id   long,
  name string
);

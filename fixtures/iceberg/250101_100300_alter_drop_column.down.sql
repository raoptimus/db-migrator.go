-- Re-add the source column dropped in 250101_100300
ALTER TABLE iceberg.analytics.events ADD COLUMN source STRING;

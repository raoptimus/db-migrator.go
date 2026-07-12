-- Remove the bucket partition field added in 250101_100600
ALTER TABLE iceberg.analytics.events DROP PARTITION FIELD bucket(16, user_id);

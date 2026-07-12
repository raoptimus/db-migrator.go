-- Add an additional partition field to bucket by user_id for more granular access patterns
ALTER TABLE iceberg.analytics.events ADD PARTITION FIELD bucket(16, user_id);

CREATE DATABASE IF NOT EXISTS raw
ENGINE = Replicated('/clickhouse/databases/{shard}/raw', '{shard}', '{replica}');

CREATE TABLE IF NOT EXISTS raw.test (
    time DateTime DEFAULT now(),
    value UInt32
)
ENGINE = ReplicatedMergeTree (
    '/clickhouse/tables/{shard}/raw_test_cluster_test',
    '{replica}'
)
PARTITION BY toYYYYMM(time)
ORDER BY (time, value);

CREATE DATABASE IF NOT EXISTS raw ON CLUSTER test_cluster;

CREATE TABLE raw.test ON CLUSTER test_cluster (
    time DateTime DEFAULT now(),
    value UInt32
)
ENGINE = ReplicatedMergeTree (
    '/clickhouse/tables/{shard}/raw_test_cluster_test',
    '{replica}'
)
PARTITION BY toYYYYMM(time)
ORDER BY (time, value);

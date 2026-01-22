CREATE DATABASE IF NOT EXISTS raw ON CLUSTER test_cluster;

CREATE TABLE raw.test ON CLUSTER test_cluster (
    time DateTime DEFAULT now(),
    value UInt32
)
ENGINE = ReplicatedMergeTree
PARTITION BY toYYYYMM(time)
ORDER BY (time, value);

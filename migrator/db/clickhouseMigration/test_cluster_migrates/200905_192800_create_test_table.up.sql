CREATE TABLE test ON CLUSTER test_cluster (
    time DateTime DEFAULT now(),
    value UInt32

)
ENGINE = ReplicatedMergeTree (
    '/clickhouse/tables/{shard}/test_cluster_test',
    '{replica}'
)
PARTITION BY toYYYYMM(time)
ORDER BY (time, value);

ALTER TABLE test ADD COLUMN value2 UInt8;


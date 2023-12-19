CREATE TABLE test (
    time DateTime DEFAULT now(),
    value UInt32

) ENGINE = MergeTree
PARTITION BY toYYYYMM(time)
ORDER BY (time, value);

ALTER TABLE test ADD COLUMN value2 UInt8;


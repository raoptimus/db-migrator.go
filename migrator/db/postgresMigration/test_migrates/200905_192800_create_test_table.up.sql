CREATE TABLE test (
    id SERIAL,
    value int
);

ALTER TABLE test
    ADD COLUMN created_at timestamp NOT NULL DEFAULT now();

CREATE INDEX test_created_at_idx ON test USING btree (created_at DESC)
;

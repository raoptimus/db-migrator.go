ALTER TABLE test ON CLUSTER test_cluster
	ADD COLUMN text String;

INSERT INTO test (value, text) VALUES (1, 'Hello');

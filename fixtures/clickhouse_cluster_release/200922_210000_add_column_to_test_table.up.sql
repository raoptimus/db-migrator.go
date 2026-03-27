ALTER TABLE raw.test ON CLUSTER test_cluster
	ADD COLUMN text String;

INSERT INTO raw.test (value, text) VALUES (1, 'Hello');

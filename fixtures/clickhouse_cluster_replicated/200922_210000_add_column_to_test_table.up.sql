ALTER TABLE raw.test ADD COLUMN text String;

INSERT INTO raw.test (value, text) VALUES (1, 'Hello');

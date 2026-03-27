CREATE TABLE test2 (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

insert into test (id, top, title, updated_at)
values (1, 0, 'some title 1', CURRENT_TIMESTAMP);

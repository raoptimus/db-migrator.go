
CREATE TABLE test (
    id INT PRIMARY KEY,
    rank SMALLINT,
    title text,
    synonyms text[],
    updated_at TIMESTAMP
);

create index test_rank_idx ON test(rank);

CREATE TABLE test2 (
    id int PRIMARY KEY REFERENCES test(id) ON DELETE CASCADE ON UPDATE CASCADE,
    document tsvector,
    updated_at timestamp
);
